package nacos

import (
	"net/url"
	"testing"
)

func TestCreateNamingClient(t *testing.T) {
	_, err := CreateNamingClient(5000,
		"8763a0db-d780-4b5e-a6e3-7299a299e845",
		"csms",
		"FsUGjIdnXcNflJdk",
		url.URL{
			Scheme: "http",
			Host:   "10.43.0.12:8848",
		},
	)
	if err != nil {
		t.Error(err)
		return
	}
	// client.RegisterInstance()
}

func TestCreateConfigClient(t *testing.T) {
	_, err := CreateConfigClient(5000,
		"8763a0db-d780-4b5e-a6e3-7299a299e845",
		"csms",
		"FsUGjIdnXcNflJdk",
		url.URL{
			Scheme: "http",
			Host:   "10.43.0.12:8848",
		},
	)
	if err != nil {
		t.Error(err)
		return
	}
	// client.GetConfig()	
}
