package utils

import (
	"bytes"
	"encoding/json"
)

// JSONString object 转string
func JSONString(obj interface{}) string {
	if obj == nil {
		return ""
	}
	bytes, err := json.Marshal(obj)
	if err != nil {
		return ""
	}
	return string(bytes)
}

// JSONMarshal 转成json 字符串
func JSONMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}
