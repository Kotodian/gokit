package utils

import (
	"time"
)

type Circuit struct {
	//hystrix.Opener
}

// start health check when circuit is opened
func (c *Circuit) Opened(now time.Time) {
}

// stop health check when circuit is closed
func (c *Circuit) Closed(now time.Time) {
}

func (c *Circuit) Success(now time.Time, duration time.Duration) {
}

func (c *Circuit) ErrBadRequest(now time.Time, duration time.Duration) {
}

func (c *Circuit) ErrInterrupt(now time.Time, duration time.Duration) {
}

func (c *Circuit) ErrConcurrencyLimitReject(now time.Time) {
}

func (c *Circuit) ErrShortCircuit(now time.Time) {
}

func (c *Circuit) ErrFailure(now time.Time, duration time.Duration) {
}

func (c *Circuit) ErrTimeout(now time.Time, duration time.Duration) {
}

func (c *Circuit) ShouldClose(now time.Time) bool {
	return false
}
