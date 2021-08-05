package redis

import (
	"encoding/json"

	"github.com/gomodule/redigo/redis"
)

type IBase interface {
	InstanceName() string
	BeforeSave() error
	SetKey(string)
	AfterSave()
}

type Base struct {
	Key string `redis:"-"`
}

func (b Base) InstanceName() string {
	return b.Key
}

func (b *Base) BeforeSave() error {
	return nil
}

func (b *Base) AfterSave() {

}

func (b *Base) SetKey(k string) {
	b.Key = k
}

func Get(o IBase, key string) (bool, error) {
	rd := pool.Get()
	defer rd.Close()

	o.SetKey(key)
	v, err := redis.Values(rd.Do("HGETALL", key))
	if err != nil {
		return false, err
	}

	if len(v) == 0 {
		return false, nil
	}
	// fmt.Println("\n\n", key, v)
	err = redis.ScanStruct(v, o)
	if err != nil {
		return false, err
	}

	return true, nil
}

func Save(o IBase, ttl ...int32) error {
	rd := pool.Get()
	defer rd.Close()

	err := o.BeforeSave()
	if err != nil {
		return err
	}
	if len(ttl) == 0 {
		_, err = rd.Do("HMSET", redis.Args{}.Add(o.InstanceName()).AddFlat(o)...)
		if err != nil {
			return err
		}
	} else {
		rd.Send("MULTI")
		rd.Send("HMSET", redis.Args{}.Add(o.InstanceName()).AddFlat(o)...)
		rd.Send("EXPIRE", o.InstanceName(), ttl[0])
		_, err = rd.Do("Exec")
		if err != nil {
			return err
		}
	}
	o.AfterSave()
	return nil
}

//
//func PublishStream(name string, obj interface{}, maxlen ...int32) error {
//	rd := pool.Get()
//	defer rd.Close()
//
//	_maxlen := int32(0)
//	if len(maxlen) > 0 && maxlen[0] != 0 {
//		_maxlen = maxlen[0]
//	} else {
//		_maxlen = 100
//	}
//
//	var str string
//
//	if s, err := json.Marshal(obj); err != nil {
//		return err
//	} else {
//		str = string(s)
//	}
//
//	args := redis.Args{}.Add(name).AddFlat(map[string]interface{}{"maxlen": _maxlen}).Add("*").AddFlat(map[string]string{
//		"obj": str,
//	})
//	//fmt.Println(args)
//	_, err := rd.Do("xadd", args...)
//	return err
//}

// PublishStreamWithMaxlen
func PublishStreamWithMaxlen(name string, maxlen int32, obj interface{}) error {
	//xadd codehole maxlen 3 * name xiaorui age 1
	// args := redis.Args{}.Add(name).AddFlat(map[string]interface{}{"maxlen": maxlen}).Add("*").AddFlat(obj)
	//var str string
	//
	//if s, err := json.Marshal(obj); err != nil {
	//	return err
	//} else {
	//	str = string(s)
	//}

	var str string

	if s, err := json.Marshal(obj); err != nil {
		return err
	} else {
		str = string(s)
	}

	args := redis.Args{}.Add(name).AddFlat(map[string]interface{}{"maxlen": maxlen}).Add("*").AddFlat(map[string]string{
		"obj": str,
	})
	//fmt.Println("pub stream", args)
	if _, err := Do("xadd", args...); err != nil {
		return err
	}
	return nil
}

// PublishStreamWithJSON
func PublishStreamWithJSON(name string, obj interface{}, maxlen ...int32) error {
	rd := pool.Get()
	defer rd.Close()
	//xadd codehole maxlen 3 * name xiaorui age 1
	// args := redis.Args{}.Add(name).AddFlat(map[string]interface{}{"maxlen": maxlen}).Add("*").AddFlat(obj)
	var str string

	if s, err := json.Marshal(obj); err != nil {
		return err
	} else {
		str = string(s)
	}

	_maxlen := int32(0)
	if len(maxlen) > 0 && maxlen[0] != 0 {
		_maxlen = maxlen[0]
	} else {
		_maxlen = 100
	}

	args := redis.Args{}.Add(name).AddFlat(map[string]interface{}{"maxlen": _maxlen}).Add("*").AddFlat(map[string]string{
		"obj": str,
	})
	if _, err := rd.Do("xadd", args...); err != nil {
		return err
	}
	return nil
}
