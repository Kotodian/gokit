package context

import (
	"strings"

	"golang.org/x/net/context"
)

type ctxDataInKey struct{}
type CtxData map[string]interface{}

func NewCtxData() CtxData {
	cd := CtxData{}
	return cd
}

func (cd CtxData) Set(k string, val interface{}) {
	k = strings.ToLower(k)
	cd[k] = val
}
func (cd CtxData) Get(k string) interface{} {
	k = strings.ToLower(k)
	return cd[k]
}

func NewLocalContext(ctx context.Context, cd CtxData) context.Context {
	return context.WithValue(ctx, ctxDataInKey{}, cd)
}

func FromLocalContext(ctx context.Context) (cd CtxData, ok bool) {
	cd, ok = ctx.Value(ctxDataInKey{}).(CtxData)
	return
}
