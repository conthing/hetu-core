package redis

import "testing"

func before() {
	Connect()
}

func Test_ReviewPendingKeys(t *testing.T) {
	Connect()
}
