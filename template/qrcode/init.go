package qrcode

import (
	"bytes"
	"errors"
	"regexp"
	"text/template"
)

// 模版已支持内容
var supports = []string{"{Domain}", "{OperatorID}", "{EvseID}", "{ConnectorID}", "{ConnectorNo}"}

func Valid(str string) bool {
	tlmps := regexp.MustCompile(`{([^}]+)}`).FindAllString(str, -1)
	for _, tlmp := range tlmps {
		marker := false
		for _, v := range supports {
			if v == tlmp {
				marker = true
				break
			}
		}
		if !marker {
			return false
		}
	}
	return true
}

type tpls map[string]interface{}

func New() tpls {
	return make(tpls, 0)
}

func (t tpls) Domain(v interface{}) tpls {
	t["Domain"] = v
	return t
}

func (t tpls) OperatorID(v interface{}) tpls {
	t["OperatorID"] = v
	return t
}

func (t tpls) EvseID(v interface{}) tpls {
	t["EvseID"] = v
	return t
}

func (t tpls) ConnectorID(v interface{}) tpls {
	t["ConnectorID"] = v
	return t
}

func (t tpls) ConnectorNo(v interface{}) tpls {
	t["ConnectorNo"] = v
	return t
}

func (t tpls) Format(v string) (string, error) {
	if ok := Valid(v); !ok {
		return "", errors.New("二维码模版不支持")
	}
	var buf bytes.Buffer
	if tpl, err := template.New("q").Parse(regexp.MustCompile(`{([^}]+)}`).ReplaceAllString(v, `{{.$1}}`)); err != nil {
		return "1", err
	} else if err = tpl.Execute(&buf, map[string]interface{}(t)); err != nil {
		return "2", err
	}
	return string(buf.Bytes()), nil
}
