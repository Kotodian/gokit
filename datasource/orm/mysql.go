package orm

import (
	"fmt"
	"go.uber.org/zap"
	"os"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"moul.io/zapgorm2"
)

const (
	EnvMaxIdleConns = "MYSQL_MAX_IDLE_CONNS"
	EnvMaxOpenConns = "MYSQL_MAX_OPEN_CONNS"
)

var mysqlDB *gorm.DB

func InitMysql(dns string, logger *zap.Logger) {
	var err error
	mysqlDB, err = NewMysql(dns, logger)
	if err != nil {
		panic("init mysql error:" + err.Error())
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

func InitMysqlWithEnvAndDB(db string, logger *zap.Logger) {
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
	InitMysql(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, passwd, host, port, db), logger)
}

func GetMysql() *gorm.DB {
	return mysqlDB
}

func NewMysql(dns string, logger *zap.Logger) (*gorm.DB, error) {
	condition := "timeout=60s&parseTime=true&charset=utf8mb4,utf8&loc=Local"
	if strings.Contains(dns, "?") {
		dns = dns + "&" + condition
	} else {
		dns = dns + "?" + condition
	}
	db, err := gorm.Open(mysql.Open(dns), &gorm.Config{
		Logger: zapgorm2.New(logger),
	})
	if err != nil {
		return nil, err
	}

	d, _ := db.DB()
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
