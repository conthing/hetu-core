package hetu

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"hetu-core/ezsp/ezsp"

	"encoding/binary"
	"hetu-core/ezsp/zcl"

	"github.com/conthing/utils/common"
)

const (
	ZDO_PROFILE = uint16(0x0000)

	//C4入网进程的几个状态
	C4_STATE_NULL       = byte(0) //初始化时
	C4_STATE_CONNECTING = byte(1) //为NUL时有report
	C4_STATE_ONLINE     = byte(2) //announce完成，offline后又收到报文，其中TC是新join的情况下announce完成会触发newnode的应用层事件
	C4_STATE_OFFLINE    = byte(3) //接收超时

	C4_MAX_OFFLINE_TIMEOUT = 30

	C4_NODE_STATUS_OFFLINE = byte(0)
	C4_NODE_STATUS_ONLINE  = byte(1)
	C4_NODE_STATUS_REBOOT  = byte(2)
	C4_NODE_STATUS_DELETED = byte(3)
)

var ErrMeshNotExist = errors.New("Mesh not exist")
var ErrMeshAlreadyExist = errors.New("Mesh already exist")
var ErrMeshNotEmpty = errors.New("Not empty mesh")

type StNode struct {
	NodeID       uint16
	Eui64        uint64
	LastRecvTime time.Time
	State        byte
	Newjoin      bool
	ToBeDeleted  bool
}

//eui64 要转成 16进制 mac
type StC4Callbacks struct {
	C4MessageSentHandler     func(eui64 uint64, profileId uint16, clusterId uint16, localEndpoint byte, remoteEndpoint byte, message []byte, success bool)
	C4IncomingMessageHandler func(eui64 uint64, profileId uint16, clusterId uint16, localEndpoint byte, remoteEndpoint byte, message []byte)
	C4NodeStatusHandler      func(eui64 uint64, nodeID uint16, status byte, deviceType byte)
}

var C4Callbacks StC4Callbacks

var Nodes sync.Map

// LoadNodesMap 加载 Map
func LoadNodesMap(m map[uint64]StNode) {
	for _, node := range m {
		StoreNode(&node)
		common.Log.Info("1 LoadNodesMap: ", node.NodeID)
	}
}

// StoreNode 保存节点到Nodes中，如果有重复的eui64，更新
func StoreNode(node *StNode) {
	nodeID := findNodeIDbyEui64(node.Eui64)
	if nodeID == ezsp.EMBER_NULL_NODE_ID {
		Nodes.Store(node.NodeID, *node) // map中存储
	} else {
		Nodes.Delete(nodeID)            // map中原来的删掉
		Nodes.Store(node.NodeID, *node) // map中存储

	}
}

//在Nodes中找到匹配的eui64
func findNodeIDbyEui64(eui64 uint64) (nodeID uint16) {
	nodeID = ezsp.EMBER_NULL_NODE_ID
	Nodes.Range(func(key, value interface{}) bool {
		if node, ok := value.(StNode); ok {
			if node.Eui64 == eui64 {
				nodeID = node.NodeID
				return false
			}
		}
		return true
	})
	return
}

var lastTimeStamp = int64(0)

func HetuTick() {
	select {
	case cbs := <-ezsp.CallbackCh:
		for _, cb := range cbs {
			ezsp.EzspCallbackDispatch(cb)
		}
	case <-time.After(time.Millisecond * 500):

	}
	now := time.Now().Unix()
	if now-lastTimeStamp >= 10 {
		lastTimeStamp = now
		Nodes.Range(func(key, value interface{}) bool {
			if node, ok := value.(StNode); ok {
				node.RefreshHandle()
			}
			return true
		})
		err := HetuBroadcast()
		common.Log.Info("hetu tick: ", err)

	}
}

func (_ *StNode) UnsupportClusterCommandHandle(z *zcl.ZclContext, cluster uint16, direction bool, disableDefaultResponse bool, sequenceNumber byte,
	commandIdentifier byte, data interface{}) (resp []byte, err error) {
	return nil, nil
}

func (node *StNode) getState() byte {
	now := time.Now()

	timeout := C4_MAX_OFFLINE_TIMEOUT * time.Second
	if now.Sub(node.LastRecvTime) > timeout {
		return C4_STATE_OFFLINE
	}
	return C4_STATE_ONLINE

}

func removeDeviceAndNode(node *StNode) {
	err := ezsp.EzspRemoveDevice(node.NodeID, node.Eui64, node.Eui64)
	if err != nil {
		common.Log.Errorf("EzspRemoveDevice failed: %v", err)
	}

	Nodes.Delete(node.NodeID)
}

// RefreshHandle 收到报文发生变化，或定时刷新时调用
func (node *StNode) RefreshHandle() {
	if node.ToBeDeleted {
		common.Log.Debugf("HetuNodeStatusHandler delete")
		if C4Callbacks.C4NodeStatusHandler != nil {
			C4Callbacks.C4NodeStatusHandler(node.Eui64, node.NodeID, C4_NODE_STATUS_DELETED, 0)
		}
		Nodes.Delete(node.NodeID)
		common.Log.Infof("node map delete 0x%016x", node.Eui64)
		return
	}

	newState := node.getState()
	common.Log.Debugf("newState:%d", newState)

	if newState != node.State {
		if newState == C4_STATE_ONLINE {
			if node.State < C4_STATE_ONLINE && node.Newjoin { //初次入网，且NULL、CONNECTING变成ONLINE，要检查passport，不允许的踢出
				if passportMatch(fmt.Sprintf("%016x", node.Eui64)) >= 0 {
					common.Log.Infof("node 0x%016x join network", node.Eui64)
				} else {
					common.Log.Errorf("reject node 0x%016x, remove it", node.Eui64)
					removeDeviceAndNode(node)
					return
				}
			} else {
				common.Log.Infof("node 0x%016x reonline", node.Eui64)
			}
			common.Log.Debugf("HetuNodeStatusHandler online")
			if C4Callbacks.C4NodeStatusHandler != nil {
				C4Callbacks.C4NodeStatusHandler(node.Eui64, node.NodeID, C4_NODE_STATUS_ONLINE, 0)
			}
		} else if newState == C4_STATE_OFFLINE {
			common.Log.Infof("node 0x%016x offline", node.Eui64)
			common.Log.Debugf("HetuNodeStatusHandler offline")
			if C4Callbacks.C4NodeStatusHandler != nil {
				C4Callbacks.C4NodeStatusHandler(node.Eui64, node.NodeID, C4_NODE_STATUS_OFFLINE, 0)
			}
		}
		node.State = newState
	}
	Nodes.Store(node.NodeID, *node) // map中存储
	common.Log.Info("4 RefreshHandle: ", node.NodeID)

}

// StPermission 发送SetPermission请求时参数的结构
type StPermission struct {
	Duration  byte          `json:"duration"`
	Passports []*StPassport `json:"passports"`
}

type StPassport struct {
	MAC string `json:"mac"`
}

var allPassPorts []*StPassport

//返回Match的字符个数，不含x
func checkPassportMAC(mac string, ppMac string) (match int) {
	if len([]rune(mac)) != 16 {
		common.Log.Errorf("MAC len not 16: %s", mac)
		return -1
	}
	for i, c := range mac {
		ppc := []rune(ppMac)[i]
		if !(((c >= '0') && (c <= '9')) || ((c >= 'a') && (c <= 'f'))) {
			return -1
		}
		if ppc == 'x' {
			continue
		}
		if ppc != c {
			return -1
		}
		match++
	}
	return
}

func passportMatch(mac string) (maxHit int) {
	maxHit = -1
	if allPassPorts != nil {
		for _, p := range allPassPorts {
			match := checkPassportMAC(mac, p.MAC)
			if match > maxHit {
				maxHit = match
				if maxHit >= 16 {
					break
				}
			}
		}
	}
	return
}

//16位hex字符或前置若干个x都是合法的
func isPassportMACValid(mac string) bool {
	leadingx := true
	if len([]rune(mac)) != 16 {
		common.Log.Errorf("MAC len not 16: %s", mac)
		return false
	}
	for _, c := range mac {
		//hex字符一定正确
		if ((c >= '0') && (c <= '9')) || ((c >= 'a') && (c <= 'f')) {
			leadingx = false //出现了hex字符
			continue
		}
		//
		if (c != 'x') || (leadingx == false) {
			return false
		}
	}
	return true
}

func SetPermission(permission *StPermission) (err error) {
	common.Log.Debugf("SetPermission %+v", *permission)
	if permission == nil {
		err = fmt.Errorf("C4SetPassports permission=NULL")
		return
	}
	if permission.Passports == nil || len(permission.Passports) == 0 {
		err = fmt.Errorf("C4SetPassports passports=NULL")
		return
	}
	common.Log.Debugf("permision to ...")
	for i, p := range permission.Passports {
		if p != nil {
			common.Log.Debugf("%d: MAC=%s", i, p.MAC)
			if isPassportMACValid(p.MAC) == false {
				err = fmt.Errorf("C4SetPassports passport %d mac -%s- invalid", i, p.MAC)
				return
			}
		}
	}
	allPassPorts = permission.Passports

	err = ezsp.EzspPermitJoining(permission.Duration)
	if err != nil {
		err = fmt.Errorf("EzspPermitJoining failed: %v", err)
	}

	return
}

func C4Init() {
	ezsp.NcpCallbacks.NcpTrustCenterJoinHandler = TrustCenterJoinHandler
	ezsp.NcpCallbacks.NcpMessageSentHandler = MessageSentHandler
	ezsp.NcpCallbacks.NcpIncomingMessageHandler = IncomingMessageHandler
	ezsp.NcpCallbacks.NcpIncomingSenderEui64Handler = IncomingSenderEui64Handler

}

func HetuBroadcast() error {
	apsFrame := ezsp.EmberApsFrame{ProfileId: 0xabcd, ClusterId: 0xabef, SourceEndpoint: 2, DestinationEndpoint: 2}
	message := []byte{0x78, 0x87, 0x1b}
	_, err := ezsp.EzspSendBroadcast(ezsp.EMBER_BROADCAST_ADDRESS, &apsFrame, 30, 0, message)
	return err
}

func TrustCenterJoinHandler(newNodeId uint16,
	newNodeEui64 uint64,
	deviceUpdateStatus byte,
	joinDecision byte,
	parentOfNewNode uint16) {

	fmt.Printf("TrustCenterJoinHandler")

	now := time.Now()
	var node StNode
	value, ok := Nodes.Load(newNodeId) // 从map中加载
	if ok {
		if node, ok = value.(StNode); !ok {
			common.Log.Errorf("Nodes map unsupported type")
			return
		}
		node.Newjoin = false //已经在node表里的，不认为Newjoin
		node.LastRecvTime = now
	} else {
		node = StNode{NodeID: newNodeId, Eui64: newNodeEui64, LastRecvTime: now}
		node.Newjoin = true
	}

	if (deviceUpdateStatus == ezsp.EMBER_STANDARD_SECURITY_UNSECURED_JOIN) ||
		(deviceUpdateStatus == ezsp.EMBER_HIGH_SECURITY_UNSECURED_JOIN) {
	} else if (deviceUpdateStatus == ezsp.EMBER_STANDARD_SECURITY_SECURED_REJOIN) ||
		(deviceUpdateStatus == ezsp.EMBER_STANDARD_SECURITY_UNSECURED_REJOIN) ||
		(deviceUpdateStatus == ezsp.EMBER_HIGH_SECURITY_SECURED_REJOIN) ||
		(deviceUpdateStatus == ezsp.EMBER_HIGH_SECURITY_UNSECURED_REJOIN) {
		node.Newjoin = false
	} else if deviceUpdateStatus == ezsp.EMBER_DEVICE_LEFT {
		node.ToBeDeleted = true
	} else {
		common.Log.Errorf("unknown deviceUpdateStatus = 0x%x newNodeId = 0x%04x\r\n", deviceUpdateStatus, newNodeId)
		return
	}
	node.RefreshHandle()
}

func MessageSentHandler(outgoingMessageType byte,
	indexOrDestination uint16,
	apsFrame *ezsp.EmberApsFrame,
	messageTag byte,
	emberStatus byte,
	message []byte) {
	if apsFrame.ProfileId == ZDO_PROFILE {
		return
	}
	if messageTag == 0 { //应用层不关心时tag==0
		return
	}
	var node StNode
	value, ok := Nodes.Load(indexOrDestination) // 从map中加载
	if ok {
		if node, ok = value.(StNode); !ok {
			common.Log.Errorf("Nodes map unsupported type")
			return
		}
	} else {
		common.Log.Errorf("0x%04x not found in Nodes map", indexOrDestination)
		return
	}
	if C4Callbacks.C4MessageSentHandler != nil {
		C4Callbacks.C4MessageSentHandler(node.Eui64, apsFrame.ProfileId, apsFrame.ClusterId, apsFrame.SourceEndpoint, apsFrame.DestinationEndpoint, message, emberStatus == ezsp.EMBER_SUCCESS)
	}
}

//存储
var orphanEui64 uint64
var orphanEui64RecvTime time.Time

func IncomingSenderEui64Handler(eui64 uint64) {
	now := time.Now()

	common.Log.Debugf("IncomingSenderEui64Handler 0x%x", eui64)
	nodeID := findNodeIDbyEui64(eui64)

	if nodeID == ezsp.EMBER_NULL_NODE_ID { //在Nodes中找不到的话，查询NCP是否有保存
		var err error
		nodeID, err = ezsp.EzspLookupNodeIdByEui64(eui64)
		if err != nil { //NCP也没有的话，作为orphan保存，与即将收到的incomingmessage匹配
			orphanEui64 = eui64
			orphanEui64RecvTime = now
			common.Log.Errorf("Incoming message lookup nodeID failed: %v", err)
			return
		}
	}
	var node StNode
	value, ok := Nodes.Load(nodeID) // 从map中加载
	if ok {
		if node, ok = value.(StNode); !ok {
			common.Log.Errorf("Nodes map unsupported type")
			return
		}
		node.Eui64 = eui64
		node.LastRecvTime = now
	} else {
		node = StNode{NodeID: nodeID, Eui64: eui64, LastRecvTime: now}
	}
	Nodes.Store(node.NodeID, node) // map中存储
	common.Log.Info("5 IncomingSenderEui64Handler: ", node.NodeID)
}

func IncomingMessageHandler(incomingMessageType byte,
	apsFrame *ezsp.EmberApsFrame,
	lastHopLqi byte,
	lastHopRssi int8,
	sender uint16,
	bindingIndex byte,
	addressIndex byte,
	message []byte) {

	now := time.Now()
	if apsFrame.ProfileId == ZDO_PROFILE {
		if apsFrame.ClusterId == 0x0013 { //device announce
			nodeID := binary.LittleEndian.Uint16(message[1:])
			eui64 := binary.LittleEndian.Uint64(message[3:])
			StoreNode(&StNode{NodeID: nodeID, Eui64: eui64, LastRecvTime: now})
			common.Log.Debugf("2 zdo announce: 0x%04x,%016x", nodeID, eui64)
		}
	} else {
		if incomingMessageType == ezsp.EMBER_INCOMING_UNICAST {
			var node StNode
			value, ok := Nodes.Load(sender) // 从map中加载
			if ok {
				if node, ok = value.(StNode); !ok {
					common.Log.Errorf("Nodes map unsupported type")
					return
				}
				common.Log.Debugf("Nodes map get %016x", node.Eui64)
				node.LastRecvTime = now
			} else {
				eui64, err := ezsp.EzspLookupEui64ByNodeId(sender)
				if err != nil {
					common.Log.Errorf("Incoming message lookup eui64 failed: %v", err)
					//orphanEui64是200ms以内的，认为是同一个node
					if now.Sub(orphanEui64RecvTime) > time.Millisecond*200 {
						return
					}
					common.Log.Warnf("match with EUI64=%016x recved %v ago", orphanEui64, now.Sub(orphanEui64RecvTime))
					eui64 = orphanEui64
				}

				common.Log.Debugf("EzspLookupEui64ByNodeId get %016x", eui64)
				node = StNode{NodeID: sender, Eui64: eui64, LastRecvTime: now}
			}
			StoreNode(&node)
			common.Log.Debugf("3 HetuIncomingMessageHandler: %d", node.NodeID)
			if C4Callbacks.C4IncomingMessageHandler != nil {
				if node.Eui64 != 0 {
					C4Callbacks.C4IncomingMessageHandler(node.Eui64, apsFrame.ProfileId, apsFrame.ClusterId, apsFrame.DestinationEndpoint, apsFrame.SourceEndpoint, message)
				} else {
					common.Log.Errorf("recv msg from NodeID 0x%04x without EUI64", node.NodeID)
				}
			}
		}
	}
}

var unicastTagSequence = byte(0)

func nextSequence() byte {
	unicastTagSequence++
	if unicastTagSequence == 0 || unicastTagSequence == 0xff {
		unicastTagSequence = 1
	}
	return unicastTagSequence
}

func SendUnicast(eui64 uint64, profileId uint16, clusterId uint16,
	localEndpoint byte, remoteEndpoint byte,
	message []byte, needConfirm bool) (err error) {
	common.Log.Debugf("SendUnicast %016x ...", eui64)

	nodeID := findNodeIDbyEui64(eui64)
	if nodeID == ezsp.EMBER_NULL_NODE_ID {
		return fmt.Errorf("unknow EUI64 %016x", eui64)
	}
	var apsFrame ezsp.EmberApsFrame
	apsFrame.ProfileId = profileId
	apsFrame.ClusterId = clusterId
	apsFrame.SourceEndpoint = localEndpoint
	apsFrame.DestinationEndpoint = remoteEndpoint
	apsFrame.Options = getSendOptions(nodeID, profileId, clusterId, byte(len(message)))
	tag := byte(0)
	if needConfirm {
		tag = nextSequence()
	}

	// todo 设置路由表
	err = ezsp.NcpSetSourceRoute(nodeID)
	_, err = ezsp.EzspSendUnicast(ezsp.EMBER_OUTGOING_DIRECT, nodeID, &apsFrame, tag, message)
	return
}

func getSendOptions(destination uint16, profileId uint16, clusterId uint16, messageLength byte) (options uint16) {
	if profileId == 0xc25d && clusterId == 0x0001 {
		if destination >= ezsp.EMBER_BROADCAST_ADDRESS {
			options = /*ezsp.EMBER_APS_OPTION_RETRY |*/ ezsp.EMBER_APS_OPTION_SOURCE_EUI64
		} else {
			options = ezsp.EMBER_APS_OPTION_RETRY | ezsp.EMBER_APS_OPTION_ENABLE_ROUTE_DISCOVERY | ezsp.EMBER_APS_OPTION_ENABLE_ADDRESS_DISCOVERY | ezsp.EMBER_APS_OPTION_SOURCE_EUI64 | ezsp.EMBER_APS_OPTION_DESTINATION_EUI64
		}
	} else {

		if messageLength <= 66 { /*66不溢*/
			options = /*ezsp.EMBER_APS_OPTION_RETRY | */ ezsp.EMBER_APS_OPTION_SOURCE_EUI64 | ezsp.EMBER_APS_OPTION_DESTINATION_EUI64
		} else if messageLength <= 74 { /*67~74*/
			options = /*ezsp.EMBER_APS_OPTION_RETRY | */ ezsp.EMBER_APS_OPTION_SOURCE_EUI64
		} else {
			options = ezsp.EMBER_APS_OPTION_NONE
		}
	}
	return
}

func RemoveDevice(eui64 uint64) (err error) {
	common.Log.Debugf("RemoveDevice %016x", eui64)
	nodeID := findNodeIDbyEui64(eui64)
	if nodeID == ezsp.EMBER_NULL_NODE_ID {
		return fmt.Errorf("unknow EUI64 %016x", eui64)
	}
	err = ezsp.EzspRemoveDevice(nodeID, eui64, eui64)
	if err != nil {
		common.Log.Errorf("EzspRemoveDevice failed: %v", err)
	}
	return
}

func RemoveNetwork() (err error) {
	common.Log.Debugf("RemoveNetwork()")
	if !ezsp.MeshStatusUp {
		return ErrMeshNotExist
	}
	notEmpty := false
	Nodes.Range(func(key, value interface{}) bool {
		notEmpty = true
		return false
	})
	if notEmpty {
		return ErrMeshNotEmpty
	}
	err = ezsp.EzspLeaveNetwork()
	return
}

func SetRadioChannel(channel byte) (err error) {
	common.Log.Debugf("SetRadioChannel(%d)", channel)
	return ezsp.EzspSetRadioChannel(channel)
}

func FormNetwork(radioChannel byte) (err error) {
	common.Log.Debugf("FormNetwork(%d)", radioChannel)
	if ezsp.MeshStatusUp {
		return ErrMeshAlreadyExist
	} else {
		return ezsp.NcpFormNetwork(radioChannel)
	}
}
