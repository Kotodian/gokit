package beanstalk

import (
	"context"
	"fmt"
	"strconv"
	"time"

	//"fmt"

	"strings"

	// beanstalkConn "github.com/Kotodian/gokit/datasource/beanstalk"
	rdlib "github.com/garyburd/redigo/redis"
	"github.com/sirupsen/logrus"

	"github.com/Kotodian/gokit/datasource/redis"
	libBeanstalk "github.com/beanstalkd/go-beanstalk"
)

// type CallbackFunc func(*libBeanstalk.Conn, uint64, []byte, error) CallbackAction

func GetTryTimes(key string) (uint32, error) {
	rd := redis.GetRedis()
	defer rd.Close()
	var err error
	var retryTimes int
	if retryTimes, err = rdlib.Int(rd.Do("get", fmt.Sprintf("trytimes:%s:task", key))); err != nil {
		if err != rdlib.ErrNil {
			return 0, err
		}
		return 0, nil
	}
	return uint32(retryTimes), nil
}

//
//func SetTryTimes(key string, times uint32) error {
//	rd := redis.GetRedis()
//	defer rd.Close()
//	var err error
//	_, err = rd.Do("set", fmt.Sprintf("trytimes:%s:task", key), times)
//	return err
//}

//----------------------------------------------------------------------------
//type ITask interface {
//	GetBID() uint64
//	SetBID(uint64)
//	GetTTL() TTL
//	GetWeight() uint32
//	GetDelay() time.Duration
//	GetTTR() time.Duration
//	GetName() string
//	GetRetryInteval() uint32
//	GetMaxRetryTimes() uint32
//	GetTryTimes() uint32
//	SetTryTimes(uint32)
//	Save() error
//	Delete() error
//	GetID() string
//	SetID(string)
//}

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

//type CallbackRet struct {
//	Action CallbackAction
//	Err    error
//}

type CallbackFuncWithControl func(ctx context.Context, task *Task) (action CallbackAction, err error)

//
//func NewCallbackRetWithITask(ctx context.Context, task ITask, defaultAction CallbackAction, fn func(ctx context.Context, ret *CallbackRet) error) (_task ITask, act chan CallbackRet) {
//	act = make(chan CallbackRet, 1)
//	go func(ctx context.Context) {
//		ret := CallbackRet{
//			Action: defaultAction,
//		}
//		fmt.Println("进入 grouting")
//		var err error
//		defer func() {
//			fmt.Println("退出 grouting")
//			close(act)
//			if err != nil {
//				ret.Err = err
//			}
//		}()
//		err = fn(ctx, &ret)
//		if err != nil {
//			ret.Err = err
//		}
//		act <- ret
//	}(ctx)
//	return task, act
//}

type Task struct {
	TTL           TTL    `redis:"ttl" json:"ttl"`                                       //过期时间，单位秒，0为永不过时，慎用0
	TTR           uint32 `redis:"ttr" json:"ttr"`                                       //最大执行时间，单位秒
	RetryInterval string `redis:"retry_interval_v2" json:"retry_interval_v2,omitempty"` //重试间隔,单位秒,以逗号作为间隔的时间序列
	MaxRetryTimes uint32 `redis:"max_retry_times" json:"max_retry_times"`               //最大重试次数
	TryTimes      uint32 `redis:"try_times" json:"try_times"`                           //当前重试次数
	Weight        uint32 `redis:"weight" json:"weight"`                                 //权重
	ExecTime      uint32 `redis:"exec_time" json:"exec_time"`                           //延迟多少秒执行
	Bid           uint64 `redis:"bid" json:"bid"`                                       //beanstalk.ID  //beanstalk id
	ID            uint64 `redis:"id" json:"id"`                                         //唯一ID
	Payload       []byte `redis:"payload" json:"payload"`                               //数据
	Name          string `redis:"name" json:"name"`                                     //beanstalk的tube set名称
	redis.Base
}

func (t Task) InstanceName() string {
	return fmt.Sprintf("%d:%s:task:beanstalk", t.GetID(), t.GetName())
}

func (t Task) GetTTL() TTL {
	return t.TTL
}

func (t Task) GetWeight() uint32 {
	return t.Weight
}

func (t Task) GetDelay() time.Duration {
	_t := t.ExecTime - uint32(time.Now().Unix())
	if _t <= 0 {
		return 0
	}
	return time.Duration(_t) * time.Second
}

func (t Task) GetTTR() time.Duration {
	return time.Duration(t.TTR) * time.Second
}

func (t Task) GetName() string {
	return t.Name
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

func (t Task) GetMaxRetryTimes() uint32 {
	return t.MaxRetryTimes
}

func (t Task) GetTryTimes() uint32 {
	return t.TryTimes
}

func (t *Task) SetTryTimes(_t uint32) {
	t.TryTimes = _t
}

func (t Task) GetBID() uint64 {
	return t.Bid
}

func (t *Task) SetBID(id uint64) {
	t.Bid = id
}

func (t *Task) GetID() uint64 {
	return t.ID
}
func (t *Task) SetID(id uint64) {
	t.ID = id
}

func (t *Task) SetName(name string) {
	t.Name = name
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
			//rd.Do("set", fmt.Sprintf("%s:%s:task", t.GetID(), t.GetName()), t.TryTimes, "ex", uint32(t.TTL)-uint32(time.Now().Local().Unix()))
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

func Publish(task *Task) error {
	conn, err := Conn()
	if err != nil {
		return err
	}
	defer Close(conn)
	return PublishWithConn(conn, task)
}

func PublishWithConn(conn *libBeanstalk.Conn, task *Task) error {
	//fmt.Println("\n\n", task.GetID(), "\n\n")
	tube := libBeanstalk.Tube{
		Conn: conn,
		Name: task.GetName(),
	}
	if err := task.Save(); err != nil {
		return err
	}
	b := []byte(fmt.Sprintf("%d", task.GetID()))
	id, err := tube.Put(b, task.GetWeight(), task.GetDelay(), task.GetTTR())
	if err != nil {
		return err
	}
	task.SetBID(id)
	return nil
}

//PeekAndPut 先判断记录是否存在，如果不存在就put进去，存在就不做任何操作
func PeekAndPublish(task *Task) error {
	//var (
	//	conn *libBeanstalk.Conn
	//	err  error
	//)
	conn, err := Conn()
	if err != nil {
		return err
	}
	defer Close(conn)

	bid := task.GetBID()
	if bid > 0 {
		_, err := conn.Peek(bid)
		if err != nil {
			if err.(libBeanstalk.ConnError).Err != libBeanstalk.ErrNotFound {
				return err
			}
		}
		//return err
	}
	return PublishWithConn(conn, task)
}

func KickAndPublish(task *Task) error {
	//var (
	//	conn *libBeanstalk.Conn
	//	err  error
	//)
	conn, err := Conn()
	if err != nil {
		return err
	}
	defer Close(conn)

	bid := task.GetBID()
	if bid > 0 {
		if err := conn.KickJob(bid); err == nil {
			return nil
		} else if err.(libBeanstalk.ConnError).Err == libBeanstalk.ErrNotFound {
			goto PUBLISH
		}
		return err
	}
PUBLISH:
	return PublishWithConn(conn, task)
}

func PeekAndPutWithConn(conn *libBeanstalk.Conn, task *Task) error {
	bid := task.GetBID()
	if bid > 0 {
		_, err := conn.Peek(bid)
		if err == nil || err.(libBeanstalk.ConnError).Err == libBeanstalk.ErrNotFound {
			return nil
		}
		return err
	}
	return PublishWithConn(conn, task)
}

func ListenWithCallback(ctx context.Context, name string, t time.Duration, callback CallbackFuncWithControl) error {
	log := logrus.WithFields(logrus.Fields{})
	ctx = context.WithValue(ctx, "log", log)
	//var (
	//	conn *libBeanstalk.Conn
	//	err  error
	//)
	conn, err := Conn()
	if err != nil {
		return err
	}
	defer Close(conn)

	_ts := libBeanstalk.NewTubeSet(conn, name)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		func() {
			bid, body, err := _ts.Reserve(t)
			if err != nil {
				errString := err.Error()
				if strings.Contains(errString, "timeout") {
					return
				} else if strings.Contains(errString, "deadline") {
					return
				}
				log.Error(err)
				return
			}
			log.Data["body"] = string(body)
			log.Data["bid"] = bid
			var id uint64
			id, err = strconv.ParseUint(string(body), 10, 64)
			if err != nil {
				log.Errorf("parse task id error, err:%s", err.Error())
				if err = conn.Delete(bid); err != nil {
					log.Error(err)
				}
				return
			} else if id == 0 {
				log.Errorf("task id is zero")
				if err = conn.Delete(bid); err != nil {
					log.Error(err)
				}
				return
			}
			log.Data["id"] = id
			//log.Data["bid"] = bid

			var task Task
			instanceName := fmt.Sprintf("%d:%s:task:beanstalk", id, name)

			{
				var exists bool
				var err error
				if exists, err = redis.Get(&task, instanceName); err != nil {
					log.Error(err)
					return
				}
				//utils.D(task, "task")
				if !exists {
					task.SetBID(bid)
					task.SetID(id)
					task.SetName(name)
					//task.SetKey(instanceName)
				} else if task.GetTTL().IsExpired() { //如果任务已经过期
					log.Warnf("task expired")
					if err = conn.Delete(id); err != nil {
						log.Error(err)
					}
					_ = task.Delete()
					return
				}
			}

			//fmt.Println(task)
			//设置已经尝试的次数，从redis中去加载
			//var times uint32
			//times, err = GetTryTimes(fmt.Sprintf("%s:%s", task.GetID(), task.GetName()))
			//if err != nil {
			//	_ = conn.Release(id, task.GetWeight(), time.Duration(task.RetryInterval)*time.Second)
			//	return
			//}
			//task.SetTryTimes(times)
			//fmt.Println("\n\n eeeeeeeeeeeeeeee ", fmt.Sprintf("%+v", task), "\n\n")
			action := CallbackActionRelease
			if string(task.Payload) == "" {
				action = CallbackActionDelete
			} else {
				if action, err = callback(ctx, &task); err != nil {
					//_ = conn.Delete(id)
					log.Error(err)
					//goto Release
					if action == CallbackActionRelease {
						goto Release
					}
				} else if action == CallbackActionRelease {
					goto Release
				}
			}

			log.Infof("action:%d", action)
			if action == CallbackActionDelete || action == CallbackActionBury {
				if _err := conn.Delete(bid); _err != nil {
					log.Error(_err)
					return
				}
				_ = task.Delete()
				return
			}

			if action == CallbackActionNothing {
				return
			}
			return
		Release:
			//log.Info("release")
			//if task.GetMaxRetryTimes() == 0 {
			//	return
			if task.GetMaxRetryTimes() > 0 && task.GetTryTimes() >= task.GetMaxRetryTimes() { //若超过最大重试次数
				//action = CallbackActionDelete
				if _err := conn.Delete(bid); _err != nil {
					log.Error(_err)
					return
				}
				_ = task.Delete()
				return
			}

			interval := 300
			if intervals := task.GetRetryInteval(); len(intervals) > int(task.GetTryTimes()) {
				interval = intervals[task.GetTryTimes()]
			}

			//err = ret.Err
			_ = conn.Release(bid, task.GetWeight(), time.Duration(interval)*time.Second)
			//utils.D(task, "xxxxx")
			log.Infof("delay:%d, trytimes:%d", interval, task.GetTryTimes())
			task.SetTryTimes(task.GetTryTimes() + 1)
			if _err := task.Save(); _err != nil {
				log.Error(_err)
				return
			}
			//fmt.Println("\n\n ddddddddddddddd retry ", task.GetID(), task.GetRetryInteval(), "\n\n")
			//SetTryTimes(fmt.Sprintf("%s:%s", task.GetID(), task.GetName()), task.GetTryTimes())
		}()
	}
}
