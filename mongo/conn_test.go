package mongo

import (
	"os"
	"testing"
)

func TestConn(t *testing.T) {
	os.Setenv("MONGO_AUTH_DB", "admin")
	os.Setenv("MONGO_DB", "jx-csms")
	os.Setenv("MONGO_USER", "root")
	os.Setenv("MONGO_PASSWD", "abc123,.")
	os.Setenv("MONGO_HOST", "10.43.0.13")
	os.Setenv("MONGO_PORT", "27017")
	InitEnv()
	_, err := Connect()
	if err != nil {
		t.Error(err)
		return
	}
}
