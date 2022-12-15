package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPushInterval(t *testing.T) {
	Init()
	pushIntervalResponse, err := PushInterval(&PushIntervalRequest{EquipmentID: 586069658767531})
	if err != nil {
		t.Error(err)
	} else {
		t.Log(pushIntervalResponse)
	}
}

func TestServiceQRCode(t *testing.T) {
	Init()
	qrCode, err := ServiceQRCode(&ServiceQRCodeRequest{
		ConnectorId: 363981844624133,
	})
	t.Log(qrCode)
	assert.Nil(t, err)
	assert.Condition(t, func() (success bool) {
		return len(qrCode) > 0
	})
}
