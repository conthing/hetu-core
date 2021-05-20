package main

import (
	"flag"
	"fmt"
	"hetu-core/backup"
	"hetu-core/db"
	"hetu-core/handler"
	mqtt "hetu-core/mqtt/client"
	"hetu-core/proxy"
	"hetu-core/redis"
	"hetu-core/router"
	"os"
	"os/signal"
	"syscall"

	"github.com/conthing/ezsp/zgb"

	"github.com/conthing/ezsp/ash"
	"github.com/conthing/ezsp/hetu"
	"github.com/conthing/utils/common"
)

// Config 配置模型
type Config struct {
	Ezsp EZSPConfig  `yaml:"ezsp"`
	HTTP HTTPConfig  `yaml:"http"`
	DB   db.DBConfig `yaml:"db"`
}
type EZSPConfig struct {
	Serial          Serial                `yaml:"serial"`
	TraceSettings   zgb.StTraceSettings   `yaml:"trace_settings"`
	NetworkSettings zgb.StNetworkSettings `yaml:"network_settings"`
}

// Serial 串口
type Serial struct {
	Name   string `yaml:"name"`
	Baud   uint   `yaml:"baud"`
	RtsCts bool   `yaml:"rtscts"`
}

// HTTP 配置
type HTTPConfig struct {
	Port int
}

var config = Config{}

func boot(_ interface{}) (needRetry bool, err error) {

	// 初始化数据库
	err = db.Init(&config.DB)
	if err != nil {
		return true, fmt.Errorf("failed to init database: %v", err)
	}
	common.Log.Debug("database init success")

	err = redis.Connect()
	if err != nil {
		return true, fmt.Errorf("failed to init redis: %v", err)
	}
	common.Log.Debug("redis init success")

	return
}

func main() {
	var cfgfile string

	// 解析命令行参数 -c <dir>
	flag.StringVar(&cfgfile, "c", "config.yaml", "Specify a config file other than default.")
	flag.Parse()

	common.InitLogger(&common.LoggerConfig{Level: "DEBUG", SkipCaller: true})
	common.Log.Infof("VERSION %s build at %s", common.Version, common.BuildTime)

	err := common.LoadYaml(cfgfile, &config)
	if err != nil {
		common.Log.Errorf("Failed to load config %w", err)
		os.Exit(1)
	}
	common.Log.Infof("Load config success %+v", config)

	if common.Bootstrap(boot, nil, 30000, 1000) != nil {
		return
	}

	redis.Subscribe(proxy.Post)
	InitEzspModule(&config.Ezsp)
	initInfo := redis.GetPubMQTTInfo()
	mqtt.Init(initInfo, proxy.Down)

	errs := make(chan error, 4)

	// 监听终端退出
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go zgb.TickRunning(errs)
	go backup.ConsumeBackupQueue()
	go router.Run(52040)

	// recv error channel
	c := <-errs
	common.Log.Errorf("terminating: %v", c)
	os.Exit(0)

}

// InitEzspModule 初始化Ezsp
func InitEzspModule(cfg *EZSPConfig) {
	hetu.HetuCallbacks = hetu.StHetuCallbacks{
		HetuMessageSentHandler:     handler.SentMessage,
		HetuIncomingMessageHandler: handler.ReceiveMessage,
		HetuNodeStatusHandler:      handler.NodeStatus,
	}

	zgb.TraceSet(&cfg.TraceSettings)
	zgb.NetworkSet(&cfg.NetworkSettings)

	err := ash.AshSerialOpen(cfg.Serial.Name, cfg.Serial.Baud, cfg.Serial.RtsCts)
	if err != nil {
		common.Log.Errorf("failed to open serial %v", cfg.Serial.Name)
	}

	// Time it took to start service
	common.Log.Infof("Open Serial success port=%s baud=%d", cfg.Serial.Name, cfg.Serial.Baud)
	// 初始化 长短地址对应表
	nodesMap := redis.ReadSaveZigbeeNodeTable()
	hetu.LoadNodesMap(nodesMap)
}
