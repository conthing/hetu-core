package router

import (
	"fmt"
	"hetu/handler"

	"github.com/gin-gonic/gin"
)

// Run 启动 HTTP 服务
func Run(port int) {
	r := gin.Default()
	r.Group("/api/v1")
	{
		r.PUT("/network", handler.NetworkHandler)
	}
	r.Run(fmt.Sprintf(":%d", port))
}
