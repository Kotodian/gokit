package lib

import (
	"fmt"
	"testing"

	"github.com/magiconair/properties/assert"
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
	text := "123456"
	cbcEncrypt := NewCBCEncrypt()
	data, err := cbcEncrypt.Encode([]byte(text), []byte("12345678"))
	if err != nil {
		t.Error(err)
		return
	}
	text2, err := cbcEncrypt.Decode(data, []byte("12345678"))
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, text, string(text2))
	fmt.Println(text2)
}

func TestECBEncrypt(t *testing.T) {
	text := "123456"
	encrypt := NewECBEncrypt()
	data, err := encrypt.Encode([]byte(text), []byte("12345678"))
	if err != nil {
		t.Error(err)
		return
	}
	text2, err := encrypt.Decode(data, []byte("12345678"))
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, text, string(text2))
	fmt.Println(text2)
}
