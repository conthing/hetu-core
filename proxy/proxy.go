package proxy

import (
	"hetu-core/config"
	"hetu-core/dto"
	"hetu-core/http"
	mqtt "hetu-core/mqtt/client"
	"hetu-core/redis"
)

// Down 下行消息
func Down(rm dto.ReceiveMessageDTO) {
	if rm.Type == "zigbee" {
		// 底层发送 zigbee 报文
		// todo
	} else {
		redis.Publish(rm.Type, rm)
	}

}

// Post 上行消息
func Post(data []byte) {

	httpInfo := redis.GetPubHTTPInfo()
	if httpInfo.Enable {
		http.Publish(httpInfo, data)
	}

	mqttInfo := redis.GetPubMQTTInfo()
	if mqttInfo.Enable {
		topic := "/hetu/" + config.Mac + "/report"
		mqtt.Publish(topic, data)
	}

}
