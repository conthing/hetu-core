package client

import (
	"fmt"
	"hetu-core/config"
	"hetu-core/dto"
	"sync"
	"time"

	"github.com/conthing/utils/common"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var client MQTT.Client
var rw sync.RWMutex

// Init 初始化 MQTT 连接
func Init(info *dto.PubMQTTInfo) {
	// 不会报 nil 错误
	if info.Enable {
		Connect(info)
		// topic := fmt.Sprintf("/hetu/%s/command", config.Mac)
		// Subscribe(topic)
	}
}

// ReConnect 重新连接
func ReConnect(info *dto.PubMQTTInfo) {
	rw.Lock()
	client.Disconnect(100)
	Connect(info)
	rw.Unlock()
}

// Connect 连接
func Connect(info *dto.PubMQTTInfo) {
	server := fmt.Sprintf("%s:%d", info.Address, info.Port)
	opts := MQTT.NewClientOptions().AddBroker(server)
	// ClientID 无需配置
	opts.SetClientID(config.Mac + time.Now().String())
	client = MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		common.Log.Errorf("client连接失败:%s", token.Error())
	}
}

// Publish 发布消息
// 读锁锁住
func Publish(topic string, payload interface{}) {
	rw.RLock()
	err := client.Publish(topic, 0, false, payload).Error()
	rw.RUnlock()
	if err != nil {
		common.Log.Error("mqtt 发送失败")
		return
	}
	common.Log.Infof("topic:%s 发布成功", topic)
}

// Subscribe 订阅消息
func Subscribe(topic string, callback MQTT.MessageHandler) {
	if token := client.Subscribe(topic, 0, callback); token.Wait() && token.Error() != nil {
		common.Log.Errorf("订阅失败:%s", token.Error())
	}
	common.Log.Infof("topic:%s 订阅成功", topic)
}
