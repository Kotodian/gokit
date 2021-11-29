package redis

import (
	"encoding/json"
	"github.com/Kotodian/gokit/datasource"
	"github.com/Kotodian/protocol/golang/keys"
	"github.com/gomodule/redigo/redis"
	"os"
	"testing"
	"time"
)

type GormModel struct {
	Id        datasource.UUID `gorm:"column:id;not null;primary_key;" json:"id"` //ID
	Version   int             `gorm:"column:version;default:0;" json:"-"`        // 乐观锁
	Created   time.Time       `gorm:"column:created_at;autoCreateTime" json:"-"` // 创建时间
	CreatedID datasource.UUID `gorm:"column:created_by" json:"-"`
	Updated   time.Time       `gorm:"column:updated_at;autoUpdateTime" json:"-"` // 更新时间
	UpdatedID datasource.UUID `gorm:"column:updated_by" json:"-"`
}
type Equipment struct {
	GormModel           `json:",inline"` // 基础元素
	EquipmentInfo       *EquipmentInfo   `gorm:"-"`
	SerialNumber        string           `gorm:"column:sn;type:varchar(20)" json:"sn"`                              // sn号
	Product             string           `gorm:"column:product;type:varchar(18)" json:"product"`                    // 产品型号
	OperatorId          datasource.UUID  `gorm:"column:operator_id;not null" json:"operator_id"`                    // 运营商
	StationId           datasource.UUID  `gorm:"column:station_id;not null" json:"station_id"`                      // 站点
	Password            string           `gorm:"column:account_password;type:varchar(50)" json:"account_password"`  // 设备password
	AccessPorts         string           `gorm:"column:access_ports;type:varchar(50);not null" json:"access_ports"` // 可访问端口
	Enabled             bool             `gorm:"column:enabled" json:"enabled"`                                     // 是否允许接入
	PowerZoneID         *datasource.UUID `gorm:"column:power_zone_id" json:"power_zone_id"`                         // 私桩电力台区
	Protocol            string           `gorm:"column:protocol;type:varchar(20)" json:"protocol"`                  // 协议
	ProtocolVersion     string           `gorm:"column:protocol_version;type:varchar(20)" json:"protocol_version"`  // 协议版本
	EquipmentType       string           `gorm:"column:category;type:varchar(20)" json:"category"`                  // 公桩还是私桩
	AccountVerification bool             `gorm:"column:account_verification" json:"account_verification"`           // 是否校验密钥
	KeepAlive           int              `gorm:"column:keepalive" json:"keepalive"`                                 // 心跳时间
	ProductionDate      *time.Time       `gorm:"column:production_date" json:"production_date"`
}
type EquipmentInfo struct {
	GormModel
	EquipmentID          datasource.UUID `gorm:"column:equipment_id" json:"-"`                                     // 设备id
	SerialNumber         string          `gorm:"column:equipment_sn;type:varchar(20)" json:"equipment_sn"`         // sn号
	ModelID              datasource.UUID `gorm:"column:model_id;type:varchar(18)" json:"model_id"`                 // 产品系列
	FirmwareVersion      string          `gorm:"column:firmware_version;type:varchar(16)" json:"firmware_version"` // 固件版本
	FirmwareExtraVersion string          `gorm:"column:firmware_extra_version" json:"firmware_extra_version"`      // 扩展字段
	Iccid                string          `gorm:"column:iccid;type:varchar(48)" json:"iccid"`                       // iccid
	RemoteAddress        string          `gorm:"column:remote_address;type:varchar(18)" json:"remote_address"`     // 设备ip
	BtMacAddr            string          `gorm:"column:bt_mac_addr;type:varchar(24)" json:"bt_mac_addr"`           // 蓝牙mac地址
	BtPassword           string          `gorm:"column:bt_password;type:varchar(32)" json:"bt_password"`           // 蓝牙密码
	EvseNumber           uint            `gorm:"column:evse_number" json:"evse_number"`                            // 设备数量
	State                bool            `gorm:"column:state;type:tinyint" json:"state"`                           // 状态
	Blocked              bool            `gorm:"column:blocked" json:"blocked"`                                    // 是否屏蔽
	AlarmNumber          uint            `gorm:"column:alarm_number" json:"alarm_nums"`                            //告警数量
	AccessPod            string          `gorm:"column:access_pod" json:"access_pod"`                              // 连接的pod
	ConnectorNumber      uint            `gorm:"-" json:"-"`                                                       // 枪数量
	LastLostConnReason   *string         `gorm:"column:last_lost_conn_reason" json:"last_lost_conn_reason"`        // 上次连接断开原因
	LastLostConnTime     *time.Time      `gorm:"column:last_lost_conn_time" json:"last_lost_conn_time"`            // 上次连接断开时间
	RegTime              *time.Time      `gorm:"column:register_datetime" json:"register_datetime"`
	ConnTime             *time.Time      `gorm:"column:conn_time" json:"conn_time"`
}

func TestInit(t *testing.T) {
	os.Setenv("REDIS_POOL", "10.43.0.20:6379")
	os.Setenv("REDIS_AUTH", "LhBIOQumQdgIm4ro")
	Init()
	redisConn := GetRedis()
	defer redisConn.Close()
	equipment := new(Equipment)
	//var equipment *Equipment

	bytes, err := redis.Bytes(redisConn.Do("get", keys.Equipment("586069658769111")))
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(bytes, equipment)
	if err != nil {
		panic(err)
	}

	t.Log(equipment.EquipmentInfo)
}
