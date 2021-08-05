package stopreason

import (
	"context"
	"fmt"
	"time"

	"github.com/Kotodian/gokit/boot"
	"github.com/Kotodian/gokit/datasource/grpc"
	"github.com/Kotodian/protocol/golang/configmap"
	"google.golang.org/grpc/metadata"
)

var (
	// 停止理由
	stopReason map[string]string
)

func init() {
	stopReason = make(map[string]string)

	boot.RegisterInit("stopreason", bootInit)
}

// Get 获取值
func Get(code int32) (ret string) {
	//defer func() {
	//	fmt.Println("xxxxxxxxxxxx", code, ret)
	//}()
	var ok bool
	if ret, ok = stopReason[fmt.Sprintf("%d", code)]; ok {
		return
	}
	ret = fmt.Sprintf("%d", code)
	return
}

// GetAll 返回所有
func GetAll() map[string]string {
	return stopReason
}

func bootInit() error {
	in := &configmap.QueryReq{
		Tos: "stop_reason",
	}
	out := &configmap.QueryResp{}
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("requestID", fmt.Sprintf("%d", time.Now().Unix())))
	if err := grpc.Invoke(ctx, configmap.ConfigServicesClient.Query, in, out); err != nil {
		return err
	}

	for _, v := range out.Node.Child {
		stopReason[v.Id] = v.Data["reason"]
	}
	//fmt.Println("1111111111", stopReason)
	return nil
}
