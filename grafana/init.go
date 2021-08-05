package grafana

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Kotodian/gokit/boot"
	"github.com/parnurzeal/gorequest"
)

var apiKey string

func init() {
	boot.RegisterInit("grafana", func() error {
		apiKey = os.Getenv("GRAFANA_API_KEY")
		return nil
	})
}

func Request(req *gorequest.SuperAgent, params interface{}, out interface{}) error {
	//req.Debug = true
	header := http.Header{}
	header.Add("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	// req := gorequest.New()
	// req.Method = method
	req.Header = header
	// req.Url = url
	resp, body, errs := req.
		Timeout(time.Second * 3).
		Type("json").
		Send(params).
		End()
	// fmt.Println("\n\n", resp, body, errs)
	if len(errs) > 0 {
		return fmt.Errorf("%v", errs)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("http response code=%d", resp.StatusCode)
	}
	if out != nil {
		err := json.Unmarshal([]byte(body), out)
		if err != nil {
			return err
		}
	}
	return nil
}
