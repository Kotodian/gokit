package api

import (
	"encoding/json"
	"errors"

	"github.com/Kotodian/gokit/datasource"
)

type PushIntervalRequest struct {
	EquipmentID datasource.UUID `json:"equipmentId"`
}

type PushIntervalResponse struct {
	OrderPushInterval int `json:"orderPushInterval"`
}

func PushInterval(req *PushIntervalRequest) (*PushIntervalResponse, error) {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	body, err := sendPostRequest("http://jx-services:8080/equip/v1/getEquipmentCallerPushOrderInterval", reqBytes, nil)
	if err != nil {
		return nil, err
	}
	response := &PushIntervalResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

type ServiceQRCodeRequest struct {
	ConnectorId datasource.UUID `json:"connectorId"`
}

type ServiceQRCodeResponse struct {
	Response
	Data string `json:"data"`
}

func ServiceQRCode(request *ServiceQRCodeRequest) (string, error) {
	reqBytes, err := json.Marshal(request)
	if err != nil {
		return "", err
	}
	body, err := sendPostRequest("http://jx-services:8080/connector/v1/generateQRCode", reqBytes, nil)
	if err != nil {
		return "", err
	}
	response := &ServiceQRCodeResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}
	if response.Status == 1 {
		return "", errors.New(response.Msg)
	}
	return response.Data, nil
}
