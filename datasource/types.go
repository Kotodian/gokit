package datasource

import (
	"bytes"
	"database/sql/driver"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/tidwall/gjson"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"reflect"
	"strconv"
)

func JsonValue(l interface{}) (driver.Value, error) {
	bytes, err := json.Marshal(l)
	return string(bytes), err
}

func JsonScan(input interface{}, l interface{}) (err error) {
	switch value := input.(type) {
	case string:
		if value == "" {
			return
		}
		err = json.Unmarshal([]byte(value), l)
	case []byte:
		if len(value) == 0 {
			return
		}
		err = json.Unmarshal(value, l)
	default:
		err = errors.New("not supported type")
	}
	return
}

type KindPassword string

func (c KindPassword) Encode() (string, error) {
	s, err := bcrypt.GenerateFromPassword([]byte(c), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(s), nil
}

func (c KindPassword) IsMatch(s string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(c), []byte(s))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
			//c.Error(http.StatusBadRequest, "密码不正确")
		} else {
			return false, err
		}
	}
	return true, nil
}

// type KindPrice float64

// func (c KindPrice) Value() (driver.Value, error) {
// 	fmt.Println("---------->kp v", c)
// 	return int64(float64(c) * 100), nil
// }

// func (c *KindPrice) Scan(input interface{}) error {
// 	fmt.Println("---------->kp s", *c)

// 	switch input.(type) {
// 	case int64:
// 		*c = KindPrice(input.(int64)) / 100
// 	case float64:
// 		*c = KindPrice(input.(float64)) / 100
// 	}
// 	return nil
// }

// func (c KindPrice) Float64() float64 {
// 	return float64(c)
// }

// func (c KindPrice) Int32() int32 {
// 	return int32(float64(c) * 1000)
// }

type UUID int64

func ParseUUID(in string) (UUID, error) {
	id, err := strconv.ParseUint(in, 10, 64)
	if err != nil {
		return 0, err
	}
	return UUID(id), nil
}

func (u UUID) String() string {
	return strconv.FormatInt(int64(u), 10)
}

func (u UUID) Int64() int64 {
	return int64(u)
}

func (u UUID) Timestamp() int64 {
	return 1582136402000 + int64(u>>12)
}

//
//func (u UUID) MarshalJSON() ([]byte, error) {
//	return json.Marshal(u.String())
//}
//
//func (u *UUID) UnmarshalJSON(b []byte) (err error) {
//	var ext UUID
//	if b[0] == 34 {
//		ext, err = ParseUUID(string(b[1 : len(b)-1]))
//	} else {
//		ext, err = ParseUUID(string(b))
//	}
//	if err != nil {
//		return err
//	}
//	*u = ext
//	return nil
//}

type KindProtocol struct {
	Type     reflect.Type
	Template *template.Template
}

var KindProtocolMap map[string]*KindProtocol

func init() {
	KindProtocolMap = make(map[string]*KindProtocol)
	// gob.Register(UUID(1))
	gob.Register(UUID(1))
}

func RegisterKindProtoMap(s string, obj proto.Message, tpl string) {
	name := reflect.TypeOf(obj).Elem().Name()
	tmpl, _ := template.New(name).Parse(tpl)
	o := &KindProtocol{
		Type:     reflect.TypeOf(obj),
		Template: tmpl,
	}
	KindProtocolMap[name] = o
}

type KindProtocolObj struct {
	Type    string        `json:"type"`
	Message proto.Message `json:"-"`
	Bytes   []byte        `json:"bytes"`
}

// MarshalJSON
// 序列化成Json，一般用在cache里
func (c KindProtocolObj) MarshalJSON() ([]byte, error) {
	if c.Message == nil {
		return nil, fmt.Errorf("empty message")
	}
	var err error
	c.Type = reflect.TypeOf(c.Message).Elem().Name()
	if c.Bytes, err = proto.Marshal(c.Message); err != nil {
		return nil, err
	}
	return json.Marshal(map[string]string{
		"type":  c.Type,
		"bytes": base64.StdEncoding.EncodeToString(c.Bytes),
	})
	//return []byte(c.Type + "|" + string(c.Bytes)), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *KindProtocolObj) UnmarshalJSON(data []byte) error {
	c.Type = gjson.GetBytes(data, "type").Str
	bytes, err := base64.StdEncoding.DecodeString(gjson.GetBytes(data, "bytes").Str)
	if err != nil {
		return err
	}
	input := fmt.Sprintf("%s|%s", c.Type, string(bytes))
	return c.Scan([]byte(input))
}

// Value
// 存进数据库的值
func (c *KindProtocolObj) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	var err error
	c.Type = reflect.TypeOf(c.Message).Elem().Name()
	c.Bytes, err = proto.Marshal(c.Message)
	if err != nil {
		return nil, err
	}
	return c.Type + "|" + string(c.Bytes), nil
}

func (c *KindProtocolObj) Scan(input interface{}) error {
	*c = KindProtocolObj{
		Bytes: input.([]byte),
	}
	q := bytes.Index(c.Bytes, []byte("|"))
	if q == -1 {
		return fmt.Errorf("field value unvalid")
	} else {
		c.Type = string(c.Bytes[:q])
	}
	if f, ok := KindProtocolMap[c.Type]; ok {
		c.Message = reflect.New(f.Type.Elem()).Interface().(proto.Message)
		return proto.Unmarshal(c.Bytes[q+1:], c.Message)
	}
	return fmt.Errorf("cannot found scan callback func, type:%s", c.Type)
}

func (c KindProtocolObj) Desc() string {
	if f, ok := KindProtocolMap[c.Type]; ok {
		var s bytes.Buffer
		f.Template.Execute(&s, c.Message)
		return s.String()
	}
	return fmt.Sprintln(c.Message)
}

type KindCoordinate float64

func (c KindCoordinate) Value() (driver.Value, error) {
	return int64(float64(c) * 1000000), nil
}

func (c *KindCoordinate) Scan(input interface{}) error {
	switch input.(type) {
	case int64:
		*c = KindCoordinate(input.(int64)) / 1000000
	case float64:
		*c = KindCoordinate(input.(float64)) / 1000000
	case []byte:
		if i, err := strconv.Atoi(string(input.([]byte))); err != nil {
			return err
		} else {
			*c = KindCoordinate(float64(i) / 1000000)
		}
	}
	return nil
}

func NewUUID(id interface{}) (ret *UUID) {
	var _id UUID
	switch id.(type) {
	case int:
		_id = UUID(id.(int))
		ret = &_id
	case int64:
		_id = UUID(id.(int64))
		ret = &_id
	case int32:
		_id = UUID(id.(int32))
		ret = &_id
	case uint32:
		_id = UUID(id.(uint32))
		ret = &_id
	case uint:
		_id = UUID(id.(uint))
		ret = &_id
	case uint64:
		_id = UUID(id.(uint64))
		ret = &_id
	case UUID:
		_id = id.(UUID)
		ret = &_id
	}
	if ret == nil || *ret == 0 {
		return nil
	}
	return
}
