package nonce

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

type ethBackend interface {
	ethereum.PendingStateReader
	ethereum.ChainStateReader
	ethereum.GasPricer
	ethereum.GasEstimator
}

type Manager interface {
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	PeekNonce(account common.Address) (uint64, error)
	ResetNonce(ctx context.Context, account common.Address) error
	EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	SetNonceAt(nonceAt NonceAtFunc)
}

type NonceAtFunc func(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error)
