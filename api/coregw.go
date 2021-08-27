package api

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type KickRequest struct {
	SN     string `json:"sn"`
	Host   string `json:"host"`
	Reason string `json:"reason"`
}

func Kick(req *KickRequest) error {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return err
	}
	client := http.DefaultClient
	reader := bytes.NewReader(reqBytes)
	resp, err := client.Post("http://jx-coregw:8080/ac/v1/kick", defaultContentType, reader)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
