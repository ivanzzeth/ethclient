package nonce

import (
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

type Storage interface {
	NonceLockFrom(from common.Address) sync.Locker
	// without locks
	GetNonce(account common.Address) (uint64, error)
	// without locks
	SetNonce(account common.Address, nonce uint64) error
}
