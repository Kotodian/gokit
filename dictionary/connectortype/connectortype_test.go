package connectortype

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Kotodian/gokit/datasource/grpc"
	"github.com/Kotodian/protocol/golang/configmap"
	"google.golang.org/grpc/metadata"
)

func Test_Get(t *testing.T) {
	in := &configmap.QueryReq{
		Tos: "connector_type",
	}
	out := &configmap.QueryResp{}
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("requestID", fmt.Sprintf("%d", time.Now().Unix())))
	if err := grpc.Invoke(ctx, configmap.ConfigServicesClient.Query, in, out); err != nil {
		t.Fatal(err)
	}

	for _, v := range out.Node.Child {
		fmt.Println(v)
	}

	time.Sleep(time.Second)

	if err := grpc.Invoke(ctx, configmap.ConfigServicesClient.Query, in, out); err != nil {
		t.Fatal(err)
	}

	for _, v := range out.Node.Child {
		fmt.Println(v)
	}
}
