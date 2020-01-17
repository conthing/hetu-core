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
		v1.PUT("/network", handler.NetworkHandler)
		v1.GET("/nodes", handler.GetZigbeeNodes)

	}
	r.Run(fmt.Sprintf(":%d", port))
}
