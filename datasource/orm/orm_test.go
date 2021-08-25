package orm

import (
	"github.com/Kotodian/gokit/datasource"
	"github.com/Kotodian/protocol/golang/coregw"
	"gorm.io/gorm"
	"os"
	"testing"
	"time"
)

type GormModel struct {
	Id      datasource.UUID `gorm:"column:id;not null;primary_key;" json:"id"`          //ID
	Version int             `gorm:"column:version;default:0;" json:"version"`           // 乐观锁
	Created time.Time       `gorm:"column:created_at;autoCreateTime" json:"created_at"` // 创建时间
	Updated time.Time       `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"` // 更新时间
	Deleted time.Time       `gorm:"column:deleted_at;default:null;" json:"deleted_at"`  // 删除时间
}
type EquipmentInfo struct {
	GormModel            `json:",inline"`
	EquipmentID          datasource.UUID     `gorm:"column:equipment_id" json:"equipment_id"`                          // 设备id
	SerialNumber         string              `gorm:"column:equipment_sn;type:varchar(20)" json:"equipment_sn"`         // sn号
	Vendor               string              `gorm:"column:vendor;type:varchar(18)" json:"vendor"`                     // 厂商
	Model                string              `gorm:"column:model;type:varchar(18)" json:"model"`                       // 产品系列
	FirmwareVersion      string              `gorm:"column:firmware_version;type:varchar(16)" json:"firmware_version"` // 固件版本
	FirmwareExtraVersion string              `gorm:"column:firmware_extra_version" json:"firmware_extra_version"`      // 扩展字段
	Iccid                string              `gorm:"column:iccid;type:varchar(48)" json:"iccid"`                       // iccid
	RemoteAddress        string              `gorm:"column:remote_address;type:varchar(18)" json:"remote_address"`     // 设备ip
	BtMacAddr            string              `gorm:"column:bt_mac_addr;type:varchar(24)" json:"bt_mac_addr"`           // 蓝牙mac地址
	BtPassword           string              `gorm:"column:bt_password;type:varchar(32)" json:"bt_password"`           // 蓝牙密码
	EvseNumber           uint                `gorm:"column:evse_number" json:"evse_number"`                            // 设备数量
	State                int                 `gorm:"column:state;type:tinyint" json:"state"`                           // 状态
	Blocked              bool                `gorm:"column:blocked" json:"blocked"`                                    // 是否屏蔽
	AlarmNums            uint                `gorm:"column:alarm_nums" json:"alarm_nums"`                              //告警数量
	KindVehicleType      *coregw.VehicleType `gorm:"column:vehicle_type" json:"vehicle_type"`                          // 使用车型
	AccessPod            string              `gorm:"column:access_pod" json:"access_pod"`                              // 连接的pod
}

func (e EquipmentInfo) TableName() string {
	return "base_equipment_extra"
}

// 缓存的键
func (e *EquipmentInfo) Key() string {
	return e.SerialNumber
}

// 钩子函数
func (e *EquipmentInfo) AfterCreate(db *gorm.DB) error {
	return nil
}

func (e *EquipmentInfo) AfterUpdate(db *gorm.DB) error {
	return nil
}
func (e *EquipmentInfo) AfterFind(db *gorm.DB) error {
	return nil
}

func (m *GormModel) ID() datasource.UUID {
	return m.Id
}

func (m *GormModel) CreatedAt() int64 {
	return m.Created.Unix()
}

func (m *GormModel) UpdatedAt() int64 {
	return m.Updated.Unix()
}

func (m *GormModel) DeletedAt() int64 {
	return m.Deleted.Unix()
}

func (m *GormModel) GetVersion() int {
	return m.Version
}

func (m *GormModel) SetVersion(version int) {
	m.Version = version
}

func (m *GormModel) Exists() bool {
	return m.Id != 0
}

func TestUpdateWithOptimistic(t *testing.T) {
	os.Setenv("DB_USER", "root")
	os.Setenv("DB_PASSWD", "jqcsms@uat123")
	os.Setenv("DB_HOST", "192.168.0.4")
	os.Setenv("DB_PORT", "3306")

	InitMysqlWithEnvAndDB("jx-csms")
	SetDB(mysqlDB)
	updates := make(map[string]interface{})
	updates["firmware_version"] = "1.4.3"
	updates["access_pod"] = "jx-ocpp-cluster-1"
	updates["remote_address"] = "localhost:8080"
	var obj = new(EquipmentInfo)
	err := GetByID(obj, datasource.UUID(195114960216133))
	if err != nil {
		t.Error(err)
		return
	}
	err = UpdateWithOptimistic(obj, updates)
	if err != nil {
		t.Error(err)
		return
	}

}
