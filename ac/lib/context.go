package lib

import "context"

func GetClientFromCtx(ctx context.Context) ClientInterface {
	return ctx.Value("client").(ClientInterface)
}

func GetTRDataFromCtx(ctx context.Context) *TRData {
	data := ctx.Value("trData").(*TRData)
	if data.Data == nil {
		data.Data = make(map[string]interface{})
	}
	return data
}
