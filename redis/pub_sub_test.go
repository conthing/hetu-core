package redis

import (
	"encoding/json"
	"hetu-core/dto"
	"testing"
	"time"

	"github.com/conthing/utils/common"
)

func Test_Publish(t *testing.T) {
	Connect()
	fn := func(data []byte) {
		common.Log.Info(string(data))
	}
	Subscribe(fn)
	time.Sleep(time.Second * 2)
	rm := dto.ReceiveMessageDTO{}
	data, _ := json.Marshal(rm)
	Publish("hetu-core", data)
	time.Sleep(time.Minute * 1)
}

func Test_Subscribe(t *testing.T) {
	Connect()
	fn := func(data []byte) {
		common.Log.Info(data)
	}
	Subscribe(fn)
	time.Sleep(time.Minute * 1)
}
