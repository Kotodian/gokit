package workpool

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type Flag int64

const (
	FLAG_OK    Flag = 1 << iota
	FLAG_RETRY Flag = 1 << iota
)

type TaskFunc func(w *WorkPool, args ...interface{}) Flag

type Task struct {
	F    TaskFunc
	Args []interface{}
}

type WorkPool struct {
	pool        chan Task
	workerCount int

	// stop相关
	stopCtx        context.Context
	stopCancelFunc context.CancelFunc
	wg             sync.WaitGroup

	// sleep相关
	sleepCtx        context.Context
	sleepCancelFunc context.CancelFunc
	sleepSeconds    int64
	sleepNotify     chan bool
}

func (t *Task) Execute(w *WorkPool) Flag {
	return t.F(w, t.Args...)
}

func New(workerCount, poolLen int) *WorkPool {
	return &WorkPool{
		workerCount: workerCount,
		pool:        make(chan Task, poolLen),
		sleepNotify: make(chan bool),
	}
}

func (w *WorkPool) PushTask(t Task) {
	w.pool <- t
}

func (w *WorkPool) PushTaskFunc(f TaskFunc, args ...interface{}) {
	w.pool <- Task{
		F:    f,
		Args: args,
	}
}

func (w *WorkPool) work(i int) {
	for {
		select {
		case <-w.stopCtx.Done():
			w.wg.Done()
			return
		case <-w.sleepCtx.Done():
			time.Sleep(time.Duration(w.sleepSeconds) * time.Second)
		case t := <-w.pool:
			flag := t.Execute(w)
			if flag&FLAG_RETRY != 0 {
				w.PushTask(t)
				//fmt.Printf("work %v PushTask,pool length %v\n", i, len(w.pool))
			}
		}
	}
}

func (w *WorkPool) Start() *WorkPool {
	//fmt.Printf("workpool run %d worker\n", w.workerCount)
	w.wg.Add(w.workerCount + 1)
	w.stopCtx, w.stopCancelFunc = context.WithCancel(context.Background())
	w.sleepCtx, w.sleepCancelFunc = context.WithCancel(context.Background())
	go w.sleepControl()
	for i := 0; i < w.workerCount; i++ {
		go w.work(i)
	}
	return w
}

func (w *WorkPool) Stop() {
	w.stopCancelFunc()
	w.wg.Wait()
}

func (w *WorkPool) sleepControl() {
	//fmt.Println("sleepControl start...")
	for {
		select {
		case <-w.stopCtx.Done():
			w.wg.Done()
			return
		case <-w.sleepNotify:
			//fmt.Printf("receive sleep notify start...\n")
			w.sleepCtx, w.sleepCancelFunc = context.WithCancel(context.Background())
			w.sleepCancelFunc()
			//fmt.Printf("sleepControl will star sleep %v s\n", w.sleepSeconds)
			time.Sleep(time.Duration(w.sleepSeconds) * time.Second)
			w.sleepSeconds = 0
			//fmt.Println("sleepControl was end sleep")
		}
	}
}

func (w *WorkPool) SleepNotify(seconds int64) {
	// 因为需要CAS操作，所以sleepSeconds没有采用time.Duration类型
	// 成功设置后才发出通知
	if atomic.CompareAndSwapInt64(&w.sleepSeconds, 0, seconds) {
		//fmt.Printf("sleepSeconds set %v\n", seconds)
		w.sleepNotify <- true
	}
}
