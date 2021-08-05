package redis

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	redisLib "github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"
)

var (
	logEntry = log.WithFields(log.Fields{
		"module": "stream",
	})
)

// StreamData 的单条数据
type StreamData struct {
	ID   string
	Vals interface{}
	Err  error
}

// Stream 整个stream的数据
type Stream struct {
	Name string
	Data []StreamData
}

type Pedding struct {
	ID             string
	ConsumerName   string
	IdleTime       int64
	DeliveredTimes int32
}

func Encode(data interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Decode(data []byte, to interface{}) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	return dec.Decode(to)
}

// ScanStream 解析整个stream
func ScanStream(val []interface{}, s *Stream, f interface{}) {
	streamName := string(val[0].([]interface{})[0].([]byte))
	if s == nil {
		s = &Stream{}
	}
	s.Name = streamName
	data1, ok := val[0].([]interface{})
	if !ok {
		return
	}
	data, ok := data1[1].([]interface{})
	if !ok {
		return
	}
	s.Data = make([]StreamData, len(data))
	_f := reflect.ValueOf(f)
	kind := _f.Kind()
	for idx, v := range data {
		// s.Data[idx].ID = string(v.([]interface{})[0].([]uint8))
		// fmt.Println("---------------->", v.([]interface{})[0])
		if kind == reflect.Ptr {
			if v != nil {
				s.Data[idx].Vals = reflect.New(_f.Elem().Type()).Interface()

				d := json.NewDecoder(bytes.NewReader(v.([]interface{})[1].([]interface{})[1].([]byte)))
				d.UseNumber()
				if err := d.Decode(s.Data[idx].Vals); err != nil {
					s.Data[idx].Err = err
					continue
				}

				//if err := json.Unmarshal(v.([]interface{})[1].([]interface{})[1].([]byte), s.Data[idx].Vals); err != nil {
				//	s.Data[idx].Err = err
				//	continue
				//}
			} else {
				continue
			}
		} else {
			_v := v.([]interface{})[1].([]interface{})[1]
			switch kind {
			case reflect.String:
				_str, _ := redisLib.String(_v, nil)
				s.Data[idx].Vals = strings.Trim(_str, `"`)
			case reflect.Uint64:
				s.Data[idx].Vals, _ = redisLib.Uint64(_v, nil)
			case reflect.Int64:
				s.Data[idx].Vals, _ = redisLib.Int64(_v, nil)
			default:
				s.Data[idx].Err = fmt.Errorf("not support, kind:%d", kind)
			}
		}
		//fmt.Println("qqqqqqqqqq", )
		s.Data[idx].ID = string(v.([]interface{})[0].([]byte))
		//
		//
		//if kind == reflect.String {
		//	str := string(v.([]interface{})[1].([]interface{})[1].([]byte))
		//	if string(str[0]) == "\"" {
		//		s.Data[idx].Vals = str[1 : len(str)-1]
		//	} else {
		//		s.Data[idx].Vals = str
		//	}
		//	//if _, err := redisLib.Scan(v.([]interface{})[1].([]interface{}), s.Data[idx].Vals); err != nil {
		//	//	return err
		//	//}
		//} else {
		//	if kind == reflect.Ptr {
		//		s.Data[idx].Vals = reflect.New(f.Elem().Type()).Interface()
		//		if err := json.Unmarshal(v.([]interface{})[1].([]interface{})[1].([]byte), s.Data[idx].Vals); err != nil {
		//			return err
		//		}
		//	} else {
		//		redisLib.Uint64(v.([]interface{})[1].([]interface{})[1], nil)
		//		//s.Data[idx].Vals = reflect.New(f.Type()).Interface()
		//		//s.Data[idx].Vals = reflect.New(f.Type()).Interface()
		//		//fmt.Println()
		//		//
		//		//fmt.Println("ffffffffffff",s.Data[idx].Vals)
		//		//if err := json.Unmarshal(v.([]interface{})[1].([]interface{})[1].([]byte), &s.Data[idx].Vals); err != nil {
		//		//	return err
		//		//}
		//	}
		//	fmt.Println("dddddddddd", kind, s.Data[idx].Vals)
		//	// if err := redisLib.ScanStruct(, s.Data[idx].Vals); err != nil {
		//	// 	return err
		//	// }
		//}
	}
	return
}

// StreamHandle redis stream
type StreamHandle struct {
	Name     string
	Group    string
	Consumer string
	PreLimit int
	callback func(context.Context, string, interface{}) error
	//retry    bool
}

func NewStreamHandle(name, group, consumer string, limit int, cb func(context.Context, string, interface{}) error) (*StreamHandle, error) {
	s := &StreamHandle{
		Name:     name,
		Group:    group,
		Consumer: consumer,
		PreLimit: limit,
		callback: cb,
		//msgCh:    make(chan StreamData, 1000),
		//lastOffset: "$",
		//newMsg:   make(chan Stream, 1000),
	}
	return s, nil
}

// loadMsg 接受消息，包括旧的消息以及新的消息，done是结束通道通知，t是数据的结构体定义
func (s *StreamHandle) handleUnAckMsg(ctx context.Context, t interface{}) {
	//wg := &sync.WaitGroup{}
	//wg.Add(2)
	//oldDone := make(chan struct{}, 1)
	//go func() { //消费之前的消息
	logEntry := log.WithFields(log.Fields{
		"module": "stream-unackmsg",
	})
	for {
		if ret := func() (ret int) {
			ret = 0 //0 continue, 1 break
			rd := GetRedis()
			defer rd.Close()
			select {
			case <-ctx.Done():
				ret = 1
				return
			default:
			}
			newConsumerID := "_xclaim"

			//先删除 cleam 的队列
			//if _, err := rd.Do("XGROUP", "DELCONSUMER", s.Name, s.Group, newConsumerID); err != nil {
			//	logEntry.Error(err)
			//	time.Sleep(3 * time.Second)
			//}

			//把旧消息传递给其他消费组处理
			resp, err := redisLib.Values(rd.Do("XPENDING", s.Name, s.Group, "-", "+", 100))
			if err != nil {
				if err != redisLib.ErrNil {
					if strings.Contains(err.Error(), "No such key") {
						return
					} else if strings.Contains(err.Error(), "timeout") {
						return
					} else {
						logEntry.Error("xpending", err)
						time.Sleep(3 * time.Second)
					}
				}
			} else if len(resp) > 0 {
				ids := make([]interface{}, 0)
				for _, data := range resp {
					m := data.([]interface{})
					ids = append(ids, m[0])
				}

				//newConsumerID := fmt.Sprintf("%s-%d-xclaim", s.Consumer, time.Now().Unix())
				args := redisLib.Args{}.Add(s.Name).Add(s.Group).Add(newConsumerID).Add(60000).Add(ids...)
				if _, err = rd.Do("XCLAIM", args...); err != nil {
					logEntry.Error(err)
				}
				s.handle(ctx, newConsumerID, t, ">")
				time.Sleep(1 * time.Second)
			}
			return
		}(); ret == 1 {
			break
		}
		time.Sleep(30 * time.Second)
	}
}

func (s *StreamHandle) handle(ctx context.Context, consumerID string, t interface{}, offset string) {
	timeout := 1000
	for {
		if ret := func() (ret int) {
			ret = 0 //0 continue, 1 break
			rd := GetRedis()
			defer func() {
				rd.Close()
			}()
			select {
			case <-ctx.Done():
				return 1
			default:
			}

			logEntry := log.WithFields(log.Fields{
				//"module":   "stream-order-daemon",
				"consumer": consumerID,
				"group":    s.Group,
				"name":     s.Name,
			})

			var msgs Stream
			//offset := "0-0"
			//if i != 0 {
			//	offset = ">"
			//}
			//offset := ">"

			logEntry.Data["offset"] = offset
			resp, err := redisLib.Values(rd.Do("xreadgroup", "group", s.Group, consumerID, "block", timeout, "count", 10, "streams", s.Name, offset))
			//fmt.Println(s.Name, s.Group, consumerID, offset, resp, err)
			if err != nil {
				if err == redisLib.ErrNil {
					offset = ">"
					timeout = 10000
				} else if strings.Contains(err.Error(), "No such key") {
					//创建一条新的记录，已防止出错
					//rd.Do("xadd", s.Name, "*", "obj", "null")
					offset = ">"
					timeout = 10000
				} else if strings.Contains(err.Error(), "timeout") {
					//logEntry.Error("xreadgroup timeout ", err)
					time.Sleep(3 * time.Second)
					//offset = ">"
					//timeout = 10000
					//logEntry.Error(err)
					//return
					//offset = ">"
					//timeout = 10000
				} else {
					logEntry.Error("xreadgroup", err)
					time.Sleep(10 * time.Second)
				}
				return
			}

			ScanStream(resp, &msgs, t)
			if len(msgs.Data) == 0 {
				offset = ">"
				timeout = 10000
				return
			}

			for _, m := range msgs.Data {
				if m.Err != nil {
					logEntry.Error(m.Err)
					if _, err := rd.Do("xdel", s.Name, m.ID); err != nil {
						logEntry.Error(err)
					}
					continue
				}
				offset = m.ID
				//if i != 0 {
				//logEntry.Info("from new", s.Name, consumerID, m, offset)
				//}
				// fmt.Println("\n\nbefore for....", msg.Data, "\n\n")
				// time.Sleep(time.Second)
				//for _, v := range msg.Data {
				//fmt.Println(m, i)
				//time.Sleep(10)
				if err := s.callback(context.WithValue(context.Background(), "logEntry", logEntry), m.ID, m.Vals); err != nil {
					//fmt.Println("enter callback err", err)
					logEntry.Error(err)
					continue
				}
				//fmt.Println("dddddddddd", s.Name, s.Group, m.ID, m)
				if _, err := rd.Do("xack", s.Name, s.Group, m.ID); err != nil {
					logEntry.Error(err)
					continue
				}
			}
			return
		}(); ret == 1 {
			//fmt.Println("break")
			break
		}
	}
}

// Daemon 后台允许
func (s *StreamHandle) Daemon(ctx context.Context, t interface{}, workerNums ...int) {
	rd := GetRedis()
	defer rd.Close()

	//创建stream group
	if _, err := rd.Do("xgroup", "create", s.Name, s.Group, "$"); err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "no such key") {
			// log.Warning(err, s.Name)
			time.Sleep(10 * time.Second)
			return
		} else if !strings.Contains(errStr, "already exists") {
			// log.Error(err, s.Name)
			time.Sleep(10 * time.Second)
			return
		}
	}

	wg := &sync.WaitGroup{}
	//接受数据并把数据转到channel
	go s.handleUnAckMsg(ctx, t)
	// })

	if len(workerNums) == 0 {
		workerNums = make([]int, 1)
		workerNums[0] = 1
	}

	wg.Add(workerNums[0])
	for i := 0; i < workerNums[0]; i++ {
		go func(i int) {
			defer wg.Done()
			s.handle(ctx, fmt.Sprintf("%s-%d", s.Consumer, i), t, ">")
		}(i)
	}
	wg.Wait()
}
