package orm_bak

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/didi/gendry/builder"
	"github.com/edwardhey/gorm"
	log "github.com/sirupsen/logrus"
)

// GetByID ...
func GetByIDWithContext(ctx context.Context, obj IBase, ID interface{}) (err error) {
	// elem := reflect.ValueOf(obj).Elem()
	db := GetDBWithContext(ctx)
	defer func() {
		if err != nil {
			err = fmt.Errorf("%s get error, %s", obj.TableName(), err.Error())
		}
	}()
	keyName, ok := getCacheEnabledWithContext(ctx, obj, ID)
	if ok {
		if MC.Clone(keyName, obj) {
			return
		}
	}
	scope := db.NewScope(obj)
	if !scope.HasColumn("id") {
		err = errors.New("id column not found")
		return
	}
	err = ErrIgnoreNotFound(db.Model(obj).Where("id=?", ID).Scan(obj).Error)
	if err != nil {
		return err
	}
	_ = scope.SetColumn("id", ID)
	if ok && obj.IsExists() {
		SaveToMC(obj, time.Hour)
	}
	return err
}

// MustGetByID 根据id获取对象(必须获取)
func MustGetByIDWithContext(ctx context.Context, obj IBase, ID interface{}) (err error) {
	db := GetDBWithContext(ctx)
	defer func() {
		if err != nil {
			err = fmt.Errorf("%s get error, %s", obj.TableName(), err.Error())
		}
	}()
	keyName, ok := getCacheEnabledWithContext(ctx, obj, ID)
	if ok {
		if MC.Clone(keyName, obj) {
			return
		}
	}
	scope := db.NewScope(obj)
	if !scope.HasColumn("id") {
		return errors.New("id column not found")
	}
	err = Err(db.Model(obj).First(obj, "id=?", ID).Error)
	if err != nil {
		return
		// if err == gorm.errrecordnotfound {
		// 	scope.setcolumn("id", id)
		// 	return nil
		// k}
	}
	if ok && obj.IsExists() {
		SaveToMC(obj, time.Hour)
	}
	return err
}

// GetByRaw 适用于查询多个字段多条记录
func GetByRawWithContext(ctx context.Context, out interface{}, sql string, params ...interface{}) (err error) {
	db := GetDBWithContext(ctx)
	if err = ErrIgnoreNotFound(db.Raw(sql, params...).Scan(out).Error); err != nil {
		return
	}

	_typer := reflect.TypeOf((*IBase)(nil)).Elem()
	//_typer2 := reflect.TypeOf(IBase).Elem()
	refV := reflect.ValueOf(out)

	if refV.Elem().Kind() == reflect.Slice {
		var ok bool
		outElem := refV.Elem()
		for i := 0; i < outElem.Len(); i++ {
			elem := outElem.Index(i)
			if i == 0 {
				if refV.Elem().Index(i).Kind() == reflect.Ptr {
					ok = elem.Type().Implements(_typer)
				} else {
					ok = elem.Addr().Type().Implements(_typer)
				}
			}
			if ok {
				if elem.Kind() == reflect.Ptr {
					elem.MethodByName("AfterFind").Call([]reflect.Value{})
				} else {
					elem.Addr().MethodByName("AfterFind").Call([]reflect.Value{})
				}
			}
		}
	} else if refV.Type().Implements(_typer) {
		refV.MethodByName("AfterFind").Call([]reflect.Value{})
	}
	return err
}

// GetPluckByRaw 适用于只查询单个字段多条记录
func GetPluckByRawWithContext(ctx context.Context, out interface{}, colName string, sql string, params ...interface{}) (err error) {
	db := GetDBWithContext(ctx)
	err = ErrIgnoreNotFound(db.Raw(sql, params...).Pluck(colName, out).Error)
	return err
}

// GetID 获取对象id
func GetIDWithContext(ctx context.Context, obj IBase, where ...interface{}) (id interface{}, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%s get id error, %s", obj.TableName(), err.Error())
		}
	}()

	db := GetDBWithContext(ctx)
	var ID struct {
		ID interface{}
	}
	err = ErrIgnoreNotFound(db.Table(obj.TableName()).Select("id").First(&ID, where...).Error)
	if err != nil {
		return 0, err
	}
	return ID.ID, nil
}

// CountWithContext 计数
func CountWithContext(ctx context.Context, obj IBase, where map[string]interface{}) (n int, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%s count error, %s", obj.TableName(), err.Error())
		}
	}()
	db := GetDBWithContext(ctx)

	cond, values, _ := builder.BuildSelect(obj.TableName(), where, []string{"count(1)"})
	if err = ErrIgnoreNotFound(db.Raw(cond, values...).Count(&n).Error); err != nil {
		return
	}
	return
}

// GetByContext 根据where条件获取对象
func GetByContext(ctx context.Context, obj IBase, cond string, where ...interface{}) (err error) {
	db := GetDBWithContext(ctx)
	defer func() {
		if err != nil {
			err = fmt.Errorf("%s get error, %s", obj.TableName(), err.Error())
		}
	}()
	if obj.GetCacheKey() != "" {
		cacheKeyName := GetCacheKeyName(obj, cond, where...)
		if cacheKeyName != "" && MC.Clone(cacheKeyName, obj) {
			return nil
		}
		defer func() {
			if obj.IsExists() {
				_ = MC.Put(cacheKeyName, obj, time.Hour)
			}
		}()
	}

	err = ErrIgnoreNotFound(db.Model(obj).Where(cond, where...).Scan(obj).Error)
	return
}

// MustGetByContext 根据where条件获取对象(必须获取)
func MustGetByContext(ctx context.Context, obj IBase, cond string, where ...interface{}) (err error) {
	db := GetDBWithContext(ctx)
	defer func() {
		if err != nil {
			err = fmt.Errorf("%s get error, %s", obj.TableName(), err.Error())
		}
	}()
	if obj.GetCacheKey() != "" {
		cacheKeyName := GetCacheKeyName(obj, cond, where...)
		if cacheKeyName != "" && MC.Clone(cacheKeyName, obj) {
			return nil
		}
		defer func() {
			if err == nil {
				_ = MC.Put(cacheKeyName, obj, time.Hour)
			}
		}()
	}
	err = Err(db.Model(obj).Where(cond, where...).Scan(obj).Error)
	return
}

// GetByCacheWithContext 根据where条件在内存中获取对象
func GetByCacheWithContext(ctx context.Context, obj IBase, ttl time.Duration, cond string, where ...interface{}) (err error) {
	db := GetDBWithContext(ctx)
	defer func() {
		if err != nil {
			err = fmt.Errorf("%s get error, %s", obj.TableName(), err.Error())
		}
	}()
	if obj.GetCacheKey() != "" {
		cacheKeyName := GetCacheKeyName(obj, cond, where...)
		if cacheKeyName != "" && MC.Clone(cacheKeyName, obj) {
			//fmt.Println("load from cache", cacheKeyName, obj.TableName())
			return nil
		}
		defer func() {
			if obj.IsExists() {
				_ = MC.Put(cacheKeyName, obj, time.Hour)
			}
		}()
	}
	//fmt.Println(db.Dialect().CurrentDatabase())
	err = ErrIgnoreNotFound(db.Model(obj).Where(cond, where...).Scan(obj).Error)
	return err
}

// GetByCacheWithRawSQLWithContext ...
func GetByCacheWithRawSQLWithContext(ctx context.Context, obj IBase, ttl time.Duration, sql string, params ...interface{}) (err error) {
	db := GetDBWithContext(ctx)
	defer func() {
		if err != nil {
			err = fmt.Errorf("%s get error, %s", obj.TableName(), err.Error())
		}
	}()
	if obj.GetCacheKey() != "" {
		cacheKeyName := GetCacheKeyName(obj, sql, params...)
		//fmt.Println("search cache", cacheKeyName, obj.TableName())
		if cacheKeyName != "" && MC.Clone(cacheKeyName, obj) {
			//fmt.Println("load from cache", cacheKeyName, obj.TableName())
			return nil
		}
		defer func() {
			if obj.IsExists() {
				_ = MC.Put(cacheKeyName, obj, time.Hour)
			}
		}()
	}
	err = ErrIgnoreNotFound(db.Raw(sql, params...).Find(obj).Error)
	return
}

// GetByCacheWithDefaultTime 根据where条件在内存中获取对象
func GetByCacheWithDefaultTimeWithContext(ctx context.Context, obj IBase, cond string, where ...interface{}) error {
	return GetByCacheWithContext(ctx, obj, time.Hour, cond, where...)
}

// AddWithContext ...
func AddWithContext(ctx context.Context, objs ...IBase) (err error) {
	db := GetDBWithContext(ctx)
	if len(objs) == 0 {
		return
	}
	if len(objs) == 1 {
		err = db.Create(objs[0]).Error
		if err != nil {
			return
		}
		SaveToMC(objs[0], time.Hour)
		//SetGroupCacheID(objs[0])
		//if keyName, ok := getCacheEnabled(objs[0]); ok {
		//
		//	MC.Delete(keyName)
		//}
	}
	return
}

// DelWithContext 删除对象(会导致panic)
// @param {[type]} obj interface{}  需要删除的对象的指针
// @return  error
func DelWithContext(ctx context.Context, objs ...IBase) (err error) {
	db := GetDBWithContext(ctx)
	if len(objs) == 0 {
		return nil
	}
	if len(objs) == 1 {
		// _, err = ORM.Delete(objs[0])
		// elem := reflect.ValueOf(objs).Elem()
		DeleteFromMC(objs[0])
		//if keyName, ok := getCacheEnabled(objs[0]); ok {
		//	_ = MC.Delete(keyName)
		//}
		//SetGroupCacheID(objs)
		err = db.Delete(objs[0]).Error
		objs[0].ResetAfterCallbackFunc()
	} else {
		var ok bool
		var tx *gorm.DB
		if _, ok = interface{}(db.DB).(*sql.Tx); !ok {
			tx = db.Begin()
		} else {
			tx = db
		}
		if !ok {
			defer func() {
				if err != nil {
					_ = tx.Rollback().Error
				} else {
					err = tx.Commit().Error
				}
			}()
		}
		// err = ORM.Begin()
		//if err != nil {
		//	return
		//}
		for _, obj := range objs {
			// elem := reflect.ValueOf(obj).Elem()
			//if _, ok := getCacheEnabled(obj); ok {
			//_ = MC.Delete(keyName)
			DeleteFromMC(obj)
			//}
			err = tx.Delete(obj).Error
			if err != nil {
				break
			}
			obj.ResetAfterCallbackFunc()
			// tx.Delete(obj)

		}

	}
	if err != nil {
		return
	}
	for _, obj := range objs {
		ref := reflect.ValueOf(obj)
		if resetFunc := ref.MethodByName("Reset"); resetFunc.IsValid() {
			resetFunc.Call(nil)
		} else {
			reset(obj)
		}
	}
	return
}

// NewDBContext 创建db对象的ctx
func NewDBContext(ctx context.Context, _db *gorm.DB) context.Context {
	return context.WithValue(ctx, "db", _db)
}

// GetDBWithContext 获取db对象
func GetDBWithContext(ctx context.Context) (_db *gorm.DB) {
	if _db := ctx.Value("db"); _db != nil {
		return _db.(*gorm.DB)
	}
	return db
}

func FirstOrCreatedWithContext(ctx context.Context, obj IBase, where ...interface{}) (err error) {
	db := GetDBWithContext(ctx)
	if err = Err(db.FirstOrCreate(obj, where...).Error); err == nil {
		DeleteFromMC(obj)
	}
	return
}

//SaveWithCondAndContext
func SaveWithCondAndContext(ctx context.Context, obj IBase, query interface{}, cond ...interface{}) (err error) {
	db := GetDBWithContext(ctx)
	obj.AddAfterSaveCallback("_tx_delete_mc", func() {
		DeleteFromMC(obj)
	})
	udb := db.Model(obj).Where(query, cond...).Save(obj)
	if err = udb.Error; err != nil {
		return
	} else if udb.RowsAffected != 1 {
		err = fmt.Errorf("更新失败,受影响的条数为%d", udb.RowsAffected)
		return
	}
	DeleteFromMC(obj)
	return
}

// SaveWithContext 保存对象
// @param {[type]} obj interface{}) (err error [description]
func SaveWithContext(ctx context.Context, objList ...IBase) (err error) {
	db := GetDBWithContext(ctx)
	if len(objList) == 0 {
		return nil
	} else if len(objList) == 1 {
		obj := objList[0]
		// elem := reflect.ValueOf(obj).Elem()
		if obj.IsDeleted() == true {
			return nil
		}
		//obj.AddAfterSaveCallback("orm", func() {
		//	fmt.Println("这里保存到mc")
		//
		//})
		obj.AddAfterSaveCallback("_cache", func() {
			//SaveToMC(obj, time.Hour)
			DeleteFromMC(obj)
		})
		if obj.IsExists() == true {
			// _, err = ORM.Update(obj)
			//fmt.Println("-----------save---------------", obj.InstanceName())
			err = db.Save(obj).Error
		} else {
			//_, err = ORM.Insert(obj)
			//fmt.Println("-----------create---------------", )
			err = db.Create(obj).Error
		}
		if err != nil {
			log.Errorf("save db error:%s, table:%s", err.Error(), obj.TableName())
			return
		}
		DeleteFromMC(obj)
		//obj.ResetAfterCallbackFunc()
		//SaveToMC(obj)
		//if obj.GetCacheKey() != "" {
		//	SetGroupCacheID(obj)
		//}
		//if keyName, ok := getCacheEnabled(obj); ok {
		//	// MC.Delete(keyName)
		//	saveToMC(keyName, obj, time.Second*3600)
		//}
	} else {
		var ok bool
		var tx *gorm.DB

		if _, ok = db.CommonDB().(*sql.Tx); !ok {
			tx = db.Begin()
		} else {
			tx = db
		}
		if !ok {
			defer func() {
				if err != nil {
					_ = tx.Rollback().Error
				} else {
					for _, obj := range objList {
						DeleteFromMC(obj)
					}
					err = tx.Commit().Error
				}
			}()
		}
		for _, obj := range objList {
			// elem := reflect.ValueOf(obj).Elem()
			if obj.IsDeleted() == true {
				continue
				// return fmt.Errorf("%s %v is already deleted", db.NewScope(obj).GetModelStruct().ModelType.Name(), elem.FieldByName("Id"))
			}
			//obj.AddAfterSaveCallback("txSave", func() {
			//	SaveToMC(obj, time.Hour)
			//})
			obj.AddAfterSaveCallback("_cache", func() {
				DeleteFromMC(obj)
			})
			if obj.IsExists() == true {
				//fmt.Println("-----------save---------------", obj.InstanceName())
				err = tx.Save(obj).Error
				// _, err = ORM.Update(obj)
			} else {
				// _, err = ORM.Insert(obj)
				//fmt.Println("-----------create---------------", obj.InstanceName())
				err = tx.Create(obj).Error
			}
			if err != nil {
				return
			}
			DeleteFromMC(obj)

			//obj.ResetAfterCallbackFunc()
			//SaveToMC(obj)
			//if obj.GetCacheKey() != "" {
			//	SetGroupCacheID(obj)
			//}
			//if keyName, ok := getCacheEnabled(obj); ok {
			//	// MC.Delete(keyName)
			//	saveToMC(keyName, obj, time.Second*3600)
			//	//SetGroupCacheID(obj)
			//}

		}
	}
	return
}

// UpdateColumn ...
func UpdateColumnWitContext(ctx context.Context, obj IBase, f map[string]interface{}) error {
	db := GetDBWithContext(ctx)
	obj.AddAfterSaveCallback("_cache", func() {
		//SaveToMC(obj, time.Hour)
		DeleteFromMC(obj)
	})
	err := db.Model(obj).Updates(f).Error
	if err != nil {
		return err
	}
	DeleteFromMC(obj)
	return nil
}
