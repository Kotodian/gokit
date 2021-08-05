package lib

import (
	"context"

	"github.com/golang/protobuf/proto"

	//"gitcsms.joysonquin.com/csms/ac-ocpp/ws/protocol"
	pCharger "github.com/Kotodian/protocol/golang/hardware/charger"
)

type TRData struct {
	Ignore      bool                   //忽略
	Retained    bool                   //MQTT的遗留信息
	Topic       string                 //MQTT的标题
	Data        map[string]interface{} //数据域
	IsTelemetry bool                   //是否遥测数据，如果是遥测数据可以不理会回复的报文
	APDU        *pCharger.APDU         //平台的报文
	ActionName  string                 //下发给设备的Action名称
	IsError     bool                   //是否错误
}

type ITranslate interface {
	//ToAPDU 发送到core的protobuf的消息，包括设备主动发送以及设备回复的
	ToAPDU(ctx context.Context, msg []byte) (to proto.Message, err error)

	//FromAPDU 发送、回复给设备的消息
	FromAPDU(ctx context.Context, apdu *pCharger.APDU) (to interface{}, err error)
}
