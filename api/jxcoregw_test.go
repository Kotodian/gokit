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
