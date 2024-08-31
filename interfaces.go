package ethclient

import (
	"context"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
)

// Subscriber represents a set of methods about chain subscription
type Subscriber interface {
	SubscribeFilterlogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) error
	SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) error
	FilterLogs(ctx context.Context, q ethereum.FilterQuery) (logs []types.Log, err error)
}

// TransactFunc represents the transact call of Smart Contract.
type TransactFunc func() (*types.Transaction, error)

// ExpectedEventsFunc returns true if event is expected.
type ExpectedEventsFunc func(event interface{}) bool

type resubscribeFunc func() (ethereum.Subscription, error)
