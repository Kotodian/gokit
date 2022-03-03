package strategy

import (
	backoff2 "github.com/Kotodian/gokit/retry/strategy/backoff"
	"time"
)

type Strategy func(attempt uint) bool

func Limit(attemptLimit uint) Strategy {
	return func(attempt uint) bool {
		return attempt < attemptLimit
	}
}

func Delay(duration time.Duration) Strategy {
	return func(attempt uint) bool {
		if attempt == 0 {
			time.Sleep(duration)
		}
		return true
	}
}

func Wait(durations ...time.Duration) Strategy {
	return func(attempt uint) bool {
		if attempt > 0 && len(durations) > 0 {
			durationIndex := int(attempt - 1)

			if len(durations) <= durationIndex {
				durationIndex = len(durations) - 1
			}

			time.Sleep(durations[durationIndex])
		}
		return true
	}
}

func BackOff(algorithm backoff2.Algorithm) Strategy {
	return func(attempt uint) bool {
		if attempt > 0 {
			time.Sleep(algorithm(attempt))
		}
		return true
	}
}
