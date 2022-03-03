package backoff

import (
	"math"
	"time"
)

type Algorithm func(attempt uint) time.Duration

func Incremental(initial, increment time.Duration) Algorithm {
	return func(attempt uint) time.Duration {
		return initial + (increment * time.Duration(attempt))
	}
}

func Linear(factor time.Duration) Algorithm {
	return func(attempt uint) time.Duration {
		return factor * time.Duration(attempt)
	}
}

func Exponential(factor time.Duration, base float64) Algorithm {
	return func(attempt uint) time.Duration {
		return factor * time.Duration(math.Pow(base, float64(attempt)))
	}
}
