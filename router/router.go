package router

import (
	"fmt"
	"hetu-core/handler"

	"github.com/gin-gonic/gin"
)

// Run 启动 HTTP 服务
func Run(port int) {
	r := gin.Default()
	v1 := r.Group("/api/v1")
	{
		v1.GET("/ping", handler.Ping)
		v1.PUT("/network", handler.NetworkHandler)
		v1.GET("/network", handler.GetMeshInfo)
		v1.GET("/nodes", handler.GetZigbeeNodes)
		v1.GET("/nodes/:mac", handler.GetNodeLatestMessage)
		v1.DELETE("/nodes/:mac", handler.DeleteNode)
		v1.GET("/version", handler.GetVersionInfo)

		// 上报设置
		v1.POST("/PubHTTP", handler.SetPubHTTP)
		v1.GET("/PubHTTP", handler.GetPubHTTPInfo)
		v1.POST("/PubMQTT", handler.SetPubMQTT)
		v1.GET("/PubMQTT", handler.GetPubMQTTInfo)

		// 指令接受
		v1.POST("/command", handler.HandleCommand)
	}
	r.Run(fmt.Sprintf(":%d", port))
}
