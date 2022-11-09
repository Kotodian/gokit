package workqueue

import (
	"container/heap"
	"sync"
	"time"
)

type DelayingInterface interface {
	Interface
	AddAfter(item any, duration time.Duration)
}

type delayingType struct {
	Interface

	stopCh   chan struct{}
	stopOnce sync.Once

	hearbeat *time.Ticker

	waitingForAddCh chan *waitFor
}

func newDelayingQueue(q Interface) *delayingType {
	ret := &delayingType{
		Interface:       q,
		hearbeat:        time.NewTicker(maxWait),
		stopCh:          make(chan struct{}),
		waitingForAddCh: make(chan *waitFor, 1000),
	}
	go ret.waitingLoop()
	return ret
}

type waitFor struct {
	data    any
	readyAt time.Time
	index   int
}

type waitForPriorityQueue []*waitFor

func (pq waitForPriorityQueue) Len() int {
	return len(pq)
}

func (pq waitForPriorityQueue) Less(i, j int) bool {
	return pq[i].readyAt.Before(pq[j].readyAt)
}

func (pq waitForPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *waitForPriorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*waitFor)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *waitForPriorityQueue) Pop() any {
	n := len(*pq)
	item := (*pq)[n-1]
	item.index = -1
	*pq = (*pq)[0:(n - 1)]
	return item
}

func (pq waitForPriorityQueue) Peek() any {
	return pq[0]
}

func (q *delayingType) Shutdown() {
	q.stopOnce.Do(func() {
		q.Interface.Shutdown()
		close(q.stopCh)
		q.hearbeat.Stop()
	})
}

func (q *delayingType) AddAfter(item any, duration time.Duration) {
	if q.ShuttingDown() {
		return
	}

	if duration <= 0 {
		q.Add(item)
		return
	}

	select {
	case <-q.stopCh:
	case q.waitingForAddCh <- &waitFor{data: item, readyAt: time.Now().Add(duration)}:
	}
}

const maxWait = 10 * time.Second

func (q *delayingType) waitingLoop() {
	never := make(<-chan time.Time)

	var nextReadyAtTimer *time.Timer
	waitingForQueue := &waitForPriorityQueue{}
	heap.Init(waitingForQueue)

	waitingEntryByData := make(map[any]*waitFor)

	for {
		if q.Interface.ShuttingDown() {
			return
		}

		now := time.Now()

		for waitingForQueue.Len() > 0 {
			entry := waitingForQueue.Peek().(*waitFor)
			if entry.readyAt.After(now) {
				break
			}
			entry = heap.Pop(waitingForQueue).(*waitFor)
			q.Add(entry)
			delete(waitingEntryByData, entry.data)
		}

		nextReadyAt := never
		if waitingForQueue.Len() > 0 {
			if nextReadyAtTimer != nil {
				nextReadyAtTimer.Stop()
			}
			entry := waitingForQueue.Peek().(*waitFor)
			nextReadyAtTimer = time.NewTimer(entry.readyAt.Sub(now))
			nextReadyAt = nextReadyAtTimer.C
		}

		select {
		case <-q.stopCh:
			return
		case <-q.hearbeat.C:
		case <-nextReadyAt:
		case waitEntry := <-q.waitingForAddCh:
			if waitEntry.readyAt.After(time.Now()) {
				existing, exists := waitingEntryByData[waitEntry.data]
				if exists {
					if existing.readyAt.After(waitEntry.readyAt) {
						existing.readyAt = waitEntry.readyAt
						heap.Fix(waitingForQueue, existing.index)
					}
					return
				}
				heap.Push(waitingForQueue, waitEntry)
				waitingEntryByData[waitEntry.data] = waitEntry
			} else {
				q.Add(waitEntry.data)
			}
			drained := false
			for !drained {
				select {
				case waitEntry := <-q.waitingForAddCh:
					if waitEntry.readyAt.After(time.Now()) {
						existing, exists := waitingEntryByData[waitEntry.data]
						if exists {
							if existing.readyAt.After(waitEntry.readyAt) {
								existing.readyAt = waitEntry.readyAt
								heap.Fix(waitingForQueue, existing.index)
							}
							return
						}
						heap.Push(waitingForQueue, waitEntry)
						waitingEntryByData[waitEntry.data] = waitEntry
					} else {
						q.Add(waitEntry.data)
					}
				default:
					drained = false
				}
			}
		}
	}
}
