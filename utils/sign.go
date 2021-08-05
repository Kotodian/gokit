package utils

import (
	"compress/gzip"
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"sort"
	"strings"
)

// CheckSign 检查签名
func CheckSign(request *http.Request, key string) error {
	signature := request.Header.Get("signature")
	sign, err := HTTPSign(request, key)
	if err != nil {
		return err
	}
	if signature == sign {
		return nil
	}
	return errors.New("签名不合法")
}

// HTTPSign 基于原生的http.Request生成签名
// query+form+postbody+timeStamp
func HTTPSign(request *http.Request, key string) (string, error) {
	timeStamp := request.Header.Get("TimeStamp")
	//检验时间
	delta := Now() - StrToInt(timeStamp)
	if delta < 0 || delta > 5 {
		return "", errors.New("时间错误")
	}
	method := request.Method
	if method == "GET" {
		signStr := request.URL.Query().Encode() + timeStamp
		return Sign(signStr, key), nil
	}
	contentType := request.Header.Get("Content-Type")
	signStr := ""
	switch {
	case strings.HasPrefix(contentType, "multipart/form-data"):
		{
			signStr = signStr + request.Form.Encode()
		}
	case strings.HasPrefix(contentType, "application/x-www-form-urlencoded"):
		{
			signStr = signStr + request.Form.Encode()
		}
	case strings.HasPrefix(contentType, "application/json"),
		strings.HasPrefix(contentType, "application/javascript"):
		{
			body := requestBody(request)
			signStr = signStr + request.Form.Encode() + JSONSorted(string(body))
		}
	default:
		{
			signStr = signStr + request.Form.Encode()
		}
	}
	signStr = signStr + timeStamp
	return Sign(signStr, key), nil
}

func requestBody(req *http.Request) []byte {
	if req.Body == nil {
		return []byte{}
	}
	var requestbody []byte
	var maxMemory = 1 << 26 //64MB
	safe := &io.LimitedReader{R: req.Body, N: int64(maxMemory)}
	if req.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(safe)
		if err != nil {
			return []byte{}
		}
		requestbody, _ = ioutil.ReadAll(reader)
	} else {
		requestbody, _ = ioutil.ReadAll(safe)
	}
	return requestbody
}

// Sign 签名
func Sign(str, key string) string {
	hmac := func(key, data string) string {
		hmac := hmac.New(md5.New, []byte(key))
		hmac.Write([]byte(data))
		return hex.EncodeToString(hmac.Sum([]byte("")))
	}
	return hmac(key, JSONSorted(str))
}

// JSONSorted json排序字符串
func JSONSorted(str string) string {
	if len(str) <= 0 {
		return ""
	}
	b := []byte(str)
	//不是json字符串直接返回
	if !json.Valid(b) {
		return str
	}
	// 数组
	if strings.HasPrefix(str, "[") {
		o := []interface{}{}
		if e := json.Unmarshal(b, &o); e != nil {
			return str
		}
		return JSONWithSlice(o)
	}
	//对象
	o := map[string]interface{}{}
	if e := json.Unmarshal(b, &o); e != nil {
		return str
	}
	return JSONWithMap(o)
}

// JSONWithMap 把map排序并转成字符串
func JSONWithMap(m map[string]interface{}) string {
	v := reflect.ValueOf(m)
	if len(v.MapKeys()) <= 0 {
		return ""
	}
	return jsonWithMap(v)
}

func sortedKeys(keys []reflect.Value) []reflect.Value {
	sort.Slice(keys, func(i, j int) bool {
		return strings.Compare(keys[i].Interface().(string), keys[j].Interface().(string)) < 0
	})
	return keys
}

func jsonWithMap(v reflect.Value) string {
	items := []interface{}{}
	sortKeys := sortedKeys(v.MapKeys())
	for _, k := range sortKeys {
		v := v.MapIndex(k)
		t := v.Type()
		for t.Kind() == reflect.Ptr || t.Kind() == reflect.Interface {
			v = v.Elem()
			t = v.Type()
		}
		kind := t.Kind()
		switch kind {
		case reflect.Map:
			{
				str := fmt.Sprintf("\"%v\":%s", k.Interface(), jsonWithMap(v))
				items = append(items, str)
			}
		case reflect.Slice, reflect.Array:
			{
				str := fmt.Sprintf("\"%v\":%s", k.Interface(), JSONWithSlice(v.Interface().([]interface{})))
				items = append(items, str)
			}
		case reflect.String:
			{
				str := fmt.Sprintf("\"%v\":\"%s\"", k.Interface(), v.Interface().(string))
				items = append(items, str)
			}
		default:
			{
				str := fmt.Sprintf("\"%v\":%v", k.Interface(), v.Interface())
				items = append(items, str)
			}
		}
	}
	return fmt.Sprintf("{%s}", SliceJoin(items, ","))
}

// JSONWithSlice 转成字符串
func JSONWithSlice(s []interface{}) string {
	items := []interface{}{}
	for _, item := range s {
		v := reflect.ValueOf(item)
		t := reflect.TypeOf(item)
		for t.Kind() == reflect.Ptr || t.Kind() == reflect.Interface {
			v = v.Elem()
			t = v.Type()
		}
		kind := t.Kind()
		switch kind {
		case reflect.Map:
			{
				items = append(items, jsonWithMap(v))
			}
		case reflect.Slice, reflect.Array:
			{
				items = append(items, JSONWithSlice(v.Interface().([]interface{})))
			}
		case reflect.String:
			{
				str := fmt.Sprintf("\"%s\"", v.Interface().(string))
				items = append(items, str)
			}
		default:
			{
				items = append(items, v.Interface())
			}
		}
	}
	return fmt.Sprintf("[%s]", SliceJoin(items, ","))
}
