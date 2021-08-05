package tracing

import (
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"

	// "github.com/cxr29/log"
	// "github.com/cxr29/log"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	log "github.com/opentracing/opentracing-go/log"
	jaeger "github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"google.golang.org/grpc/metadata"
)

var cfg config.Configuration

func init() {
	cfg = config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort:  "jaeger:5775", // localhost:5775
		},
	}
}

func New(serviceName string) (opentracing.Tracer, io.Closer, error) {
	return cfg.New(
		serviceName,
		config.Logger(jaeger.NullLogger),
		// config.Logger(jaeger.StdLogger),
	)
	// if err != nil {
	// log.Fatal(err)
	// }

	// return tracer, closer, err
}

// StartSpanWithParent will start a new span with a parent span.
// example:
//      span:= StartSpanWithParent(c.Get("tracing-context"),
func StartSpanWithParent(parent opentracing.SpanContext, operationName, method, path string) opentracing.Span {
	options := []opentracing.StartSpanOption{
		opentracing.Tag{Key: ext.SpanKindRPCServer.Key, Value: ext.SpanKindRPCServer.Value},
		opentracing.Tag{Key: string(ext.HTTPMethod), Value: method},
		opentracing.Tag{Key: string(ext.HTTPUrl), Value: path},
		opentracing.Tag{Key: "current-goroutines", Value: runtime.NumGoroutine()},
	}

	if parent != nil {
		options = append(options, opentracing.ChildOf(parent))
	}

	return opentracing.StartSpan(operationName, options...)
}

// StartSpanWithHeader will look in the headers to look for a parent span before starting the new span.
// example:
//  func handleGet(c *gin.Context) {
//     span := StartSpanWithHeader(&c.Request.Header, "api-request", method, path)
//     defer span.Finish()
//     c.Set("tracing-context", span) // add the span to the context so it can be used for the duration of the request.
//     bosePersonID := c.Param("bosePersonID")
//     span.SetTag("bosePersonID", bosePersonID)
//
func StartSpanWithHeader(header *http.Header, operationName, method, path string) opentracing.Span {
	var wireContext opentracing.SpanContext
	if header != nil {
		wireContext, _ = opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(*header))
	}
	return StartSpanWithParent(wireContext, operationName, method, path)
	// return span
	// return StartSpanWithParent(wireContext, operationName, method, path)
}

// InjectTraceID injects the span ID into the provided HTTP header object, so that the
// current span will be propogated downstream to the server responding to an HTTP request.
// Specifying the span ID in this way will allow the tracing system to connect spans
// between servers.
//
//  Usage:
//          // resty example
// 	    r := resty.R()
//	    injectTraceID(span, r.Header)
//	    resp, err := r.Get(fmt.Sprintf("http://localhost:8000/users/%s", bosePersonID))
//
//          // galapagos_clients example
//          c := galapagos_clients.GetHTTPClient()
//          req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:8000/users/%s", bosePersonID))
//          injectTraceID(span, req.Header)
//          c.Do(req)
func InjectTraceID(ctx opentracing.SpanContext, header http.Header) {
	opentracing.GlobalTracer().Inject(
		ctx,
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(header))
}

func Error(span opentracing.Span, err error) {
	span.SetTag(string(ext.Error), true)
	span.LogFields(
		log.Error(err),
	)
}

//MDReaderWriter metadata Reader and Writer
type MDReaderWriter struct {
	metadata.MD
}

// ForeachKey implements ForeachKey of opentracing.TextMapReader
func (c MDReaderWriter) ForeachKey(handler func(key, val string) error) error {
	for k, vs := range c.MD {
		for _, v := range vs {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

// Set implements Set() of opentracing.TextMapWriter
func (c MDReaderWriter) Set(key, val string) {
	key = strings.ToLower(key)
	c.MD[key] = append(c.MD[key], val)
}
