package api

import (
	"encoding/json"

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
