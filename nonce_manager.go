package ethclient

import (
	"context"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type NonceManager interface {
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	PeekNonce(account common.Address) uint64
	ResetNonce(ctx context.Context, account common.Address) error
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
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

func (nm *SimpleNonceManager) SuggestGasPrice(ctx context.Context) (gasPrice *big.Int, err error) {
	gasPrice, err = nm.client.SuggestGasPrice(ctx)
	if err != nil {
		return
	}

	// Multiplier 1.5
	gasPrice.Mul(gasPrice, big.NewInt(1500))
	gasPrice.Div(gasPrice, big.NewInt(1000))

	return
}

func (nm *SimpleNonceManager) PeekNonce(account common.Address) uint64 {
	locker := nm.NonceLockFrom(account)
	locker.Lock()
	defer locker.Unlock()

	nonce := nm.nonceMap[account]
	return nonce
}

func (nm *SimpleNonceManager) ResetNonce(ctx context.Context, account common.Address) error {
	locker := nm.NonceLockFrom(account)
	locker.Lock()
	defer locker.Unlock()

	nonceInLatest, err := nm.client.NonceAt(ctx, account, nil)
	if err != nil {
		return err
	}

	nm.nonceMap[account] = nonceInLatest

	return nil
}
