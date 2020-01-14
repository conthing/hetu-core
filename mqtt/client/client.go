package client

import (
	"github.com/conthing/utils/common"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var client MQTT.Client

// MQTTURL MQTTURL
var MQTTURL = "tcp://mqtt.conthing.com:1883"

// Connect 连接 id 可以是本机的 MAC
func Connect(id string) {
	opts := MQTT.NewClientOptions().AddBroker(MQTTURL)
	opts.SetClientID(id)

	client = MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		common.Log.Errorf("client连接失败:%s", token.Error())
	}
}

// Publish 发布消息
func Publish(topic string, payload interface{}) error {
	err := client.Publish(topic, 0, false, payload).Error()
	if err != nil {
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
