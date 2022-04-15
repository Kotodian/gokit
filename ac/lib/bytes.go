package lib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Kotodian/gokit/lodash/types"
)

func BytesToInt(bys []byte) int {
	bytebuff := bytes.NewBuffer(bys)
	var data int64
	binary.Read(bytebuff, binary.LittleEndian, &data)
	return int(data)
}

func IntToBytes(data int) []byte {
	if data == 0 {
		return []byte{0x00}
	}
	body := make([]byte, 0)
	buf := bytes.NewBuffer(body)
	binary.Write(buf, binary.LittleEndian, data)
	return buf.Bytes()
}

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
	// 全部转换成utc时间发送
	t = t.UTC()
	// 换算成毫秒
	msec := t.Nanosecond()/int(time.Millisecond) + t.Second()*1000
	return []byte{byte(t.Year() - 2000), byte(t.Month()), byte(t.Weekday()<<5) | byte(t.Day()), byte(t.Hour()), byte(t.Minute()), byte(msec >> 8), byte(msec)}
	// return []byte{byte(msec), byte(msec >> 8), byte(t.Minute()), byte(t.Hour()),
	// 	byte(t.Weekday()<<5) | byte(t.Day()), byte(t.Month()), byte(t.Year() - 2000)}
}

func ParseCP56Time2a(b []byte) time.Time {
	b = ReserveBytes(b)
	if len(b) < 7 || b[2]&0x80 == 0x80 {
		return time.Time{}
	}
	// 取出前两个字节代表毫秒
	x := int(binary.LittleEndian.Uint16(b))
	msec := x % 1000
	// 毫秒除1000为秒
	sec := x / 1000
	// 截取后6位为分钟
	// 0011 1111
	min := int(b[2] & 0x3f)
	// 截取后5位为小时
	hour := int(b[3] & 0x1f)
	// 截取后5位为天
	day := int(b[4] & 0x1f)
	// 截取后4位为天
	month := time.Month(b[5] & 0x0f)
	// 截取后7位+2000为年份
	// 0111 1111
	year := 2000 + int(b[6]&0x7f)

	nsec := msec * int(time.Millisecond)
	date := time.Date(year, month, day, hour, min, sec, nsec, time.UTC)
	return date.Local()
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
