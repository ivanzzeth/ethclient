package subscriber

import (
	"encoding/json"
	"fmt"
	"math/big"
	"slices"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type Query struct {
	ChainID *big.Int
	ethereum.FilterQuery
}

func NewQuery(chainId *big.Int, q ethereum.FilterQuery) Query {
	return Query{
		ChainID:     big.NewInt(0).Set(chainId),
		FilterQuery: q,
	}
}

func (q Query) Hash() common.Hash {
	return GetQueryHash(q.ChainID, q.FilterQuery)
}

func GetQueryKey(chainId *big.Int, query ethereum.FilterQuery) string {
	hash := GetQueryHash(chainId, query)

	return hash.Hex()
}

func GetQueryHash(chainId *big.Int, query ethereum.FilterQuery) common.Hash {
	type js struct {
		ChainId string
		ethereum.FilterQuery
	}

	// Ordering addresses for consistent hash
	slices.SortFunc(query.Addresses, func(a, b common.Address) int {
		if a.Hex() < b.Hex() {
			return -1
		}
		if a.Hex() > b.Hex() {
			return 1
		}
		return 0
	})

	var obj js = js{
		ChainId:     chainId.String(),
		FilterQuery: query,
	}
	json, _ := json.Marshal(obj)
	hash := crypto.Keccak256Hash(json)
	return hash
}

type QueryWithChannel struct {
	Query
	out chan etypes.Log // channel to send logs to
}

func distributeLogs(allLogs []etypes.Log, queries []ethereum.FilterQuery) [][]etypes.Log {
	logs := make([][]etypes.Log, len(queries))
	for _, l := range allLogs {
		for qi, q := range queries {
			// Address condition
			addressCondition := true
			if len(q.Addresses) > 0 {
				addressCondition = false
				for _, addr := range q.Addresses {
					if addr == l.Address {
						addressCondition = true
						break
					}
				}
			}

			// Block condition
			blockHashCondition := q.BlockHash != nil && l.BlockHash.Cmp(*q.BlockHash) == 0
			fromBlockCondition := q.FromBlock == nil || l.BlockNumber >= q.FromBlock.Uint64()
			toBlockCondition := q.ToBlock == nil || l.BlockNumber <= q.ToBlock.Uint64()
			blockCondition := true
			if q.BlockHash != nil {
				blockCondition = blockHashCondition
			} else {
				blockCondition = fromBlockCondition && toBlockCondition
			}

			getTopicCondition := func(logTopic common.Hash, queryTopic []common.Hash) bool {
				if len(queryTopic) == 0 {
					return true
				}

				for _, t := range queryTopic {
					if t.Cmp(logTopic) == 0 {
						return true
					}
				}

				return false
			}

			topicCondition := true
			if len(l.Topics) > 0 {
				if len(q.Topics) > 0 {
					topic0Condition := getTopicCondition(l.Topics[0], q.Topics[0])
					topicCondition = topicCondition && topic0Condition
				}
			} else if len(q.Topics) > 0 && len(q.Topics[0]) > 0 {
				topicCondition = false
			}

			if len(l.Topics) > 1 {
				if len(q.Topics) > 1 {
					topic1Condition := getTopicCondition(l.Topics[1], q.Topics[1])
					topicCondition = topicCondition && topic1Condition
				}
			} else if len(q.Topics) > 1 && len(q.Topics[1]) > 0 {
				topicCondition = false
			}

			if len(l.Topics) > 2 {
				if len(q.Topics) > 2 {
					topic2Condition := getTopicCondition(l.Topics[2], q.Topics[2])
					topicCondition = topicCondition && topic2Condition
				}
			} else if len(q.Topics) > 2 && len(q.Topics[2]) > 0 {
				topicCondition = false
			}

			if len(l.Topics) > 3 {
				if len(q.Topics) > 3 {
					topic3Condition := getTopicCondition(l.Topics[3], q.Topics[3])
					topicCondition = topicCondition && topic3Condition
				}
			} else if len(q.Topics) > 3 && len(q.Topics[3]) > 0 {
				topicCondition = false
			}

			if addressCondition && blockCondition && topicCondition {
				logs[qi] = append(logs[qi], l)
			}
		}
	}

	return logs
}

func splitFilterQuery(queryIncoming ethereum.FilterQuery, maxAddressesPerQuery int) (queries []ethereum.FilterQuery, err error) {
	if maxAddressesPerQuery <= 0 {
		return nil, fmt.Errorf("maxAddressesPerQuery must be positive")
	}

	if len(queryIncoming.Addresses) <= maxAddressesPerQuery {
		return []ethereum.FilterQuery{queryIncoming}, nil
	}

	addressChunks := slices.Chunk(queryIncoming.Addresses, maxAddressesPerQuery)

	for chunk := range addressChunks {
		newQuery := queryIncoming
		newQuery.Addresses = chunk
		queries = append(queries, newQuery)
	}

	return queries, nil
}

func mergeFilterQueriesWithMaxAddressesPerQuery(queries []ethereum.FilterQuery, maxAddressesPerQuery int, mergeOverlappingBlockRange bool) (queriesOut []ethereum.FilterQuery, err error) {
	var tmpQueries []ethereum.FilterQuery
	tmpQueries, err = mergeFilterQueries(queries, mergeOverlappingBlockRange)
	if err != nil {
		return nil, err
	}

	for _, q := range tmpQueries {
		if len(q.Addresses) <= maxAddressesPerQuery {
			queriesOut = append(queriesOut, q)
			continue
		}

		queriesAfterSplit, err := splitFilterQuery(q, maxAddressesPerQuery)
		if err != nil {
			return nil, err
		}

		queriesOut = append(queriesOut, queriesAfterSplit...)
	}

	return queriesOut, nil
}

// Pacts multiple queries into one as many as possible
func mergeFilterQueries(queries []ethereum.FilterQuery, mergeOverlappingBlockRange bool) ([]ethereum.FilterQuery, error) {
	// Phase 1: Group by BlockHash strictly (highest priority)
	blockHashGroups := make(map[common.Hash]*mergedQuery)

	// Phase 2: Group by block range
	blockRangeGroups := make(map[blockRange]*mergedQuery)

	for _, q := range queries {
		// Handle BlockHash queries
		if q.BlockHash != nil {
			hash := *q.BlockHash
			if _, exists := blockHashGroups[hash]; !exists {
				group := &mergedQuery{
					Addresses: make(map[common.Address]struct{}),
				}
				// Initialize Topics array
				if len(q.Topics) > 0 {
					group.Topics = make([]map[common.Hash]struct{}, len(q.Topics))
					for i := range group.Topics {
						group.Topics[i] = make(map[common.Hash]struct{})
					}
				}
				blockHashGroups[hash] = group
			}
			mergeIntoGroup(blockHashGroups[hash], q)
			continue
		}

		// Handle block range queries
		key := blockRange{from: q.FromBlock, to: q.ToBlock}
		if existing, found := findMergeableRange(blockRangeGroups, key, mergeOverlappingBlockRange); found {
			mergeIntoGroup(existing, q) // Merge into existing group
		} else {
			group := &mergedQuery{
				FromBlock: q.FromBlock,
				ToBlock:   q.ToBlock,
				Addresses: make(map[common.Address]struct{}),
			}
			// Initialize Topics array
			if len(q.Topics) > 0 {
				group.Topics = make([]map[common.Hash]struct{}, len(q.Topics))
				for i := range group.Topics {
					group.Topics[i] = make(map[common.Hash]struct{})
				}
			}
			blockRangeGroups[key] = group
			mergeIntoGroup(group, q)
		}
	}

	// Generate final query list
	return flattenGroups(blockHashGroups, blockRangeGroups), nil
}

// Core data structures
type blockRange struct{ from, to *big.Int }
type mergedQuery struct {
	FromBlock *big.Int
	ToBlock   *big.Int
	Addresses map[common.Address]struct{}
	Topics    []map[common.Hash]struct{} // Topic set for each position
}

// Merge query into group (handle addresses and topics)
func mergeIntoGroup(group *mergedQuery, q ethereum.FilterQuery) {
	// Merge addresses
	for _, addr := range q.Addresses {
		group.Addresses[addr] = struct{}{}
	}

	// Merge Topics (hierarchical inclusion logic)
	for i, topics := range q.Topics {
		if i >= len(group.Topics) {
			continue // Prevent out of bounds
		}

		if topics == nil {
			// nil means wildcard, keep nil after merging
			group.Topics[i] = nil
		} else {
			// For non-nil, record all topics that have appeared
			if group.Topics[i] == nil {
				// If current is nil (wildcard), no need to process (wildcard already includes all)
				continue
			}
			for _, topic := range topics {
				group.Topics[i][topic] = struct{}{}
			}
		}
	}

	// Expand block range
	if q.FromBlock != nil {
		if group.FromBlock == nil || q.FromBlock.Cmp(group.FromBlock) < 0 {
			group.FromBlock = q.FromBlock
		}
	}
	if q.ToBlock != nil {
		if group.ToBlock == nil || q.ToBlock.Cmp(group.ToBlock) > 0 {
			group.ToBlock = q.ToBlock
		}
	}
}

// Find mergeable block range group (allows range expansion)
func findMergeableRange(groups map[blockRange]*mergedQuery, key blockRange, mergeOverlappingBlockRange bool) (*mergedQuery, bool) {
	for br, group := range groups {
		if canMergeRanges(br, key, mergeOverlappingBlockRange) {
			return group, true
		}
	}
	return nil, false
}

// Determine if two block ranges can be merged
func canMergeRanges(a, b blockRange, mergeOverlappingBlockRange bool) bool {
	// Handle nil value cases
	if a.from == nil || b.from == nil || a.to == nil || b.to == nil {
		return true // If any end is open, allow merging
	}

	if mergeOverlappingBlockRange {
		// Allow merging overlapping or adjacent block ranges
		// Check for overlap or adjacency: a.from <= b.to + 1 && a.to + 1 >= b.from
		return a.from.Cmp(big.NewInt(0).Add(b.to, big.NewInt(1))) <= 0 &&
			big.NewInt(0).Add(a.to, big.NewInt(1)).Cmp(b.from) >= 0
	} else {
		// Only allow merging completely identical block ranges
		return a.from.Cmp(b.from) == 0 && a.to.Cmp(b.to) == 0
	}
}

// Convert grouped structure to final query list
func flattenGroups(hashGroups map[common.Hash]*mergedQuery, rangeGroups map[blockRange]*mergedQuery) []ethereum.FilterQuery {
	var result []ethereum.FilterQuery

	// Process BlockHash groups
	for hash, group := range hashGroups {
		// Fix: explicitly convert maps.Keys result to slice
		addresses := make([]common.Address, 0, len(group.Addresses))
		for addr := range group.Addresses {
			addresses = append(addresses, addr)
		}

		query := ethereum.FilterQuery{
			BlockHash: &hash,
			Addresses: addresses, // Use converted slice
		}
		if len(group.Topics) > 0 {
			query.Topics = convertTopicMaps(group.Topics)
		}
		result = append(result, query)
	}

	// Process block range groups
	for _, group := range rangeGroups {
		// Similarly fix address conversion
		addresses := make([]common.Address, 0, len(group.Addresses))
		for addr := range group.Addresses {
			addresses = append(addresses, addr)
		}

		query := ethereum.FilterQuery{
			FromBlock: group.FromBlock,
			ToBlock:   group.ToBlock,
			Addresses: addresses,
		}
		if len(group.Topics) > 0 {
			query.Topics = convertTopicMaps(group.Topics)
		}
		result = append(result, query)
	}

	return result
}

// Convert Topics from map form back to original array format
func convertTopicMaps(topicMaps []map[common.Hash]struct{}) [][]common.Hash {
	var topics [][]common.Hash
	for _, m := range topicMaps {
		if m == nil {
			topics = append(topics, nil) // Preserve nil wildcard
		} else {
			keys := make([]common.Hash, 0, len(m))
			for k := range m {
				keys = append(keys, k)
			}
			topics = append(topics, keys)
		}
	}
	return topics
}
