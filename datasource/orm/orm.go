package orm

import (
	"errors"
	"github.com/Kotodian/gokit/datasource"
	"github.com/Kotodian/gokit/retry"
	"github.com/Kotodian/gokit/retry/strategy"
	"github.com/didi/gendry/builder"
	"gorm.io/gorm"
	"time"
)

var (
	db *gorm.DB
)
var (
	ErrRetryMax = errors.New("maximum number of retries exceeded")
)

func GetDB() *gorm.DB {
	return db
}

func SetDB(_db *gorm.DB) {
	db = _db
}

type DeleteFunc func(conn *gorm.DB, object Object) error
type CreateFunc func(conn *gorm.DB, object Object) error

type Object interface {
	// Exists 判断是否存在
	Exists() bool
	// ID 数据库中的唯一值
	ID() datasource.UUID
	// Key 缓存的键
	Key() string
	// TableName 数据库中的表名
	TableName() string
	// CreatedAt 创建时间
	CreatedAt() int64
	// UpdatedAt 更新时间
	UpdatedAt() int64
	// CreatedBy 创建者
	CreatedBy() datasource.UUID
	// UpdatedBy 更新者
	UpdatedBy() datasource.UUID
	// GetVersion 获取版本
	GetVersion() int
	// SetVersion 设置版本
	SetVersion(version int)
	// AfterCreate 钩子函数
	AfterCreate(db *gorm.DB) error
	// AfterUpdate 钩子函数
	AfterUpdate(db *gorm.DB) error
	// UpdateHook 自定义钩子函数
	UpdateHook()
	// AfterFind 钩子函数
	AfterFind(db *gorm.DB) error
}

func GetByID(conn *gorm.DB, obj Object, id datasource.UUID) (err error) {
	return Get(conn, obj, "id = ?", id)
}

func Get(conn *gorm.DB, obj Object, cond string, where ...interface{}) (err error) {
	err = conn.Where(cond, where...).Take(obj).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return err
}

func UpdateColumn(conn *gorm.DB, obj Object, f map[string]interface{}) error {
	err := conn.Model(obj).Updates(f).Error
	if err != nil {
		return err
	}
	obj.UpdateHook()
	return nil
}

func UpdateWithOptimistic(conn *gorm.DB, obj Object, f map[string]interface{}) error {
	if f == nil {
		return nil
	}
	err := updateWithOptimistic(conn, obj, f, 3, 0)
	if err != nil {
		return err
	}
	obj.UpdateHook()
	return nil
}

func updateWithOptimistic(conn *gorm.DB, obj Object, f map[string]interface{}, retryCount, currentRetryCount int) error {
	for currentRetryCount < retryCount {
		currentVersion := obj.GetVersion()
		obj.SetVersion(currentVersion + 1)
		f["version"] = currentVersion + 1
		column := conn.Model(obj).Where("version", currentVersion).Updates(f)
		if column.Error != nil {
			return column.Error
		}
		if column.RowsAffected == 0 {
			time.Sleep(100 * time.Millisecond)
			id := obj.ID()
			err := conn.First(obj, id).Error
			if err != nil {
				if err != gorm.ErrRecordNotFound {
					return err
				}
			}
			currentRetryCount++
		} else {
			return nil
		}
	}
	return nil
}

func Create(conn *gorm.DB, obj Object) error {
	err := conn.Create(obj).Error
	return err
}

func Updates(conn *gorm.DB, tableName string, updates map[string]interface{}, cond string, where ...interface{}) error {
	return conn.Table(tableName).Where(cond, where...).Updates(updates).Error
}
func UpdatesModel(conn *gorm.DB, obj Object, updates map[string]interface{}, cond string, where ...interface{}) error {
	return conn.Model(obj).Where(cond, where...).Updates(updates).Error
}

func Count(conn *gorm.DB, tableName string, cond string, where ...interface{}) (count int64, err error) {
	count = 0
	err = conn.Table(tableName).Where(cond, where...).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func Find(conn *gorm.DB, tableName string, condition map[string]interface{}, fields []string, dest interface{}) error {
	cond, vals, err := builder.BuildSelect(tableName, condition, fields)
	if err != nil {
		return err
	}
	return conn.Raw(cond, vals...).Scan(dest).Error
}

func FirstOrCreate(conn *gorm.DB, object Object, condition interface{}) error {
	return conn.FirstOrCreate(object, condition).Error
}

func FindInBatches(conn *gorm.DB, dest interface{}, limit int, fc func(tx *gorm.DB, batch int) error, where string, cond ...interface{}) error {
	return conn.Where(where, cond...).FindInBatches(dest, limit, fc).Error
}

func wrapCreateFunc(conn *gorm.DB, object Object) retry.Action {
	return func(attempt uint) error {
		return Create(conn, object)
	}
}

func RetryCreate(conn *gorm.DB, object Object) error {
	return retry.Retry(wrapCreateFunc(conn, object), strategy.Limit(3))
}
