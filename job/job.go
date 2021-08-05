package job

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type jobOnMessageFunc func(context.Context, interface{}) (isBreak bool, err error)
type jobOnTimeoutFunc func(context.Context) (isBreak bool, err error)

var jobMaps sync.Map

// GetJob 获取job对象
func GetJob(s string) *Job {
	j, loaded := jobMaps.LoadOrStore(s, &Job{
		//Done:   make(chan struct{}, 1),
		mapKey: s,
		ref:    1,
	})
	job := j.(*Job)
	if loaded {
		job.AddRef()
	}
	return job
}

// Job job
type Job struct {
	mapKey      string
	isListening int32
	Lock        sync.RWMutex
	ref         int32
	process     sync.Map
	//Done        chan struct{}
	onceUnlock sync.Once
}

// JobItem jobitem
type JobItem struct {
	CH          chan interface{}
	ErrCH       chan interface{}
	Timeout     time.Duration
	Func        jobOnMessageFunc
	TimeoutFunc jobOnTimeoutFunc
}

// AddRef 添加计数(原子性)
func (j *Job) AddRef() {
	atomic.AddInt32(&j.ref, 1)
}

// GetRef 返回引用
func (j *Job) GetRef() int32 {
	return atomic.LoadInt32(&j.ref)
}

// Release 释放计数,当计数等于0时测试，清理当前job,并关闭
func (j *Job) Release() {
	if atomic.AddInt32(&j.ref, -1) == 0 {
		jobMaps.Delete(j.mapKey)
		//close(j.Done)
		j.SetIsListening(false)
	}
}

// Unlock 解锁任务
//func (j *Job) Unlock() {
//	if atomic.LoadInt32(&j.ref) == 0 {
//		//j.OnceUnlock.Do(func() {
//		j.Lock.Unlock()
//		//})
//	}
//}

// LockByProcess ...
func (j *Job) LockByProcess(name string) {
	data, loaded := j.process.LoadOrStore(name, int32(0))
	ref := data.(int32)
	if !loaded {
		j.Lock.Lock()
	}
	j.process.Store(name, atomic.AddInt32(&ref, 1))
}

// UnlockByProcess ...
func (j *Job) UnlockByProcess(name string) {
	if data, ok := j.process.Load(name); ok {
		ref := data.(int32)
		ref1 := atomic.AddInt32(&ref, -1)
		if ref1 == 0 {
			j.Lock.Unlock()
			j.process.Delete(name)
		} else {
			j.process.Store(name, ref1)
		}
	}
}

// SetIsListening 设置有监听的
func (j *Job) SetIsListening(t bool) bool {
	//j.lock.Lock()
	//defer j.lock.Unlock()
	//j.isListening = t
	//atomic.StoreInt32(&j.isListening, 1)
	if t {
		if atomic.SwapInt32(&j.isListening, 1) == 0 {
			return true
		}
	} else {
		atomic.StoreInt32(&j.isListening, 0)
	}
	return false
}

func (j *Job) IsListening() bool {
	return atomic.LoadInt32(&j.isListening) == 1
}

func (j *Job) Unlock() {
	j.onceUnlock.Do(func() {
		j.Lock.Unlock()
	})
}
