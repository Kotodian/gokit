package workqueue

type Interface interface {
	Add(item any)
	Len() int
	Get() (item any, shutdown bool)
	Done(item any)
	Shutdown()
	ShutDownWithDrain()
	ShuttingDown() bool
}


