package serverinfo

import (
	"fmt"
)

var (
	// serverinfos 服务信息
	serverinfos map[string]map[string]interface{}
)

func init() {
	//boot.RegisterInit("server_info", bootInit)
	serverinfos = map[string]map[string]interface{}{
		"open": map[string]interface{}{
			"grpc":     8031,
			"redirect": "admin",
		},
		"api": map[string]interface{}{
			"grpc":     8031, // 暂时在admin里实现
			"redirect": "admin",
		},
		"admin": map[string]interface{}{
			"http": 8030,
			"grpc": 8031,
		},
		"coregw": map[string]interface{}{
			"http": 8032,
			"grpc": 8034,
		},
		"configmap": map[string]interface{}{
			"grpc": 8035,
		},
		"customer": map[string]interface{}{
			"grpc": 8038,
		},
		"grpc-proxy": map[string]interface{}{
			"grpc": 8071,
		},
		"groupadmin": map[string]interface{}{
			"http": 8040,
			"grpc": 8041,
		},
		"ac-szunit": map[string]interface{}{
			"tcp": 8043,
		},
		"ac-sinexcel": map[string]interface{}{
			"tcp": 8045,
		},
		"ac-yunkuaichong": map[string]interface{}{
			"tcp": 8047,
		},
		"ac-nandian": map[string]interface{}{
			"tcp": 8048,
		},
		"pay": map[string]interface{}{
			"grpc": 8051,
			"http": 9092,
		},
		"message": map[string]interface{}{
			"http": 9093,
		},
		"xa": map[string]interface{}{
			"grpc": 8060,
		},
		"front": map[string]interface{}{
			"http": 9090,
		},
		"verifycode": map[string]interface{}{
			"grpc": 8052,
		},
		"init": map[string]interface{}{
			"http": 8055,
			"tcp":  8056,
		},
		"sn": map[string]interface{}{
			"grpc": 8057,
		},
		"pay-agg": map[string]interface{}{
			"http": 9094,
		},
	}

}

func GetTCPListenAddr(srvName string) string {
	if v, ok := serverinfos[srvName]; ok {
		if port, ok := v["tcp"]; ok {
			return fmt.Sprintf("0.0.0.0:%d", port)
		}
	}
	return "0.0.0.0:8070"
}

func GetGRPCListenAddr(srvName string) string {
	if v, ok := serverinfos[srvName]; ok {
		if port, ok := v["grpc"]; ok {
			return fmt.Sprintf("0.0.0.0:%d", port)
		}
	}
	return "0.0.0.0:8070"
}

func GetHTTPListenAddr(srvName string) string {
	if v, ok := serverinfos[srvName]; ok {
		if port, ok := v["http"]; ok {
			return fmt.Sprintf("0.0.0.0:%d", port)
		}
	}
	return "0.0.0.0:8090"
}

func GetHTTPAddr(srvName string) string {
	host := srvName
	if v, ok := serverinfos[srvName]; ok {
		if vv, ok := v["redirect"]; ok {
			host = vv.(string)
		}
		if port, ok := v["http"]; ok {
			return fmt.Sprintf("%s:%d", host, port)
		}
	}
	return fmt.Sprintf("%s:8090", host)
}

func GetGRPCAddr(srvName string) string {
	host := srvName
	if v, ok := serverinfos[srvName]; ok {
		if vv, ok := v["redirect"]; ok {
			host = vv.(string)
		}
		if port, ok := v["grpc"]; ok {
			return fmt.Sprintf("%s:%d", host, port)
		}
	}
	return fmt.Sprintf("%s:8070", host)
}
