package id

import (
	"testing"
)

func Test_NewID(t *testing.T) {
	evseID, _ := NewEvseID()
	if evseID<<BitsTimestamp != 0 {
		t.Error("genr error evseID")
	}

	connectorID, _ := GetPosID(evseID, KindTypeConnector, 2)
	if GetPosByID(connectorID) != 2 {
		t.Error("get pos id error")
	}

	if KindTypeConnector != GetTypeByID(connectorID) {
		t.Error("get type id error")
	}
}
