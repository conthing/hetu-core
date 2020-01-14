package handler

import (
	"hetu/dto"
	"hetu/ezsp/hetu"
	"net/http"

	"github.com/gin-gonic/gin"
)

// NetworkHandler 控制 Zigbee 主机网络的开关
func NetworkHandler(c *gin.Context) {
	var net dto.Network
	err := c.ShouldBindJSON(&net)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid json")
		return
	}
	switch net.Command {
	case "PermitJoin":
		// -------------
		p := new(hetu.StPermission)
		p.Duration = byte(255)
		p.Passports = make([]*hetu.StPassport, 2)
		p.Passports[0].MAC = "xxxxxxxxxxxxce73"
		p.Passports[1].MAC = "xxxxxxxxxxxxce73"
		hetu.SetPermission(p)
		// -------------
		c.JSON(http.StatusOK, &dto.Resp{Code: 0, Message: "入网开始"})
	case "RemoveAll":
		hetu.RemoveNetwork()
		c.JSON(http.StatusOK, &dto.Resp{Code: 0, Message: "离网成功"})
	default:
		c.JSON(http.StatusBadRequest, &dto.Resp{Code: 0, Message: "invalid json"})
	}

}
