package api

import "testing"

func TestKick(t *testing.T) {
	req := &KickRequest{
		Reason: "why",
		CoreID: "T1641735210",
	}
	err := Kick(req)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestAuthorize(t *testing.T) {

}

func TestNotifyReport(t *testing.T) {

}

func TestBootNotification(t *testing.T) {

}

func TestReportChargingProfile(t *testing.T) {

}

func TestLogStatusNotification(t *testing.T) {

}

func TestReservationStatusUpdate(t *testing.T) {

}

func TestTransactionEventEnd(t *testing.T) {

}

func TestTransactionEventEndOffline(t *testing.T) {

}

func TestTransactionEventStart(t *testing.T) {

}

func TestFirmwareStatusNotification(t *testing.T) {

}
