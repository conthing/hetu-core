package redis

import (
	"github.com/mediocregopher/radix/v3"
)

var _pubSubConn radix.PubSubConn

// InitPubSubConn 初始化 PubSubConn
func InitPubSubConn() {
	conn, err := radix.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		panic(err)
	}

	_pubSubConn = radix.PubSub(conn)
}

// Publish redis 队列发送消息
func Publish() {
}

// SubScribe redis 队列订阅消息
func SubScribe() {
	// _pubSubConn.Subscribe()

}
