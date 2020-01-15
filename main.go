package main

import (
	"hetu/config"
	"hetu/ezsp/zgb"
	mqtt "hetu/mqtt/client"
	"hetu/redis"
	"hetu/router"
)

func main() {
	config.Service()
	redis.Connect()
	initEzspModule()

	mqtt.Connect("hetu_mqtt_post")
	errs := make(chan error, 3)

	go zgb.TickRunning(errs)
	router.Run(8080)

}
