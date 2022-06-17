package lib

import (
	"fmt"
	"testing"
)

func TestCBCEncrypt(t *testing.T) {
	cbcEncrypt := NewAESEncrypt(CBC)

	data, err := cbcEncrypt.Encode([]byte("0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF66"), []byte("12345678"))
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("%02X", data)
}

func TestCBCEncrypt2(t *testing.T) {
	cbcEncrypt := NewCBCEncrypt()
	data, err := cbcEncrypt.Encode([]byte("0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF66"), []byte("12345678"))
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("%02X", data)
}
