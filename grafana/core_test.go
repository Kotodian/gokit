package grafana

import "testing"

func TestCreateOrg(t *testing.T) {
	if _, err := CreateOrganisation("aaaaaaaaaaaaa"); err != nil {
		t.Error(err)
	}
}
