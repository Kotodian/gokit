package http

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	libhttp "net/http"
	"time"

	"github.com/Kotodian/gokit/utils"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	"github.com/edwardhey/gorequest"
)

const (
	KindRequestMethodGet  string = "get"
	KindRequestMethodPost string = "post"
)

func Request(url string, method string, data interface{}, header libhttp.Header, out interface{}) (err error) {
	header.Set("Content-Type", "application/json")
	curl := gorequest.New()
	if method == KindRequestMethodGet {
		curl.Get(url)
	} else {
		curl.Post(url)
	}
	curl.Header = header
	curl.Timeout(time.Second * 3).
		Type("json").
		SendStruct(data)

	var resp gorequest.Response
	var body string
	var errs []error
	resp, body, errs = curl.End()
	if len(errs) > 0 {
		err = fmt.Errorf("%v", errs)
		return
	}
	defer resp.Body.Close()

	// fmt.Printf("-------->[%+v][%+v]\r\n", resp, body)

	if resp.StatusCode != 200 {
		err = fmt.Errorf("http response error code=%d", resp.StatusCode)
		return
	} else if !gjson.Valid(body) {
		// errMsg = body
		logrus.Errorf("http resp body:[%s]\r\n", body)
		return errors.New("is not valid json")
	}

	if out == nil {
		var errMsg string
		var respCode int64
		// var dataString string

		jsonRes := gjson.GetMany(body, "code", "msg", "data")
		errMsg = jsonRes[1].String()
		respCode = jsonRes[0].Int()
		// dataString = jsonRes[2].String()

		if respCode != 0 {
			err = errors.New(errMsg)
			return
		}
	} else {
		if err = json.Unmarshal([]byte(body), out); err != nil {
			return
		}
	}

	return nil
}

type SuccessCondJsonBody struct {
	Field string      `json:"field_name"`
	Val   interface{} `json:"val"`
}

type KindDecryptor int

const (
	KindDecryptorNone KindDecryptor = 0
	KindDecryptorAES  KindDecryptor = 1
)

type SuccessCond struct {
	Status         []int                `json:"status"`
	Body           *SuccessCondJsonBody `json:"body"`
	Decryptor      string               `json:"decryptor,omitempty"`
	DecryptorField string               `json:"decryptor_field,omitempty"`
	KindDecryptor  KindDecryptor        `json:"kind_decryptor,omitempty"`
}

func NewDecryptor(kind KindDecryptor, str string) (de IDecryptor, err error) {
	switch kind {
	case KindDecryptorAES:
		de = &DecryptorAES{}
	default:
		return nil, fmt.Errorf("not support")
	}
	if err = json.Unmarshal([]byte(str), de); err != nil {
		return
	}
	return
}

type IDecryptor interface {
	Decrypt(src string) (data string, err error)
}

type DecryptorAES struct {
	DataSecret   string `json:"data_secret"`    //消息秘钥：用于对所有接口中的Data信息进行加密
	DataSecretIV string `json:"data_secret_iv"` //消息秘钥初始化向量：固定16位，用于AES机密过程的混合加密
}

func (d *DecryptorAES) Decrypt(src string) (data string, err error) {
	var sb []byte
	sb, err = base64.StdEncoding.DecodeString(src)
	if err != nil {
		return "", err
	}
	decrypted := make([]byte, len(sb))
	var aesBlockDecrypter cipher.Block
	aesBlockDecrypter, err = aes.NewCipher([]byte(d.DataSecret))
	if err != nil {
		return "", err
	}
	aesDecrypter := cipher.NewCBCDecrypter(aesBlockDecrypter, []byte(d.DataSecretIV))
	aesDecrypter.CryptBlocks(decrypted, sb)
	return string(utils.PKCS5Trimming(decrypted)), nil
}

type Out struct {
	IsSuccessed bool
	Body        string
	Err         string
}

//ParamsCert 证书
type ParamsCert struct {
	CA         string `json:"ca"`          //CA证书
	ClientCert string `json:"client_cert"` //客户端证书
	ClientKey  string `json:"client_key"`  //客户端秘钥
}

type Params struct {
	IsJson      bool                   `json:"is_json"`
	Method      string                 `json:"method"`
	Timeout     time.Duration          `json:"timeout"`
	Header      map[string]string      `json:"header,omitempty"`
	Data        interface{}            `json:"data,omitempty"`
	Json        interface{}            `json:"json,omitempty"`
	IgnoreBody  bool                   `json:"ignore_body"`
	SuccessCond *SuccessCond           `json:"success_cond,omitempty"`
	Addon       map[string]interface{} `json:"addon,omitempty"` //附加的内容，会在队列
	Debug       bool                   `json:"debug"`           //调试
	Cert        *ParamsCert            `json:"cert,omitempty"`  //证书
	//LoadBody    bool                   `json:"load_body"`
}

type Respnse struct {
	Addon       map[string]interface{} `json:"addmon,omitempty"`
	Data        string                 `json:"data"` //返回的内容
	Err         string                 `json:"err"`
	IsSuccessed bool                   `json:"is_successed"`
}

func RequestWithCondCheck(url string, data Params, out *Out) (err error) {
	curl := gorequest.New()
	if data.Debug {
		curl.Debug = true
	}
	if data.Cert != nil {
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM([]byte(data.Cert.CA))

		var cert tls.Certificate
		if cert, err = tls.X509KeyPair([]byte(data.Cert.ClientCert), []byte(data.Cert.ClientKey)); err != nil {
			return
		}

		curl.TLSClientConfig(&tls.Config{
			InsecureSkipVerify: true,
			RootCAs:            pool,
			Certificates:       []tls.Certificate{cert},
		})
	}
	if data.Method == KindRequestMethodGet {
		curl.Get(url)
	} else {
		curl.Post(url)

	}
	header := libhttp.Header{}
	defer func() {
		if out != nil {
			if err != nil {
				out.IsSuccessed = false
			}
		}
	}()

	if data.Header != nil {
		for k, v := range data.Header {
			header.Set(k, v)
		}
	}

	if data.IsJson {
		header.Set("Content-Type", "application/json")
		header.Set("Accept", "application/json")
		curl.Type("json")
		curl.SendStruct(data.Json)
	}

	if data.Data != nil {
		curl.Query(data.Data)
	}
	//curl.Debug = true

	//if data.Json != nil {
	//
	//}

	curl.Header = header

	to := time.Second * 3
	if data.Timeout > 0 {
		to = data.Timeout
	}
	curl.Timeout(to)

	resp, body, errs := curl.EndBytes()
	if len(errs) > 0 {
		err = fmt.Errorf("%v", errs)
		return
	}
	defer resp.Body.Close()

	//defer func() {
	//	fmt.Println("---------------")
	//	fmt.Println(resp.StatusCode)
	//	fmt.Println(fmt.Sprintf("%+v", out))
	//	fmt.Println("---------------")
	//}()

	//去除bom头
	body = bytes.TrimPrefix(bytes.TrimSpace(body), []byte{239, 187, 191})
	out.Body = string(body)

	if data.SuccessCond != nil {
		if len(data.SuccessCond.Status) > 0 { //校验状态码
			data.SuccessCond.Status = append(data.SuccessCond.Status, 200)
			isSuccess := false
			for _, v := range data.SuccessCond.Status {
				if v == resp.StatusCode {
					isSuccess = true
					if data.IgnoreBody {
						out.IsSuccessed = true
						return
						//return nil
					}
					break
					//out.IsSuccessed = true
					//break
					//}
				}
			}
			//如果HTTP状态码都不对，就直接返回错误了
			if !isSuccess {
				out.Err = fmt.Sprintf("http response code unvalid, code:%d", resp.StatusCode)
				return
			}
		} else { //校验200
			if resp.StatusCode == 200 {
				if data.IgnoreBody {
					out.IsSuccessed = true
					return
				}
			} else {
				//if data.IgnoreBody {
				out.Err = fmt.Sprintf("http response code unvalid, code:%d", resp.StatusCode)
				return
				//}
			}
		}
	}

	//再检查内容
	if data.IsJson {
		if data.SuccessCond != nil && data.SuccessCond.Body != nil {
			if data.SuccessCond.KindDecryptor != KindDecryptorNone {
				//再判断内容
				decryptorFieldVal := gjson.GetBytes(body, data.SuccessCond.DecryptorField)
				if decryptorFieldBody := decryptorFieldVal.String(); decryptorFieldBody != "" {
					var de IDecryptor
					var bs string
					if de, err = NewDecryptor(data.SuccessCond.KindDecryptor, data.SuccessCond.Decryptor); err != nil {
						out.IsSuccessed = false
						out.Err = "构造解码方式错误," + err.Error()
						return
					} else if bs, err = de.Decrypt(decryptorFieldVal.String()); err != nil {
						out.IsSuccessed = false
						out.Err = "解码失败," + err.Error()
						return
					}
					if data.Debug {
						fmt.Println("decrypted body:", bs)
					}
					body = []byte(bs)
				} else {
					err = fmt.Errorf("decrypt body is empty")
					return
				}
			}
			if len(body) == 0 {
				return fmt.Errorf("resp empty body")
			} else if !gjson.ValidBytes(body) {
				return fmt.Errorf("resp body is not valid json, %s", string(body))
			}

			//再判断内容
			jsonRes := gjson.GetManyBytes(body, data.SuccessCond.Body.Field)

			if len(jsonRes) > 0 && jsonRes[0].String() == fmt.Sprintf("%v", data.SuccessCond.Body.Val) {
				out.IsSuccessed = true
			} else {
				out.IsSuccessed = false
				out.Err = fmt.Errorf("http response check fail, %s is '%v' not '%v'", data.SuccessCond.Body.Field, jsonRes[0], data.SuccessCond.Body.Val).Error()
			}
		}
	}
	return
}
