package utils

import (
	"testing"
)

func TestBcryptPwd(t *testing.T) {
	ori := "123456"

	pwd := NewBcryptPwd(ori)

	// expect match
	if !pwd.Match(ori) {
		t.Fatal(ori, pwd)
	}

	// expect not match
	if pwd.Match(ori + ori) {
		t.Fatal(ori+ori, pwd)
	}
}
