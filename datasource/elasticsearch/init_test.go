package elasticsearch

import (
	"context"
	"testing"
	"time"
)

func TestIndexCreate(t *testing.T) {
	err := IndexCreate(context.Background(), "equipment")
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateByQuery(t *testing.T) {
	script := `ctx._source.lastLostConnReason=params.lastLostConnReason;ctx._source.lastLostConnTime=params.lastLostConnTime;`
	params := map[string]interface{}{
		"lastLostConnReason": "websocket 1009",
		"lastLostConnTime":   time.Now().Format("2006-01-02 15:04:05"),
	}
	err := UpdateByQuery(context.Background(), "equipment", script, params, "sn", "T1641735210", true)
	if err != nil {
		t.Error(err)
	}
}
