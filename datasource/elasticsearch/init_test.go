package elasticsearch

import (
	"context"
	"testing"
	"time"
)

func TestIndexCreate(t *testing.T) {
	Init()
	err := IndexCreate(context.Background(), "test2")
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateByQuery(t *testing.T) {
	//script := `ctx._source.lastLostConnReason=params.lastLostConnReason;ctx._source.lastLostConnTime=params.lastLostConnTime;ctx._source.state=params.state`
	params := map[string]interface{}{
		"lastLostConnReason": "websocket断开",
		"lastLostConnTime":   time.Now().Format("2006-01-02 15:04:05"),
		"state":              false,
	}
	err := UpdateByQuery(context.Background(), "equipment", params, map[string]interface{}{
		"accessPod": "jx-ac-ocpp-cluster-0",
	})
	if err != nil {
		t.Error(err)
	}
}
