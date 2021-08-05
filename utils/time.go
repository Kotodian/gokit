package utils

import (
	"fmt"
	"time"
)

//LastMonth 上个月
func LastMonth(t time.Time) time.Time {
	deltaDay := 0
	for {
		lastMonth := t.AddDate(0, -1, deltaDay)
		if lastMonth.Month() == t.Month() {
			deltaDay--
			continue
		}
		return lastMonth
	}
	return time.Time{}
}

// Now 获取1970到现在的秒数
func Now() int {
	local, err := time.LoadLocation("Local") //服务器设置的时区
	if err != nil {
		return int(time.Now().Unix())
	}
	now := time.Now().In(local)
	return int(now.Unix())
}

// MescStr 毫秒字符串
func MescStr() string {
	return fmt.Sprintf("%d", time.Now().UnixNano()/1e6)
}

// MescMsgID 毫秒一天前消息ID
func MescMsgID(days ...int) string {
	day := 1
	if len(days) > 0 {
		day = days[0]
	}
	return fmt.Sprintf("%d-0", time.Now().UnixNano()/1e6-86400000*int64(day))
}

// GetDuration 计算订单时长
func GetDuration(startTime, endTime time.Time) string {
	var c time.Duration
	c = endTime.Sub(startTime)
	ret := ""
	if h := int32(c.Hours()); h != 0 {
		ret += fmt.Sprintf("%d小时 ", h)
	}
	if m := int32(c.Minutes()); m != 0 {
		ret += fmt.Sprintf("%d分 ", m%60)
	}
	if s := int32(c.Seconds()); s != 0 {
		ret += fmt.Sprintf("%d秒", s%60)
	}
	return ret
}
