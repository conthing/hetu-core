package handler

import (
	"hetu-core/dto"
	"hetu-core/ezsp/hetu"
	"net/http"

	"github.com/conthing/utils/common"
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
		c.JSON(http.StatusOK, &dto.Resp{Code: dto.Success, Message: "入网开始"})
	case "CreateZigbeeNet":
		err = hetu.FormNetwork(0xff)
		if err != nil {
			c.JSON(http.StatusBadGateway, &dto.Resp{
				Code:    dto.CreateZigbeeNetFailed,
				Message: "CreateZigbeeNet Failed",
			})
			common.Log.Errorf("FormNetwork failed: %v", err)
			return
		}
		c.JSON(http.StatusOK, &dto.Resp{Code: dto.Success, Message: "建网成功"})
	case "RemoveZigbeeNet":
		err = hetu.RemoveNetwork()
		if err != nil {
			c.JSON(http.StatusBadGateway, &dto.Resp{
				Code:    dto.RemoveZigbeeNetFailed,
				Message: "RemoveZigbeeNet Failed",
			})
			common.Log.Errorf("RemoveZigbeeNet failed: %v", err)
			return
		}
		c.JSON(http.StatusOK, &dto.Resp{Code: dto.Success, Message: "删网成功"})

	default:
		c.JSON(http.StatusBadRequest, &dto.Resp{Code: dto.InvalidJSON, Message: "invalid json"})
	}

}
