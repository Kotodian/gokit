package prometheus

import (
	"bytes"
	"fmt"
	"text/template"

	protocol "github.com/Kotodian/protocol/golang/prometheus"
)

var templateFiles map[string]*template.Template

func init() {
	//template.ParseFiles()
	templateFiles = make(map[string]*template.Template)
}

func RegisterTemplate(alertType string, f string) (err error) {
	if templateFiles[alertType], err = template.ParseFiles(f); err != nil {
		return
	}
	return
}

func Render(alert *protocol.Alert) (string, error) {
	var tpl bytes.Buffer
	t, ok := templateFiles[alert.Labels["alertname"]]
	if !ok {
		return fmt.Sprintf("%s=%v", alert.Labels["alertname"], alert.Annotations["value"]), nil
	}
	if err := t.Execute(&tpl, alert); err != nil {
		return "", err
	}

	return tpl.String(), nil
	// return templateFiles[alert.Labels["alertname"]].
}
