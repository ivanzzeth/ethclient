package ethclient

import (
	"context"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type NonceManager interface {
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
}

type SimpleNonceManager struct {
	nonceMap map[common.Address]uint64
	lockMap  sync.Map
	// lock     sync.Mutex
	client *ethclient.Client
}

var snm *SimpleNonceManager
var snmOnce sync.Once

func NewSimpleNonceManager(client *ethclient.Client) (*SimpleNonceManager, error) {
	snmOnce.Do(func() {
		snm = &SimpleNonceManager{
			nonceMap: make(map[common.Address]uint64),
			lockMap:  sync.Map{},
			client:   client,
		}
	})

	return snm, nil
}

func (nm *SimpleNonceManager) NonceLockFrom(from common.Address) *sync.Mutex {
	lock, _ := nm.lockMap.LoadOrStore(from, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

func (nm *SimpleNonceManager) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	// nm.lock.Lock()
	// defer nm.lock.Unlock()

	locker := nm.NonceLockFrom(account)
	locker.Lock()
	defer locker.Unlock()

	var (
		nonce uint64
		err   error
	)

	nonce, ok := nm.nonceMap[account]
	if !ok {
		nonce, err = nm.client.NonceAt(ctx, account, nil)
		if err != nil {
			return 0, err
		}
	}

	nm.nonceMap[account] = nonce + 1

	return nonce, nil
}
