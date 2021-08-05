package bosun

import (
	"fmt"
	"time"
)

const (
	BosunTimeFormat string = "2006-01-02 15:04:05.99999999 -0700 MST"
)

type BosunTime time.Time

func (bt BosunTime) MarshalJSON() ([]byte, error) {
	//fmt.Println("bosun time", bt, time.Time(bt).Local().Format("2006-01-02 15:04:05"))
	rs := []byte(fmt.Sprintf(`"%s"`, time.Time(bt).Local().Format(BosunTimeFormat)))
	return rs, nil
}

func (bt *BosunTime) UnmarshalJSON(b []byte) (err error) {
	timeStr := string(b[1 : len(b)-1])
	// fmt.Println(BosunTimeFormat, timeStr)
	var t time.Time
	t, err = time.Parse(BosunTimeFormat, timeStr)
	if err != nil {
		return
	}
	//local, err := time.LoadLocation("Local") // 服务器设置的时区
	//if err != nil {
	//	return err
	//}
	//*bt = BosunTime(t.In(local))
	*bt = BosunTime(t)
	return nil
}

func (bt BosunTime) Time() time.Time {
	return time.Time(bt)
}

type ReqAction struct {
	Type    string     `json:"type"`
	Message string     `json:"message"`
	Keys    []string   `json:"keys"`
	Ids     []int64    `josn:"ids"`
	Notify  bool       `json:"notify"`
	User    string     `json:"user"`
	Time    *time.Time `json:"time"`
}
