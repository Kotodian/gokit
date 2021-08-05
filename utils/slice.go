package utils

import "fmt"

// SliceJoin 合并字符串
func SliceJoin(s []interface{}, sep string) string {
	if len(s) == 0 {
		return ""
	}
	str := ""
	for _, item := range s {
		str = fmt.Sprintf("%s%s%v", str, sep, item)
	}
	return str[1:]
}
