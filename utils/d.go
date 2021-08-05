package utils

import (
	"fmt"
)

func D(obj interface{}, label ...string) {
	if obj != nil {
		fmt.Println("\n\n")
		if len(label) > 0 {
			fmt.Println(label[0], fmt.Sprintf("%+v", obj))
		} else {
			fmt.Println(fmt.Sprintf("%+v", obj))
		}
	}
}
