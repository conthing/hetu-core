package proxy

import (
	"encoding/json"
	"hetu-core/config"
	"hetu-core/dto"
	"hetu-core/http"
	mqtt "hetu-core/mqtt/client"
	"hetu-core/redis"

	"github.com/conthing/utils/common"
)

// Down 下行消息
func Down(rm *dto.ReceiveMessageDTO) {
	if rm.Type == "zigbee" {
		// 底层发送 zigbee 报文
		common.Log.Info("TODO: 底层报文下发")
		// todo
	} else {
		rmJSON, err := json.Marshal(rm)
		if err != nil {
			common.Log.Error("序列化错误")
			return
		}
		redis.Publish(rm.Type, rmJSON)
		common.Log.Info("redis 下发消息成功")
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
