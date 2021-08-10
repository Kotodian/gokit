package orm

import (
	"errors"

	"github.com/Kotodian/gokit/datasource"
	"gorm.io/gorm"
)


var (
	db *gorm.DB
)

type DeleteFunc func(object Object) error

func RealDelete(object Object) error {
	return db.Delete(object).Error
}

type Object interface {
	// ID 数据库中的唯一值
	ID() datasource.UUID
	// TableName 数据库中的表名
	TableName() string
	// 创建时间
	CreatedAt() int64
	// 更新时间
	UpdatedAt() int64
	// 删除时间
	DeletedAt() int64
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

func Delete(obj Object, deleteFunc DeleteFunc) error {
	return deleteFunc(obj)
}

func UpdateColumn(obj Object, f map[string]interface{}) error {
	return db.Model(obj).Updates(f).Error
}
