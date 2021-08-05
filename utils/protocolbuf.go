package utils

import (
	"encoding/json"

	"github.com/golang/protobuf/ptypes/any"
)

func EncodeToAny(data interface{}) (a *any.Any, err error) {
	a = new(any.Any)
	if a.Value, err = json.Marshal(data); err != nil {
		return
	}
	return
}

func DecodeFromAny(a *any.Any, data interface{}) error {
	return json.Unmarshal(a.Value, a)
}
