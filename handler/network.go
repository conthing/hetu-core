package handler

import (
	"hetu-core/dto"
	"hetu-core/redis"
	"net/http"

	"github.com/conthing/ezsp/ezsp"
	"github.com/conthing/ezsp/hetu"

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

		err = hetu.SetPermission(255)
		if err != nil {
			common.Log.Errorf("SetPermission failed: %v", err)
			c.JSON(http.StatusBadGateway, &dto.Resp{
				Code:    dto.CreateZigbeeNetFailed,
				Message: "PermitJoin Failed",
			})
			return
		}
		common.Log.Infof("SetPermission OK")
		// -------------
		c.JSON(http.StatusOK, &dto.Resp{Code: dto.Success, Message: "入网开始"})
	case "StopJoin":
		err = hetu.SetPermission(0)
		if err != nil {
			common.Log.Errorf("StopJoin failed: %v", err)
			c.JSON(http.StatusBadGateway, &dto.Resp{
				Code:    dto.CreateZigbeeNetFailed,
				Message: "StopJoin Failed",
			})
			return
		}
		common.Log.Infof("StopJoin OK")
		// -------------
		c.JSON(http.StatusOK, &dto.Resp{Code: dto.Success, Message: "StopJoin"})
	case "CreateZigbeeNet":
		err = hetu.FormNetwork(net.Channel)
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
			c.JSON(http.StatusInternalServerError, &dto.Resp{
				Code:    dto.RemoveZigbeeNetFailed,
				Message: "RemoveZigbeeNet Failed",
			})
			common.Log.Errorf("RemoveZigbeeNet failed: %v", err)
			return
		}
		err = redis.DeleteNodeList()
		if err != nil {
			c.JSON(http.StatusInternalServerError, &dto.Resp{
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

// GetMeshInfo 获取频道号
func GetMeshInfo(c *gin.Context) {
	if !ezsp.MeshStatusUp {
		c.JSON(http.StatusBadRequest, "网络未开启")
		return
	}
	messInfo := ezsp.MeshInfo
	c.JSON(http.StatusOK, &messInfo)
}
