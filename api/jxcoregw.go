package api

import (
	"bytes"
	"encoding/json"
)

type KickRequest struct {
	CoreID string `json:"core_id"`
	Host   string `json:"host"`
	Reason string `json:"reason"`
}

func Kick(req *KickRequest) error {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return err
	}
	client := NewClient()
	reader := bytes.NewReader(reqBytes)
	resp, err := client.Post("http://jx-coregw:8080/ac/v1/kickOffline", defaultContentType, reader)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
