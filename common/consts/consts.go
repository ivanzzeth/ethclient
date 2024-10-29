package consts

import "time"

const (
	RetryInterval        = 3 * time.Second
	DefaultMsgBuffer     = 1000
	DefaultBlocksPerScan = uint64(100)
	MaxBlocksPerScan     = uint64(10000000)
)
