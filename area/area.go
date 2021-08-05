/*
 * Copyright (c) 2019.
 */

package area

import (
	"strings"

	"github.com/edwardhey/gb2260"
)

var GB2260 gb2260.GB2260

func init() {
	GB2260 = gb2260.NewGB2260("2018")
}

func GetFullAreaName(division *gb2260.Division) string {
	if division == nil {
		return ""
	}
	var names []string
	stacks := division.Stack()
	for _, div := range stacks {
		names = append(names, div.Name)
	}

	return strings.Join(names, "/")
}
