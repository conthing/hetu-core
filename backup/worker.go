package backup

import (
	"hetu-core/proxy"
	"hetu-core/redis"

	"github.com/conthing/utils/common"
)

func working(next chan bool, done chan bool, fail chan bool) {

	for {
		select {
		case <-next:
			consumeMessage(done, fail)
		}
	}
}

// consumeMessage 消费尾部消息
func consumeMessage(done chan bool, fail chan bool) {
	res, err := redis.ReadBackupQueue()
	if err != nil {
		common.Log.Error("read queue tail failed: ", err)
		fail <- true
		return
	}

	err = proxy.Post([]byte(res[1])) // TODO 为什么下标是1？
	if err != nil {
		fail <- true
		return
	}

	common.Log.Info("[OK] consume message success", res[1])
	done <- true
}
