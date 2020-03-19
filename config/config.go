package config

import (
	"bytes"
	"io/ioutil"
	"os"

	"github.com/conthing/ezsp/zgb"

	"github.com/conthing/utils/common"
	"gopkg.in/yaml.v2"
)

const configFile = "config.yaml"

// Mac 全局变量 Mac 只读
var Mac = common.GetSerialNumber()

// Config 配置模型
type Config struct {
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

var Conf = Config{
	Serial: Serial{
		Name:   "S1",
		Baud:   57600,
		RtsCts: false,
	},
	TraceSettings: zgb.StTraceSettings{
		EzspCallbackTraceOn: true,
		NcpTraceOn:          true,
		NcpFormTraceOn:      true,
	},
	NetworkSettings: zgb.StNetworkSettings{
		NetworkType:   "hetu",
		SecurityLevel: 0,
	},
}

// Service 配置服务
func Service() {

	if !exists(configFile) {
		createConfigFile()
	}
	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		common.Log.Error("读取配置文件失败: ", err)
	}

	err = yaml.Unmarshal(yamlFile, &Conf)
	if err != nil {
		common.Log.Error("配置文件序列化失败: ", err)
	}

}

// exists 判断所给路径文件/文件夹是否存在
func exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func createConfigFile() {
	buf := new(bytes.Buffer)
	err := yaml.NewEncoder(buf).Encode(Conf)
	if err != nil {
		common.Log.Error("配置文件编码失败: ", err)
	}

	f, err := os.Create(configFile)
	if err != nil {
		common.Log.Error("配置文件创建失败: ", err)
	}
	defer func() {
		err := f.Close()
		if err != nil {
			common.Log.Error("配置文件关闭失败: ", err)
		}

	}()

	_, err = f.Write(buf.Bytes())
	if err != nil {
		common.Log.Error("配置文件写入失败", err)
	}
}
