package main

import (
	"hetu-core/backup"
	"hetu-core/config"
	"hetu-core/handler"
	mqtt "hetu-core/mqtt/client"
	"hetu-core/proxy"
	"hetu-core/redis"
	"hetu-core/router"
	"os"

	"github.com/conthing/ezsp/zgb"

	"github.com/conthing/ezsp/ash"
	"github.com/conthing/ezsp/hetu"
	"github.com/conthing/utils/common"
)

func main() {
	config.Service()
	redis.Connect()
	redis.Subscribe(proxy.Post)
	InitEzspModule()
	initInfo := redis.GetPubMQTTInfo()
	mqtt.Init(initInfo, proxy.Down)

	errs := make(chan error, 3)
	common.Log.Infof("VERSION %s build at %s", common.Version, common.BuildTime)
	go zgb.TickRunning(errs)
	go backup.ConsumeBackupQueue()
	go router.Run(52040)

	// recv error channel
	c := <-errs
	common.Log.Errorf("terminating: %v", c)
	os.Exit(0)

}

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
	nodesMap := redis.ReadSaveZigbeeNodeTable()
	hetu.LoadNodesMap(nodesMap)
}
