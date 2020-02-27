package ezsp

import (
	"hetu-core/config"
	"hetu-core/ezsp/ash"
	"hetu-core/ezsp/hetu"
	"hetu-core/ezsp/zgb"
	"hetu-core/handler"
	"hetu-core/redis"

	"github.com/conthing/utils/common"
)

// InitEzspModule 初始化Ezsp
func InitEzspModule() {
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
	nodesMap = redis.ReadSaveZigbeeNodeTable()
	hetu.LoadNodesMap(nodesMap)
}
