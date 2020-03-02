package handler

import (
	"net/http"

	"github.com/conthing/utils/common"
	"github.com/gin-gonic/gin"
)

// GetVersionInfo 获取服务版本
func GetVersionInfo(c *gin.Context) {
	c.String(http.StatusOK, "%s %s", common.Version, common.BuildTime)
}
