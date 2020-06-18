package redis

import (
	"hetu-core/dto"

	"github.com/conthing/utils/common"

	"github.com/mediocregopher/radix/v3"
)

// SavePubHTTP 保存 HTTP 上报方法
func SavePubHTTP(info *dto.PubHTTPInfo) error {

	key := "PubHTTPInfo"
	err := Client.Do(radix.FlatCmd(nil, "HMSET", key, info))
	if err != nil {
		common.Log.Error("设置 PubHTTPInfo 错误", err)
		return err
	}
	return nil
}

// GetPubHTTPInfo 获取 PubHTTPInfo
func GetPubHTTPInfo() *dto.PubHTTPInfo {
	key := "PubHTTPInfo"
	var info dto.PubHTTPInfo
	err := Client.Do(radix.FlatCmd(&info, "HGETALL", key))
	if err != nil {
		common.Log.Error("获取 PubHTTPInfo 错误", err)
	}
	return &info
}

// SavePubMQTT 保存 MQTT 上报方法
func SavePubMQTT(info *dto.PubMQTTInfo) error {
	key := "PubMQTTInfo"
	err := Client.Do(radix.FlatCmd(nil, "HMSET", key, info))
	if err != nil {
		common.Log.Error("设置 PubMQTTInfo 错误", err)
		return err
	}
	return nil
}

// GetPubMQTTInfo 获取 PubMQTTInfo
func GetPubMQTTInfo() *dto.PubMQTTInfo {

	key := "PubMQTTInfo"
	var info dto.PubMQTTInfo
	err := Client.Do(radix.FlatCmd(&info, "HGETALL", key))
	if err != nil {
		common.Log.Error("获取 PubMQTTInfo 错误", err)
	}

	return &info

}
