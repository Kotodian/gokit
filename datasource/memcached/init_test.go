package memcached

import (
	"fmt"
	"testing"
)

type T struct {
	Name string `json:"name"`
}

func Test_Get(t *testing.T) {

	err := Save("12111", 11111, 3000)
	fmt.Println("save err:", err)
	{
		var ret int
		_, err := Get("12111", &ret)
		fmt.Printf("------->[%+v][%+v]\r\n", err, ret)
	}

	// save := map[string]interface{}{
	// 	"count":   3333,
	// 	"balance": 22222,
	// }

	// err := Save("12111", save, 3000)
	// fmt.Println("save err:", err)
	// {
	// 	var ret map[string]interface{}
	// 	err := Get("12111", &ret)
	// 	fmt.Printf("------->[%+v][%+v][%+v]\r\n", err, ret["count"].(json.Number), ret["balance"].(json.Number))
	// }

	// ----------------------
	// err := Save("12111", &T{Name: "1111"}, 3)
	// fmt.Println("save err:", err)
	// time.Sleep(5 * time.Second)
	// {
	// 	_t := &T{}
	// 	err := Get("test1111", _t)
	// 	fmt.Printf("------->[%+v]\r\n", err)
	// }

	// {
	// 	_t := &T{}
	// 	err := Get("12111", _t)
	// 	fmt.Printf("------->[%+v]\r\n", err)
	// }
}
