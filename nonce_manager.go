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
	lock     sync.Mutex
	client   *ethclient.Client
}

func NewNonceManager(client *ethclient.Client) (*SimpleNonceManager, error) {
	return &SimpleNonceManager{
		nonceMap: make(map[common.Address]uint64),
		client:   client,
	}, nil
}

func (nm *SimpleNonceManager) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	nm.lock.Lock()
	defer nm.lock.Unlock()

	var (
		nonce uint64
		err   error
	)

	nonce, ok := nm.nonceMap[account]
	if !ok {
		nonce, err = nm.client.PendingNonceAt(ctx, account)
		if err != nil {
			return 0, err
		}
	}

	nm.nonceMap[account] = nonce + 1

	return nonce, nil
}
