package dto

import "time"

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
	NodeID       uint16
	Eui64        uint64
	LastRecvTime time.Time
	State        byte
	Mac          string
}
