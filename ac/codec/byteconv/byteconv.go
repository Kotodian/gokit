package byteconv

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"unsafe"
)

var endian binary.ByteOrder
var isLittleEndianFlag bool

func init() {
	endian = binary.LittleEndian
	isLittleEndianFlag = true
}

// SetLittleEndian 设置小端
func SetLittleEndian() {
	endian = binary.LittleEndian
	isLittleEndianFlag = true
}

// SetBigEndian 设置大端
func SetBigEndian() {
	endian = binary.BigEndian
	isLittleEndianFlag = false
}

// Convert 二进制取反
// 输入: src  二进制数据
//      length  待转换的二进制数据长度
// 返回: dst   取反后的二进制数据
func convert(src []byte, length int) []byte {
	dst := make([]byte, length)
	for i := 0; i < length; i++ {
		dst[i] = src[i] ^ 0xff
	}
	return dst
}

// Power 求权
// 输入: base    进制基数
//      times   权级数
// 返回: uint64  当前数据位的权
func power(base, times int) uint64 {
	var rslt uint64 = 1
	for i := 0; i < times; i++ {
		rslt *= uint64(base)
	}
	return rslt
}

// BcdToDec BCD转10进制
func BcdToDec(bcd []byte) uint64 {
	if isLittleEndianFlag {
		return BcdToDecLittle(bcd)
	}
	return BcdToDecBig(bcd)
}

// BcdToDecBig BCD转10进制 (大端)
// 输入: bcd     待转换的BCD码
// 返回: uint64  转换后的数字
// 思路:压缩BCD码一个字符所表示的十进制数据范围为0 ~ 99,进制为100  先求每个字符所表示的十进制值，然后乘以权
func BcdToDecBig(bcd []byte) (ret uint64) {
	var nLen = len(bcd)
	for i := 0; i < nLen; i++ {
		tmp := uint64(((bcd[i]>>4)&0x0f)*10 + (bcd[i] & 0x0f))
		ret += uint64(tmp * power(100, nLen-1-i))
	}
	return ret
}

// BcdToDecLittle BCD转10进制 (小端)
func BcdToDecLittle(bcd []byte) (ret uint64) {
	var nLen = len(bcd)
	for i := 0; i < nLen; i++ {
		tmp := uint64(((bcd[i]>>4)&0x0f)*10 + (bcd[i] & 0x0f))
		ret += uint64(tmp * power(100, i))
	}
	return ret
}

// DecToBcd 十进制转BCD码
func DecToBcd(dec uint64) []byte {
	if isLittleEndianFlag {
		return DecToBcdLittle(dec)
	}
	return DecToBcdBig(dec)
}

// DecToBcdBig 十进制转BCD码 (大端)
func DecToBcdBig(dec uint64) []byte {
	l := len(fmt.Sprintf("%d", dec))
	l = l/2 + l%2
	bcd := make([]byte, l)
	for i := l - 1; i >= 0; i-- {
		temp := uint8(dec % 100)
		bcd[i] = byte(uint8(((temp / 10) << 4) + ((temp % 10) & 0x0F)))
		dec /= 100
	}
	return bcd
}

// DecToBcdLittle 十进制转BCD码 (小端)
func DecToBcdLittle(dec uint64) []byte {
	l := len(fmt.Sprintf("%d", dec))
	l = l/2 + l%2
	bcd := make([]byte, l)
	for i := 0; i < l; i++ {
		temp := uint8(dec % 100)
		bcd[i] = byte(uint8(((temp / 10) << 4) + ((temp % 10) & 0x0F)))
		dec /= 100
	}
	return bcd
}

// BCDToByte bcd转换成byte
func BCDToByte(dec []byte) (ret []byte) {
	ret = make([]byte, len(dec)*2)
	for i, j := 0, 0; i < len(dec); i++ {
		if cur := dec[i] >> 4; cur > 9 {
			ret[j] = cur - 10 + 'A'
		} else {
			ret[j] = cur + '0'
		}
		if cur := dec[i] & 0x0F; cur > 9 {
			ret[j+1] = cur - 10 + 'A'
		} else {
			ret[j+1] = cur + '0'
		}
		j = j + 2
	}
	return
}

// ByteToBcd 字节转换为BCD
func ByteToBcd(dec []byte) (ret []byte) {
	l := len(dec)
	for i := 0; i < l; i++ {
		if !(dec[i] >= '0' && dec[i] <= '9' || dec[i] >= 'a' && dec[i] <= 'f' || dec[i] >= 'A' && dec[i] <= 'F') {
			return
		}
	}

	l = (l - l%2) / 2
	ret = make([]byte, l)
	for i, j := 0, 0; i < l; i++ {
		if dec[j] > '9' {
			if dec[j] > 'F' {
				ret[i] = dec[j] - 'a' + 10
			} else {
				ret[i] = dec[j] - 'A' + 10
			}
		} else {
			ret[i] = dec[j] - '0'
		}
		if dec[j+1] > '9' {
			if dec[j+1] > 'F' {
				ret[i] = (ret[i] << 4) + dec[j+1] - 'a' + 10
			} else {
				ret[i] = (ret[i] << 4) + dec[j+1] - 'A' + 10
			}
		} else {
			ret[i] = (ret[i] << 4) + dec[j+1] - '0'
		}

		j += 2
	}
	return
}

// ByteToBcdByte 字节转换为BCD
// todo: 大小端区分
func ByteToBcdByte(dec []byte) []byte {
	for i := 0; i < len(dec); i++ {
		nLow := dec[i] & 0x0f
		nHigh := dec[i] >> 4 & 0x0f
		dec[i] = uint8(nHigh*10 + nLow)
	}
	return dec
}

// B2S 二进制转化字符串
func B2S(buf []byte) string {
	return *(*string)(unsafe.Pointer(&buf))
}

// S2B 字符串转化二进制
func S2B(s *string) []byte {
	return *(*[]byte)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(s))))
}

// Byte2String 二进制转化字符串
func Byte2String(buf []byte) (ret string) {
	for k := range buf {
		if buf[k] == 0x00 {
			break
		}
		ret += string(buf[k])
	}
	return ret
}

// Trim 二进制转化字符串
func Trim(buf []byte) (ret []byte) {
	for k := range buf {
		if buf[k] != 0x00 {
			return buf[k:]
		}
	}
	return buf
}
