package redis

import (
	"log"

	"github.com/mediocregopher/radix/v3"
)

var pool *radix.Pool

// Connect 初始化连接池
func Connect() {
	var err error
	pool, err = radix.NewPool("tcp", "127.0.0.1:6379", 10)
	if err != nil {
		log.Fatal("数据库连接失败", err)
	}
}

// SaveToPreparedQueue 未发出去的队列
func SaveToPreparedQueue() {
	// var length string

	// pool.Do(radix.Cmd(nil, "LPUSH", "prepared_list", string(m)))
}

// SaveZigbeeDeviceList 保存到数据库
func SaveZigbeeDeviceList(m []byte) {
	pool.Do(radix.Cmd(nil, "LPUSH", "zigbee_device_list", string(m)))
}

// SaveZigbeeNode 存设备地址、短地址关系对应节点
func SaveZigbeeNode(eui64 uint64, m []byte) {
	pool.Do(radix.Cmd(nil, "HSET", "zigbee_device_table", string(eui64), string(m)))
}

// ReadSaveZigbeeNodeTable 读取对应表
func ReadSaveZigbeeNodeTable() map[string]string {
	var m map[string]string
	pool.Do(radix.Cmd(&m, "HGETALL", "zigbee_device_table"))
	return m
}
