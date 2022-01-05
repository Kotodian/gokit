package id

import (
	"testing"
)

func TestNext(t *testing.T) {
	next := Next()
	t.Log(next.String())
}
