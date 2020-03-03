package dto

import (
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
)

// Network Zigbee 网络控制
type Network struct {
	Command string
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
	ID      string `json:"id"`
	Enable  bool   `json:"enable"`
}
