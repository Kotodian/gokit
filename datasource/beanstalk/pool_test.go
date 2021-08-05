package beanstalk

import (
	"fmt"
	"testing"
	"time"

	bkd "github.com/edwardhey/beanstalk"
)

func Test_Use(t *testing.T) {
	bconn, err := p.Get()
	defer p.Put(bconn)
	if err != nil {
		t.Log("get bconn error ", err)
	}
	bc := bconn.(*bkd.Conn)

	fmt.Printf("pool len:[%d] conn:[%v]\r\n", p.Len(), bconn)
	time.Sleep(16 * time.Second)
	fmt.Println("pool len: ", p.Len())

	tube := &bkd.Tube{bc, "remoteControl"}
	stats, _ := tube.Stats()
	fmt.Println("stats: ", stats)
}
