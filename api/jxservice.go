package api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

type pushIntervalRequest struct {
}

type pushIntervalResponse struct {
}

func PushInterval(req *pushIntervalRequest) (*pushIntervalResponse, error) {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	client := http.DefaultClient
	client.Timeout = 10 * time.Second
	reader := bytes.NewReader(reqBytes)
	resp, err := client.Post("http://jx-services:8080/", defaultContentType, reader)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	response := &pushIntervalResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}