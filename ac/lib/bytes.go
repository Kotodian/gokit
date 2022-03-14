package lib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func BytesToInt(bys []byte) int {
	bytebuff := bytes.NewBuffer(bys)
	var data int64
	binary.Read(bytebuff, binary.BigEndian, &data)
	return int(data)
}

func IntToBytes(data int) []byte {
	if data == 0 {
		return []byte{0x00}
	}
	body := make([]byte, 0)
	buf := bytes.NewBuffer(body)
	binary.Write(buf, binary.BigEndian, data)
	return buf.Bytes()
}

func BCDToNumber(bcd []byte) string {
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
	bcd = RemoveZero(bcd)
	byteBuff := bytes.NewBuffer(bcd)
	var data int64
	binary.Read(byteBuff, binary.LittleEndian, &data)
	if data == 0 {
		return false
	}
	return true
}

func AsciiByteToString(body []byte) string {
	body = RemoveZero(body)
	return string(body)
}

func RemoveZero(body []byte) []byte {
	return bytes.TrimRight(body, "\x00")
}

func NumberToBCD(number string) []byte {
	var rNumber = number
	for i := 0; i < 8-len(number); i++ {
		rNumber = "f" + rNumber
	}
	bcd := Hex2Byte(rNumber)
	return bcd
}

func Hex2Byte(str string) []byte {
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
	for i := 0; i < length; i++ {
		payload = append(payload, 0x00)
	}
	return payload
}

func BytesToInt16(buf []byte) int {
	return int(binary.BigEndian.Uint16(buf))
}

func Int16ToBytes(n int) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(n))
	return b
}

func CP56Time2a(t time.Time) []byte {
	// 全部转换成utc时间发送
	t = t.UTC()
	// 换算成毫秒
	msec := t.Nanosecond()/int(time.Millisecond) + t.Second()*1000
	return []byte{byte(msec), byte(msec >> 8), byte(t.Minute()), byte(t.Hour()),
		byte(t.Weekday()<<5) | byte(t.Day()), byte(t.Month()), byte(t.Year() - 2000)}
}

func ParseCP56Time2a(b []byte) time.Time {
	if len(b) < 7 || b[2]&0x80 == 0x80 {
		return time.Time{}
	}
	x := int(binary.LittleEndian.Uint16(b))
	msec := x % 1000
	sec := x / 1000
	min := int(b[2] & 0x3f)
	hour := int(b[3] & 0x1f)
	day := int(b[4] & 0x1f)
	month := time.Month(b[5] & 0x0f)
	year := 2000 + int(b[6]&0x7f)

	nsec := msec * int(time.Millisecond)
	date := time.Date(year, month, day, hour, min, sec, nsec, time.UTC)
	return date.Local()
}
