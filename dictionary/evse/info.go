package evse

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Kotodian/gokit/boot"
	"github.com/Kotodian/gokit/datasource/grpc"
	"github.com/Kotodian/protocol/golang/configmap"
	libgrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

var (
	infos map[string][]string
)

func init() {
	infos = make(map[string][]string)
	boot.RegisterInit("evse_info", bootInit)
}

// Get 获取值
func Get(code string) []string {
	if v, ok := infos[code]; ok {
		return v
	}
	return nil
}

// GetAll 返回所有
func GetAll() map[string][]string {
	return infos
}

func bootInit() error {
	in := &configmap.QueryReq{
		Tos: "evse_info",
	}
	out := &configmap.QueryResp{}
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("requestID", fmt.Sprintf("%d", time.Now().Unix())))
	if err := grpc.Invoke(ctx, configmap.ConfigServicesClient.Query, in, out); err != nil {
		if codes.NotFound == libgrpc.Code(err) {
			return nil
		}
		return err
	}

	for _, v := range out.Node.Child {
		infos[v.Id] = strings.Split(v.Data["items"], ",")
	}
	return nil
}
