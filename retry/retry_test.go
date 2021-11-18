package retry

import (
	"errors"
	"github.com/Kotodian/gokit/retry/strategy"
	"testing"
)

func TestRetry(t *testing.T) {
	err := Retry(func(attempt uint) error {
		if attempt == 3 {
			t.Log("success")
			return nil
		}
		t.Log("error")
		return errors.New("sss")
	}, DefaultIgnore, strategy.Limit(3))
	if err != nil {
		t.Error(err)
	}
}
