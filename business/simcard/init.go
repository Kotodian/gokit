package simcard

import "github.com/Kotodian/gokit/datasource"

type Queue struct {
	OperatorID datasource.UUID `json:"operator_id,string"`
	ICCID      string          `json:"iccid"`
}
