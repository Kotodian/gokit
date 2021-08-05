package backendauth

import (
	"context"
	"fmt"
	"testing"
)

func Test_Auth(t *testing.T) {
	ctx := context.Background()
	// ctx = Auth(ctx)
	fmt.Printf("----->[%+v]\r\n", ctx)
	fmt.Printf("----->[%+v]\r\n", Check(ctx))

}
