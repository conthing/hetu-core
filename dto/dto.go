package dto

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	// Success 成功状态码
	Success = iota
	// InvalidJSON 无效的 JSON
	InvalidJSON
	// CreateZigbeeNetFailed 创建 Zigbee 网络失败
	CreateZigbeeNetFailed
	// RemoveZigbeeNetFailed 删除 Zigbee 网络失败
	RemoveZigbeeNetFailed

	GetNodeListFailed

	GetNodeLatestMessageFailed
	DeleteNodeFailed
)

// Network Zigbee 网络控制
type Network struct {
	Command string
	Channel uint8 `binding:"-"`
}

// Resp 回复
type Resp struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// ZigbeeDeviceMessage MQTT、Redis 存储模型
type ZigbeeDeviceMessage struct {
	Eui64        uint64    `json:"eui64"`
	Addr         uint16    `json:"addr"`
	Message      []byte    `json:"message"`
	LastRecvTime time.Time `json:"time"`
	UUID         uuid.UUID `json:"uuid"`
	HostMac      string    `json:"mac"`
	HostAlias    string    `json:"alias"`
}

func (m *ZigbeeDeviceMessage) String() (string, error) {
	array, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("Marshal:%w", err)
	}
	return string(array), nil
}

// PostMessageDTO 上行消息结构体
type PostMessageDTO struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// UnicastData 单波发送报文结构体
type UnicastData struct {
	Eui64   uint64 `json:"eui64"`
	Message []byte `json:"message"`
}

// ReceiveMessageDTO 下行消息结构体
type ReceiveMessageDTO struct {
	Type string      `json:"type"`
	Data UnicastData `json:"data"`
}

// ZigbeeNode 设备节点
// NodeID     短地址
// Addr       播码地址
// State      1 connecting
//            2 online
//            3 offline
type ZigbeeNode struct {
	Eui64  uint64
	State  byte
	NodeID uint16
	Addr   byte
}

// ZNode 客户端接受
type ZNode struct {
	Mac  string
	Node ZigbeeNode
}

// PubHTTPInfo HTTP上报设置结构体
type PubHTTPInfo struct {
	Address string `json:"address"`
	Port    uint16 `json:"port"`
	URL     string `json:"URL"`
	Enable  bool   `json:"enable"`
}

// PubMQTTInfo MQTT上报设置
type PubMQTTInfo struct {
	Address string `json:"address"`
	Port    uint16 `json:"port"`
	Enable  bool   `json:"enable"`
}
