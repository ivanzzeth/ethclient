package nonce

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

type ethBackend interface {
	ethereum.PendingStateReader
	ethereum.GasPricer
}

type Manager interface {
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	PeekNonce(account common.Address) (uint64, error)
	ResetNonce(ctx context.Context, account common.Address) error
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
}
