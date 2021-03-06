package client

import (
	"encoding/json"
	"fmt"
	"hetu-core/dto"
	"sync"
	"time"

	"github.com/conthing/utils/common"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var client MQTT.Client
var rw sync.RWMutex

// Init 初始化 MQTT 连接
func Init(info *dto.PubMQTTInfo, down func(rm *dto.ReceiveMessageDTO)) {
	// 不会报 nil 错误
	if info.Enable {
		Connect(info)
		topic := fmt.Sprintf("/hetu/%s/command", common.GetSerialNumber())

		fn := func(client MQTT.Client, message MQTT.Message) {
			common.Log.Infof("MQTT 订阅消息报文 [%s]", message.Topic())
			rm := new(dto.ReceiveMessageDTO)
			err := json.Unmarshal(message.Payload(), rm)
			if err != nil {
				common.Log.Info("序列化失败")
				return
			}
			if rm.Type == "" {
				common.Log.Error("报文 type 为空")
				return
			}
			down(rm)
		}

		Subscribe(topic, fn)
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
	opts.SetClientID(common.GetSerialNumber() + time.Now().String())
	client = MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		common.Log.Errorf("client连接失败:%s", token.Error())
	}
}

// Publish 发布消息
// 读锁锁住
func Publish(topic string, payload interface{}) error {
	rw.RLock()
	err := client.Publish(topic, 0, false, payload).Error()
	rw.RUnlock()
	if err != nil {
		common.Log.Error("mqtt 发送失败")
		return err
	}
	common.Log.Infof("topic:%s 发布成功", topic)
	return nil
}

// Subscribe 订阅消息
func Subscribe(topic string, callback MQTT.MessageHandler) {
	if token := client.Subscribe(topic, 0, callback); token.Wait() && token.Error() != nil {
		common.Log.Errorf("订阅失败:%s", token.Error())
	}
	common.Log.Infof("topic:%s 订阅成功", topic)
}
