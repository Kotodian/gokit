package interceptor

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/Kotodian/gokit/tracing"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var before InterceptorFunc

func SetInterceptorFunc(fn InterceptorFunc) {
	before = fn
}

type InterceptorFunc func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, resp interface{}) (context.Context, error)

// RPCServerInterceptor GRPC服务拦截器
func RPCServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	spanContext, err := opentracing.GlobalTracer().Extract(opentracing.TextMap, tracing.MDReaderWriter{md})
	if err != nil && err != opentracing.ErrSpanContextNotFound {
		return nil, err
	}
	span := opentracing.StartSpan(
		info.FullMethod,
		ext.RPCServerOption(spanContext),
		opentracing.Tag{Key: string(ext.Component), Value: "gRPC"},
		ext.SpanKindRPCServer,
	)
	if span.BaggageItem("opentracing") == "1" {
		defer func() {
			if err != nil {
				span.SetTag(string(ext.Error), true)
				span.LogFields(log.Error(err))
			}
			span.Finish()
		}()
	}
	outCtx := ctx
	if before != nil {
		if outCtx, err = before(ctx, req, info, resp); err != nil {
			return
		}
	}
	resp, err = handler(opentracing.ContextWithSpan(outCtx, span), req)
	if err != nil {
		logrus.Errorf("req Method:[%v] param:[%+v] err:[%+v]", info.FullMethod, req, err)
	}
	return
}
