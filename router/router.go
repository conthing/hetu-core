package router

import (
	"fmt"
	"hetu-core/handler"

	"github.com/gin-gonic/gin"
)

// Run 启动 HTTP 服务
func Run(port int) {
	r := gin.Default()
	r.Group("/api/v1")
	{
		r.PUT("/network", handler.NetworkHandler)
		r.GET("/nodes", handler.GetZigbeeNodes)

	}
	r.Run(fmt.Sprintf(":%d", port))
}
