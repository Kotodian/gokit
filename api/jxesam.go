package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
)

type AccessVerifyRequest struct {
	DeviceSerialNumber     string `json:"deviceSerialNumber"`
	DeviceProtocol         string `json:"deviceProtocol"`
	DeviceProtocolVersion  string `json:"deviceProtocolVersion"`
	RequestPort            string `json:"requestPort"`
	RemoteAddress          string `json:"remoteAddress"`
	CertDeviceSerialNumber string `json:"certDeviceSerialNumber"`
	Username               string `json:"username"`
	Password               string `json:"password"`
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
	Resp response `json:",inline"`
	Data string   `json:"data,omitempty"`
}

const defaultURL = "http://jx-esam:8080"
const device = "device"
const defaultContentType = "application/json"
const defaultVersion = "v1"

// 设备接入校验接口
func AccessVerify(request *AccessVerifyRequest) (uint64, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return 0, err
	}
	client := http.DefaultClient
	r := bytes.NewReader(body)
	resp, err := client.Post(path.Join(defaultURL, device, defaultVersion, "accessVerify"), defaultContentType, r)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	response := &registerStatusResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, err
	}

	if response.Resp.Code != "0" {
		return 0, errors.New(response.Resp.Msg)
	}

	coreID, err := strconv.ParseUint(response.Data, 10, 64)
	if err != nil {
		return 0, err
	}
	return coreID, nil
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
