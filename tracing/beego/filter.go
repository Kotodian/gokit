package beego

import (
	"fmt"
	"runtime"

	"github.com/astaxie/beego"
	beeContext "github.com/astaxie/beego/context"

	//"github.com/cxr29/log"
	"github.com/Kotodian/gokit/boot"
	"github.com/Kotodian/gokit/tracing"
	"github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
)

var (
	tracer opentracing.Tracer
)

func init() {
	boot.RegisterInit("opentracing", bootInit)
}

func bootInit() error {
	var err error
	tracer, _, err = tracing.New("admin")
	if err != nil {
		return err
	}
	opentracing.SetGlobalTracer(tracer)

	beego.BConfig.RecoverFunc = recoverPanic
	return nil
}

func recoverPanic(ctx *beeContext.Context) {
	if d := ctx.Input.GetData("span"); d != nil {
		span := d.(opentracing.Span)
		if span.BaggageItem("opentracing") == "1" {
			span.Finish()
		}
	}
	if err := recover(); err != nil {
		if err == beego.ErrAbort {
			return
		}
		//if !beego.BConfig.RecoverPanic {
		//	//errStr := fmt.Sprint(err)
		//	//packet := raven.NewPacket(errStr, raven.NewException(errors.New(errStr), raven.NewStacktrace(2, 3, nil)), raven.NewHttp(ctx.Request))
		//	//raven.Capture(packet, nil)
		//	ctx.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		//}

		//if beego.BConfig.EnableErrorsShow {
		//  if _, ok := beego.ErrorMaps[fmt.Sprint(err)]; ok {
		//      exception(fmt.Sprint(err), ctx)
		//      return
		//  }
		//}
		var stack string
		log.Error("the request url is ", ctx.Input.URL())
		log.Error("Handler crashed with error", err)
		for i := 1; ; i++ {
			_, file, line, ok := runtime.Caller(i)
			if !ok {
				break
			}
			log.Errorf(fmt.Sprintf("%s:%d", file, line))
			stack = stack + fmt.Sprintln(fmt.Sprintf("%s:%d", file, line))
		}
		//if beego.BConfig.RunMode == beego.DEV && beego.BConfig.EnableErrorsRender {
		//	showErr(err, ctx, stack)
		//}
	}
}

func BeforeRouterFilter(ctx *beeContext.Context) {
	//fmt.Println("tracing start")
	method := ctx.Request.Method
	url := ctx.Request.URL.String()
	span := tracing.StartSpanWithHeader(&ctx.Request.Header, fmt.Sprintf("%s:%s", method, url), method, url)
	ctx.Input.SetData("span", span)
	// span.LogFields(
	// log.String("ip", ctx.Input.IP()),
	// )
	// ctx.Input.SetData("rawCtx", newCtx)
	//defer span.Finish()

	// requestID := ctx.Input.Header("requestID")
	// if requestID == "" {
	// 	requestID = (<-UUID).String()
	// }
	// // ctx.Input.SetData("rawCtx", _ctx)
	// span, newCtx := opentracing.StartSpanFromContext(context.WithValue(context.Background(), "beeContext", ctx), requestID)
	// ctx.Input.SetData("rawCtx", newCtx)
	// // span.LogFields(
	// // log.String("event", "DNS start"),
	// // log.String("xxxxxxxxx", "1111111111111111"),
	// // log.String("url", ctx.Request.URL.String()),
	// // )
	// fmt.Println("start", span, ctx.Request.URL, ctx.Input.IsAjax())
}

func AfterExecFilter(ctx *beeContext.Context) {
	// fmt.Println("tracing end")
	// if ctx.Output.Status
	if d := ctx.Input.GetData("span"); d != nil {
		span := d.(opentracing.Span)
		if span.BaggageItem("opentracing") == "1" {
			span.Finish()
		}
	}
	// span.SetTag("")
	// .Finish()
}
