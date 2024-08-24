package nonce

import (
	"context"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

var _ Manager = &SimpleManager{}

type SimpleManager struct {
	Storage
	client *ethclient.Client
}

var snm *SimpleManager
var snmOnce sync.Once

func GetSimpleNonceManager(client *ethclient.Client, storage Storage) (*SimpleManager, error) {
	snmOnce.Do(func() {
		snm = &SimpleManager{
			Storage: storage,
			client:  client,
		}
	})

	return snm, nil
}

func NewSimpleNonceManager(client *ethclient.Client, storage Storage) (*SimpleManager, error) {
	return &SimpleManager{
		Storage: storage,
		client:  client,
	}, nil
}

func (nm *SimpleManager) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	locker := nm.NonceLockFrom(account)
	locker.Lock()
	defer locker.Unlock()

	var (
		nonce uint64
		err   error
	)

	nonce, err = nm.GetNonce(account)
	if err != nil {
		return 0, err
	}

	nonceInLatest, err := nm.client.PendingNonceAt(ctx, account)
	if err != nil {
		return 0, err
	}

	if nonce == 0 || nonceInLatest > nonce {
		nonce, err = nm.client.PendingNonceAt(ctx, account)
		if err != nil {
			return 0, err
		}
	}

	err = nm.SetNonce(account, nonce+1)
	if err != nil {
		return 0, err
	}

	return nonce, nil
}

func (nm *SimpleManager) SuggestGasPrice(ctx context.Context) (gasPrice *big.Int, err error) {
	gasPrice, err = nm.client.SuggestGasPrice(ctx)
	if err != nil {
		return
	}

	// Multiplier 1.5
	gasPrice.Mul(gasPrice, big.NewInt(1500))
	gasPrice.Div(gasPrice, big.NewInt(1000))

	return
}

func (nm *SimpleManager) PeekNonce(account common.Address) (uint64, error) {
	locker := nm.NonceLockFrom(account)
	locker.Lock()
	defer locker.Unlock()

	nonce, err := nm.GetNonce(account)
	if err != nil {
		return 0, err
	}

	return nonce, nil
}

func (nm *SimpleManager) ResetNonce(ctx context.Context, account common.Address) error {
	locker := nm.NonceLockFrom(account)
	locker.Lock()
	defer locker.Unlock()

	nonceInLatest, err := nm.client.NonceAt(ctx, account, nil)
	if err != nil {
		return err
	}

	err = nm.SetNonce(account, nonceInLatest)
	if err != nil {
		return err
	}

	return nil
}
