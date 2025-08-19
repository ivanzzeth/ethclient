package subscriber

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ivanzzeth/ethclient/types"
)

// Subscriber represents a set of methods about chain subscription
type Subscriber interface {
	Close()
	GetQueryHandler() QueryHandler
	GetBlockConfirmationsOnSubscription() uint64

	SetBuffer(buffer int)
	SetBlockConfirmationsOnSubscription(confirmations uint64)
	SetQueryHandler(handler QueryHandler) // use QueryHandler instead of SubscriberStorage if handler set
	SetFetchMissingHeaders(enable bool)

	// Provided for handler submitting query.
	SubmitQuery(query ethereum.FilterQuery) error
	SubscribeNewHead(ctx context.Context, ch chan<- *etypes.Header) (ethereum.Subscription, error)
	SubscribeFilterFullTransactions(ctx context.Context, filter FilterTransaction, ch chan<- *etypes.Transaction) (ethereum.Subscription, error)
	// SubscribeFullPendingTransactions subscribes to new pending transactions.
	SubscribeFullPendingTransactions(ctx context.Context, ch chan<- *etypes.Transaction) (*rpc.ClientSubscription, error)
	// SubscribePendingTransactions subscribes to new pending transaction hashes.
	SubscribePendingTransactions(ctx context.Context, ch chan<- common.Hash) (*rpc.ClientSubscription, error)
	SubscribeFilterFullPendingTransactions(ctx context.Context, filter FilterTransaction, ch chan<- *etypes.Transaction) (*rpc.ClientSubscription, error)
	SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- etypes.Log) (ethereum.Subscription, error)
	FilterLogs(ctx context.Context, q ethereum.FilterQuery) (logs []etypes.Log, err error)
}

type FilterTransaction struct {
	FromBlock *big.Int // beginning of the queried range, nil means genesis block. only used for historical data
	ToBlock   *big.Int // end of the range, nil means latest block. only used for historical data

	// Any of these conditions meet, the transaction is considered.
	From           []common.Address
	To             []common.Address
	MethodSelector []types.MethodSelector
}

// Used only for function `SubscribeFilterlogs` && query.ToBlock == nil
type SubscriberStorage interface {
	QueryStateReader
	QueryStateWriter
}

type QueryStateReader interface {
	LatestBlockForQuery(ctx context.Context, query ethereum.FilterQuery) (uint64, error)
	LatestLogForQuery(ctx context.Context, query ethereum.FilterQuery) (etypes.Log, error)

	// Save query result to save network io
	FilterLogs(ctx context.Context, q ethereum.FilterQuery) (logs []etypes.Log, err error)
	// Report whether client can use `FilterLogs` in the storage instead of ethclient.FilterLogs
	IsFilterLogsSupported(q ethereum.FilterQuery) bool

	FilterLogsBatch(ctx context.Context, queries []ethereum.FilterQuery) (logs [][]etypes.Log, err error)
}

type QueryStateWriter interface {
	// Must call the function after all logs was handled for the block.
	SaveLatestBlockForQuery(ctx context.Context, query ethereum.FilterQuery, blockNum uint64) error
	// Must call the function after each log was handled .
	SaveLatestLogForQuery(ctx context.Context, query ethereum.FilterQuery, log etypes.Log) error

	// Save query result to save network io
	SaveFilterLogs(q ethereum.FilterQuery, logs []etypes.Log) (err error)
}

// Used only for handler set && query.ToBlock == nil
type QueryHandler interface {
	// Please update query states by handler-self, otherwise
	// logs may be replayed
	SubscriberStorage
	// Subscriber will call back it for handling when incoming logs are ready.
	// If log.Address is address(0), just for updating block number
	HandleQuery(ctx context.Context, query Query, log etypes.Log) error
}

type resubscribeFunc func() (ethereum.Subscription, error)
type FilterLogsFunc func(ctx context.Context, q ethereum.FilterQuery) (logs []etypes.Log, err error)
