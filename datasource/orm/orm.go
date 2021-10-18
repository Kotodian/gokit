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

func GetByID(conn *gorm.DB, obj Object, id datasource.UUID) (err error) {
	return Get(conn, obj, "id = ?", id)
}

func Get(conn *gorm.DB, obj Object, cond string, where ...interface{}) (err error) {
	err = conn.Where(cond, where...).First(obj).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return err
}

func GetByMoreCond(conn *gorm.DB, obj Object, condVal map[string]interface{}) (err error) {
	for k, v := range condVal {
		conn = conn.Where(k, v)
	}
	err = conn.First(obj).Error
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

func UpdateColumn(conn *gorm.DB, obj Object, f map[string]interface{}) error {
	return conn.Model(obj).Updates(f).Error
}

func UpdateWithOptimistic(conn *gorm.DB, obj Object, f map[string]interface{}) error {
	if f == nil {
		return nil
	}
	return updateWithOptimistic(conn, obj, f, 3, 0)
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
	err = db.Table(tableName).Where(cond, where...).Count(&count).Error
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
