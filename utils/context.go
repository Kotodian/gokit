/*
 * Copyright (c) 2019.
 */

package utils

import (
	"context"
	"time"
)

func ContextWithTimeoutIfNotSet(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return ctx, nil
	}
	return context.WithTimeout(ctx, timeout)
}
