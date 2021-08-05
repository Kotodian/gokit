package orm_bak

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/edwardhey/gorm"
	//_ "github.com/go-sql-driver/mysql"
	_ "github.com/edwardhey/mysql"
)

const (
	EnvMaxIdleConns = "MYSQL_MAX_IDLE_CONNS"
	EnvMaxOpenConns = "MYSQL_MAX_OPEN_CONNS"
)

var mysqlDB *gorm.DB

func InitMysql(dns string) {
	var err error
	mysqlDB, err = NewMysql(dns)
	if err != nil {
		panic("init mysql error:" + err.Error())
	}
	//mysqlDB.SetLogger(logrus.WithFields(logrus.Fields{}))
	mysqlDB.Callback().Create().
		After("gorm:commit_or_rollback_transaction").
		Register("after_create_commit", afterCommitCallback)
	mysqlDB.Callback().Update().
		After("gorm:commit_or_rollback_transaction").
		Register("after_update_commit", afterCommitCallback)

	//mysqlDB.Callback().Delete().
	//	After("gorm:commit_or_rollback_transaction").
	//	Register("after_delete_commit", afterDeleteCommitCallback)
}

// afterCreateCommitCallback will invoke `AfterCreateCommit`, `AfterSaveCommit` method
func afterCommitCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		if _, ok := scope.Get("gorm:tx:objs"); !ok {
			//scope.
			scope.CallMethod("AfterSaveCommit")
		}
		//syncTableVer.Delete(utils.MD5(fmt.Sprintf("%s:ver:table:gcache:orm", scope.TableName())))
	}
}

//
//// afterDeleteCommitCallback will invoke `AfterDeleteCommit` method
//func afterDeleteCommitCallback(scope *gorm.Scope) {
//	//logrus.Info("after delete callback commit")
//	if !scope.HasError() {
//		scope.CallMethod("AfterDeleteCommit")
//	}
//}

func InitMysqlWithEnvAndDB(db string) {
	user := os.Getenv("DB_USER")
	passwd := os.Getenv("DB_PASSWD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	if host == "" {
		host = "mysql"
	}
	if port == "" {
		port = "3306"
	}
	InitMysql(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, passwd, host, port, db))
}

func GetMysql() *gorm.DB {
	return mysqlDB
}

func NewMysql(dns string) (*gorm.DB, error) {
	condition := "timeout=60s&parseTime=true&charset=utf8mb4,utf8&loc=Local"
	if strings.Contains(dns, "?") {
		dns = dns + "&" + condition
	} else {
		dns = dns + "?" + condition
	}
	db, err := gorm.Open("mysql", dns)
	if err != nil {
		return nil, err
	}
	d := db.DB()
	d.SetConnMaxLifetime(300 * time.Second)
	d.SetMaxIdleConns(getIntEnv(EnvMaxIdleConns, 100))
	d.SetMaxOpenConns(getIntEnv(EnvMaxOpenConns, 2000))
	return db, nil
}

func getIntEnv(key string, def int) int {
	envstr := os.Getenv(key)
	if envstr != "" {
		if tmp, _ := strconv.ParseInt(envstr, 10, 64); tmp > 0 {
			return int(tmp)
		}
	}
	return def
}
