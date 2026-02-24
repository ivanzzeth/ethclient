// Package subscriber: batch log filterer to merge multiple eth_getLogs into one JSON-RPC batch request.
package subscriber

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common/hexutil"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
)

// LogFiltererBatch extends log filtering with a batch call to reduce RPC round-trips.
// When multiple subscriptions share the same chain and block range, one FilterLogsBatch
// can replace N separate FilterLogs calls (e.g. one HTTP request via JSON-RPC batch).
type LogFiltererBatch interface {
	FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]etypes.Log, error)
	FilterLogsBatch(ctx context.Context, queries []ethereum.FilterQuery) ([][]etypes.Log, error)
}

// batchLogFilterer wraps go-ethereum ethclient and adds FilterLogsBatch using rpc.BatchCallContext.
type batchLogFilterer struct {
	*ethclient.Client
	rpc *rpc.Client
}

// NewBatchLogFilterer returns a LogFiltererBatch that uses BatchCallContext for FilterLogsBatch.
func NewBatchLogFilterer(c *ethclient.Client, rpcCli *rpc.Client) LogFiltererBatch {
	if c == nil || rpcCli == nil {
		panic("batchLogFilterer: client and rpc must be non-nil")
	}
	return &batchLogFilterer{Client: c, rpc: rpcCli}
}

// FilterLogs delegates to the embedded client.
func (b *batchLogFilterer) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]etypes.Log, error) {
	return b.Client.FilterLogs(ctx, q)
}

// FilterLogsBatch sends multiple eth_getLogs in one JSON-RPC batch and returns results in order.
func (b *batchLogFilterer) FilterLogsBatch(ctx context.Context, queries []ethereum.FilterQuery) ([][]etypes.Log, error) {
	if len(queries) == 0 {
		return nil, nil
	}
	results := make([][]etypes.Log, len(queries))
	elems := make([]rpc.BatchElem, len(queries))
	for i, q := range queries {
		arg, err := toFilterArg(q)
		if err != nil {
			return nil, fmt.Errorf("query %d: %w", i, err)
		}
		elems[i] = rpc.BatchElem{
			Method: "eth_getLogs",
			Args:   []interface{}{arg},
			Result: &results[i],
		}
	}
	if err := b.rpc.BatchCallContext(ctx, elems); err != nil {
		return nil, err
	}
	for i, e := range elems {
		if e.Error != nil {
			return nil, fmt.Errorf("query %d: %w", i, e.Error)
		}
	}
	log.Debug("FilterLogsBatch completed", "queries", len(queries))
	return results, nil
}

// toFilterArg builds the argument map for eth_getLogs (same shape as go-ethereum ethclient).
func toFilterArg(q ethereum.FilterQuery) (interface{}, error) {
	arg := map[string]interface{}{}
	if q.Addresses != nil {
		arg["address"] = q.Addresses
	}
	if q.Topics != nil {
		arg["topics"] = q.Topics
	}
	if q.BlockHash != nil {
		arg["blockHash"] = *q.BlockHash
		if q.FromBlock != nil || q.ToBlock != nil {
			return nil, errors.New("cannot specify both BlockHash and FromBlock/ToBlock")
		}
	} else {
		if q.FromBlock == nil {
			arg["fromBlock"] = "0x0"
		} else {
			arg["fromBlock"] = toBlockNumArg(q.FromBlock)
		}
		arg["toBlock"] = toBlockNumArg(q.ToBlock)
	}
	return arg, nil
}

func toBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}
	if number.Sign() >= 0 {
		return hexutil.EncodeBig(number)
	}
	if number.IsInt64() {
		return rpc.BlockNumber(number.Int64()).String()
	}
	return fmt.Sprintf("<invalid %d>", number)
}
