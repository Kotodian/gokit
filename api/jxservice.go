package api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"

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
	client := NewClient()
	reader := bytes.NewReader(reqBytes)
	resp, err := client.Post("http://jx-services:8080/equip/v1/getEquipmentCallerPushOrderInterval", defaultContentType, reader)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
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
