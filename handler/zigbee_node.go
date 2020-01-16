package handler

import (
	"encoding/json"
	"hetu-core/dto"
	"hetu-core/redis"

	"github.com/conthing/utils/common"
	"github.com/gin-gonic/gin"
)

// GetZigbeeNodes 获取ZigbeeNodes
func GetZigbeeNodes(c *gin.Context) {
	nodes := make([]dto.ZigbeeNode, 0)

	m := redis.ReadSaveZigbeeNodeTable()

	for _, nodeStr := range m {
		var node dto.ZigbeeNode
		err := json.Unmarshal([]byte(nodeStr), &node)
		if err != nil {
			common.Log.Error("node 序列化错误", err)
			continue
		}
		if node.Message == nil {
			continue
		}
		nodes = append(nodes, node)
	}

	c.JSON(200, nodes)
}
