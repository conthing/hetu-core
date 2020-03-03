package handler

import (
	"hetu-core/dto"
	mqtt "hetu-core/mqtt/client"
	"hetu-core/redis"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetPubHTTP HTTP 上报设置
// http 从数据库里查
func SetPubHTTP(c *gin.Context) {
	var info dto.PubHTTPInfo
	err := c.ShouldBindJSON(&info)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	redis.SavePubHTTP(&info)
}

// GetPubHTTPInfo 获取
func GetPubHTTPInfo(c *gin.Context) {
	info := redis.GetPubMQTTInfo()
	c.JSON(http.StatusOK, info)
}

// SetPubMQTT MQTT 上报设置
// mqtt 重制连接
func SetPubMQTT(c *gin.Context) {
	var info dto.PubMQTTInfo
	err := c.ShouldBindJSON(&info)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	err = redis.SavePubMQTT(&info)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	mqtt.ReConnect(&info)
}

// GetPubMQTTInfo 获取
func GetPubMQTTInfo(c *gin.Context) {
	info := redis.GetPubMQTTInfo()
	c.JSON(http.StatusOK, info)
}
