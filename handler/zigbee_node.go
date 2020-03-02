package handler

import (
	"encoding/hex"
	"hetu-core/dto"
	"hetu-core/redis"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetZigbeeNodes 获取ZigbeeNodes
func GetZigbeeNodes(c *gin.Context) {
	nodes, err := redis.GetNodeList()
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Resp{
			Code:    dto.GetNodeListFailed,
			Message: err.Error(),
		})
	}
	znodes := make([]dto.ZNode, len(nodes))
	for index, node := range nodes {
		znodes[index].Node = node
		znodes[index].Mac = strconv.FormatUint(node.Eui64, 16)
	}
	c.JSON(http.StatusOK, dto.Resp{
		Code: dto.Success,
		Data: znodes,
	})
}

// GetNodeLatestMessage 获取最新的Node
func GetNodeLatestMessage(c *gin.Context) {
	mac := c.Param("mac")
	message, err := redis.GetNodeLatestMessage(mac)

	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Resp{
			Code:    dto.GetNodeLatestMessageFailed,
			Message: err.Error(),
		})
		return
	}

	str := hex.EncodeToString(message.Message)

	c.JSON(http.StatusOK, dto.Resp{
		Code: dto.Success,
		Data: str,
	})

}
