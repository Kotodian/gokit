package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/Kotodian/gokit/datasource/grpc"
	"github.com/Kotodian/gokit/exerr"
	"github.com/gogo/protobuf/proto"
	gorpc "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// NewContextWithTimeout xxx
func NewContextWithTimeout(requestID string, clientID string, timeout time.Duration) (context.Context, context.CancelFunc) {
	md := metadata.New(map[string]string{
		"requestid": requestID,
		"clientid":  clientID,
	})
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	ctx = metadata.NewOutgoingContext(ctx, md)
	//set timeout
	return ctx, cancel
}

// CodeRPC 调用
func CodeRPC(in interface{}, out interface{}, fn interface{}) *exerr.Error {
	// requestid 要随机生成
	staticMD := metadata.Pairs("requestid", fmt.Sprintf("%d", Now()))
	ctx := metadata.NewOutgoingContext(context.Background(), staticMD)
	inMsg := in.(proto.Message)
	outMsg := out.(proto.Message)
	err := grpc.Invoke(ctx, fn, inMsg, outMsg)
	if err != nil {
		code := gorpc.Code(err)
		return exerr.Msg(gorpc.ErrorDesc(err), int(code))
	}
	return nil
}
