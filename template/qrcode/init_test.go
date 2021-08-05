package qrcode

import "testing"

func Test_Add(t *testing.T) {
	tplStr := "{Domain}/g/{OperatorID}/{EvseID}-{ConnectorNo}"
	t.Log("--->valid: ", Valid(tplStr))

	ql := New().Domain("http://goiot.net").OperatorID("1111").EvseID("3333").ConnectorNo("1").ConnectorID("000")
	str, err := ql.Format(tplStr)
	t.Logf("--->qrocde:[%s] err:[%+v] ", str, err)
}
