/*
 * Copyright (c) 2019.
 */

package area

import (
	"fmt"
	"os"
	"sort"
	"testing"
)

type Strings []string

//Len()
func (s Strings) Len() int {
	return len(s)
}

//Less():成绩将有低到高排序
func (s Strings) Less(i, j int) bool {
	return s[i] < s[j]
}

//Swap()
func (s Strings) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type DD struct {
	Head string
	Area Strings
}

type DDS []*DD

//Len()
func (s DDS) Len() int {
	return len(s)
}

//Less():成绩将有低到高排序
func (s DDS) Less(i, j int) bool {
	// h1, _ := strconv.ParseInt(s[i].Head, 10, 64)
	// h2, _ := strconv.ParseInt(s[j].Head, 10, 64)
	// return h1 < h2
	return s[i].Head < s[j].Head
}

//Swap()
func (s DDS) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func Test_area(t *testing.T) {
	s := GB2260.Store

	areas := make(map[string][]string, 0)

	nc := 0
	for k, v := range s {
		if len(k) != 6 {
			nc++
			continue
		}
		head1 := k[:2]
		if k[4:] == "00" {
			head1 += "0000"
			if k != head1 {
				areas[head1] = append(areas[head1], fmt.Sprintf("\"%s\":\"%s\"", k, v))
			}
		}

		head2 := k[:4] + "00"
		if k != head2 {
			areas[head2] = append(areas[head2], fmt.Sprintf("\"%s\":\"%s\"", k, v))
		}
	}

	var dds DDS
	for k, v := range areas {
		dd := &DD{Head: k}
		dd.Area = append(dd.Area, v...)
		sort.Sort(dd.Area)
		dds = append(dds, dd)
	}
	sort.Sort(dds)

	fileObj, err := os.OpenFile("./area.text", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println("Failed to open the file", err.Error())
		os.Exit(2)
	}
	defer fileObj.Close()

	for k := range dds {
		content := "\r\n\"" + dds[k].Head + "\": {"
		for i := range dds[k].Area {
			content += "\r\n\t" + dds[k].Area[i] + ","
		}
		content += "\r\n},"
		if _, err := fileObj.WriteString(content); err != nil {
			fmt.Println("write error:" + err.Error())
			os.Exit(2)
		}
	}

	fmt.Printf("[%d][%d]\r\n", nc, len(areas))
}
