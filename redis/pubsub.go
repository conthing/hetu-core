package redis

import (
	"github.com/conthing/utils/common"
	"github.com/mediocregopher/radix/v3"
)

var _pubSubConn radix.PubSubConn

// InitPubSubConn 初始化 PubSubConn
func InitPubSubConn() error {
	conn, err := radix.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		return err
	}

	_pubSubConn = radix.PubSub(conn)
	return nil
}

// Publish redis 队列发送消息
func Publish(topic string, data interface{}) {
	err := Client.Do(radix.FlatCmd(nil, "PUBLISH", topic, data))
	if err != nil {
		common.Log.Error("Publish redis 队列错误", err)
	}
}

// Subscribe redis 队列订阅消息
func Subscribe(fn func(data []byte)) {
	msgCh := make(chan radix.PubSubMessage)
	if err := _pubSubConn.Subscribe(msgCh, "hetu-core"); err != nil {
		common.Log.Error("redis 队列订阅失败", err)
		return
	}
	go func() {
		for {
			select {
			case msg := <-msgCh:
				common.Log.Infof("Redis 队列消息 [%s] [%s] ", msg.Channel, msg.Type)
				fn(msg.Message)
			}
		}
	}()

}
