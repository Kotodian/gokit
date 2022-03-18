package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
)

type AccessVerifyRequest struct {
	DeviceSerialNumber    string `json:"equipmentSN"`
	DeviceProtocol        string `json:"deviceProtocol,omitempty"`
	DeviceProtocolVersion string `json:"deviceProtocolVersion,omitempty"`
	RequestPort           string `json:"requestPort"`
	RemoteAddress         string `json:"remoteAddress"`
	CertSerialNumber      string `json:"certSerialNumber,omitempty"`
	Username              string `json:"account_code,omitempty"`
	Password              string `json:"account_password,omitempty"`
}

type RegisterStatusRequest struct {
	DeviceSerialNumber string `json:"deviceSerialNumber"`
	TimeStamp          int64  `json:"timeStamp"`
}

type registerStatusResponse struct {
	Response
	Data *Equipment `json:"data,omitempty"`
}

type Equipment struct {
	KeepAlive  int    `json:"keepalive"`
	CoreID     string `json:"id"`
	Registered bool   `json:"registered"`
}

const defaultURL = "http://jx-esam:8080"

//const defaultURL = "http://10.43.0.51:8080"
const device = "/device"
const defaultContentType = "application/json"
const defaultVersion = "/v1"
const verify = "/verify"

// AccessVerify 设备接入校验接口
func AccessVerify(request *AccessVerifyRequest) (*Equipment, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	client := NewClient()
	r := bytes.NewReader(body)
	url := defaultURL + device + verify + defaultVersion + "/accessVerify"
	resp, err := client.Post(url, defaultContentType, r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	response := &registerStatusResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	if response.Status != 0 {
		return nil, errors.New(response.Msg)
	}

	return response.Data, nil
}

//// 设备是否注册接口
//func RegisterStatus(request *RegisterStatusRequest) (uint64, error) {
//	body, err := json.Marshal(request)
//	if err != nil {
//		return 0, err
//	}
//	client := http.DefaultClient
//	r := bytes.NewReader(body)
//	resp, err := client.Post(path.Join(defaultURL, device, defaultVersion, "registerStatus"), defaultContentType, r)
//	if err != nil {
//		return 0, err
//	}
//	defer resp.Body.Close()
//	body, err = ioutil.ReadAll(resp.Body)
//	if err != nil {
//		return 0, err
//	}
//	response := &registerStatusResponse{}
//	err = json.Unmarshal(body, &response)
//	if err != nil {
//		return 0, err
//	}
//
//	if response.Resp.Code != "0" {
//		return 0, errors.New(response.Resp.Msg)
//	}
//
//	coreID, err := strconv.ParseUint(response.Data, 10, 64)
//	if err != nil {
//		return 0, err
//	}
//	return coreID, nil
//}
//
//// 设备签发证书接口
//func GetCertificate() {
//
//}
