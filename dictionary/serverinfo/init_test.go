package serverinfo

import (
	"fmt"
	"testing"
)

func Test_Info(t *testing.T) {
	fmt.Println("---->1", GetGRPCListenAddr("admin"))
	fmt.Println("---->2", GetHTTPListenAddr("admin"))
	fmt.Println("---->3", GetHTTPAddr("api"))
	fmt.Println("---->4", GetGRPCAddr("api"))
}
