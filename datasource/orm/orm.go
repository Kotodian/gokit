package orm

import (
	"errors"
	"time"

	"github.com/Kotodian/gokit/datasource"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
)

func GetDB() *gorm.DB {
	return db
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
	// 钩子函数
	AfterCreate(db *gorm.DB) error
	AfterUpdate(db *gorm.DB) error
	AfterFind(db *gorm.DB) error
}

func GetByID(obj Object, id datasource.UUID) (err error) {
	err = db.Where("id = ?", id).First(obj).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		obj = nil
		err = nil
	}
	return err
}

func Get(obj Object, cond string, where ...interface{}) (err error) {
	err = db.Where(cond, where).First(obj).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		obj = nil
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
