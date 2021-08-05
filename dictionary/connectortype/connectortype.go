package connectortype

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
	connectorTypes map[string]string
)

func init() {
	connectorTypes = make(map[string]string)
	boot.RegisterInit("connectortype", bootInit)
}

// Get 获取值
func Get(code string) string {
	if v, ok := connectorTypes[code]; ok {
		return v
	}
	return "-"
}

// GetAll 返回所有
func GetAll() map[string]string {
	return connectorTypes
}

func bootInit() error {
	in := &configmap.QueryReq{
		Tos: "connector_type",
	}
	out := &configmap.QueryResp{}
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("requestID", fmt.Sprintf("%d", time.Now().Unix())))
	if err := grpc.Invoke(ctx, configmap.ConfigServicesClient.Query, in, out); err != nil {
		return err
	}

	for _, v := range out.Node.Child {
		connectorTypes[v.Id] = v.Data["type"]
	}
	return nil
}
