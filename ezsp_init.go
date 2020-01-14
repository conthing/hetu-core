package main

import (
	"hetu/config"
	"hetu/ezsp/ash"
	"hetu/ezsp/hetu"
	"hetu/ezsp/zgb"
	"hetu/handler"

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

}
