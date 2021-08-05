package feed

import (
	"time"

	"github.com/Kotodian/gokit/datasource"
	"google.golang.org/grpc/status"
)

type Kind int

const (
	KindCharge Kind = 0
	KindEvse   Kind = 1
)

type TemplateFunc func(queue *Queue) (string, error)

var templates map[Kind]TemplateFunc

func init() {
	templates = make(map[Kind]TemplateFunc)
}

func RegisterTempalte(key Kind, fn TemplateFunc) {
	templates[key] = fn
}

func GetTemplate(key Kind) (fn TemplateFunc, ok bool) {
	fn, ok = templates[key]
	return
}

type Queue struct {
	EvseID      *datasource.UUID       `json:"evse_id,omitempty,string"`      //设备ID
	StationID   *datasource.UUID       `json:"station_id,omitempty,string"`   //站点ID
	OperatorID  *datasource.UUID       `json:"operator_id,omitempty,string"`  //运营商ID
	ManagerID   *datasource.UUID       `json:"manager_id,omitempty,string"`   //管理员ID
	OrderID     *datasource.UUID       `json:"order_id,omitempty,string"`     //订单ID
	ConnectorID *datasource.UUID       `json:"connector_id,omitemtpy,string"` //枪ID
	Kind        Kind                   `json:"kind"`                          //类型
	Payload     map[string]interface{} `json:"payload"`                       //数据载体
	Master      string                 `json:"master"`                        //主人
	Time        time.Time              `json:"time"`                          //时间
	Msg         string                 `json:"msg"`                           //消息
	Error       *string                `json:"error,omitempty"`               //错误信息
}

func FormatError(err error) *string {
	var errMsg string
	if err == nil {
		return nil
	} else if rpcErr, ok := status.FromError(err); ok {
		errMsg = rpcErr.Message()
		return &errMsg
	}
	errMsg = err.Error()
	return &errMsg
}
