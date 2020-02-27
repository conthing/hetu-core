package handler

import (
	"hetu-core/dto"
	"hetu-core/redis"
	"net/http"

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
	c.JSON(http.StatusOK, dto.Resp{
		Code: dto.Success,
		Data: nodes,
	})
}
