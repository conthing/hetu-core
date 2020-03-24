package handler

import (
	"hetu-core/dto"
	"hetu-core/proxy"

	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleCommand 处理远程服务器下发指令
func HandleCommand(c *gin.Context) {
	var info dto.ReceiveMessageDTO
	err := c.ShouldBindJSON(&info)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	proxy.Down(&info)
}
