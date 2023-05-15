package lib

import (
	"fmt"
	"testing"
)

func TestCRC(t *testing.T) {
	data := []byte{
		0x68, 0x01, 0x00, 0xC0, 0x07, 0x00, 0x01, 0x00, 0x02, 0x00, 0x14, 0x02, 0xD3, 0x0D, 0x24, 0x46, 0x50,
	}
	t.Logf("%X", CheckSum(data))
}

func TestMD5(t *testing.T) {
	data := []byte{
		0x18, 0x7d, 0x02, 0x20, 0xd1, 0x74, 0x03, 0x00, 0x11, 0x75, 0x03, 0x00, 0x09, 0x82, 0x00, 0x00, 0x15, 0x75, 0x03, 0x00, 0x17, 0x75, 0x03,
		0x00, 0x19, 0x75, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x25, 0x82,
		0x00, 0x00, 0x1d, 0x75, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x81, 0x82, 0x00, 0x00, 0xef, 0xc9, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23,
		0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00,
		0x23, 0x75, 0x03, 0x00, 0x83, 0x84, 0x00, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03,
		0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75,
		0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23,
		0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00,
		0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x81, 0xb0, 0x00, 0x00, 0x91, 0xb1, 0x00, 0x00, 0x07, 0xb2, 0x00, 0x00, 0x23, 0x75, 0x03,
		0x00, 0x23, 0x75, 0x03, 0x00, 0x9f, 0xb9, 0x00, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75,
		0x03, 0x00, 0xab, 0xb4, 0x00, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23,
		0x75, 0x03, 0x00, 0xc5, 0xb9, 0x00, 0x00, 0x01, 0xba, 0x00, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00,
		0x23, 0x75, 0x03, 0x00, 0xeb, 0xd1, 0x04, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03,
		0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75,
		0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x87, 0xd3, 0x03, 0x00, 0x41, 0xd3, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23,
		0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0xad, 0xba, 0x00, 0x00, 0x8f, 0xbb, 0x00, 0x00,
		0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03,
		0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75,
		0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x7d, 0xa9, 0x04, 0x00, 0x23,
		0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00,
		0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03,
		0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0x23, 0x75, 0x03, 0x00, 0xdf, 0xf8, 0x0c, 0xd0, 0x4c, 0xf0,
		0xac, 0xf8, 0x00, 0x48, 0x00, 0x47, 0xb7, 0x84, 0x00, 0x00, 0x18, 0x7d, 0x02, 0x20, 0x04, 0x20, 0x71, 0x46, 0x08, 0x42, 0x02, 0xd0, 0xef,
		0xf3, 0x09, 0x80, 0x01, 0xe0, 0xef, 0xf3, 0x08, 0x80, 0x71, 0x46, 0x00, 0x4a, 0x10, 0x47, 0x49, 0x88, 0x03, 0x00, 0x2c, 0x4b, 0x19, 0x68,
		0x08, 0x68, 0xb0, 0xe8, 0xf0, 0x4f, 0x80, 0xf3, 0x09, 0x88, 0xbf, 0xf3, 0x6f, 0x8f, 0x4f, 0xf0, 0x00, 0x00, 0x80, 0xf3, 0x11, 0x88, 0x70,
		0x47, 0x00, 0x00, 0x00, 0x00, 0x08, 0x48, 0x00, 0x68, 0x00, 0x68, 0x80, 0xf3, 0x08, 0x88, 0x4f, 0xf0, 0x00, 0x00, 0x80, 0xf3, 0x14, 0x88,
		0x62, 0xb6, 0x61, 0xb6, 0xbf, 0xf3, 0x4f, 0x8f, 0xbf, 0xf3, 0x6f, 0x8f, 0x00, 0xdf, 0x00, 0xbf, 0x00, 0xbf, 0x08, 0xed, 0x00, 0xe0, 0xdf,
		0xf8, 0x0c, 0x00, 0x01, 0x68, 0x41, 0xf4, 0x70, 0x01, 0x01, 0x60, 0x70, 0x47, 0x00, 0xbf, 0x88, 0xed, 0x00, 0xe0, 0xef, 0xf3, 0x09, 0x80,
		0xbf, 0xf3, 0x6f, 0x8f, 0x13, 0x4b, 0x1a, 0x68, 0x1e, 0xf0, 0x10, 0x0f, 0x08, 0xbf, 0x20, 0xed, 0x10, 0x8a, 0x20, 0xe9, 0xf0, 0x4f, 0x10,
		0x60, 0x09, 0xb4, 0x4f, 0xf0, 0x50, 0x00, 0x80, 0xf3, 0x11, 0x88, 0xbf, 0xf3, 0x4f, 0x8f, 0xbf, 0xf3, 0x6f, 0x8f, 0x33, 0xf0, 0xeb, 0xf9,
		0x4f, 0xf0, 0x00, 0x00, 0x80, 0xf3, 0x11, 0x88, 0x09, 0xbc, 0x19, 0x68, 0x08, 0x68, 0xb0, 0xe8, 0xf0, 0x4f, 0x1e, 0xf0, 0x10, 0x0f, 0x08,
		0xbf, 0xb0, 0xec, 0x10, 0x8a, 0x80, 0xf3, 0x09, 0x88, 0xbf, 0xf3, 0x6f, 0x8f, 0x70, 0x47, 0x08, 0x27, 0x00, 0x20, 0xef, 0xf3, 0x05, 0x80,
		0x70, 0x47, 0x00, 0x00, 0x2d, 0xe9, 0xf0, 0x5f, 0x59, 0x20, 0x16, 0x21, 0x88, 0x22, 0x84, 0x07, 0xc4, 0xf8, 0x00, 0x01, 0xc4, 0xf8, 0x00,
		0x11, 0xc4, 0xf8, 0x00, 0x21, 0xd4, 0xf8, 0x00, 0x31, 0x00, 0x2b, 0xf5, 0xd0, 0x4f, 0xf0, 0x40, 0x20, 0xd0, 0xf8, 0x40, 0x11, 0x21, 0xf0,
		0xf0, 0x01, 0xc0, 0xf8, 0x40, 0x11, 0x01, 0x20, 0x2f, 0xf0, 0xf5, 0xfb, 0x02, 0x20, 0x2f, 0xf0, 0xf2, 0xfb, 0x01, 0x20, 0x2f, 0xf0, 0x82,
		0xfa, 0xfe, 0x48, 0x2f, 0xf0, 0xa2, 0xfb, 0x00, 0x21, 0x02, 0x20, 0x2f, 0xf0, 0x8e, 0xfa, 0x11, 0x20, 0xc4, 0xf8, 0x34, 0x02, 0xfa, 0x4d,
		0x28, 0x46, 0x2f, 0xf0, 0xb1, 0xf9, 0xf9, 0x4e, 0x30, 0x46, 0x2f, 0xf0, 0xad, 0xf9, 0xf8, 0x4f, 0x38, 0x46, 0x2f, 0xf0, 0xa9, 0xf9, 0xf7,
		0x48, 0x2f, 0xf0, 0xa6, 0xf9, 0xdf, 0xf8, 0xd8, 0x83, 0x40, 0x46, 0x2f, 0xf0, 0xa1, 0xf9, 0xdf, 0xf8, 0xd4, 0x93, 0x48, 0x46, 0x2f, 0xf0,
		0x9c, 0xf9, 0xdf, 0xf8, 0xcc, 0xa3, 0x50, 0x46, 0x2f, 0xf0, 0x97, 0xf9, 0xdf, 0xf8, 0xc8, 0xb3, 0x58, 0x46, 0x2f, 0xf0, 0x92, 0xf9, 0xf0,
		0x48, 0x2f, 0xf0, 0x8f, 0xf9, 0xf0, 0x48, 0x2f, 0xf0, 0x8c, 0xf9, 0x01, 0x20, 0x2f, 0xf0, 0x89, 0xf9, 0xee, 0x48, 0x2f, 0xf0, 0x86, 0xf9,
		0xed, 0x48, 0x2f, 0xf0, 0x83, 0xf9, 0xed, 0x48, 0x2f, 0xf0, 0x80, 0xf9, 0x0e, 0x20, 0x2f, 0xf0, 0x7d, 0xf9, 0xeb, 0x48, 0x2f, 0xf0, 0x7a,
		0xf9, 0x4f, 0xf0, 0xac, 0x40, 0x2f, 0xf0, 0x76, 0xf9, 0x10, 0x20, 0x2f, 0xf0, 0x73, 0xf9, 0x00, 0x22, 0x11, 0x46, 0x28, 0x46, 0x2f, 0xf0,
		0x26, 0xf9, 0x00, 0x22, 0x11, 0x46, 0xdd, 0x48, 0x2f, 0xf0, 0x21, 0xf9, 0x00, 0x22, 0x11, 0x46, 0x58, 0x46, 0x2f, 0xf0, 0x1c, 0xf9, 0x00,
		0x22, 0x11, 0x46, 0x38, 0x46,
	}
	b := MD5(data)
	fmt.Printf("%X\n", b)
}

func TestMD5Chunck(t *testing.T) {
	MD5Chunck("remote_TOF22000003_0.0.10.bin")
}
