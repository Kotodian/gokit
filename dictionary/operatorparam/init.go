package operatorparam

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Kotodian/gokit/boot"
	"github.com/Kotodian/gokit/datasource/grpc"
	"github.com/Kotodian/protocol/golang/configmap"
	libgrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

var (
	operatorparams map[string]string
)

func init() {
	operatorparams = make(map[string]string)
	boot.RegisterInit("operatorparams", bootInit)
}

// Get 获取值
func Get(k string) string {
	if v, ok := operatorparams[k]; ok {
		return v
	}
	return "-"
}

// 是否禁用提现
func GetIsDisableWithdraw() (bool, error) {
	v, ok := operatorparams["is_disable_withdraw"]
	if !ok {
		// return false, errors.New("禁用提现配置未找到")
		return false, nil
	}

	return strconv.ParseBool(v)
}

// 是否禁用账单
func GetIsDisableRoyalty() (bool, error) {
	v, ok := operatorparams["is_disable_royalty"]
	if !ok {
		// return false, errors.New("禁用账单配置未找到")
		return false, nil
	}

	return strconv.ParseBool(v)
}

// GetAll 返回所有
func GetAll() map[string]string {
	return operatorparams
}

// Set 设置运营参数
func Set(ctx context.Context, kvs map[string]interface{}) error {
	in := &configmap.AddReq{
		Tos:  "system-seting",
		Node: &configmap.Node{},
	}
	out := &configmap.AddResp{}

	in.Node.Id = "operate-seting"
	in.Node.Data = make(map[string]string, 0)

	for k, v := range kvs {
		operatorparams[k] = fmt.Sprintf("%v", v)
	}
	in.Node.Data = operatorparams

	if err := grpc.Invoke(ctx, configmap.ConfigServicesServer.Add, in, out); err != nil {
		return err
	}

	return nil
}

func bootInit() error {
	in := &configmap.QueryReq{
		Tos: "system-seting",
		Id:  "operate-seting",
	}

	out := &configmap.QueryResp{}

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("requestID", fmt.Sprintf("%d", time.Now().Unix())))
	if err := grpc.Invoke(ctx, configmap.ConfigServicesClient.Query, in, out); err != nil {
		if libgrpc.Code(err) == codes.NotFound {
			return nil
		}
		return fmt.Errorf("获取运营参数错误: " + err.Error())
	}

	for k, v := range out.Node.Data {
		operatorparams[k] = v
	}
	// fmt.Printf("---------->[%+v]\r\n", operatorparams)

	return nil
}
