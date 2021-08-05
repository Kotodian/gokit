package id

import "testing"

func Test_ID(t *testing.T) {
	id, _ := IDGenr.Next()
	t.Log("--->id:", id)

	timestamp := id >> 24
	t.Log("--->timestamp:", timestamp)

}
