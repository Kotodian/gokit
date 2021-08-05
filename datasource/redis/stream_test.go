package redis

import (
	"context"
	"fmt"
)

type user struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func cb(ctx context.Context, model interface{}) error {
	fmt.Printf("model:[%+v]\r\n", model)
	return nil
}

// func Test_stream(t *testing.T) {
// 	sg, _ := NewStreamHandle("stream_test", "group1", "ssy", 10, cb)

// 	done := make(chan struct{})
// 	go sg.Daemon(done, &user{})

// 	for i := 100; ; i++ {
// 		u := &user{
// 			Name: fmt.Sprintf("ssy_%d", i),
// 			Age:  i,
// 		}
// 		if err := PublishStreamWithMaxlen("stream_test", 10, u); err != nil {
// 			fmt.Println(" publish error: ", err)
// 		} else {
// 			fmt.Println(" publish success")
// 		}

// 		time.Sleep(1 * time.Second)
// 	}
// }
