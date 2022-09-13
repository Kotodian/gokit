package api

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/Kotodian/protocol/golang/hardware/charger"
)

const (
	proto  = "http://"
	host   = "jx-coregw:8080"
	prefix = "/ac/v1"

	coregwUrlPrefix = proto + host + prefix
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

func Authorize(clientID string, req *charger.AuthorizeReq) error {
	url := coregwUrlPrefix + "/authorize/" + clientID

	message, err := sendRequest(url, req, nil)

	if err != nil {
		return err
	}
	resp := &Response{}
	err = json.Unmarshal(message, resp)
	if err != nil {
		return err
	}

	if resp.Status == 1 {
		return errors.New(resp.Msg)
	}
	return nil
}

func NotifyReport(clientID string, req *charger.NotifyReportReq) error {
	url := coregwUrlPrefix + "/notifyReport/" + clientID

	message, err := sendRequest(url, req, nil)

	if err != nil {
		return err
	}

	resp := &Response{}
	err = json.Unmarshal(message, resp)
	if err != nil {
		return err
	}

	if resp.Status == 1 {
		return errors.New(resp.Msg)
	}
	return nil
}

func DeviceRegistration(clientID string, req *charger.DeviceRegistrationReq) error {
	url := coregwUrlPrefix + "/deviceRegistration/" + clientID
	message, err := sendRequest(url, req, nil)
	if err != nil {
		return err
	}
	resp := &Response{}
	err = json.Unmarshal(message, resp)
	if err != nil {
		return err
	}

	if resp.Status == 1 {
		return errors.New(resp.Msg)
	}
	return nil
}

func BootNotification(clientID string, req *charger.BootNotificationReq) error {
	url := coregwUrlPrefix + "/bootNotification/" + clientID

	message, err := sendRequest(url, req, nil)
	if err != nil {
		return err
	}
	resp := &Response{}
	err = json.Unmarshal(message, resp)
	if err != nil {
		return err
	}

	if resp.Status == 1 {
		return errors.New(resp.Msg)
	}
	return nil
}

func Heartbeat(clientID string, req *charger.HeartbeatReq) error {
	url := coregwUrlPrefix + "/heartbeat/" + clientID

	message, err := sendRequest(url, req, nil)
	if err != nil {
		return err
	}
	resp := &Response{}
	err = json.Unmarshal(message, resp)
	if err != nil {
		return err
	}

	if resp.Status == 1 {
		return errors.New(resp.Msg)
	}
	return nil
}

func ReportChargingProfile(clientID string, req *charger.ReportChargingProfilesReq) error {
	url := coregwUrlPrefix + "/reportChargingProfile/" + clientID

	message, err := sendRequest(url, req, nil)
	if err != nil {
		return err
	}
	resp := &Response{}
	err = json.Unmarshal(message, resp)
	if err != nil {
		return err
	}

	if resp.Status == 1 {
		return errors.New(resp.Msg)
	}
	return nil
}

func LogStatusNotification(clientID string, req *charger.LogStatusNotificationReq) error {
	url := coregwUrlPrefix + "/logStatusNotification/" + clientID

	message, err := sendRequest(url, req, nil)
	if err != nil {
		return err
	}
	resp := &Response{}
	err = json.Unmarshal(message, resp)
	if err != nil {
		return err
	}

	if resp.Status == 1 {
		return errors.New(resp.Msg)
	}
	return nil
}

func ReservationStatusUpdate(clientID string, req *charger.ReservationStatusUpdate) error {
	url := coregwUrlPrefix + "/reservationStatusUpdate/" + clientID
	message, err := sendRequest(url, req, nil)
	if err != nil {
		return err
	}
	resp := &Response{}
	err = json.Unmarshal(message, resp)
	if err != nil {
		return err
	}

	if resp.Status == 1 {
		return errors.New(resp.Msg)
	}
	return nil
}

func TransactionEventEnd(clientID string, req *charger.StopTransactionReq) error {
	url := coregwUrlPrefix + "/transactionEventEnd/" + clientID
	message, err := sendRequest(url, req, nil)
	if err != nil {
		return err
	}
	resp := &Response{}
	err = json.Unmarshal(message, resp)
	if err != nil {
		return err
	}

	if resp.Status == 1 {
		return errors.New(resp.Msg)
	}
	return nil
}

func TransactionEventEndOffline(clientID string, req *charger.TransactionReq) error {
	url := coregwUrlPrefix + "/transactionEventEndOffline/" + clientID
	message, err := sendRequest(url, req, nil)
	if err != nil {
		return err
	}
	resp := &Response{}
	err = json.Unmarshal(message, resp)
	if err != nil {
		return err
	}

	if resp.Status == 1 {
		return errors.New(resp.Msg)
	}
	return nil
}

func TransactionEventStart(clientID string, req *charger.StartTransactionReq) error {
	url := coregwUrlPrefix + "/transactionEventStart/" + clientID
	message, err := sendRequest(url, req, nil)
	if err != nil {
		return err
	}
	resp := &Response{}
	err = json.Unmarshal(message, resp)
	if err != nil {
		return err
	}

	if resp.Status == 1 {
		return errors.New(resp.Msg)
	}
	return nil
}

func FirmwareStatusNotification(clientID string, req *charger.FirmwareStatusNotificationReq) error {
	url := coregwUrlPrefix + "/firmwareStatusNotification/" + clientID
	message, err := sendRequest(url, req, nil)
	if err != nil {
		return err
	}
	resp := &Response{}
	err = json.Unmarshal(message, resp)
	if err != nil {
		return err
	}

	if resp.Status == 1 {
		return errors.New(resp.Msg)
	}
	return nil
}

func NotifyEvent(clientID string, req *charger.WarningReq) error {
	url := coregwUrlPrefix + "/notifyEvent/" + clientID
	message, err := sendRequest(url, req, nil)
	if err != nil {
		return err
	}
	resp := &Response{}
	err = json.Unmarshal(message, resp)
	if err != nil {
		return err
	}

	if resp.Status == 1 {
		return errors.New(resp.Msg)
	}
	return nil
}
