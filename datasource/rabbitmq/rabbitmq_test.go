package rabbitmq

import (
	"context"
	"os"
	"testing"
)

func TestPublish(t *testing.T) {
	os.Setenv("RABBITMQ_USER", "admin")
	os.Setenv("RABBITMQ_PASS", "3hCjyu5RXOX1xW6W")
	os.Setenv("RABBITMQ_POOL", "10.43.0.14:5672")

	Init()
	err := Publish(context.Background(), "joyson_email", nil, []byte(
		`{
    "id": "uuid", 
    "appId": "jx-services", 
    "subject": "测试", 
    "target": "ping.ling@joysonquin.com,lphj07@163.com",
    "cc": "ping.ling@joysonquin.com",
    "content": "测试MQ消息和邮件以及cc"
}`))

	if err != nil {
		t.Error(err.Error())
	}
}
