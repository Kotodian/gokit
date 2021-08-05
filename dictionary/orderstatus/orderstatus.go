package orderstatus

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
	// orderStatus 订单状态
	orderStatus map[string]string
)

func init() {
	orderStatus = make(map[string]string)
	boot.RegisterInit("orderstatus", bootInit)
}

// Get 获取值
func Get(code string) string {
	if v, ok := orderStatus[code]; ok {
		return v
	}
	return "-"
}

func bootInit() error {
	in := &configmap.QueryReq{
		Tos: "order_status",
	}
	out := &configmap.QueryResp{}
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("requestID", fmt.Sprintf("%d", time.Now().Unix())))
	if err := grpc.Invoke(ctx, configmap.ConfigServicesClient.Query, in, out); err != nil {
		return err
	}

	for _, v := range out.Node.Child {
		orderStatus[v.Id] = v.Data["status"]
	}
	return nil
}
