package rabbitmq

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Kotodian/gokit/datasource/redis"
	rdlib "github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

var scheduleExchange Exchange

func init() {
	scheduleExchange = Exchange{
		Name:    "schedulejob-delayed-exchange",
		Type:    ExchangeTypeXDelayedMessage,
		Durable: true,
		Delayed: true,
	}
}

type TTL uint32

func (t TTL) IsExpired() bool {
	if t == 0 {
		return false
	}
	return uint32(time.Now().Local().Unix()) >= uint32(t)
}

type CallbackAction int

const (
	CallbackActionBury    CallbackAction = -1
	CallbackActionRelease CallbackAction = 1
	CallbackActionDelete  CallbackAction = 0
	CallbackActionNothing CallbackAction = 2
)

type CallbackFuncWithControl func(ctx context.Context, task *Task) (action CallbackAction, err error)

type Task struct {
	Weight        uint8  `redis:"weight" json:"weight"`
	Name          string `redis:"name" json:"name"`                              //task name
	TTL           TTL    `redis:"ttl" json:"ttl"`                                //过期时间，单位秒，0为永不过时，慎用0
	TTR           uint32 `redis:"ttr" json:"ttr"`                                //最大执行时间，单位秒
	RetryInterval string `redis:"retry_interval" json:"retryInterval,omitempty"` //重试间隔,单位秒,以逗号作为间隔的时间序列
	MaxRetryTimes uint32 `redis:"max_retry_times" json:"maxRetryTimes"`          //最大重试次数
	RetryTimes    uint32 `redis:"retry_times" json:"retryTimes"`                 //当前重试次数
	ExecTime      uint32 `redis:"exec_time" json:"execTime"`                     //延迟多少秒执行
	Payload       []byte `redis:"payload" json:"payload"`                        //数据
	ID            uint64 `redis:"id" json:"id"`                                  //唯一ID
	redis.Base
}

func (t Task) GetWeight() uint8 {
	return t.Weight
}
func (t Task) GetTTL() TTL {
	return t.TTL
}
func (t Task) GetID() uint64 {
	return t.ID
}
func (t *Task) SetID(id uint64) {
	t.ID = id
}

func (t *Task) GetName() string {
	return t.Name
}
func (t *Task) SetName(name string) {
	t.Name = name
}

func (t Task) InstanceName() string {
	return fmt.Sprintf("%d:%s:rabbitmq", t.GetID(), t.GetName())
}

func (t Task) GetRetryInteval() (ret []int) {
	for _, v := range strings.Split(t.RetryInterval, ",") {
		i, _ := strconv.Atoi(v)
		if i == 0 {
			i = 300
		}
		ret = append(ret, i)
	}
	return
}

func (t Task) GetDelay() time.Duration {
	_t := t.ExecTime - uint32(time.Now().Unix())
	if _t <= 0 {
		return 0
	}
	return time.Duration(_t) * time.Second
}
func (t Task) GetRetryTimes() uint32 {
	return t.RetryTimes
}
func (t *Task) SetRetryTimes(_t uint32) {
	t.RetryTimes = _t
}
func (t Task) GetMaxRetryTimes() uint32 {
	return t.MaxRetryTimes
}
func (t *Task) Save() error {
	rd := redis.GetRedis()
	defer rd.Close()
	name := t.InstanceName()
	if t.TTL > 0 {
		if !t.TTL.IsExpired() {
			if err := rd.Send("MULTI"); err != nil {
				return err
			}
			if err := rd.Send("HMSET", rdlib.Args{}.Add(name).AddFlat(t)...); err != nil {
				return err
			} else if err := rd.Send("EXPIRE", name, t.TTL); err != nil {
				return err
			}
			if _, err := rd.Do("EXEC"); err != nil {
				return err
			}
		} else {
			if _, err := rd.Do("DEL", name); err != nil {
				return err
			}
			//t.Delete()
		}
	} else {
		if _, err := rd.Do("HMSET", rdlib.Args{}.Add(name).AddFlat(t)...); err != nil {
			return err
		}
	}
	return nil
}
func (t *Task) Delete() error {
	rd := redis.GetRedis()
	defer rd.Close()
	_, err := rd.Do("del", t.InstanceName())
	return err
}

func PublishScheduleJob(ctx context.Context, t *Task, delayTime ...int64) (err error) {
	if err := t.Save(); err != nil {
		return err
	}
	b := []byte(fmt.Sprintf("%d", t.GetID()))
	return scheduleExchange.Publish(ctx, "schedulejob.delayed", amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "text/plain",
		Priority:     t.Weight,
		Body:         b,
	}, delayTime...)

}

func ConsumeWithCallback(ctx context.Context, name string, t time.Duration, callback CallbackFuncWithControl) {
	logEntry := logrus.WithFields(logrus.Fields{"method": "ConsumeWithCallback"})
	scheduleExchange.Start(ctx, 10, Receiver{
		Name:      "schedulejob_delayed_queue",
		RouterKey: "schedulejob.delayed",
		AutoAck:   false,
		OnErrorFn: func(r Receiver, err error) {
			logEntry.Dup().Errorf("consume scheduleJob delayed exchange error:%s", err.Error())
		},
		OnReceiveFn: func(r Receiver, msg amqp.Delivery) (ret bool) {
			var err error
			_log := logEntry.Dup()
			ctx = context.WithValue(ctx, "log", _log)
			defer func() {
				if err != nil {
					_log.Error(err)
				}
				ret = true
			}()
			id, err := strconv.ParseUint(string(msg.Body), 10, 64)
			if err != nil {
				_log.Errorf("parse task id error, err:%s", err.Error())
				return
			} else if id == 0 {
				_log.Errorf("task id is zero")
				return
			}
			_log.Data["id"] = id
			var task Task
			instanceName := fmt.Sprintf("%d:%s:rabbitmq", id, name)
			if exists, _err := redis.Get(&task, instanceName); _err != nil {
				_log.Error(_err)
				return
			} else if exists && task.GetTTL().IsExpired() { //如果任务已经过期
				_log.Warnf("task expired")
				_ = task.Delete()
				return
			}
			action := CallbackActionRelease
			if string(task.Payload) == "" {
				action = CallbackActionDelete
			} else {
				if action, err = callback(ctx, &task); err != nil {
					_log.Error(err)
				}
			}
			_log.Infof("action:%d", action)
			if action == CallbackActionDelete || action == CallbackActionBury {
				_ = task.Delete()
				return
			}
			if action == CallbackActionNothing {
				return
			}
			if action == CallbackActionRelease {
				if task.GetMaxRetryTimes() > 0 && task.GetRetryTimes() >= task.GetMaxRetryTimes() { //若超过最大重试次数
					_ = task.Delete()
					return
				}
				interval := 300
				if intervals := task.GetRetryInteval(); len(intervals) > int(task.GetRetryTimes()) {
					interval = intervals[task.GetRetryTimes()]
				}
				//将需重试job放入rabbitmq 延时队列
				PublishScheduleJob(ctx, &task, int64(interval))
				_log.Infof("delay:%d, retrytimes:%d", interval, task.GetRetryTimes())
				task.SetRetryTimes(task.GetRetryTimes() + 1)
				if _err := task.Save(); _err != nil {
					_log.Error(_err)
					return
				}
			}
			return
		},
	})
}
