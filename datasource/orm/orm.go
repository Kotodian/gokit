package orm

import (
	"errors"
	"github.com/Kotodian/gokit/datasource"
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

type DeleteFunc func(object Object) error
type CreateFunc func(object Object) error

func RealDelete(object Object) error {
	return db.Delete(object).Error
}

func FakeDelete(object Object) error {
	return db.Table(object.TableName()).Where("id = ?", object.ID()).UpdateColumn("deleted_at", time.Now().Unix()).Error
}

type Object interface {
	// Exists 判断是否存在
	Exists() bool
	// ID 数据库中的唯一值
	ID() datasource.UUID
	// 缓存的键
	Key() string
	// TableName 数据库中的表名
	TableName() string
	// 创建时间
	CreatedAt() int64
	// 更新时间
	UpdatedAt() int64
	// 删除时间
	DeletedAt() int64
	GetVersion() int
	SetVersion(version int)

	// 钩子函数
	AfterCreate(db *gorm.DB) error
	AfterUpdate(db *gorm.DB) error
	AfterFind(db *gorm.DB) error
}

func GetByID(obj Object, id datasource.UUID) (err error) {
	return Get(obj, "id = ?", id)
}

func Get(obj Object, cond string, where ...interface{}) (err error) {
	err = db.Where(cond, where).First(obj).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return err
}

func Delete(obj Object, deleteFunc ...DeleteFunc) error {
	if len(deleteFunc) > 0 {
		return deleteFunc[0](obj)
	} else {
		return FakeDelete(obj)
	}
}

func UpdateColumn(obj Object, f map[string]interface{}) error {
	return db.Model(obj).Updates(f).Error
}

func UpdateWithOptimistic(obj Object, f map[string]interface{}) error {
	if f == nil {
		return nil
	}
	return updateWithOptimistic(obj, f, 3, 0)
}

func updateWithOptimistic(obj Object, f map[string]interface{}, retryCount, currentRetryCount int) error {
	if currentRetryCount > retryCount {
		return ErrRetryMax
	}
	currentVersion := obj.GetVersion()
	obj.SetVersion(currentVersion + 1)
	f["version"] = currentVersion + 1
	column := db.Model(obj).Where("version", currentVersion).Updates(f)
	affected := column.RowsAffected
	if affected == 0 {
		time.Sleep(100 * time.Millisecond)
		id := obj.ID()
		db.First(obj, id)
		currentRetryCount++
		err := updateWithOptimistic(obj, f, retryCount, currentRetryCount)
		if err != nil {
			return err
		}
	}
	return column.Error
}

func Create(obj Object) error {
	err := db.Create(obj).Error
	return err
}

func Updates(tableName string, updates map[string]interface{}, cond string, where ...interface{}) error {
	return db.Table(tableName).Where(cond, where).Updates(updates).Error
}

func Count(tableName string, cond string, where ...interface{}) (count int64, err error) {
	count = 0
	err = db.Table(tableName).Where(cond, where).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func Find(tableName string, condition map[string]interface{}, fields []string, dest interface{}) error {
	cond, vals, err := builder.BuildSelect(tableName, condition, fields)
	if err != nil {
		return err
	}
	return db.Raw(cond, vals).Scan(dest).Error
}
