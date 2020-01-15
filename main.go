package main

import (
	"hetu-core/config"
	"hetu-core/ezsp"
	"hetu-core/ezsp/zgb"
	mqtt "hetu-core/mqtt/client"
	"hetu-core/redis"
	"hetu-core/router"
)

func main() {
	config.Service()
	redis.Connect()
	ezsp.InitEzspModule()

	mqtt.Connect("hetu_mqtt_post")
	errs := make(chan error, 3)

	go zgb.TickRunning(errs)
	router.Run(8080)

}
