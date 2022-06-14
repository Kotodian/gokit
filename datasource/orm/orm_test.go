package orm

import (
	"context"
	"os"
	"testing"

	"github.com/Kotodian/gokit/datasource"
	"github.com/Kotodian/protocol/interfaces"
	"gorm.io/gorm"
)

type GormModel struct {
	Id        datasource.UUID `gorm:"column:id;not null;primary_key;" json:"id"` //ID
	Version   int             `gorm:"column:version;default:0;" json:"version"`  // 乐观锁
	Created   int64           `gorm:"column:created_at;autoCreateTime" json:"-"` // 创建时间
	CreatedID datasource.UUID `gorm:"column:created_by" json:"-"`
	Updated   int64           `gorm:"column:updated_at;autoUpdateTime" json:"-"` // 更新时间
	UpdatedID datasource.UUID `gorm:"column:updated_by" json:"-"`
}

func (m *GormModel) CreatedAt() int64 {
	return m.Created
}

func (m *GormModel) UpdatedAt() int64 {
	return m.Updated
}

func (m *GormModel) CreatedBy() datasource.UUID {
	return m.CreatedID
}

func (m *GormModel) UpdatedBy() datasource.UUID {
	return m.UpdatedID
}

func (m *GormModel) ID() datasource.UUID {
	return m.Id
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

func (m *GormModel) BeforeCreate(db *gorm.DB) error {
	return nil
}

func (m *GormModel) BeforeUpdate(db *gorm.DB) error {
	return nil
}

func DBUpdates() map[string]interface{} {
	return nil
}
func (m *GormModel) ESUpdates() map[string]interface{} {
	return nil
}

func (m *GormModel) UpdateHook() {
	return
}

type Equipment struct {
	GormModel           `json:",inline"` // 基础元素
	SerialNumber        string           `gorm:"column:sn;type:varchar(20)" json:"sn"`           // sn号
	Product             string           `gorm:"column:product;type:varchar(18)" json:"product"` // 产品型号
	OperatorId          datasource.UUID  `gorm:"column:operator_id;not null" json:"operatorId"`  // 运营商
	OperatorName        string           `gorm:"-" json:"-"`
	StationId           datasource.UUID  `gorm:"column:station_id;not null" json:"stationId"` // 站点
	StationName         string           `gorm:"-" json:"-"`
	Password            string           `gorm:"column:account_password;type:varchar(50)" json:"-"`               // 设备password
	AccessPorts         string           `gorm:"column:access_ports;type:varchar(50);not null" json:"-"`          // 可访问端口
	Enabled             bool             `gorm:"column:enabled" json:"enabled"`                                   // 是否允许接入
	PowerZoneID         *datasource.UUID `gorm:"column:power_zone_id" json:"powerZoneId"`                         // 私桩电力台区
	Protocol            string           `gorm:"column:protocol;type:varchar(20)" json:"protocol"`                // 协议
	ProtocolVersion     string           `gorm:"column:protocol_version;type:varchar(20)" json:"protocolVersion"` // 协议版本
	EquipmentType       string           `gorm:"column:category;type:varchar(20)" json:"category"`                // 公桩还是私桩
	AccountVerification bool             `gorm:"column:account_verification" json:"-"`                            // 是否校验密钥
	KeepAlive           int              `gorm:"column:keepalive" json:"keepalive"`                               // 心跳时间
	ProductionDate      *int64           `gorm:"column:production_date" json:"-"`
}

func (e Equipment) TableName() string {
	return "base_equipment"
}

// Key 缓存的键
func (e *Equipment) Key() string {
	return ""
}

// AfterCreate 钩子函数
func (e *Equipment) AfterCreate(db *gorm.DB) error {
	return nil
}

func (e *Equipment) AfterUpdate(db *gorm.DB) error {
	return nil
}

func (e *Equipment) AfterFind(db *gorm.DB) error {
	return nil
}

type SmallEquipment struct {
	GormModel    `json:",inline"` // 基础元素
	SerialNumber string           `gorm:"column:sn;type:varchar(20)" json:"sn"`
}

func (e SmallEquipment) TableName() string {
	return "base_equipment"
}

// Key 缓存的键
func (e *SmallEquipment) Key() string {
	return ""
}

// AfterCreate 钩子函数
func (e *SmallEquipment) AfterCreate(db *gorm.DB) error {
	return nil
}

func (e *SmallEquipment) AfterUpdate(db *gorm.DB) error {
	return nil
}

func (e *SmallEquipment) AfterFind(db *gorm.DB) error {
	return nil
}

type Connector struct {
	GormModel
	// 需要设置redis的时候必须要赋值
	EvseSerial      string `gorm:"column:evse_serial" json:"evse_serial"`
	EquipmentSerial string `gorm:"column:equipment_sn" json:"equipment_sn"`

	EvseID          datasource.UUID  `gorm:"column:evse_id;not null" json:"evse_id"`           // evse_id
	EquipmentID     datasource.UUID  `gorm:"column:equipment_id;not null" json:"equipment_id"` // equipment_id
	ConnectorSerial string           `gorm:"column:serial;type:varchar(6)" json:"serial"`      // 枪序号
	OrderID         *datasource.UUID `gorm:"column:order_id" json:"order_id"`                  // 当前订单id
	// RecordID           *datasource.UUID                      `gorm:"column:record_id" json:"record_id"`                     // 当前报告记录id
	CurrentState  interfaces.KindConnectorState         `gorm:"column:current_state;default:0" json:"current_state"`   // 枪当前状态
	BeforeState   interfaces.KindConnectorState         `gorm:"column:before_state;default:0" json:"before_state"`     // 之前的状态
	ChargingState interfaces.KindConnectorChargingState `gorm:"column:charging_state;default:0" json:"charging_state"` // 充电状态
	ReservationID *int64                                `gorm:"column:reservation_id" json:"reservation_id"`
	// ConnectorType      int                                   `gorm:"column:category;default:0" json:"category"` // 接口类型
	// VoltageUpperLimits int                                   `gorm:"column:voltage_upper_limits" json:"voltage_upper_limits"`
	// VoltageLowerLimits int                                   `gorm:"column:voltage_lower_limits" json:"voltage_lower_limits"`
	// NationalStandard   int                                   `gorm:"column:national_standard" json:"national_standard"`
	ParkNo       string                     `gorm:"column:park_no" json:"park_no"`
	PushInterval int                        `gorm:"-" json:"push_interval"`
	LastPushTime int64                      `gorm:"-" json:"last_push_time"`
	OrderState   *interfaces.KindOrderState `gorm:"-" json:"order_state"`
	StationID    datasource.UUID            `gorm:"-" json:"station_id"`
}

func (e Connector) TableName() string {
	return "base_connector"
}

// Key 缓存的键
func (e *Connector) Key() string {
	return ""
}

// AfterCreate 钩子函数
func (e *Connector) AfterCreate(db *gorm.DB) error {
	return nil
}

func (e *Connector) AfterUpdate(db *gorm.DB) error {
	return nil
}

func (e *Connector) AfterFind(db *gorm.DB) error {
	return nil
}

func TestSmallGet(t *testing.T) {
	os.Setenv("DB_HOST", "192.168.0.4")
	os.Setenv("DB_USER", "root")
	os.Setenv("DB_PASSWD", "jqcsms@uat123")
	InitMysqlWithEnvAndDB("jx-csms", nil)
	SetDB(mysqlDB)
	// equipment := &Equipment{}
	// smallEquipment := &SmallEquipment{}
	// err := SmallGet(db.WithContext(context.Background()), equipment, smallEquipment, "sn = ?", "T1641735211")
	// if err != nil {
	// 	t.Error(err)
	// }
	// t.Log(smallEquipment.SerialNumber)
	GetByID(db.WithContext(context.Background()), &Connector{}, 296409967734917)
}

func TestFindInBatch(t *testing.T) {
	os.Setenv("DB_HOST", "192.168.0.4")
	os.Setenv("DB_USER", "root")
	os.Setenv("DB_PASSWD", "jqcsms@uat123")
	InitMysqlWithEnvAndDB("jx-csms", nil)
	SetDB(mysqlDB)
	equipments := make([]*SmallEquipment, 0)
	err := FindInBatches(db.WithContext(context.Background()), &equipments, 100, func(tx *gorm.DB, batch int) error {
		for _, v := range equipments {
			t.Log(v.SerialNumber)
		}
		return nil
	}, "operator_id = ?", 586069660491776)
	if err != nil {
		t.Error(err)
	}
}
