package bankcode

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
	bankcodes map[string]string
	banknames map[string]string
)

func init() {
	bankcodes = make(map[string]string)
	banknames = make(map[string]string)
	boot.RegisterInit("bankcode", bootInit)
}

// GetName 获取银行名
func GetName(code string) string {
	if v, ok := banknames[code]; ok {
		return v
	}
	return "-"
}

// GetCode 获取银行编码
func GetCode(name string) string {
	if v, ok := bankcodes[name]; ok {
		return v
	}
	return "-"
}

// GetAll 返回所有(存在数据库的map格式)
func GetAll() map[string]string {
	return bankcodes
}

func bootInit() error {
	in := &configmap.QueryReq{
		Tos: "bank",
	}
	out := &configmap.QueryResp{}
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("requestID", fmt.Sprintf("%d", time.Now().Unix())))
	if err := grpc.Invoke(ctx, configmap.ConfigServicesClient.Query, in, out); err != nil {
		return err
	}

	for _, v := range out.Node.Child {
		name, code := v.Id, v.Data["code"]

		bankcodes[name] = code
		banknames[code] = name
	}
	return nil
}
