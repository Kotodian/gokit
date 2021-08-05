package backendauth

import "context"

var authkey, authval struct{}

func Auth(ctx context.Context) context.Context {
	return context.WithValue(ctx, authkey, authval)
}

func Check(ctx context.Context) bool {
	if v := ctx.Value(authkey); v == authval {
		return true
	}
	return false
}
