package retry

import "github.com/Kotodian/gokit/retry/strategy"

type Action func(attempt uint) error

type Ignore func(err error) bool

func DefaultIgnore(err error) bool {
	return true
}

func Retry(action Action, strategies ...strategy.Strategy) error {
	var err error

	for attempt := uint(0); (attempt == 0 || err != nil) && shouldAttempt(attempt, strategies...); attempt++ {
		err = action(attempt + 1)
	}
	return err
}

func shouldAttempt(attempt uint, strategies ...strategy.Strategy) bool {
	shouldAttempt := true

	for i := 0; shouldAttempt && i < len(strategies); i++ {
		shouldAttempt = shouldAttempt && strategies[i](attempt)
	}

	return shouldAttempt
}
