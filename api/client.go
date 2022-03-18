package api

import (
	"net/http"
	"time"
)


func NewClient() *http.Client {
	client := http.DefaultClient
	client.Transport = http.DefaultTransport
	client.Timeout = 10 * time.Second
	return client
}
