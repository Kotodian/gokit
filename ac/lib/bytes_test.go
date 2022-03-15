package lib

import (
	"testing"
)

func TestParseCP56Time2a(t *testing.T) {
	b := []byte{0x50, 0x46, 0x24, 0x0d, 0xd3, 0x02, 0x14}
	t.Log(ParseCP56Time2a(b).Format("2006-01-02 15:04:05"))
}
