package subscriber

import (
	"context"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
)

// Subscriber represents a set of methods about chain subscription
type Subscriber interface {
	Close()
	// Provided for handler submitting query.
	SubmitQuery(query ethereum.FilterQuery) error
	SetQueryHandler(handler QueryHandler) // use QueryHandler instead of SubscriberStorage if handler set
	GetQueryHandler() QueryHandler
	GetBlockConfirmationsOnSubscription() uint64
	SetBlockConfirmationsOnSubscription(confirmations uint64)
	SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error)
	SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error)
	FilterLogs(ctx context.Context, q ethereum.FilterQuery) (logs []types.Log, err error)
}

// Used only for function `SubscribeFilterlogs` && query.ToBlock == nil
type SubscriberStorage interface {
	QueryStateReader
	QueryStateWriter
}

type QueryStateReader interface {
	LatestBlockForQuery(ctx context.Context, query ethereum.FilterQuery) (uint64, error)
	LatestLogForQuery(ctx context.Context, query ethereum.FilterQuery) (types.Log, error)
}

type QueryStateWriter interface {
	// Must call the function after all logs was handled for the block.
	SaveLatestBlockForQuery(ctx context.Context, query ethereum.FilterQuery, blockNum uint64) error
	// Must call the function after each log was handled .
	SaveLatestLogForQuery(ctx context.Context, query ethereum.FilterQuery, log types.Log) error
}

// Used only for handler set && query.ToBlock == nil
type QueryHandler interface {
	// Please update query states by handler-self, otherwise
	// logs may be replayed
	QueryStateReader
	// Subscriber will call back it for handling when incoming logs are ready.
	HandleQuery(ctx context.Context, query Query, log types.Log) error
}

type resubscribeFunc func() (ethereum.Subscription, error)
