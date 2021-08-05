package utils

import "os"

// Env 获取环境变量
func Env(key string, def ...string) string {
	val := os.Getenv(key)
	if len(val) > 0 {
		return val
	}
	if len(def) > 0 {
		return def[0]
	}
	return ""
}
