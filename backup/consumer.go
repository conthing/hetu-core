package backup

// ConsumeBackupQueue 消费备份队列
func ConsumeBackupQueue() {
	next := make(chan bool, 0)
	fail := make(chan bool, 0)
	done := make(chan bool, 0)
	go working(next, done, fail)
	go superviseWorker(next, done, fail)

}
