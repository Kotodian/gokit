package api

import "testing"

func TestKick(t *testing.T) {
	req := &KickRequest{
		Reason: "why",
		SN:     "T1641735210",
	}
	err := Kick(req)
	if err != nil {
		t.Error(err)
		return
	}
}
