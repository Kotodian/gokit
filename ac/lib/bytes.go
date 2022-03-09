package lib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
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
