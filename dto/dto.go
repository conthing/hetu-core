package dto

import "time"

const (
	// Success 成功状态码
	Success = iota
	// InvalidJSON 无效的 JSON
	InvalidJSON
	// CreateZigbeeNetFailed 创建 Zigbee 网络失败
	CreateZigbeeNetFailed
	// RemoveZigbeeNetFailed 删除 Zigbee 网络失败
	RemoveZigbeeNetFailed
)

// Network Zigbee 网络控制
type Network struct {
	Command string
}

// Resp 回复
type Resp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ZigbeeDeviceMessage MQTT、Redis 存储模型
type ZigbeeDeviceMessage struct {
	Mac     string `json:"mac"`
	Addr    uint16 `json:"addr"`
	Message []byte `json:"message"`
	Time    int64  `json:"time"`
}

// ZigbeeNode 设备节点
type ZigbeeNode struct {
	Eui64        uint64
	LastRecvTime time.Time
	State        byte
	NodeID       uint16
	// Addr 播码地址
	Addr    uint16
	Message []byte
}
