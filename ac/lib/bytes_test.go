package lib

import (
	"fmt"
	"testing"
	"time"
)

func TestParseCP56Time2a(t *testing.T) {
	b := []byte{0x50, 0x46, 0x24, 0x0d, 0xd3, 0x02, 0x14}
	b = ReserveBytes(b)
	t.Log(ParseCP56Time2a(b).Format("2006-01-02 15:04:05"))
	b = []byte{0x15, 0x09, 0x01, 0x0A, 0x1F, 0xA1, 0xE3}
	ti := ParseCP56Time2a(b)
	b = CP56Time2a(ti)
	t.Logf("%X\n", b)
	b2 := CP56Time2a(time.Now())
	t2 := ParseCP56Time2a(b2)
	t.Log(t2.Format("2006-01-02 15:04:05"))
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

func TestCP56Time2a(t *testing.T) {
	now := time.Now()
	t.Log(now.Format("2006-01-02 15:04:05"))
	t.Logf("%X", CP56Time2a(now))
}

func TestReserveBytes(t *testing.T) {
	b := ReserveBytes([]byte{0x7F, 00, 00, 01})
	t.Logf("%X", b)
}

func TestBytesToFloat(t *testing.T) {
	d := BytesToFloat([]byte{0x01, 0x18}, 2)
	t.Log(d)
}

func TestFloatToBytes(t *testing.T) {
	b := FloatToBytes(1.23, 2)
	t.Logf("%X\n", b)
}

func TestBytesToInt(t *testing.T) {
	i := BytesToInt([]byte{0x01, 0x01, 0xff})
	t.Log(i)
}
