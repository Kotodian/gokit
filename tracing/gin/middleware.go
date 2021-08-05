package gin

import (
	"fmt"

	"github.com/Kotodian/gokit/tracing"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go/ext"
)

func Tracing(c *gin.Context) {
	method := c.Request.Method
	url := c.Request.URL.String()
	// operatorName, ok := c.Get("tracingOperator")
	// if ok {
	operatorName := fmt.Sprintf("%s:%s", method, url)
	// }
	span := tracing.StartSpanWithHeader(&c.Request.Header, operatorName, method, url)
	c.Set("span", span)
	if span.BaggageItem("opentracing") == "1" {
		// span.LogFields(log.Object("header", c.Request.Header))
		// span.LogFields(log.Object("params", c.Params))
		defer span.Finish()
	}
	c.Next()
	if len(c.Errors) > 0 {
		span.SetTag(string(ext.Error), true)
	}
}
