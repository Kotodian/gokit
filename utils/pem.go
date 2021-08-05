package utils

import (
	"bytes"
	"strings"
)

// ParseCert 转成PEM
func ParseCert(raw string) (result []byte) {
	return parseKey(raw, "-----BEGIN CERTIFICATE-----", "-----END CERTIFICATE-----")
}

// ParseKey 转成PEM
func ParseKey(raw string) (result []byte) {
	return parseKey(raw, "-----BEGIN PRIVATE KEY-----", "-----END PRIVATE KEY-----")
}

func parseKey(raw, prefix, suffix string) (result []byte) {
	if raw == "" {
		return nil
	}
	raw = strings.Replace(raw, prefix, "", 1)
	raw = strings.Replace(raw, suffix, "", 1)
	raw = strings.Replace(raw, " ", "", -1)
	raw = strings.Replace(raw, "\n", "", -1)
	raw = strings.Replace(raw, "\r", "", -1)
	raw = strings.Replace(raw, "\t", "", -1)

	var ll = 64
	var sl = len(raw)
	var c = sl / ll
	if sl%ll > 0 {
		c = c + 1
	}

	var buf bytes.Buffer
	buf.WriteString(prefix + "\n")
	for i := 0; i < c; i++ {
		var b = i * ll
		var e = b + ll
		if e > sl {
			buf.WriteString(raw[b:])
		} else {
			buf.WriteString(raw[b:e])
		}
		buf.WriteString("\n")
	}
	buf.WriteString(suffix)
	return buf.Bytes()
}

// ParseCertKey 返回字符串
func ParseCertKey(raw string) string {
	return orginKey(raw, "-----BEGIN CERTIFICATE-----", "-----END CERTIFICATE-----")
}

// ParseKeyKey 返回字符串
func ParseKeyKey(raw string) string {
	return orginKey(raw, "-----BEGIN PRIVATE KEY-----", "-----END PRIVATE KEY-----")
}

func orginKey(raw, prefix, suffix string) string {
	raw = strings.Replace(raw, prefix, "", 1)
	raw = strings.Replace(raw, suffix, "", 1)
	raw = strings.Replace(raw, " ", "", -1)
	raw = strings.Replace(raw, "\n", "", -1)
	raw = strings.Replace(raw, "\r", "", -1)
	raw = strings.Replace(raw, "\t", "", -1)
	return raw
}
