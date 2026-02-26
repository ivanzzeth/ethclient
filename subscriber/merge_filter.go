// Package subscriber: merge multiple FilterQueries into one so a single eth_getLogs
// (one JSON-RPC request) can be used instead of N, minimizing RPC cost per cycle.
package subscriber

import (
	"errors"
	"math/big"
	"slices"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
)

// GetPartitionKey returns a stable key for grouping FilterQueries that can be merged into one eth_getLogs.
// Same key ⇒ same block selector (range or blockHash) ⇒ can merge. For range queries, if FromBlock/ToBlock
// are nil, fallbackFrom/fallbackTo are used when non-nil; otherwise "nil" is used so nil range does not
// merge with a concrete range.
func GetPartitionKey(q ethereum.FilterQuery, fallbackFrom, fallbackTo *big.Int) string {
	if q.BlockHash != nil {
		return "H:" + q.BlockHash.Hex()
	}
	from := q.FromBlock
	if from == nil && fallbackFrom != nil {
		from = fallbackFrom
	}
	to := q.ToBlock
	if to == nil && fallbackTo != nil {
		to = fallbackTo
	}
	return "R:" + blockNumString(from) + ":" + blockNumString(to)
}

func blockNumString(n *big.Int) string {
	if n == nil {
		return "nil"
	}
	return n.String()
}

// PartitionQueriesByMergeKey splits queries into groups that share the same partition key (same block range
// or same BlockHash). Each group can be merged into one eth_getLogs. fallbackFrom/fallbackTo are used for
// queries with nil FromBlock/ToBlock when provided.
func PartitionQueriesByMergeKey(queries []ethereum.FilterQuery, fallbackFrom, fallbackTo *big.Int) [][]ethereum.FilterQuery {
	if len(queries) == 0 {
		return nil
	}
	keyToIndex := make(map[string]int)
	var groups [][]ethereum.FilterQuery
	for _, q := range queries {
		key := GetPartitionKey(q, fallbackFrom, fallbackTo)
		if idx, ok := keyToIndex[key]; ok {
			groups[idx] = append(groups[idx], q)
		} else {
			keyToIndex[key] = len(groups)
			groups = append(groups, []ethereum.FilterQuery{q})
		}
	}
	return groups
}

// MergeFilterQueries merges multiple FilterQueries that share the same block range
// into one FilterQuery: union of Addresses and union of Topics per index.
// The result is suitable for a single eth_getLogs call; each returned log can then
// be matched to the original queries via LogMatchesQuery.
// fromBlock and toBlock are set on the merged query; BlockHash must be nil for all.
func MergeFilterQueries(queries []ethereum.FilterQuery, fromBlock, toBlock *big.Int) (ethereum.FilterQuery, error) {
	if len(queries) == 0 {
		return ethereum.FilterQuery{}, nil
	}
	if fromBlock == nil || toBlock == nil {
		return ethereum.FilterQuery{}, errMergeBlockRange
	}
	for _, q := range queries {
		if q.BlockHash != nil {
			return ethereum.FilterQuery{}, errMergeBlockHash
		}
	}

	// If any query matches any address, merged must not restrict address (avoid missing logs).
	anyAddress := false
	addrSet := make(map[common.Address]struct{})
	var maxTopicLen int
	for _, q := range queries {
		if len(q.Topics) > maxTopicLen {
			maxTopicLen = len(q.Topics)
		}
		if q.Addresses == nil || len(q.Addresses) == 0 {
			anyAddress = true
			continue
		}
		for _, a := range q.Addresses {
			addrSet[a] = struct{}{}
		}
	}
	var addresses []common.Address
	if !anyAddress && len(addrSet) > 0 {
		addresses = make([]common.Address, 0, len(addrSet))
		for a := range addrSet {
			addresses = append(addresses, a)
		}
		slices.SortFunc(addresses, func(a, b common.Address) int { return slices.Compare(a.Bytes(), b.Bytes()) })
	}
	// else: addresses stays nil → match any address (correctness)

	topics := make([][]common.Hash, maxTopicLen)
	for i := 0; i < maxTopicLen; i++ {
		anyNil := false
		hashSet := make(map[common.Hash]struct{})
		for _, q := range queries {
			if i >= len(q.Topics) || q.Topics[i] == nil {
				anyNil = true
				break
			}
			for _, h := range q.Topics[i] {
				hashSet[h] = struct{}{}
			}
		}
		if anyNil {
			topics[i] = nil
			continue
		}
		// Empty union (e.g. all queries had Topics[i]==[]): use nil so we don't filter out logs (avoid missed dispatch).
		if len(hashSet) == 0 {
			topics[i] = nil
			continue
		}
		topics[i] = make([]common.Hash, 0, len(hashSet))
		for h := range hashSet {
			topics[i] = append(topics[i], h)
		}
		slices.SortFunc(topics[i], func(a, b common.Hash) int { return slices.Compare(a.Bytes(), b.Bytes()) })
	}

	return ethereum.FilterQuery{
		FromBlock:  fromBlock,
		ToBlock:    toBlock,
		Addresses:  addresses,
		Topics:     topics,
		BlockHash:  nil,
	}, nil
}

// MergeFilterQueriesByBlockHash merges multiple FilterQueries that share the same BlockHash
// into one FilterQuery (union of Addresses and Topics). Used for greedy minimization:
// all queries with the same BlockHash become one eth_getLogs.
func MergeFilterQueriesByBlockHash(queries []ethereum.FilterQuery, blockHash common.Hash) (ethereum.FilterQuery, error) {
	if len(queries) == 0 {
		return ethereum.FilterQuery{}, nil
	}
	for _, q := range queries {
		if q.BlockHash == nil || *q.BlockHash != blockHash {
			return ethereum.FilterQuery{}, errMergeBlockHashMismatch
		}
	}
	anyAddress := false
	addrSet := make(map[common.Address]struct{})
	var maxTopicLen int
	for _, q := range queries {
		if len(q.Topics) > maxTopicLen {
			maxTopicLen = len(q.Topics)
		}
		if q.Addresses == nil || len(q.Addresses) == 0 {
			anyAddress = true
			continue
		}
		for _, a := range q.Addresses {
			addrSet[a] = struct{}{}
		}
	}
	var addresses []common.Address
	if !anyAddress && len(addrSet) > 0 {
		addresses = make([]common.Address, 0, len(addrSet))
		for a := range addrSet {
			addresses = append(addresses, a)
		}
		slices.SortFunc(addresses, func(a, b common.Address) int { return slices.Compare(a.Bytes(), b.Bytes()) })
	}
	topics := make([][]common.Hash, maxTopicLen)
	for i := 0; i < maxTopicLen; i++ {
		anyNil := false
		hashSet := make(map[common.Hash]struct{})
		for _, q := range queries {
			if i >= len(q.Topics) || q.Topics[i] == nil {
				anyNil = true
				break
			}
			for _, h := range q.Topics[i] {
				hashSet[h] = struct{}{}
			}
		}
		if anyNil {
			topics[i] = nil
			continue
		}
		if len(hashSet) == 0 {
			topics[i] = nil
			continue
		}
		topics[i] = make([]common.Hash, 0, len(hashSet))
		for h := range hashSet {
			topics[i] = append(topics[i], h)
		}
		slices.SortFunc(topics[i], func(a, b common.Hash) int { return slices.Compare(a.Bytes(), b.Bytes()) })
	}
	hb := blockHash
	return ethereum.FilterQuery{
		BlockHash: &hb,
		Addresses: addresses,
		Topics:    topics,
	}, nil
}

var (
	errMergeBlockHash        = errors.New("merge filter: BlockHash not allowed")
	errMergeBlockRange       = errors.New("merge filter: fromBlock and toBlock required")
	errMergeBlockHashMismatch = errors.New("merge filter: all queries must have same BlockHash")
)

// LogMatchesQuery reports whether a log matches the given FilterQuery (address and topics).
func LogMatchesQuery(log *etypes.Log, q ethereum.FilterQuery) bool {
	if len(q.Addresses) > 0 {
		if !slices.Contains(q.Addresses, log.Address) {
			return false
		}
	}
	for i, filterTopics := range q.Topics {
		if filterTopics == nil || len(filterTopics) == 0 {
			continue
		}
		if i >= len(log.Topics) {
			return false
		}
		if !slices.Contains(filterTopics, log.Topics[i]) {
			return false
		}
	}
	return true
}
