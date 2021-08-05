package template

import (
	"fmt"
	"time"
)

func AuthorizationModeDesc(authMode int32) string {
	switch authMode {
	case 0: //本地即插即充启动
		return "即插即充"
	case 1: //本地管理员启动
		return "设备管理员"
	case 2: //鉴权卡刷卡本地鉴权启动
		return "本地鉴权卡"
	case 3: //鉴权卡刷卡在线鉴权启动
		return "在线鉴权卡"
	case 4: //本地钱包卡刷卡启动
		return "钱包卡"
	case 5: //车辆VIN本地鉴权启动
		return "本地VIN"
	case 6: //车辆VIN在线鉴权启动
		return "在线VIN"
	case 7: //本地通过蓝牙启动
		return "蓝牙"
	case 8: //本地通过输入校验码启动
		return "校验码"
	case 9: //远程管理员启动
		return "后台管理员"
	case 10: //
		return "远程用户启动"
	}
	return "-"
}

// TConsumeTime 计算订单时长
func TConsumeTime(orderstatus int32, startTime, endTime int64) string {
	var c time.Duration
	timeEnd := time.Now()
	if orderstatus < 0 {
		c = time.Duration(0)
		if startTime != 0 && endTime != 0 {
			c = time.Unix(endTime, 0).Sub(time.Unix(startTime, 0))
		}
	} else {
		if startTime == 0 {
			return ""
		}
		if endTime != 0 {
			timeEnd = time.Unix(endTime, 0)
		}
		c = timeEnd.Sub(time.Unix(startTime, 0))
	}
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
	return ret // ConsumeTimeDesc(c)
}
