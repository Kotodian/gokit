package log

import (
	"os"

	log "github.com/sirupsen/logrus"
)

// var ControllerEntry *log.Entry
var entry *log.Entry

func init() {
	//t := time.Now()
	//file, err := os.OpenFile(fmt.Sprintf("/var/log/goiot/panic_%s.log", t.Format("2006-01-02")), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	//if err != nil {
	//	panic(err)
	//}
	//if err = syscall.Dup2(int(file.Fd()), int(os.Stderr.Fd())); err != nil {
	//	panic(err)
	//}
	//
	log.SetOutput(os.Stdout)

	if level := os.Getenv("LOG_LEVEL"); level == "debug" {
		log.SetLevel(log.DebugLevel)
	}
	// log.SetLevel(log.DebugLevel)
	// log.SetFormatter(&log.TextFormatter{})
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	//hook, err := fluent.NewWithConfig(fluent.Config{
	//	FluentHost: "192.168.0.20",
	//	FluentPort: 5170,
	//})
	//if err != nil {
	//	panic(err)
	//}
	//log.AddHook(hook)

	// fmt.Println("", os.Getenv("ALIYUN_ACCESS_KEY"))
	//log.Info("ALIYUN_ACCESS_KEY:", slsAccessKeyID)
	//slsAccessKeyID := os.Getenv("ALIYUN_ACCESS_KEY")
	//if slsAccessKeyID != "" {
	//	log.AddHook(NewAliyunSLSLogHookFromEnv())
	//}
	//entry = log.WithFields(log.Fields{
	//	"module": "aliyun-sls",
	//})
}

// func InitControllerEntry(request_id string, ip string, userID string) {
// 	ControllerEntry = log.WithFields(log.Fields{
// 		"request_id": request_id,
// 		"ip":         ip,
// 		"user_id":    userID,
// 	})
// }
