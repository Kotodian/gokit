package elasticsearch

import (
	"context"
	"testing"
)

func TestIndexCreate(t *testing.T) {
	err := IndexCreate(context.Background(), "equipment")
	if err != nil {
		t.Error(err)
	}
}
