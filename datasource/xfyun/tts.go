package xfyun

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func init() {
	//fmt.Println("ddddddddddddddddddddddddddd", os.Getenv("XFYUN_TTS_APPID"))
	//config["TTS_APPID"] = os.Getenv("XFYUN_TTS_APPID")
	//config["TTS_KEY"] = os.Getenv("XFYUN_TTS_KEY")
}

func TTS(appID, key, text string) ([]byte, error) {
	curTime := strconv.FormatInt(time.Now().Unix(), 10)
	tt := make(map[string]string)
	//  音频编码(raw合成的音频格式pcm、wav,lame合成的音频格式MP3)
	tt["aue"] = "lame"
	//  采样率
	tt["auf"] = "audio/L16;rate=8000"
	//  发音人（登陆开放平台https://www.xfyun.cn/后--我的应用（必须为webapi类型应用）--添加在线语音合成（已添加的不用添加）--发音人管理---添加发音人--修改发音人参数）
	tt["voice_name"] = "aisxping"
	//tt["voice_name"] = "aisjiuxu"
	param, _ := json.Marshal(tt)
	base64Param := base64.StdEncoding.EncodeToString(param)

	w := md5.New()
	_, _ = io.WriteString(w, key+curTime+base64Param)
	checksum := fmt.Sprintf("%x", w.Sum(nil))
	//  待合成文本
	var data = url.Values{}
	data.Add("text", text)
	reqBody := data.Encode()

	client := &http.Client{}
	//  组装http请求头
	req, _ := http.NewRequest("POST", "https://api.xfyun.cn/v1/service/v1/tts", strings.NewReader(reqBody))

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-CurTime", curTime)
	req.Header.Set("X-Appid", appID)
	req.Header.Set("X-Param", base64Param)
	req.Header.Set("X-CheckSum", checksum)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	contentType := resp.Header.Get("Content-type")
	if contentType == "text/plain" {
		var data map[string]string
		if err = json.Unmarshal(respBody, &data); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("%s", data["desc"])
	} else if contentType != "audio/mpeg" {
		return nil, fmt.Errorf("未知错误")
	}

	//fmt.Println(resp.Header)
	//ioutil.WriteFile("./ceshi.mp3", []byte(respBody), 0666)
	return []byte(respBody), nil
	//d1 := []byte(resp_body)
	//  保存合成音频文件
}
