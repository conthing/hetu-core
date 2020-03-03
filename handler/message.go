package handler

import (
	"encoding/binary"
	"encoding/json"
	"hetu-core/dto"
	"hetu-core/http"
	mqtt "hetu-core/mqtt/client"
	"hetu-core/redis"
	"os"
	"time"

	"github.com/conthing/utils/common"
	"github.com/google/uuid"
)

// ReceiveMessage 接收到 Zigbee 报文
func ReceiveMessage(eui64 uint64, message []byte, recvTime time.Time) {

	if eui64 == 0 {
		common.Log.Error("错误的 eui64")
		os.Exit(1)
	}

	m := &dto.ZigbeeDeviceMessage{
		UUID:         uuid.New(),
		Message:      message,
		Eui64:        eui64,
		LastRecvTime: recvTime,
		Addr:         binary.LittleEndian.Uint16(message),
	}

	// 上传报文
	mJSON, err := json.Marshal(m)
	if err != nil {
		common.Log.Error("序列化错误")
		return
	}

	httpInfo := redis.GetPubHTTPInfo()
	if httpInfo.Enable {
		http.Publish(httpInfo, mJSON)
	}

	mqttInfo := redis.GetPubMQTTInfo()
	if mqttInfo.Enable {
		mqtt.Publish("zigbee_device", mJSON)
	}

	// 持久化 Message
	redis.SaveZigbeeMessage(m)

}

// NodeStatus 离线、上线
func NodeStatus(eui64 uint64, nodeID uint16, status byte, addr byte) {
	if eui64 == 0 {
		common.Log.Error("错误的 eui64")
		os.Exit(1)
	}

	node := &dto.ZigbeeNode{
		Eui64:  eui64,
		State:  status,
		NodeID: nodeID,
		Addr:   addr,
	}

	redis.SaveZigbeeNode(node)

}

// SentMessage 往 Zigbee 发数据
func SentMessage(eui64 uint64, profileID uint16, clusterID uint16, localEndpoint byte, remoteEndpoint byte, message []byte, success bool) {

}
