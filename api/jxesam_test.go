package api

import (
	"fmt"
	"testing"
)

func TestAccessVerify(t *testing.T) {
	verify, err := AccessVerify(&AccessVerifyRequest{
		DeviceSerialNumber: "T1641735210",
		RemoteAddress:      "127.0.0.1",
		RequestPort:        "8844",
	})
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(verify.CoreID, verify.KeepAlive, verify.Registered)
}
