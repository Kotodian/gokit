package orm

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/Kotodian/gokit/datasource"
	"github.com/Kotodian/gokit/ecode"
	"github.com/didi/gendry/builder"
	"github.com/edwardhey/gorm"
	"github.com/edwardhey/mysql"
	log "github.com/sirupsen/logrus"
)

var afterSaveCommit sync.Map

// IBase 接口
type IBase interface {
	IsExists() bool
	IsNew() bool
	IsDeleted() bool
	SetID(id interface{})
	TableName() string
	GetCacheKey() string
	SetIsExists(bool)
	SetIsDeleted(bool)
	ResetAfterCallbackFunc()
	AddAfterSaveCallback(name string, fn func())
	AfterSaveCommit(*gorm.Scope)
	AfterFind() error
	//HaveCASID() bool
	//Scope() *gorm.Scope
}

// Base 基类
type Base struct {
	CreatedAt         time.Time         `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt         time.Time         `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	isExists          bool              `gorm:"-" json:"-"`
	isDeleted         bool              `gorm:"-" json:"-"`
	afterSaveCallback map[string]func() `gorm:"-" json:"-"`
	//scope             *gorm.Scope       `gorm:"-" json:"-"`
	//CasID             *uint64           `gorm:"cas_id" json:"cas_id"`
	//cacheVer          uint64            `gorm:"-" json:"-"`
	//DeletedAt         *time.Time        `gorm:"column:deleted_at;type:datetime"` // need null mode
}

var (
	db                 *gorm.DB
	TTLMinute          time.Duration = time.Second * 60
	TTLDay             time.Duration = time.Second * 86400
	TTLWeek            time.Duration = TTLDay * 7
	TTLMonth           time.Duration = TTLWeek * 4
	TTLYear            time.Duration = TTLMonth * 12
	TimeFormatDateTime string        = "2006-01-02 15:04:05"
	TimeFormatDate     string        = "2006-01-02"
)

// Err 错误处理
func Err(err error) error {
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			err = ecode.NothingFound
		}
	}
	return err
}

// ErrIsDuplicate 是否数据重复错误
func ErrIsDuplicate(err error) bool {
	if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
		return true
	}
	return false
}

// ErrIgnoreNotFound 错误忽略
func ErrIgnoreNotFound(err error) error {
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
	}
	return err
}

// SetOrmDB SetOrmDB
func SetOrmDB(_db *gorm.DB) {
	db = _db
}

// GetCacheKey
func (m *Base) GetCacheKey() string {
	return ""
}

// IsExists 是否存在对象
func (m *Base) IsExists() bool {
	return m.isExists
}

// SetID ...
func (m *Base) SetID(id interface{}) {
	s := db.NewScope(m)
	if s.HasColumn("id") {
		_ = db.NewScope(m).SetColumn("id", id)
	}
}

// ResetAfterCallbackFunc ...
func (m *Base) ResetAfterCallbackFunc() {
	m.afterSaveCallback = map[string]func(){}
}

// AddAfterSaveCallback 添加保存之后
func (m *Base) AddAfterSaveCallback(name string, fn func()) {
	if m.afterSaveCallback == nil {
		m.afterSaveCallback = map[string]func(){}
	}
	//fmt.Println("----------------------------------->回调存储", fn)
	m.afterSaveCallback[name] = fn
}

// IsNew 是否新对象
func (m *Base) IsNew() bool {
	return !m.isExists
}

// IsDeleted 是否已删除
func (m *Base) IsDeleted() bool {
	return m.isDeleted
}

// SetIsExists 标记对象已存在
func (m *Base) SetIsExists(b bool) {
	m.isExists = b
}

// SetIsDeleted 标记对象是删除的
func (m *Base) SetIsDeleted(b bool) {
	m.isDeleted = b
}

// SetNew 设置成新对象
func (m *Base) SetNew() {
	m.isExists = false
}

// SetExists 标记对象已存在
func (m *Base) SetExists() {
	m.isExists = true
}

// Reset ...
func (m *Base) Reset() {
}

// BeforeSave default callbacks
func (m *Base) BeforeSave() error {
	return nil
}

func (m Base) TableName() string {
	return ""
}

// AfterSave ...
func (m *Base) AfterSave(scope *gorm.Scope) error {
	if funcs, ok := scope.Get("gorm:tx:objs"); ok {
		funcs.(map[string]*gorm.Scope)[scope.InstanceID()] = scope
		scope.Set("gorm:tx:objs", funcs)
	}
	return nil
}

// AfterSaveCommit 执行保存后的callback
func (m *Base) AfterSaveCommit(scope *gorm.Scope) {
	//TODO: 暂不知道为什么数据更新会有延迟，所以先用这种方式来处理
	if len(m.afterSaveCallback) > 0 {
		go func() {
			l, _ := afterSaveCommit.LoadOrStore(scope.InstanceID(), &sync.RWMutex{})
			l.(*sync.RWMutex).Lock()
			defer func() {
				l.(*sync.RWMutex).Unlock()
				afterSaveCommit.Delete(scope.InstanceID())
			}()
			time.Sleep(time.Second)
			for _, fn := range m.afterSaveCallback {
				fn()
				//}(fn)
				//go fn()
				//go func(fn func()) {
				//	time.Sleep(time.Second)
			}
			m.ResetAfterCallbackFunc()
		}()
	}
	//return nil
}

// BeforeCreate ...
func (m *Base) BeforeCreate() error {
	now := time.Now().Local()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	m.UpdatedAt = now
	return nil
}

// AfterCreate ...
func (m *Base) AfterCreate() error {
	m.isExists = true
	m.isDeleted = false
	return nil
}

// BeforeUpdate ...
func (m *Base) BeforeUpdate() error {
	now := time.Now().Local()
	m.UpdatedAt = now
	return nil
}

// AfterUpdate ...
func (m *Base) AfterUpdate() error {
	return nil
}

// AfterFind ...
func (m *Base) AfterFind() error {
	m.isExists = true
	m.isDeleted = false
	return nil
}

// BeforeDelete ...
func (m *Base) BeforeDelete() error {
	return nil
}

// AfterDelete ...
func (m *Base) AfterDelete() error {
	m.isExists = false
	m.isDeleted = true
	return nil
}

// GetIDChanByWhere 根据条件查询出符合记录的ID
func GetIDChanByWhere(o *gorm.DB) <-chan datasource.UUID {
	ch := make(chan datasource.UUID, 1)
	go func() {
		defer close(ch)
		rows, err := o.Select("id").Rows()
		if err != nil {
			return
		}
		defer rows.Close()
		for rows.Next() {
			var _id datasource.UUID
			_ = rows.Scan(&_id)
			ch <- _id
		}
	}()
	return ch
}

// GetByID ...
func GetByID(obj IBase, ID interface{}) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%s get id:%v error, %s", obj.TableName(), ID, err.Error())
		}
	}()
	keyName, ok := getCacheEnabled(obj, ID)
	if ok {
		if MC.Clone(keyName, obj) {
			return
		}
	}
	//scope := db.NewScope(obj)
	//if !scope.HasColumn("id") {
	//	err = errors.New("id column not found")
	//	return
	//}
	err = ErrIgnoreNotFound(db.Model(obj).Where("id=?", ID).Scan(obj).Error)
	if err != nil {
		return err
	}
	if ok && obj.IsExists() {
		SaveToMC(obj, time.Hour)
	}
	return err
}

// MustGetByID 根据id获取对象(必须获取)
func MustGetByID(obj IBase, ID interface{}) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%s get id:%v error, %s", obj.TableName(), ID, err.Error())
		}
	}()
	keyName, ok := getCacheEnabled(obj, ID)
	if ok {
		if MC.Clone(keyName, obj) {
			return
		}
	}
	//scope := db.NewScope(obj)
	//if !scope.HasColumn("id") {
	//	err = errors.New("id column not found")
	//	return
	//}
	err = Err(db.Model(obj).First(obj, "id=?", ID).Error)
	if err != nil {
		return
	}
	if ok && obj.IsExists() {
		SaveToMC(obj, time.Hour)
	}
	return err
}

//
//func Get(obj IBase, cond string, where ...interface{}) (err error) {
//	err = db.Model(obj).Where(cond, where...).Scan(obj).Error
//	if err == gorm.ErrRecordNotFound {
//		return nil
//	}
//	return err
//}
//
//func MustGet(obj IBase, cond string, where ...interface{}) (err error) {
//	return db.Model(obj).Where(cond, where...).Scan(obj).Error
//}

// GetByRaw 适用于查询多个字段多条记录
func GetByRaw(out interface{}, sql string, params ...interface{}) (err error) {
	//defer func() { err = ErrIgnoreNotFound(err) }()
	//db = db.Debug()
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
	return
}

// GetPluckByRaw 适用于只查询单个字段多条记录
func GetPluckByRaw(out interface{}, colName string, sql string, params ...interface{}) (err error) {
	//defer func() {
	//	err = ErrIgnoreNotFound(err)
	//}()
	return ErrIgnoreNotFound(db.Raw(sql, params...).Pluck(colName, out).Error)
}

// GetID 获取对象id
func GetID(obj IBase, where ...interface{}) (id interface{}, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%s get id error, %s", obj.TableName(), err.Error())
		}
	}()
	var ID struct {
		ID interface{}
	}
	err = ErrIgnoreNotFound(db.Table(obj.TableName()).Select("id").First(&ID, where...).Error)
	if err != nil {
		return 0, err
	}
	return ID.ID, nil
}

//CountByRaw
func CountByRaw(cond string, val ...interface{}) (n int, err error) {
	if err = ErrIgnoreNotFound(db.Raw(cond, val...).Count(&n).Error); err != nil {
		return
	}
	return
}

// Count 计数
func Count(obj IBase, where map[string]interface{}) (n int, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%s count error, %s", obj.TableName(), err.Error())
		}
	}()

	cond, values, _ := builder.BuildSelect(obj.TableName(), where, []string{"count(1)"})
	if err = ErrIgnoreNotFound(db.Raw(cond, values...).Count(&n).Error); err != nil {
		return
	}
	return
}

// Get 根据where条件获取对象
func Get(obj IBase, cond string, where ...interface{}) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%s get error, %s", obj.TableName(), err.Error())
		}
	}()
	if obj.GetCacheKey() != "" {
		cacheKeyName := GetCacheKeyName(obj, cond, where...)
		if cacheKeyName != "" {
			if MC.Clone(cacheKeyName, obj) {
				return nil
			} else {
				defer func() {
					if obj.IsExists() {
						_ = MC.Put(cacheKeyName, obj, TTLDay)
					}
				}()
			}
		}
	}

	err = ErrIgnoreNotFound(db.Model(obj).Where(cond, where...).Scan(obj).Error)
	return
}

// MustGet 根据where条件获取对象(必须获取)
func MustGet(obj IBase, cond string, where ...interface{}) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%s get error, %s", obj.TableName(), err.Error())
		}
	}()
	if obj.GetCacheKey() != "" {
		cacheKeyName := GetCacheKeyName(obj, cond, where...)
		if cacheKeyName != "" {
			if MC.Clone(cacheKeyName, obj) {
				return nil
			} else {
				defer func() {
					if obj.IsExists() {
						_ = MC.Put(cacheKeyName, obj, TTLDay)
					}
				}()
			}
		}
	}
	err = Err(db.Model(obj).Where(cond, where...).Scan(obj).Error)
	return
}

// GetByCache 根据where条件在内存中获取对象
func GetByCache(obj IBase, ttl time.Duration, cond string, where ...interface{}) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%s get error, %s", obj.TableName(), err.Error())
		}
	}()
	if obj.GetCacheKey() != "" {
		cacheKeyName := GetCacheKeyName(obj, cond, where...)
		if cacheKeyName != "" {
			if MC.Clone(cacheKeyName, obj) {
				return nil
			} else {
				defer func() {
					if obj.IsExists() {
						_ = MC.Put(cacheKeyName, obj, ttl)
					}
				}()
			}
		}
	}
	//fmt.Println(db.Dialect().CurrentDatabase())
	err = ErrIgnoreNotFound(db.Model(obj).Where(cond, where...).Scan(obj).Error)
	return err
}

// GetByCacheWithRawSQL ...
func GetByCacheWithRawSQL(obj IBase, ttl time.Duration, sql string, params ...interface{}) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%s get error, %s", obj.TableName(), err.Error())
		}
	}()
	if obj.GetCacheKey() != "" {
		cacheKeyName := GetCacheKeyName(obj, sql, params...)
		if cacheKeyName != "" {
			if MC.Clone(cacheKeyName, obj) {
				return nil
			} else {
				defer func() {
					if obj.IsExists() {
						_ = MC.Put(cacheKeyName, obj, ttl)
					}
				}()
			}
		}
	}
	err = ErrIgnoreNotFound(db.Raw(sql, params...).Find(obj).Error)
	return
}

// GetByCacheWithDefaultTime 根据where条件在内存中获取对象
func GetByCacheWithDefaultTime(obj IBase, cond string, where ...interface{}) error {
	return GetByCache(obj, time.Hour, cond, where...)
}

// Create ...
func Create(objs ...IBase) (err error) {
	if len(objs) == 0 {
		return
	}
	if len(objs) == 1 {
		err = db.Create(objs[0]).Error
		if err != nil {
			return
		}
		SaveToMC(objs[0], time.Hour)
		//if keyName, ok := getCacheEnabled(objs[0]); ok {
		//
		//	MC.Delete(keyName)
		//}
	}
	return
}

// Del 删除对象(会导致panic)
// @param {[type]} obj interface{}  需要删除的对象的指针
// @return  error
func Del(objs ...IBase) (err error) {
	//db = db.Debug()
	if len(objs) == 0 {
		return nil
	}
	if len(objs) == 1 {
		if err = db.Delete(objs[0]).Error; err != nil {
			return
		}
		DeleteFromMC(objs[0])
		objs[0].ResetAfterCallbackFunc()
	} else {
		tx := db.Begin()
		// err = ORM.Begin()
		//if err != nil {
		//	return
		//}
		for _, obj := range objs {
			// elem := reflect.ValueOf(obj).Elem()
			err = tx.Delete(obj).Error
			if err != nil {
				break
			}
			obj.ResetAfterCallbackFunc()
			// tx.Delete(obj)
		}
		if err != nil {
			err = tx.Rollback().Error
		} else {
			err = tx.Commit().Error
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
		DeleteFromMC(obj)
	}
	return
}

// reset 重置对象为空对象(会导致panic)
//  @param  {[type]} obj interface{} [description]
func reset(obj IBase) {
	elem := reflect.ValueOf(obj).Elem()
	elem.Set(reflect.Zero(reflect.TypeOf(elem.Interface())))
}

// GetDB 获取db对象
func GetDB() *gorm.DB {
	return db
}

// SaveToMCWithTimeout 临时保存对象
func SaveToMCWithTimeout(t time.Duration, objList ...IBase) {
	for _, obj := range objList {
		if obj.GetCacheKey() != "" {
			SaveToMC(obj, t)
		}
	}
}

func SaveWithCond(obj IBase, query interface{}, cond ...interface{}) (err error) {
	//obj.AddAfterSaveCallback("_tx_delete_mc", func() {
	//	DeleteFromMC(obj)
	//SaveToMCWithTimeout(time.Hour, obj)
	//})
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

// TxSaveWithCond ...
func TxSaveWithCond(tx *gorm.DB, obj IBase, query interface{}, cond ...interface{}) (err error) {
	obj.AddAfterSaveCallback("_tx_delete_mc", func() {
		DeleteFromMC(obj)
	})
	udb := tx.Model(obj).Where(query, cond...).Save(obj)
	if err = udb.Error; err != nil {
		return
	} else if udb.RowsAffected != 1 {
		err = fmt.Errorf("更新失败,受影响的条数为%d", udb.RowsAffected)
		return
	}
	DeleteFromMC(obj)
	return nil
}

func TxUpdateColumn(tx *gorm.DB, obj IBase, f interface{}) error {
	obj.AddAfterSaveCallback("_tx_delete_mc", func() {
		DeleteFromMC(obj)
	})
	err := tx.Model(obj).Updates(f).Error
	if err != nil {
		return err
	}
	DeleteFromMC(obj)
	return nil
}

// TxSave 事务保存
func TxSave(tx *gorm.DB, objList ...IBase) (err error) {
	for _, obj := range objList {
		// elem := reflect.ValueOf(obj).Elem()
		if obj.IsDeleted() == true {
			continue
			// return fmt.Errorf("%s %v is already deleted", db.NewScope(obj).GetModelStruct().ModelType.Name(), elem.FieldByName("Id"))
		}
		//obj.AddAfterSaveCallback("txSave", func() {
		//SaveToMC(obj, time.Hour)
		//})
		obj.AddAfterSaveCallback("_tx_delete_mc", func() {
			DeleteFromMC(obj)
		})
		if obj.IsExists() == true {
			err = tx.Save(obj).Error
			// _, err = ORM.Update(obj)
		} else {
			// _, err = ORM.Insert(obj)
			err = tx.Create(obj).Error
		}
		if err != nil {
			return
		}
		//之所以不使用SaveToMC是因为可能事务提交不成功，这里为了做数据一致性的问题，删除是最好的办法了,目前为止...
		DeleteFromMC(obj)
		//obj.ResetAfterCallbackFunc()
	}
	return
}

// TxDel 事务删除
// TODO: 这里可能会导致脏读，事务并未提交，如果缓存中没有数据，仍然有可能读取到数据再放回缓存中
func TxDel(tx *gorm.DB, objList ...IBase) (err error) {
	for _, obj := range objList {
		err = tx.Delete(obj).Error
		if err != nil {
			break
		}
		DeleteFromMC(obj)
		obj.ResetAfterCallbackFunc()
	}
	tx.Scopes().InstantSet("deleted_objs", objList)

	if err != nil {
		for _, obj := range objList {
			ref := reflect.ValueOf(obj)
			if resetFunc := ref.MethodByName("Reset"); resetFunc.IsValid() {
				resetFunc.Call(nil)
			} else {
				reset(obj)
			}
		}
	}
	return
}

// TxFirstOrCreate ...
func TxFirstOrCreate(tx *gorm.DB, obj IBase, where ...interface{}) (err error) {
	if err = Err(tx.FirstOrCreate(obj, where...).Error); err == nil {
		DeleteFromMC(obj)
		obj.AddAfterSaveCallback("_tx_delete_mc", func() {
			DeleteFromMC(obj)
		})
	}
	return
}

// TxSaveAndCommitWithErr ...
func TxSaveAndCommitWithErr(tx *gorm.DB, e error) (err error) {
	if e == nil {
		if err = tx.Commit().Error; err != nil {
			return
		}
		if v, ok := tx.Scopes().Get("deleted_objs"); ok {
			for _, obj := range v.([]IBase) {
				DeleteFromMC(obj)
			}
		}
		// err = ORM.Commit()
	} else {
		// keep err for return
		_ = tx.Rollback().Error
		err = e
	}
	return
}

// TxSaveAndCommit 使用外部事物保存对象
func TxSaveAndCommit(tx *gorm.DB, objList ...IBase) (err error) {
	defer func() {
		if err == nil {
			if err = tx.Commit().Error; err != nil {
				return
			}
			if v, ok := tx.Scopes().Get("deleted_objs"); ok {
				for _, obj := range v.([]IBase) {
					DeleteFromMC(obj)
				}
			}
			// err = ORM.Commit()
		} else {
			// keep err for return
			tx.Rollback()
		}
	}()
	for _, obj := range objList {
		// elem := reflect.ValueOf(obj).Elem()
		if obj.IsDeleted() == true {
			continue
			// return fmt.Errorf("%s %v is already deleted", db.NewScope(obj).GetModelStruct().ModelType.Name(), elem.FieldByName("Id"))
		}
		obj.AddAfterSaveCallback("txSave", func() {
			DeleteFromMC(obj)
		})
		if obj.IsExists() == true {
			err = tx.Save(obj).Error
			// _, err = ORM.Update(obj)
		} else {
			// _, err = ORM.Insert(obj)
			err = tx.Create(obj).Error
		}
		if err != nil {
			return
		}
		//之所以不使用SaveToMC是因为可能事务提交不成功，这里为了做数据一致性的问题，删除是最好的办法了,目前为止...
		DeleteFromMC(obj)
		//obj.ResetAfterCallbackFunc()
	}
	return
}

// FirstOrCreated ...
func FirstOrCreated(obj IBase, where ...interface{}) (err error) {
	obj.AddAfterSaveCallback("_tx_delete_mc", func() {
		DeleteFromMC(obj)
	})
	if err = Err(db.FirstOrCreate(obj, where...).Error); err != nil {
		return
	}
	DeleteFromMC(obj)
	return
}

// Save 保存对象
// @param {[type]} obj interface{}) (err error [description]
func Save(objList ...IBase) (err error) {
	if len(objList) == 0 {
		return nil
	} else if len(objList) == 1 {
		obj := objList[0]
		// elem := reflect.ValueOf(obj).Elem()
		if obj.IsDeleted() == true {
			return nil
		}
		//DeleteFromMC(obj)
		obj.AddAfterSaveCallback("mc", func() {
			//fmt.Println("这里保存到mc")
			DeleteFromMC(obj)
		})
		if obj.IsExists() == true {
			// _, err = ORM.Update(obj)
			//fmt.Println("-----------save---------------", obj.InstanceName())
			err = db.Save(obj).Error
		} else {
			//_, err = ORM.Insert(obj)
			//fmt.Println("-----------create---------------")
			err = db.Create(obj).Error
		}
		//SaveToMCWithTimeout(time.Hour, obj)
		if err != nil {
			log.Errorf("save db error:%s, table:%s", err.Error(), obj.TableName())
			return
		}
		DeleteFromMC(obj)
		//SaveToMC(obj, time.Hour)
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
		tx := db.Begin()
		// err = ORM.Begin()
		//if err != nil {
		//	return err
		//}
		defer func() {
			if err == nil {
				err = tx.Commit().Error
				// err = ORM.Commit()
			} else {
				// keep err for return
				tx.Rollback()
			}
		}()
		for _, obj := range objList {
			// elem := reflect.ValueOf(obj).Elem()
			if obj.IsDeleted() == true {
				continue
				// return fmt.Errorf("%s %v is already deleted", db.NewScope(obj).GetModelStruct().ModelType.Name(), elem.FieldByName("Id"))
			}
			//obj.AddAfterSaveCallback("txSave", func() {
			//	SaveToMC(obj, time.Hour)
			//})
			obj.AddAfterSaveCallback("mc", func() {
				//fmt.Println("这里保存到mc")
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
			//DeleteFromMC(obj)

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
func UpdateColumn(obj IBase, f interface{}) error {
	var v interface{}
	switch f.(type) {
	case []string:
		s := db.NewScope(obj)
		columns := map[string]interface{}{}
		for _, column := range f.([]string) {
			if f, ok := s.FieldByName(column); !ok {
				return fmt.Errorf("filed:%s not exists", column)
			} else {
				columns[column] = f.Field.Interface()
			}
		}
		v = columns
	case map[string]interface{}:
		v = f
	case IBase:
		v = f
	default:
		return fmt.Errorf("not support")
	}

	obj.AddAfterSaveCallback("_tx_delete_mc", func() {
		DeleteFromMC(obj)
	})
	err := db.Model(obj).Updates(v).Error
	if err != nil {
		return err
	}
	DeleteFromMC(obj)
	return nil
}

