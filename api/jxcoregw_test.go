package api

import (
	"strconv"
	"testing"
	"time"

	"github.com/Kotodian/protocol/golang/hardware/charger"
	"github.com/Kotodian/protocol/interfaces"
	"github.com/stretchr/testify/assert"
)

func TestKick(t *testing.T) {
	Init()
	req := &KickRequest{
		Host:   ocppDefaultHost,
		Reason: "why",
		CoreID: testCoreID,
	}
	err := Kick(req)
	if err != nil {
		t.Error(err)
		return
	}
}

const testCoreID = "336379858853895"
const ocppDefaultHost = "jx-ac-ocpp-cluster-0"

func TestAuthorize(t *testing.T) {

}

func TestDeviceRegistration(t *testing.T) {
	Init()
	req := &charger.DeviceRegistrationReq{
		SerialNumber:    "T18230000qk",
		PartNumber:      "NJC0P121B0",
		VendorId:        "MT01CNJQX0",
		FirmwareVersion: "0.0.1",
		Reason:          "Powerup",
		RemoteAddress:   "127.0.0.1",
	}
	err := DeviceRegistration(ocppDefaultHost, testCoreID, req)
	assert.Nil(t, err)

	err = Kick(&KickRequest{Reason: "websocket关闭连接", CoreID: testCoreID})
	assert.Nil(t, err)
}

func TestBootNotification(t *testing.T) {
	Init()
	req := &charger.BootNotificationReq{
		SerialNumber:    "T18220000qk",
		PartNumber:      "NJC0P121B0",
		VendorId:        "MT01CNJQX0",
		FirmwareVersion: "0.0.1",
		Reason:          "Powerup",
		RemoteAddress:   "127.0.0.1",
		Controllee:      "jx-ac-ocpp-cluster-0",
	}
	err := BootNotification(ocppDefaultHost, testCoreID, req)
	assert.Nil(t, err)

	err = Kick(&KickRequest{Reason: "websocket关闭连接", CoreID: testCoreID, Host: ocppDefaultHost})
	assert.Nil(t, err)
}

func TestHeartbeat(t *testing.T) {
	Init()
	req := &charger.HeartbeatReq{
		RemoteAddress: "127.0.0.1",
	}
	err := Heartbeat(ocppDefaultHost, testCoreID, req)
	assert.Nil(t, err)

	err = Kick(&KickRequest{Reason: "websocket关闭连接", CoreID: testCoreID, Host: ocppDefaultHost})
	assert.Nil(t, err)

}

func TestStatusNotification(t *testing.T) {
	Init()
	onlineReq := &charger.HeartbeatReq{
		RemoteAddress: "127.0.0.1",
	}
	err := Heartbeat(ocppDefaultHost, testCoreID, onlineReq)
	assert.Nil(t, err)
	req := &charger.StatusNotificationReq{
		EvseId:         "1",
		ConnectorId:    "2",
		ConnectorState: int32(interfaces.KindConnectorStateAvailable),
	}
	err = StatusNotification(ocppDefaultHost, testCoreID, req)
	assert.Nil(t, err)

	err = Kick(&KickRequest{Reason: "websocket关闭连接", CoreID: testCoreID, Host: ocppDefaultHost})
	assert.Nil(t, err)
}

func TestNotifyEvent(t *testing.T) {
	Init()
	onlineReq := &charger.HeartbeatReq{
		RemoteAddress: "127.0.0.1",
	}
	err := Heartbeat(ocppDefaultHost, testCoreID, onlineReq)
	assert.Nil(t, err)

	req := &charger.WarningReq{
		Code:          0xff0d07,
		Time:          time.Now().Unix(),
		RemoteAddress: "127.0.0.1",
	}
	err = NotifyEvent(ocppDefaultHost, testCoreID, req)
	assert.Nil(t, err)
	err = Kick(&KickRequest{Reason: "websocket关闭连接", CoreID: testCoreID, Host: ocppDefaultHost})
	assert.Nil(t, err)
}

func TestNotifyReport(t *testing.T) {
	Init()
	onlineReq := &charger.HeartbeatReq{
		RemoteAddress: "127.0.0.1",
	}
	err := Heartbeat(ocppDefaultHost, testCoreID, onlineReq)
	assert.Nil(t, err)

	req := &charger.NotifyReportReq{
		ReportData: []*charger.ReportDataType{
			{
				Component: &charger.ComponentType{
					Name: "ChargingStation",
				},
				Variable: &charger.VariableType{
					Name: "AuthFree",
				},
				VariableAttribute: []*charger.VariableAttributeType{
					{
						Value:              "0",
						VariableMutability: charger.VariableAttributeType_ReadWrite,
					},
				},
			},
		},
	}
	err = NotifyReport(ocppDefaultHost, testCoreID, req)

	assert.Nil(t, err)
	err = Kick(&KickRequest{Reason: "websocket关闭连接", CoreID: testCoreID, Host: ocppDefaultHost})
	assert.Nil(t, err)
}

func TestTransactionEvent(t *testing.T) {
	Init()
	onlineReq := &charger.HeartbeatReq{
		RemoteAddress: "127.0.0.1",
	}
	err := Heartbeat(ocppDefaultHost, testCoreID, onlineReq)
	assert.Nil(t, err)

	statusReq := &charger.StatusNotificationReq{
		EvseId:         "1",
		ConnectorId:    "1",
		ConnectorState: int32(interfaces.KindConnectorStateOccupied),
	}
	err = StatusNotification(ocppDefaultHost, testCoreID, statusReq)
	assert.Nil(t, err)
	transactionId := strconv.FormatInt(time.Now().Unix(), 10)
	req := &charger.StartTransactionReq{
		EvseId:        "1",
		ConnectorId:   "1",
		Timestamp:     time.Now().Unix(),
		MeterStart:    0,
		CurrentA:      0,
		CurrentB:      0,
		CurrentC:      0,
		VoltageA:      0,
		VoltageB:      0,
		VoltageC:      0,
		Soc:           0,
		BatteryTemp:   0,
		Power:         0,
		TransactionId: transactionId,
		Mode:          charger.AuthorizationMode_AM_LocalIdentityCard,
		IdData: &charger.IdToken{
			Id: "0000000",
		},
		Tariff: &charger.Tariff{
			Id:     -1,
			Valley: 0,
			Sharp:  0,
			Flat:   0,
			Peak:   0,
		},
		ConnectorChargingState: charger.ChargerStatus_CHS_Charging,
	}
	err = TransactionEventStart(ocppDefaultHost, testCoreID, req)
	assert.Nil(t, err)

	err = TransactionEventUpdate(ocppDefaultHost, testCoreID, &charger.ChargingInfoReq{
		EvseId:      "1",
		ConnectorId: "1",
		Timestamp:   time.Now().Unix(),
		Electricity: 0,
		CurrentA:    0,
		CurrentB:    0,
		CurrentC:    0,
		VoltageA:    0,
		VoltageB:    0,
		VoltageC:    0,
		Soc:         0,
		BatteryTemp: 0,
		Power:       0,
		RecordId:    transactionId,
		Tariff: &charger.Tariff{
			Id:     -1,
			Valley: 0,
			Sharp:  0,
			Flat:   0,
			Peak:   0,
		},
	})
	assert.Nil(t, err)

	stopTransactionReq := &charger.StopTransactionReq{
		EvseId:        "1",
		ConnectorId:   "1",
		Timestamp:     time.Now().Unix(),
		MeterStop:     0,
		CurrentA:      0,
		CurrentB:      0,
		CurrentC:      0,
		VoltageA:      0,
		VoltageB:      0,
		VoltageC:      0,
		Soc:           0,
		BatteryTemp:   0,
		Power:         0,
		TransactionId: transactionId,
		Tariff: &charger.Tariff{
			Id:     -1,
			Valley: 0,
			Sharp:  0,
			Flat:   0,
			Peak:   0,
		},
		Reason:   charger.StopReason_SR_Normal,
		RecordId: transactionId,
	}
	err = TransactionEventEnd(ocppDefaultHost, testCoreID, stopTransactionReq)
	assert.Nil(t, err)

	err = Kick(&KickRequest{Reason: "websocket关闭连接", CoreID: testCoreID, Host: ocppDefaultHost})
	assert.Nil(t, err)

}

func TestFirmwareStatusNotification(t *testing.T) {

}

func TestLogStatusNotification(t *testing.T) {

}

func TestReservationStatusUpdate(t *testing.T) {

}

func TestReportChargingProfile(t *testing.T) {

}

func TestTransactionEventEndOffline(t *testing.T) {

}
