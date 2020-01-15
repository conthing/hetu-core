package main

import (
	"encoding/json"
	"hetu/config"
	"hetu/dto"
	"hetu/ezsp/ash"
	"hetu/ezsp/hetu"
	"hetu/ezsp/zgb"
	"hetu/handler"
	"hetu/redis"
	"strconv"

	"github.com/conthing/utils/common"
)

func initEzspModule() {
	hetu.C4Callbacks = hetu.StC4Callbacks{
		C4MessageSentHandler:     handler.SentMessage,
		C4IncomingMessageHandler: handler.ReceiveMessage,
		C4NodeStatusHandler:      handler.NodeStatus,
	}

	zgb.TraceSet(&config.Conf.TraceSettings)
	zgb.NetworkSet(&config.Conf.NetworkSettings)

	err := ash.AshSerialOpen(config.Conf.Serial.Name, config.Conf.Serial.Baud, config.Conf.Serial.RtsCts)
	if err != nil {
		common.Log.Errorf("failed to open serial %v", config.Conf.Serial.Name)
	}

	// Time it took to start service
	common.Log.Infof("Open Serial success port=%s baud=%d", config.Conf.Serial.Name, config.Conf.Serial.Baud)

	// 初始化 长短地址对应表
	var nodesMap map[uint64]hetu.StNode
	m := redis.ReadSaveZigbeeNodeTable()
	for key, value := range m {
		eui64, err := strconv.ParseUint(key, 16, 64)
		if err != nil {
			common.Log.Error("eui64类型转换失败", err)
			continue
		}
		var node dto.ZigbeeNode
		err = json.Unmarshal([]byte(value), &node)
		if err != nil {
			common.Log.Error("node 序列化错误", err)
			continue
		}
		nodesMap[eui64] = hetu.StNode{
			NodeID:       node.NodeID,
			Eui64:        node.Eui64,
			LastRecvTime: node.LastRecvTime,
		}
	}
	hetu.LoadNodesMap(nodesMap)
}
