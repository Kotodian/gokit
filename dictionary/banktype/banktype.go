package banktype

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Kotodian/gokit/boot"
	"github.com/Kotodian/gokit/datasource/grpc"
	"github.com/Kotodian/gokit/utils"
	"github.com/Kotodian/protocol/golang/configmap"
	"google.golang.org/grpc/metadata"
)

// BankType 银行
type BankType struct {
	BankID   string `json:"bankId"`
	BankName string `json:"bankName"`
	BankCode string `json:"bankCode"`
}

var (
	tos       = "bankCode"
	lock      sync.Mutex
	banktypes map[string]*BankType
)

func init() {
	banktypes = make(map[string]*BankType)
	boot.RegisterInit("banktypes", bootInit)
}

// Get 获取值
func Get(bankID string) *BankType {
	if v, ok := banktypes[bankID]; ok {
		return v
	}
	return nil
}

// GetAll 返回所有
func GetAll() map[string]*BankType {
	return banktypes
}

func bootInit() error {
	in := &configmap.QueryReq{
		Tos: tos,
	}
	out := &configmap.QueryResp{}
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("requestID", fmt.Sprintf("%d", time.Now().Unix())))
	if err := grpc.Invoke(ctx, configmap.ConfigServicesClient.Query, in, out); err != nil {
		return err
	}
	lock.Lock()
	for _, v := range out.Node.Child {
		bankType := &BankType{
			BankID:   v.Data["bankId"],
			BankName: v.Data["bankName"],
			BankCode: v.Data["bankCode"],
		}
		banktypes[bankType.BankID] = bankType
	}
	lock.Unlock()
	return nil
}

// Add 添加BankType
func Add(bankType *BankType) error {
	if len(bankType.BankID) <= 0 {
		bankType.BankID = utils.MD5(bankType.BankCode + bankType.BankName)
	}
	in := &configmap.AddReq{
		Tos: tos,
		Node: &configmap.Node{
			Id: bankType.BankID,
			Data: map[string]string{
				"bankId":   bankType.BankID,
				"bankName": bankType.BankName,
				"bankCode": bankType.BankCode,
			},
		},
	}
	out := &configmap.AddResp{}
	ctx := grpc.NewContext(context.Background(), "requestID", fmt.Sprintf("%d", time.Now().Unix()))
	if err := grpc.Invoke(ctx, configmap.ConfigServicesClient.Add, in, out); err != nil {
		return err
	}
	lock.Lock()
	banktypes[bankType.BankID] = bankType
	lock.Unlock()
	return nil
}

// Del 删除一项数据
func Del(bankType *BankType) error {
	if _, ok := banktypes[bankType.BankID]; !ok {
		return errors.New("不存在")
	}
	in := &configmap.DelReq{
		Tos: tos,
		Id:  bankType.BankID,
	}
	out := &configmap.DelResp{}
	ctx := grpc.NewContext(context.Background(), "requestID", fmt.Sprintf("%d", time.Now().Unix()))
	if err := grpc.Invoke(ctx, configmap.ConfigServicesClient.Del, in, out); err != nil {
		return err
	}
	lock.Lock()
	delete(banktypes, bankType.BankID)
	lock.Unlock()
	return nil
}
