package grpc

import (
	"context"
	"runtime"

	"reflect"
	"strings"

	"github.com/Kotodian/gokit/datasource/grpc/client"
	"github.com/Kotodian/gokit/tracing"
	"github.com/gogo/protobuf/proto"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	grpcLib "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func Invoke(ctx context.Context, method interface{}, in proto.Message, out proto.Message, opts ...grpcLib.CallOption) (err error) {
	var conn *grpcLib.ClientConn
	conn, err = client.Conn()
	if err != nil {
		return err
	}
	defer client.Close(conn)

	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	} else {
		md = md.Copy()
	}

	path := runtime.FuncForPC(reflect.ValueOf(method).Pointer()).Name()
	lastDot := strings.LastIndex(path, ".")

	methodString := path[strings.LastIndex(path, "/"):lastDot-6] + "/" + path[lastDot+1:]
	methodArr := strings.Split(methodString, ".")
	methodArr[0] = strings.Replace(methodArr[0], "-", "", -1)
	methodString = strings.Join(methodArr, ".")
	{ //opentracing
		parentSpan := opentracing.SpanFromContext(ctx)
		if parentSpan != nil {
			parentCtx := parentSpan.Context()
			span := opentracing.StartSpan(
				methodString,
				opentracing.ChildOf(parentCtx),
				opentracing.Tag{Key: string(ext.Component), Value: "gRPC"},
				ext.SpanKindRPCClient,
			)
			if span.BaggageItem("opentracing") != "1" {
				goto CALL
			}
			defer func() {
				if err != nil {
					span.SetTag(string(ext.Error), true)
					span.LogFields(log.Error(err))
				}
				span.Finish()
			}()

			mdWriter := tracing.MDReaderWriter{md}
			err := opentracing.GlobalTracer().Inject(span.Context(), opentracing.TextMap, mdWriter)
			if err != nil {
				span.LogFields(log.String("inject-error", err.Error()))
			}
			// ctx = opentracing.ContextWithSpan(ctx, span)
		}
	}
CALL:
	//fmt.Println(methodString)
	ctx = metadata.NewOutgoingContext(ctx, md)
	err = grpcLib.Invoke(ctx, methodString, in, out, conn, opts...)
	// fmt.Println("invoke....................................", ctx, methodString, err)
	return
}

func NewContext(ctx context.Context, kv ...string) context.Context {
	return metadata.NewOutgoingContext(ctx, metadata.Pairs(
		kv...,
	))
}
