package api

import (
	"fmt"
	"testing"
)

func TestAccessVerify(t *testing.T) {
	Init()
	verify, err := AccessVerify("lYMFB!X#87,7woq?$C#W~z", &AccessVerifyRequest{
		DeviceSerialNumber: "T1641735213",
		RemoteAddress:      "127.0.0.1",
		RequestPort:        "32887",
	})
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(verify.CoreID, verify.KeepAlive, verify.Registered)
	fmt.Println(verify.BaseURL)
}
