package api

import "testing"

func TestPushInterval(t *testing.T) {
	Init()
	pushIntervalResponse, err := PushInterval(&PushIntervalRequest{EquipmentID: 586069658767531})
	if err != nil {
		t.Error(err)
	} else {
		t.Log(pushIntervalResponse)
	}
}
