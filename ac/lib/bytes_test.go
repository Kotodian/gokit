package lib

import (
	"fmt"
	"testing"
)

func TestParseCP56Time2a(t *testing.T) {
	b := []byte{0x50, 0x46, 0x24, 0x0d, 0xd3, 0x02, 0x14}
	t.Log(ParseCP56Time2a(b).Format("2006-01-02 15:04:05"))
}

func TestNumberToBCD(t *testing.T) {
	// t.Log(uint16(8) & (1 << 3))
	// t.Log(BCDToString([]byte{0x32, 0x01, 0x02, 0x00, 0x00, 0x00, 0x00, 0x11, 0x15, 0x11, 0x16, 0x15, 0x55, 0x35, 0x02}))
	fmt.Printf("BCDFromUint32(1648439200): %X\n", StringToBCD("1648439200"))
}


func TestBytesToInt16(t *testing.T) {
	i := BytesToInt16([]byte{0x01, 0x00})
	t.Log(i)
}