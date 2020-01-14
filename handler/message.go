package handler

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hetu/dto"
	mqtt "hetu/mqtt/client"
	"hetu/redis"
	"time"

	"github.com/conthing/utils/common"
)

// ReceiveMessage 接收到 Zigbee 报文
func ReceiveMessage(eui64 uint64, profileID uint16, clusterID uint16, localEndpoint byte, remoteEndpoint byte, message []byte) {
	m := new(dto.ZigbeeDeviceMessage)
	m.Mac = fmt.Sprintf("%016x", eui64)
	m.Addr = binary.LittleEndian.Uint16(message)
	m.Message = message
	m.Time = time.Now().Unix()
	mJSON, err := json.Marshal(m)
	if err != nil {
		common.Log.Error("序列化错误")
	}
	// MQTT
	err = mqtt.Publish("zigbee_device", mJSON)
	if err != nil {
		common.Log.Error("mqtt 发送失败")
		redis.SaveToPreparedQueue()
	}
	// Redis
	redis.Save(mJSON)

}

// SentMessage 往 Zigbee 发数据
func SentMessage(eui64 uint64, profileID uint16, clusterID uint16, localEndpoint byte, remoteEndpoint byte, message []byte, success bool) {

}

// NodeStatus ?
func NodeStatus(eui64 uint64, nodeID uint16, status byte, deviceType byte) {

}
