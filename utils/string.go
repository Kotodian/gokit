package utils

import "strconv"

// StrToInt 字符串转int
func StrToInt(key string, def ...int) int {
	num, err := strconv.Atoi(key)
	if err == nil {
		return num
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}
