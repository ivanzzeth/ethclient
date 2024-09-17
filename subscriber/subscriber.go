package subscriber

import (
	"context"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
)

// Subscriber represents a set of methods about chain subscription
type Subscriber interface {
	GetBlockConfirmationsOnSubscription() uint64
	SetBlockConfirmationsOnSubscription(confirmations uint64)
	SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error)
	SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error)
	FilterLogs(ctx context.Context, q ethereum.FilterQuery) (logs []types.Log, err error)
}

// Used only for function `SubscribeFilterlogs` and query.ToBlock == nil
type SubscriberStorage interface {
	LatestBlockForQuery(ctx context.Context, query ethereum.FilterQuery) (uint64, error)
	LatestLogForQuery(ctx context.Context, query ethereum.FilterQuery) (types.Log, error)

	SaveLatestBlockForQuery(ctx context.Context, query ethereum.FilterQuery, blockNum uint64) error
	SaveLatestLogForQuery(ctx context.Context, query ethereum.FilterQuery, log types.Log) error
}

type resubscribeFunc func() (ethereum.Subscription, error)
