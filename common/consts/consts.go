package consts

import "time"

const (
	RetryInterval        = 2 * time.Second
	DefaultMsgBuffer     = 1000
	DefaultBlocksPerScan = 100
)
