package redis

import (
	"encoding/json"
	"fmt"
	"hetu-core/dto"
	"strconv"
	"time"

	"github.com/conthing/ezsp/hetu"

	"github.com/conthing/utils/common"
	"github.com/mediocregopher/radix/v3"
)

// Client redis调用客户端
var Client *radix.Pool

// Connect 初始化连接池
func Connect() (err error) {
	Client, err = radix.NewPool("tcp", "127.0.0.1:6379", 10)
	if err != nil {
		return
	}
	err = InitPubSubConn()
	if err != nil {
		return
	}
	// redis慢热，导致zigbee节点没有同步到ezsp，这里在初始化时增加一个读取，如果出错不运行下去
	ZigbeeNodeList := make([]string, 0)
	err = Client.Do(radix.Cmd(&ZigbeeNodeList, "smembers", "ZigbeeNodeSet"))
	if err != nil {
		//if strings.Contains(err.Error(),"LOADING") {
		return err
		//}
	}
	return nil
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

		common.Log.Info("[ok] DB 保存 Zigbee 节点成功")
		// capture the output of the first transaction command, i.e. the GET
		return nil

	}))

	if err != nil {
		common.Log.Error("保存 Zigbee 节点失败")
	}

}

const zigbeeMessageQueue = "zigbee_message_queue" // todo 为什么trim保持64个，读却只读一个

// AddToZigbeeMessageQueue 加入 ZigbeeMessage 队列头部
func AddToZigbeeMessageQueue(m *dto.ZigbeeDeviceMessage) {

	str, err := m.String()
	if err != nil {
		common.Log.Error("String:", err)
		return
	}
	key := zigbeeMessageQueue + strconv.FormatUint(m.Eui64, 16)
	err = Client.Do(radix.Cmd(nil, "LPUSH", key, str))

	if err != nil {
		common.Log.Error("LPUSH:", err)
	}
}

// TrimZigbeeMessageQueue 剔除老旧数据
// 容量64
func TrimZigbeeMessageQueue(m *dto.ZigbeeDeviceMessage) {
	key := zigbeeMessageQueue + strconv.FormatUint(m.Eui64, 16)

	err := Client.Do(radix.Cmd(nil, "LTRIM", key, "0", "63"))
	if err != nil {
		common.Log.Error("TrimZigbeeMessageQueue LTRIM:", err)
		return
	}
	common.Log.Info("[success] TrimBackupQueue")

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

// DeleteNodeList 删除网络节点
func DeleteNodeList() error {
	key := "ZigbeeNodeSet"
	err := Client.Do(radix.Cmd(nil, "DEL", key))
	if err != nil {
		common.Log.Error("failed to delete ZigbeeNodeSet ", err)
		return err
	}
	return nil

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
	return NodeList, nil
}

// GetNodeLatestMessage 获取节点的最新message 根据16进制的mac
func GetNodeLatestMessage(mac string) (*dto.ZigbeeDeviceMessage, error) {
	var dto dto.ZigbeeDeviceMessage
	var str string
	err := Client.Do(radix.Cmd(&str, "LINDEX", zigbeeMessageQueue+mac, "0"))
	if err != nil {
		return nil, fmt.Errorf("LINDEX:%w", err)
	}
	err = json.Unmarshal([]byte(str), &dto)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal:%w", err)
	}
	common.Log.Info("获取节点的最新 message 成功")
	return &dto, nil
}
