package proxy

import (
	"encoding/json"
	"hetu-core/dto"
	"hetu-core/http"
	mqtt "hetu-core/mqtt/client"
	"hetu-core/redis"

	"github.com/conthing/ezsp/hetu"
	"github.com/conthing/utils/common"
)

// Down 下行消息
func Down(rm *dto.ReceiveMessageDTO) {
	if rm.Type == "zigbee" {
		// 底层发送 zigbee 报文
		err := hetu.SendUnicast(rm.Data.Eui64, rm.Data.Message)
		if err != nil {
			common.Log.Error("[FAILED to hetu SendUnicast: ", err)
		}
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
func Post(data []byte) error {

	httpInfo := redis.GetPubHTTPInfo()
	if httpInfo.Enable {
		err := http.Publish(httpInfo, data)
		if err != nil {
			// PushBack实现失败重传
			redis.RPushBackupQueue(string(data)) // todo 为什么只有HTTP这里会pushback？
			return err
		}
	}

	mqttInfo := redis.GetPubMQTTInfo()
	if mqttInfo.Enable {
		topic := "/hetu/" + common.GetSerialNumber() + "/report"
		err := mqtt.Publish(topic, data)
		if err != nil {
			common.Log.Error("[MQTT] Post data failed", err)
			return err
		}
	}

	redis.TrimBackupQueue()
	return nil

}
