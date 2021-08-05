package utils

import (
	"errors"
	"reflect"
	"strings"
)

func Contains(obj interface{}, target interface{}) (bool, error) {
	targetValue := reflect.ValueOf(target)
	switch reflect.TypeOf(target).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < targetValue.Len(); i++ {
			if targetValue.Index(i).Interface() == obj {
				return true, nil
			}
		}
		return false, nil
	case reflect.Map:
		if targetValue.MapIndex(reflect.ValueOf(obj)).IsValid() {
			return true, nil
		}
		return false, nil
	case reflect.String:
		return strings.Contains(obj.(string), target.(string)), nil
	}

	return false, errors.New("not support type match")
}
