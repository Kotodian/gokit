package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type AccessVerifyRequest struct {
	DeviceSerialNumber     string `json:"equipmentSN"`
	DeviceProtocol         string `json:"deviceProtocol"`
	DeviceProtocolVersion  string `json:"deviceProtocolVersion"`
	RequestPort            string `json:"requestPort"`
	RemoteAddress          string `json:"remoteAddress"`
	CertDeviceSerialNumber string `json:"certEquipmentSN"`
	CertSerialNumber       string `json:"certSerialNumber"`
	Username               string `json:"account_code"`
	Password               string `json:"account_password"`
}

type RegisterStatusRequest struct {
	DeviceSerialNumber string `json:"deviceSerialNumber"`
	TimeStamp          int64  `json:"timeStamp"`
}

type response struct {
	Status    int    `json:"status"`
	Rows      int    `json:"rows"`
	Code      string `json:"code"`
	Msg       string `json:"msg"`
	Timestamp string `json:"timestamp"`
}

type registerStatusResponse struct {
	Resp response   `json:",inline"`
	Data *Equipment `json:"data,omitempty"`
}

type Equipment struct {
	KeepAlive int    `json:"keepalive"`
	CoreID    string `json:"id"`
}

//const defaultURL = "http://jx-esam:8080"
const defaultURL = "http://10.43.0.51:8080"
const device = "/device"
const defaultContentType = "application/json"
const defaultVersion = "/v1"
const verify = "/verify"

// 设备接入校验接口
func AccessVerify(request *AccessVerifyRequest) (*Equipment, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	client := http.DefaultClient
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
	fmt.Println(string(body))
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	if response.Resp.Status != 0 {
		return nil, errors.New(response.Resp.Msg)
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
