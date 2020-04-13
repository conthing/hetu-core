package redis

import (
	"hetu-core/config"
	"hetu-core/dto"
	"log"
	"strconv"
	"time"

	"github.com/conthing/ezsp/hetu"

	"github.com/conthing/utils/common"
	"github.com/mediocregopher/radix/v3"
)

// Client redis调用客户端
var Client *radix.Pool

// Connect 初始化连接池
func Connect() {
	var err error
	Client, err = radix.NewPool("tcp", "127.0.0.1:6379", 10)
	if err != nil {
		log.Fatal("数据库连接失败", err)
	}
	err = InitPubSubConn()
	if err != nil {
		common.Log.Error("pubsub 连接失败", err)
		return
	}
	common.Log.Info("redis 启动成功")
	// 初始化 host alias
	alias, err := GetAlias()
	if err != nil {
		common.Log.Error("get alias failed", err)
		return
	}
	if alias != "" {
		// alias has been initialized.
		return
	}

	Maclen := len(config.Mac)
	subMac := config.Mac[Maclen-6 : Maclen]
	err = SaveAlias("fortoo_" + subMac)
	if err != nil {
		common.Log.Error("init alias failed", err)
		return
	}

}

// SaveZigbeeNode @HMSET
// 保存到数据库
func SaveZigbeeNode(node *dto.ZigbeeNode) {
	key := strconv.FormatUint(node.Eui64, 16)
	err := Client.Do(radix.WithConn(key, func(c radix.Conn) error {
		// Begin the transaction with a MULTI command
		if err := c.Do(radix.Cmd(nil, "MULTI")); err != nil {
			return err
		}

		// If any of the calls after the MULTI call error it's important that
		// the transaction is discarded. This isn't strictly necessary if the
		// error was a network error, as the connection would be closed by the
		// client anyway, but it's important otherwise.
		var err error
		defer func() {
			if err != nil {
				// The return from DISCARD doesn't matter. If it's an error then
				// it's a network error and the Conn will be closed by the
				// client.
				c.Do(radix.Cmd(nil, "DISCARD"))
			}
		}()

		// queue up the transaction's commands
		if err = c.Do(radix.FlatCmd(nil, "HMSET", key, node)); err != nil {
			return err
		}
		if err = c.Do(radix.Cmd(nil, "SADD", "ZigbeeNodeSet", key)); err != nil {
			return err
		}

		// execute the transaction, capturing the result
		var result []string
		if err = c.Do(radix.Cmd(&result, "EXEC")); err != nil {
			return err
		}

		common.Log.Info("[ok] 保存 Zigbee 节点成功")
		// capture the output of the first transaction command, i.e. the GET
		return nil

	}))

	if err != nil {
		common.Log.Error("保存 Zigbee 节点失败")
	}

}

// SaveZigbeeMessage 保存 ZigbeeMessage 信息
func SaveZigbeeMessage(m *dto.ZigbeeDeviceMessage) {
	key := m.UUID.String()
	err := Client.Do(radix.WithConn(key, func(c radix.Conn) error {
		// Begin the transaction with a MULTI command
		if err := c.Do(radix.Cmd(nil, "MULTI")); err != nil {
			return err
		}

		// If any of the calls after the MULTI call error it's important that
		// the transaction is discarded. This isn't strictly necessary if the
		// error was a network error, as the connection would be closed by the
		// client anyway, but it's important otherwise.
		var err error
		defer func() {
			if err != nil {
				// The return from DISCARD doesn't matter. If it's an error then
				// it's a network error and the Conn will be closed by the
				// client.
				c.Do(radix.Cmd(nil, "DISCARD"))
			}
		}()

		// queue up the transaction's commands
		if err = c.Do(radix.FlatCmd(nil, "HMSET", key, m)); err != nil {
			return err
		}

		macGroupKey := "mac2uuid:" + strconv.FormatUint(m.Eui64, 16)

		if err = c.Do(radix.Cmd(nil, "LPUSH", macGroupKey, key)); err != nil {
			return err
		}

		// execute the transaction, capturing the result
		var result []string
		if err = c.Do(radix.Cmd(&result, "EXEC")); err != nil {
			return err
		}

		// capture the output of the first transaction command, i.e. the GET
		return nil

	}))

	if err != nil {
		common.Log.Error("保存 Zigbee Message 失败")
	}
	common.Log.Info("[ok] 保存 Zigbee Message 成功")
}

// TrimZigbeeMessage 剔除老旧数据
// 容量2700
func TrimZigbeeMessage(m *dto.ZigbeeDeviceMessage) {
	var res []string
	macGroupKey := "mac2uuid:" + strconv.FormatUint(m.Eui64, 16)
	err := Client.Do(radix.Cmd(&res, "LRANGE", macGroupKey, "0", "-1"))
	if err != nil {
		common.Log.Error("TrimZigbeeMessage LRANGE error: ", err)
		return
	}

	common.Log.Info(res, len(res))
	if len(res) <= 2700 {
		return
	}
	var popUUID string
	err = Client.Do(radix.Cmd(&popUUID, "RPOP", macGroupKey))
	if err != nil {
		common.Log.Error("TrimZigbeeMessage LPOP error: ", err)
		return
	}

	err = Client.Do(radix.Cmd(nil, "DEL", popUUID))
	if err != nil {
		common.Log.Error("TrimZigbeeMessage DEL error: ", err)
		return
	}
	common.Log.Info("[success] TrimZigbeeMessage")

}

// ReadSaveZigbeeNodeTable @set
// 读取对应表
func ReadSaveZigbeeNodeTable() map[uint64]hetu.StNode {
	nodesMap := make(map[uint64]hetu.StNode)
	// PART 1 读取节点列表
	ZigbeeNodeList := make([]string, 0)
	key := "ZigbeeNodeSet"
	err := Client.Do(radix.Cmd(&ZigbeeNodeList, "smembers", key))
	if err != nil {
		common.Log.Error("读取节点表失败", err)
		common.Log.Warn("使用空记录")
		return nodesMap
	}
	common.Log.Info("[PART 1] 读取节点列表成功", ZigbeeNodeList)

	// PART 2 加载节点列表
	NodeList := make([]dto.ZigbeeNode, len(ZigbeeNodeList))
	CmdActionList := make([]radix.CmdAction, 0)
	for index := range NodeList {
		CmdActionList = append(CmdActionList, radix.Cmd(&NodeList[index], "HGETALL", ZigbeeNodeList[index]))
	}
	p := radix.Pipeline(CmdActionList...)
	err = Client.Do(p)
	if err != nil {
		common.Log.Error("保存 Zigbee Message 失败", err)
		return nodesMap
	}
	common.Log.Info("[PART 2] 加载节点列表成功")

	// PART 3 读取节点时间
	timeList := make([]time.Time, len(ZigbeeNodeList))
	CmdActionList = make([]radix.CmdAction, 0)
	for index := range timeList {
		CmdActionList = append(CmdActionList, radix.Cmd(&timeList[index], "HGET", ZigbeeNodeList[index], "LastRecvTime"))
	}
	p = radix.Pipeline(CmdActionList...)
	err = Client.Do(p)
	if err != nil {
		common.Log.Error("读取节点时间失败", err)
		return nodesMap
	}
	common.Log.Info("[PART 3] 读取节点时间")

	// PART 4 整合为 Map
	for index, node := range NodeList {
		var stNode hetu.StNode
		stNode.LastRecvTime = timeList[index]
		stNode.Eui64 = node.Eui64
		stNode.NodeID = node.NodeID
		stNode.Addr = node.Addr
		// stNode.Addr = node.Addr
		nodesMap[node.Eui64] = stNode
	}
	common.Log.Info("[PART 4] 整合为 Map")
	return nodesMap
}

// GetNodeList 获取节点列表
func GetNodeList() ([]dto.ZigbeeNode, error) {
	NodeList := make([]dto.ZigbeeNode, 0)
	// PART 1 读取节点列表
	var ZigbeeNodeList []string
	key := "ZigbeeNodeSet"
	err := Client.Do(radix.Cmd(&ZigbeeNodeList, "smembers", key))
	if err != nil {
		common.Log.Error("读取节点表失败", err)
		return nil, err
	}
	common.Log.Info("[PART 1] 读取节点列表成功")

	// PART 2 加载节点列表
	NodeList = make([]dto.ZigbeeNode, len(ZigbeeNodeList))
	CmdActionList := make([]radix.CmdAction, 0)
	for index := range NodeList {
		CmdActionList = append(CmdActionList, radix.Cmd(&NodeList[index], "HGETALL", ZigbeeNodeList[index]))
	}
	p := radix.Pipeline(CmdActionList...)
	err = Client.Do(p)
	if err != nil {
		common.Log.Error("加载节点列表失败", err)
		return nil, err
	}
	common.Log.Info("[PART 2] 加载节点列表成功", NodeList)
	return NodeList, nil
}

// GetNodeLatestMessage 获取节点的最新message 根据 eui64
// todo 修改获取队头元素策略
func GetNodeLatestMessage(mac string) (*dto.ZigbeeDeviceMessage, error) {

	// Stage 1 获取 uuid
	var latestUUID string
	macGroupKey := "mac2uuid:" + mac
	err := Client.Do(radix.Cmd(&latestUUID, "LINDEX", macGroupKey, "0"))
	if err != nil {
		common.Log.Error("读取 uuid 错误", err)
		return nil, err
	}

	// Stage 2 获取 message
	var message dto.ZigbeeDeviceMessage
	err = Client.Do(radix.Cmd(&message, "HGETALL", latestUUID))
	if err != nil {
		common.Log.Error("获取 message 错误", err)
		return nil, err
	}

	common.Log.Info("获取节点的最新 message 成功")
	return &message, nil

}
