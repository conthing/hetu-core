package redis

import (
	"github.com/conthing/utils/common"
	"github.com/mediocregopher/radix/v3"
)

const backupQueue = "backup_queue"

// AddToBackupQueue 加入备份队列头部
func AddToBackupQueue(data []byte) {
	err := Client.Do(radix.Cmd(nil, "LPUSH", backupQueue, string(data)))
	if err != nil {
		common.Log.Error("加入队头失败", err)
	}
	common.Log.Info("[OK] AddToBackupQueue")

}

// TrimBackupQueue 清理备份队列
// 容量为64
func TrimBackupQueue() {

	err := Client.Do(radix.Cmd(nil, "LTRIM", backupQueue, "0", "63"))
	if err != nil {
		common.Log.Error("TrimBackupQueue LTRIM error: ", err)
		return
	}
	common.Log.Info("[success] TrimBackupQueue")
}

// ReadBackupQueue 读取备份队列
func ReadBackupQueue() ([]string, error) {
	var res []string
	err := Client.Do(radix.Cmd(&res, "BRPOP", backupQueue, "0"))
	if err != nil {
		return res, err
	}
	return res, nil
}

// RPushBackupQueue 尾部消息塞回队尾
func RPushBackupQueue(data string) {
	err := Client.Do(radix.Cmd(nil, "RPUSH", backupQueue, data))
	if err != nil {
		common.Log.Error("RPush backup_queue error", err)
	}
}
