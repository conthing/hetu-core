package backup

import (
	"time"
)

// superviseWorkern 监督工人
func superviseWorker(next chan bool, done chan bool, fail chan bool) {
	// 第一次先让工人干活
	next <- true
	for {
		select {
		case <-done:
			// 等待 worker 完成一次上报
			// 让 worker 进行下一个工作
			next <- true
		case <-fail:
			// 等待 5 秒再进行下一个工作
			//common.Log.Error("等待 5 秒再进行下一个工作")
			time.Sleep(time.Second * 5)
			next <- true
		}
	}

}
