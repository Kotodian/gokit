package api

import (
	"encoding/json"
	"errors"

	"github.com/Kotodian/protocol/golang/hardware/charger"
)

const (
	proto  = "http://"
	host   = "jx-coregw:8080"
	prefix = "/ac/v1"

	coregwUrlPrefix = proto + host + prefix

	HostHeader = "JX-AC-HOST"
)

type KickRequest struct {
	CoreID string `json:"core_id"`
	Host   string `json:"host"`
	Reason string `json:"reason"`
}

func Kick(req *KickRequest) error {
	url := coregwUrlPrefix + "/kickOffline"
	_, err := sendRequest(url, req, nil)
	return err
}

func Authorize(hostname, clientID string, req *charger.AuthorizeReq) error {
	url := coregwUrlPrefix + "/authorize/" + clientID
	return handleRequest(url, hostname, req)
}

func NotifyReport(hostname, clientID string, req *charger.NotifyReportReq) error {
	url := coregwUrlPrefix + "/notifyReport/" + clientID

	return handleRequest(url, hostname, req)
}

func DeviceRegistration(hostname, clientID string, req *charger.DeviceRegistrationReq) error {
	url := coregwUrlPrefix + "/deviceRegistration/" + clientID
	return handleRequest(url, hostname, req)
}

func BootNotification(hostname, clientID string, req *charger.BootNotificationReq) error {
	url := coregwUrlPrefix + "/bootNotification/" + clientID
	return handleRequest(url, hostname, req)
}

func Heartbeat(hostname, clientID string, req *charger.HeartbeatReq) error {
	url := coregwUrlPrefix + "/heartbeat/" + clientID

	return handleRequest(url, hostname, req)
}

func StatusNotification(hostname, clientID string, req *charger.StatusNotificationReq) error {
	url := coregwUrlPrefix + "/statusNotification/" + clientID

	return handleRequest(url, hostname, req)
}

func ReportChargingProfile(hostname, clientID string, req *charger.ReportChargingProfilesReq) error {
	url := coregwUrlPrefix + "/reportChargingProfile/" + clientID

	return handleRequest(url, hostname, req)
}

func LogStatusNotification(hostname, clientID string, req *charger.LogStatusNotificationReq) error {
	url := coregwUrlPrefix + "/logStatusNotification/" + clientID

	return handleRequest(url, hostname, req)
}

func ReservationStatusUpdate(hostname, clientID string, req *charger.ReservationStatusUpdateReq) error {
	url := coregwUrlPrefix + "/reservationStatusUpdate/" + clientID
	return handleRequest(url, hostname, req)
}

func TransactionEventEnd(hostname, clientID string, req *charger.StopTransactionReq) error {
	url := coregwUrlPrefix + "/transactionEventEnd/" + clientID
	return handleRequest(url, hostname, req)
}

func TransactionEventEndOffline(hostname, clientID string, req *charger.TransactionReq) error {
	url := coregwUrlPrefix + "/transactionEventEndOffline/" + clientID
	return handleRequest(url, hostname, req)
}

func TransactionEventStart(hostname, clientID string, req *charger.StartTransactionReq) error {
	url := coregwUrlPrefix + "/transactionEventStart/" + clientID
	return handleRequest(url, hostname, req)
}

func FirmwareStatusNotification(hostname, clientID string, req *charger.FirmwareStatusNotificationReq) error {
	url := coregwUrlPrefix + "/firmwareStatusNotification/" + clientID
	return handleRequest(url, hostname, req)
}

func NotifyEvent(hostname, clientID string, req *charger.WarningReq) error {
	url := coregwUrlPrefix + "/notifyEvent/" + clientID
	return handleRequest(url, hostname, req)
}

func handleRequest(url, hostname string, req interface{}) error {
	message, err := sendRequest(url, req, map[string]string{HostHeader: hostname})
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
