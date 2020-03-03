package main

import (
	"hetu-core/config"
	"hetu-core/ezsp"
	"hetu-core/ezsp/zgb"
	mqtt "hetu-core/mqtt/client"
	"hetu-core/redis"
	"hetu-core/router"
	"os"

	"github.com/conthing/utils/common"
)

func main() {
	config.Service()
	redis.Connect()
	ezsp.InitEzspModule()
	initInfo := redis.GetPubMQTTInfo()
	mqtt.Init(initInfo)
	errs := make(chan error, 3)
	common.Log.Infof("VERSION %s build at %s", common.Version, common.BuildTime)

	go zgb.TickRunning(errs)

	go router.Run(8080)

	// recv error channel
	c := <-errs
	common.Log.Errorf("terminating: %v", c)
	os.Exit(0)

}
