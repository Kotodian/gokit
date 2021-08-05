package utils

import "reflect"

// CloneValue ...
func CloneValue(source interface{}, dest interface{}) {
	x := reflect.ValueOf(source)
	//fmt.Println("source", source)
	if x.Kind() == reflect.Ptr {
		starX := x.Elem()
		y := reflect.New(starX.Type())
		starY := y.Elem()
		starY.Set(starX)
		reflect.ValueOf(dest).Elem().Set(y.Elem())
	} else {
		reflect.ValueOf(dest).Elem().Set(x)
	}
}
