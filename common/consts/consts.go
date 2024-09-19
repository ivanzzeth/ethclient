package consts

import "time"

const (
	RetryInterval        = 3 * time.Second
	DefaultMsgBuffer     = 1000
	DefaultBlocksPerScan = 100
)
