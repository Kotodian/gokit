package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/valyala/fasthttp"
)

func NewClient() *http.Client {
	client := http.DefaultClient
	client.Transport = http.DefaultTransport
	client.Timeout = 10 * time.Second
	return client
}

var headerContentTypeJson = []byte("application/json")

var client *fasthttp.Client

var (
	ErrBodyIsNil = errors.New("body is nil")
	// service端发生异常导致未返回数据
	ErrServicesException = errors.New("services exception")
	// 404 not found err
	ErrNotFound = errors.New("not found")
)

func Init() {
	client = &fasthttp.Client{
		ReadTimeout:         10 * time.Second,
		WriteTimeout:        10 * time.Second,
		MaxIdleConnDuration: 10 * time.Second,
		Dial: (&fasthttp.TCPDialer{
			Concurrency:      8 * 1024,
			DNSCacheDuration: 1 * time.Hour,
		}).Dial,
		MaxIdemponentCallAttempts: 7,
	}
}

func sendRequest(url string, protocol interface{}, header map[string]string) ([]byte, error) {
	reqEntityBytes, err := json.Marshal(protocol)
	if err != nil {
		return nil, err
	}
	return sendPostRequest(url, reqEntityBytes, header)
}

func sendPostRequest(url string, requestBody []byte, header map[string]string) ([]byte, error) {

	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentTypeBytes(headerContentTypeJson)
	req.SetBodyRaw(requestBody)
	for k, v := range header {
		req.Header.Set(k, v)
	}

	resp := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseResponse(resp)
		fasthttp.ReleaseRequest(req)
	}()
	err := client.DoTimeout(req, resp, 3*time.Second)

	//if err != nil {
	//	if _, know := httpConnError(err); know {
	//		return nil, err
	//	} else {
	//		return nil, err
	//	}
	//}
	if err != nil {
		return nil, err
	}

	if statusCode := resp.StatusCode(); statusCode != fasthttp.StatusOK {
		if statusCode == fasthttp.StatusNotFound {
			return nil, ErrNotFound
		}
		return nil, ErrServicesException
	}

	respBody := resp.Body()
	if len(respBody) == 0 {
		return nil, ErrBodyIsNil
	}

	return respBody, nil
}
