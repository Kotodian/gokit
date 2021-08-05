package codec

import (
	"encoding/hex"
	"fmt"
	"testing"
)

type byte10 [10]byte

func (b byte10) String() string {
	return fmt.Sprintf("%s", string(b[:]))
}

type User struct {
	Name byte10
	Age  uint8
}
type TT struct {
	// Flag interface{}
	Flag uint8

	DD interface{}
	// FF []User
}

type TF struct {
	Start  uint32
	Num    uint8
	Users  []User
	Param1 int8
	Param2 uint16
	Param3 uint32
	Stop   uint64
	Param4 int32
}
type TS struct {
	Flag uint8
	VV   []uint16
}

func Test_codec_simple(t *testing.T) {
	in := &User{
		Name: [10]byte{0x67, 0x68, 0x69},
		Age:  18,
	}
	byt, err := Marshal(in)
	if err != nil {
		t.Fatal("----->", err)
	}
	t.Logf("--->[%v]\r\n", byt)

	out := &User{}
	if err := Unmarshal(byt, out); err != nil {
		t.Fatal("--->", err.Error())
	}
	t.Logf("--->[%v]", out)
}
func Test_codec_complex(t *testing.T) {
	// in := &TT{
	// 	Flag: 1,
	// }

	// // in.FF = append(in.FF, user)
	// in.DD = &TF{}
	// in.DD.(*TF).Start = 123
	// in.DD.(*TF).Num = 5
	// in.DD.(*TF).Param1 = -2
	// in.DD.(*TF).Param2 = 6666
	// in.DD.(*TF).Param3 = 8888
	// in.DD.(*TF).Stop = 123

	// for i := 0; i < 5; i++ {
	// 	user := User{}
	// 	user.Age = uint8(i)
	// 	copy(user.Name[:], []byte(fmt.Sprintf("name-%d", i))[:])
	// 	in.DD.(*TF).Users = append(in.DD.(*TF).Users, user)
	// }
	// byt, _ := Marshal(in)
	// t.Logf("--->marshal[%x]\r\n", byt)

	byt, _ := hex.DecodeString("017b000000056e616d652d3000000000006e616d652d3100000000016e616d652d3200000000026e616d652d3300000000036e616d652d340000000004fe0a1ab82200007b00000000000000")
	out := &TT{
		DD: &TF{},
	}
	if err := Unmarshal(byt, out); err != nil {
		t.Error("--->err", err.Error())
	} else {
		t.Logf("--->unmarshal[%v][%+v]", out, out.DD)
	}
}

func Test_codec_xxx(t *testing.T) {
	in := &TS{
		Flag: 1,
	}
	in.VV = append(in.VV, 2)
	in.VV = append(in.VV, 4)

	byt, _ := Marshal(in)
	t.Logf("--->[%v]\r\n", byt)

	out := &TS{}
	if err := Unmarshal(byt, out); err != nil {
		t.Error("--->", err.Error())
	} else {
		t.Logf("--->[%v]", out)
	}
}

// TransactionReq  CMD=202	充电桩上报充电记录信息
type TransactionReq struct {
	Elecs [48]uint16 // 		各个时段充电电量 0.01kwh
}

func Test_Slice(t *testing.T) {
	out := &TransactionReq{}
	buf, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000003a021d000000000000000000000000000000000000000000000000000000")

	if err := Unmarshal(buf, out); err != nil {
		t.Error("--->", err.Error())
	} else {
		t.Logf("--->[%v]", out)
	}
}
