package lib

import (
	"fmt"
	"testing"
)

func TestCBCEncrypt(t *testing.T) {
	cbcEncrypt := NewAESEncrypt(CBC)

	data, err := cbcEncrypt.Encode([]byte("0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF66"), []byte("qwertyuiopasdfgh"))
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("%02X", data)
}
