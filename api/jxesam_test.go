package api

import (
	"fmt"
	"testing"
)

func TestAccessVerify(t *testing.T) {
	verify, err := AccessVerify(&AccessVerifyRequest{
		DeviceSerialNumber: "T17210000AC",
		RemoteAddress:      "127.0.0.1",
		RequestPort:        "31887",
	})
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(verify.CoreID, verify.KeepAlive, verify.Registered)
}
