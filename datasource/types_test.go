package datasource

import (
	"encoding/json"
	"fmt"
	"testing"
)

func Test_UUID(t *testing.T) {
	type A struct {
		ID UUID `json:"id"`
	}
	a := A{
		ID: 1712616789004779520,
	}
	if b, err := json.Marshal(a); err != nil {
		t.Error(err)
	} else {
		fmt.Println(string(b))
		var _a A
		if err := json.Unmarshal([]byte(b), &_a); err != nil {
			t.Error(err)
		} else {
			fmt.Println(_a)
		}
	}

}
