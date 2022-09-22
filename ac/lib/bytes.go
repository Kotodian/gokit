package lib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/Kotodian/gokit/lodash/types"
	"github.com/thinkgos/go-iecp5/asdu"
)

func IntToBytes(data uint, len int) []byte {
	bys := make([]byte, len)
	for i := 0; i < len; i++ {
		bys[i] = byte(data >> uint(8*i))
	}
	return bys
}

func BytesToInt(bys []byte) uint {
	var data uint
	for i := 0; i < len(bys); i++ {
		data += (uint(bys[i]) << uint(8*i))
	}
	return data
}

// func IntToBytes(data int) []byte {
// 	if data == 0 {
// 		return []byte{0x00}
// 	}
// 	body := make([]byte, 0)
// 	buf := bytes.NewBuffer(body)
// 	binary.Write(buf, binary.LittleEndian, data)
// 	return buf.Bytes()
// }

func BCDToString(bcd []byte) string {
	bcd = RemoveZero(bcd)
	var number string
	for _, i := range bcd {
		number += fmt.Sprintf("%02X", i)
	}
	pos := strings.LastIndex(number, "F")
	if pos == 8 {
		return "0"
	}
	return number[pos+1:]
}

func BINToBool(bcd []byte) bool {
	bcd = RemoveMAX(bcd)
	byteBuff := bytes.NewBuffer(bcd)
	var data int64
	binary.Read(byteBuff, binary.LittleEndian, &data)
	if data == 0 {
		return false
	}
	return true
}

func AsciiByteToString(body []byte) string {
	body = RemoveMAX(body)
	return string(body)
}

func RemoveMAX(body []byte) []byte {
	return bytes.TrimRight(body, "\xFF")
}

func RemoveZero(body []byte) []byte {
	return bytes.TrimRight(body, "\x00")
}

func StringToBCD(number string) []byte {
	bcd := hex2Byte(number)
	return bcd
}

func hex2Byte(str string) []byte {
	slen := len(str)
	bHex := make([]byte, len(str)/2)
	ii := 0
	for i := 0; i < len(str); i = i + 2 {
		if slen != 1 {
			ss := string(str[i]) + string(str[i+1])
			bt, _ := strconv.ParseInt(ss, 16, 32)
			bHex[ii] = byte(bt)
			ii = ii + 1
			slen = slen - 2
		}
	}
	return bHex
}

func FillZero(payload []byte, length int) []byte {
	l := len(payload)
	if l < length {
		for i := 0; i < length; i++ {
			payload = append(payload, 0x00)
		}
	}
	return payload
}

func FillMAX(payload []byte, length int) []byte {
	l := len(payload)
	if l < length {
		for i := 0; i < length-l; i++ {
			payload = append(payload, 0xFF)
		}
	}
	return payload
}

func BytesToInt16(buf []byte) int {
	return int(binary.LittleEndian.Uint16(buf))
}

func Int16ToBytes(n int) []byte {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, uint16(n))
	return b
}

func CP56Time2a(t time.Time) []byte {
	b := asdu.CP56Time2a(t, time.UTC)
	return b
}

func ParseCP56Time2a(b []byte) time.Time {
	t := asdu.ParseCP56Time2a(b, time.UTC)
	return t.Local()
}

func pow100(power byte) uint64 {
	res := uint64(1)
	for i := byte(0); i < power; i++ {
		res *= 100
	}
	return res
}

func BCDFromUint(value uint64, size int) []byte {
	buf := make([]byte, size)
	if value > 0 {
		remainder := value
		for pos := size - 1; pos >= 0 && remainder > 0; pos-- {
			tail := byte(remainder % 100)
			hi, lo := tail/10, tail%10
			buf[pos] = byte(hi<<4 + lo)
			remainder = remainder / 100
		}
	}
	return buf
}

// Returns uint8 value in BCD format.
//
// If value > 99, function returns value for last two digits of source value
// (Example: uint8(123) = uint8(0x23)).
func BCDFromUint8(value uint8) byte {
	return BCDFromUint(uint64(value), 1)[0]
}

// Returns two-bytes array with uint16 value in BCD format
//
// If value > 9999, function returns value for last two digits of source value
// (Example: uint8(12345) = []byte{0x23, 0x45}).
func BCDFromUint16(value uint16) []byte {
	return BCDFromUint(uint64(value), 2)
}

// Returns four-bytes array with uint32 value in BCD format
//
// If value > 99999999, function returns value for last two digits of source value
// (Example: uint8(1234567890) = []byte{0x23, 0x45, 0x67, 0x89}).
func BCDFromUint32(value uint32) []byte {
	return BCDFromUint(uint64(value), 4)
}

// Returns eight-bytes array with uint64 value in BCD format
//
// If value > 9999999999999999, function returns value for last two digits of source value
// (Example: uint8(12233445566778899) = []byte{0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99}).
func BCDFromUint64(value uint64) []byte {
	return BCDFromUint(value, 8)
}

func bcdToUint(value []byte, size int) uint64 {
	vlen := len(value)
	if vlen > size {
		value = value[vlen-size:]
	}
	res := uint64(0)
	for i, b := range value {
		hi, lo := b>>4, b&0x0f
		if hi > 9 || lo > 9 {
			return 0
		}
		res += uint64(hi*10+lo) * pow100(byte(vlen-i)-1)
	}
	return res
}

// Returns uint8 value converted from bcd byte.
//
// If byte is not BCD (e.g. 0x1A), function returns zero.
func BCDToUint8(value byte) uint8 {
	return uint8(bcdToUint([]byte{value}, 1))
}

// Return uint16 value converted from at most last two bytes of bcd bytes array.
//
// If any byte of used array part is not BCD (e.g 0x1A), function returns zero.
func BCDToUint16(value []byte) uint16 {
	return uint16(bcdToUint(value, 2))
}

// Return uint32 value converted from at most last four bytes of bcd bytes array.
//
// If any byte of used array part is not BCD (e.g 0x1A), function returns zero.
func BCDToUint32(value []byte) uint32 {
	return uint32(bcdToUint(value, 4))
}

// Return uint64 value converted from at most last eight bytes of bcd bytes array.
//
// If any byte of used array part is not BCD (e.g 0x1A), function returns zero.
func BCDToUint64(value []byte) uint64 {
	return bcdToUint(value, 8)
}

func BCDToUint[T types.Unsigned](value []byte, size int) T {
	vlen := len(value)
	if vlen > size {
		value = value[vlen-size:]
	}
	res := uint64(0)
	for i, b := range value {
		hi, lo := b>>4, b&0x0f
		if hi > 9 || lo > 9 {
			return 0
		}
		res += uint64(hi*10+lo) * pow100(byte(vlen-i)-1)
	}
	return T(res)
}

func BCDFromUintG[T types.Unsigned](value T, size int) []byte {
	buf := make([]byte, size)
	if value > 0 {
		remainder := value
		for pos := size - 1; pos >= 0 && remainder > 0; pos-- {
			tail := byte(remainder % 100)
			hi, lo := tail/10, tail%10
			buf[pos] = byte(hi<<4 + lo)
			remainder = remainder / 100
		}
	}
	return buf
}

func BINToFloat[T types.Float](value []byte, bit int) T {
	return 0.00
}

func FloatToBIN[T types.Float](value T, bit int) []byte {
	return nil
}

func ReserveBytes(b []byte) []byte {
	_b := make([]byte, len(b))
	for i := 0; i < len(b); i++ {
		_b[i] = b[len(b)-1-i]
	}
	return _b
}

func BytesToFloat(b []byte, bit int) float64 {
	decimal := BytesToInt16(b)
	pow := math.Pow(10, float64(bit))
	return float64(decimal) / pow
}

func FloatToBytes(f float64, bit int) []byte {
	decimal := int(f * math.Pow(10, float64(bit)))
	return Int16ToBytes(decimal)
}

// type Endpoint struct {
// 	// parse header
// 	hu HeaderUnmarshaler
// 	// parse header
// 	bu BodyUnmarshaler

// 	hw HeaderMarshaler

// 	bw BodyMarhshaler
// 	// parse encrypted data and encrypt data
// 	en map[EncryptType]Encrypt
// }

type Reader interface {
	Next(n int) []byte
}

// type HeaderMarshalerReader interface {
// 	Reader
// 	HeaderUnmarshaler
// }

// func NewHeaderUnmarshaler(body []byte) HeaderUnmarshaler {

// }

// func Unmarshal() {

// }

// type HeaderUnmarshaler interface {
// 	ReadStart(b byte) int
// 	ReadLength(b byte) int
// 	ReadEncrypt(byte) EncryptType
// }

// type Writer interface {
// 	Write(b []byte)
// }

type BodyReaderUnmarshaler interface {
	Reader
	ReadInt16(b []byte) int16
	ReadInt32(b []byte) int32
	ReadString(b []byte) string
	ReadDateTime(b []byte) time.Time
	ReadFloat(b []byte) float64
}

type BodyUnmarshaler interface {
	ReadInt16(b []byte) int16
	ReadInt32(b []byte) int32
	ReadString(b []byte) string
	ReadDateTime(b []byte) time.Time
	ReadFloat(b []byte) float64
}

// type HeaderMarshaler interface {
// 	Writer
// 	WriteStart()
// 	WriteLength(int)
// }

// type BodyMarhshaler interface {
// 	Writer
// 	WriteInt16(int16)
// 	WriteInt32(int32)
// 	WriteString(string)
// 	WriteDateTime(time.Time)
// 	WriteFloat(float64, int)
// }

// type LoginRequest struct {
// }

// func (r *LoginRequest) Unmarshal(u BodyUnmarshaler) error {
// 	return nil
// }
