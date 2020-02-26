package handler

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hetu-core/dto"
	mqtt "hetu-core/mqtt/client"
	"hetu-core/redis"
	"os"
	"strconv"
	"time"

	"github.com/conthing/utils/common"
)

// ReceiveMessage 接收到 Zigbee 报文
func ReceiveMessage(eui64 uint64, profileID uint16, clusterID uint16, localEndpoint byte, remoteEndpoint byte, message []byte) {
	if eui64 == 0 {
		common.Log.Error("错误的 eui64")
		os.Exit(1)
	}
	m := new(dto.ZigbeeDeviceMessage)
	m.Mac = fmt.Sprintf("%016x", eui64)
	m.Addr = binary.LittleEndian.Uint16(message)
	m.Message = message
	m.Time = time.Now().Unix()
	mJSON, err := json.Marshal(m)
	if err != nil {
		common.Log.Error("序列化错误")
		return
	}
	// MQTT
	err = mqtt.Publish("zigbee_device", mJSON)
	if err != nil {
		common.Log.Error("mqtt 发送失败")
		redis.SaveToPreparedQueue()
	}
	// Redis 时间序列
	redis.SaveZigbeeDeviceList(mJSON)
	// Redis table
	euiStr := strconv.FormatUint(eui64, 16)
	if euiStr == "0" {
		common.Log.Error("错误的 eui64")
		os.Exit(1)
	}
	node, err := redis.GetZigbeeNode(euiStr)
	if err != nil {
		return
	}
	common.Log.Info("redis 读取 node", node.NodeID)
	node.LastRecvTime = time.Now()
	node.Addr = binary.LittleEndian.Uint16(message)
	node.Message = message
	node.Eui64 = eui64
	node.State = byte(1)

	data, err := json.Marshal(node)
	if err != nil {
		common.Log.Error("序列化 node 节点 失败", err)
		return
	}
	redis.SaveZigbeeNode(euiStr, data)

}

// SentMessage 往 Zigbee 发数据
func SentMessage(eui64 uint64, profileID uint16, clusterID uint16, localEndpoint byte, remoteEndpoint byte, message []byte, success bool) {

}

// NodeStatus 离线、上线
func NodeStatus(eui64 uint64, nodeID uint16, status byte, deviceType byte) {
	if eui64 == 0 {
		common.Log.Error("错误的 eui64")
		os.Exit(1)
	}
	euiStr := strconv.FormatUint(eui64, 16)
	if euiStr == "0" {
		common.Log.Error("错误的 eui64", eui64, euiStr)
		os.Exit(1)
	}

	node, _ := redis.GetZigbeeNode(euiStr)
	node.State = status
	node.NodeID = nodeID
	data, err := json.Marshal(node)
	if err != nil {
		common.Log.Error("序列化 node 节点 失败", err)
		return
	}
	redis.SaveZigbeeNode(euiStr, data)
	common.Log.Info("redis 存 node", node.NodeID)

}
