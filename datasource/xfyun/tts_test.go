package xfyun

import (
	"os"
	"testing"
)

func init() {
	os.Setenv("XFYUN_TTS_APPID", "5d6e86301")
	os.Setenv("XFYUN_TTS_KEY", "8f5c6dd953fa260d118412885837f931")
}

func TestTTS(t *testing.T) {
	// fmt.Println(TTS("1号桩1枪，尾号1881 的用户 ༄寧 靜 致 远༄ 充电完成"))
}
