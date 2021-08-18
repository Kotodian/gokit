package flux

import (
	"fmt"
	"testing"
)

func TestBuilder(t *testing.T) {
	builder := New()
	query := builder.Bucket("example-bucket").
		Range("-1h", "-10m").
		AddFilter("_measurement", Equal, "cpu").
		AddFilter("_field", Equal, "usage_system").
		AddFilter("cpu", Equal, "cpu-total").Build()
	fmt.Println(query)
}
