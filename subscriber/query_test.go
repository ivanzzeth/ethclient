package subscriber

import (
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
)

func TestDistributeLogs(t *testing.T) {
	// Create test hashes and addresses
	blockHash1 := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	blockHash2 := common.HexToHash("0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321")
	topic1 := common.HexToHash("0xabc123def456abc123def456abc123def456abc123def456abc123def456abc1")
	topic2 := common.HexToHash("0xdef456abc123def456abc123def456abc123def456abc123def456abc123def4")
	topic3 := common.HexToHash("0x123abc456def123abc456def123abc456def123abc456def123abc456def123a")
	address1 := common.HexToAddress("0x742d35Cc6634C893292Ce8bB6239C002Ad8e6b59")
	address2 := common.HexToAddress("0x853d43Cc6634C893292Ce8bB6239C002Ad8e6b60")

	tests := []struct {
		name     string
		allLogs  []etypes.Log
		queries  []ethereum.FilterQuery
		expected [][]etypes.Log
	}{
		{
			name: "Basic block number range filtering",
			allLogs: []etypes.Log{
				{BlockNumber: 100, BlockHash: blockHash1, Address: address1},
				{BlockNumber: 150, BlockHash: blockHash1, Address: address1},
				{BlockNumber: 200, BlockHash: blockHash1, Address: address1},
			},
			queries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(120),
					ToBlock:   big.NewInt(180),
					Addresses: []common.Address{address1},
				},
			},
			expected: [][]etypes.Log{
				{{BlockNumber: 150, BlockHash: blockHash1, Address: address1}},
			},
		},
		{
			name: "Multiple queries with different ranges",
			allLogs: []etypes.Log{
				{BlockNumber: 100, BlockHash: blockHash1, Address: address1},
				{BlockNumber: 200, BlockHash: blockHash1, Address: address1},
				{BlockNumber: 300, BlockHash: blockHash1, Address: address1},
			},
			queries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(50),
					ToBlock:   big.NewInt(150),
					Addresses: []common.Address{address1},
				},
				{
					FromBlock: big.NewInt(250),
					ToBlock:   big.NewInt(350),
					Addresses: []common.Address{address1},
				},
			},
			expected: [][]etypes.Log{
				{{BlockNumber: 100, BlockHash: blockHash1, Address: address1}},
				{{BlockNumber: 300, BlockHash: blockHash1, Address: address1}},
			},
		},
		{
			name: "Block hash specific filtering",
			allLogs: []etypes.Log{
				{BlockNumber: 100, BlockHash: blockHash1, Address: address1},
				{BlockNumber: 100, BlockHash: blockHash2, Address: address1},
			},
			queries: []ethereum.FilterQuery{
				{
					BlockHash: &blockHash1,
					Addresses: []common.Address{address1},
				},
			},
			expected: [][]etypes.Log{
				{{BlockNumber: 100, BlockHash: blockHash1, Address: address1}},
			},
		},
		{
			name: "Topic filtering - exact match",
			allLogs: []etypes.Log{
				{
					BlockNumber: 100,
					BlockHash:   blockHash1,
					Address:     address1,
					Topics:      []common.Hash{topic1, topic2},
				},
				{
					BlockNumber: 100,
					BlockHash:   blockHash1,
					Address:     address1,
					Topics:      []common.Hash{topic1, topic3},
				},
			},
			queries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(90),
					ToBlock:   big.NewInt(110),
					Addresses: []common.Address{address1},
					Topics:    [][]common.Hash{{topic1}, {topic2}},
				},
			},
			expected: [][]etypes.Log{
				{{
					BlockNumber: 100,
					BlockHash:   blockHash1,
					Address:     address1,
					Topics:      []common.Hash{topic1, topic2},
				}},
			},
		},
		{
			name: "Topic filtering - wildcard support",
			allLogs: []etypes.Log{
				{
					BlockNumber: 100,
					BlockHash:   blockHash1,
					Address:     address1,
					Topics:      []common.Hash{topic1, topic2},
				},
				{
					BlockNumber: 100,
					BlockHash:   blockHash1,
					Address:     address1,
					Topics:      []common.Hash{topic1, topic3},
				},
			},
			queries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(90),
					ToBlock:   big.NewInt(110),
					Addresses: []common.Address{address1},
					Topics:    [][]common.Hash{{topic1}, nil}, // Second topic is wildcard
				},
			},
			expected: [][]etypes.Log{
				{
					{
						BlockNumber: 100,
						BlockHash:   blockHash1,
						Address:     address1,
						Topics:      []common.Hash{topic1, topic2},
					},
					{
						BlockNumber: 100,
						BlockHash:   blockHash1,
						Address:     address1,
						Topics:      []common.Hash{topic1, topic3},
					},
				},
			},
		},
		{
			name: "Multiple address filtering",
			allLogs: []etypes.Log{
				{BlockNumber: 100, BlockHash: blockHash1, Address: address1},
				{BlockNumber: 100, BlockHash: blockHash1, Address: address2},
			},
			queries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(90),
					ToBlock:   big.NewInt(110),
					Addresses: []common.Address{address1, address2},
				},
			},
			expected: [][]etypes.Log{
				{
					{BlockNumber: 100, BlockHash: blockHash1, Address: address1},
					{BlockNumber: 100, BlockHash: blockHash1, Address: address2},
				},
			},
		},
		{
			name: "Nil block range conditions",
			allLogs: []etypes.Log{
				{BlockNumber: 100, BlockHash: blockHash1, Address: address1},
				{BlockNumber: 200, BlockHash: blockHash1, Address: address1},
			},
			queries: []ethereum.FilterQuery{
				{
					FromBlock: nil, // No lower bound
					ToBlock:   big.NewInt(150),
					Addresses: []common.Address{address1},
				},
				{
					FromBlock: big.NewInt(150),
					ToBlock:   nil, // No upper bound
					Addresses: []common.Address{address1},
				},
			},
			expected: [][]etypes.Log{
				{{BlockNumber: 100, BlockHash: blockHash1, Address: address1}},
				{{BlockNumber: 200, BlockHash: blockHash1, Address: address1}},
			},
		},
		{
			name:    "Empty logs input",
			allLogs: []etypes.Log{},
			queries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(100),
					ToBlock:   big.NewInt(200),
					Addresses: []common.Address{address1},
				},
			},
			expected: [][]etypes.Log{{}},
		},
		{
			name: "No matching logs",
			allLogs: []etypes.Log{
				{BlockNumber: 100, BlockHash: blockHash1, Address: address1},
			},
			queries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(200),
					ToBlock:   big.NewInt(300),
					Addresses: []common.Address{address1},
				},
			},
			expected: [][]etypes.Log{{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := distributeLogs(tt.allLogs, tt.queries)

			// Compare the length first for better error messages
			assert.Equal(t, len(tt.expected), len(result), "Number of query results mismatch")

			for i := range result {
				assert.Equal(t, len(tt.expected[i]), len(result[i]),
					"Number of logs in query %d mismatch", i)

				for j := range result[i] {
					assert.Equal(t, tt.expected[i][j].BlockNumber, result[i][j].BlockNumber,
						"BlockNumber mismatch in query %d, log %d", i, j)
					assert.Equal(t, tt.expected[i][j].BlockHash, result[i][j].BlockHash,
						"BlockHash mismatch in query %d, log %d", i, j)
					assert.Equal(t, tt.expected[i][j].Address, result[i][j].Address,
						"Address mismatch in query %d, log %d", i, j)

					// Compare topics if they exist
					if len(tt.expected[i][j].Topics) > 0 || len(result[i][j].Topics) > 0 {
						assert.Equal(t, tt.expected[i][j].Topics, result[i][j].Topics,
							"Topics mismatch in query %d, log %d", i, j)
					}
				}
			}
		})
	}
}

// TestDistributeLogsEdgeCases tests edge cases and error conditions
func TestDistributeLogsEdgeCases(t *testing.T) {
	t.Run("Nil queries slice", func(t *testing.T) {
		logs := []etypes.Log{{BlockNumber: 100}}
		result := distributeLogs(logs, nil)
		assert.Empty(t, result)
	})

	t.Run("Empty queries slice", func(t *testing.T) {
		logs := []etypes.Log{{BlockNumber: 100}}
		result := distributeLogs(logs, []ethereum.FilterQuery{})
		assert.Empty(t, result)
	})

	t.Run("Nil block pointers in query", func(t *testing.T) {
		logs := []etypes.Log{{BlockNumber: 100, Address: common.HexToAddress("0x1")}}
		queries := []ethereum.FilterQuery{
			{
				FromBlock: nil,
				ToBlock:   nil,
				Addresses: []common.Address{common.HexToAddress("0x1")},
			},
		}
		result := distributeLogs(logs, queries)
		assert.Len(t, result, 1)
		assert.Len(t, result[0], 1)
	})
}

// TestDistributeLogsAddressFiltering tests address filtering scenarios
func TestDistributeLogsAddressFiltering(t *testing.T) {
	blockHash1 := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	address1 := common.HexToAddress("0x742d35Cc6634C893292Ce8bB6239C002Ad8e6b59")
	address2 := common.HexToAddress("0x853d43Cc6634C893292Ce8bB6239C002Ad8e6b60")
	address3 := common.HexToAddress("0x964d51Dd6634C893292Ce8bB6239C002Ad8e6b71")

	tests := []struct {
		name     string
		allLogs  []etypes.Log
		queries  []ethereum.FilterQuery
		expected [][]etypes.Log
	}{
		{
			name: "Log address not in query addresses",
			allLogs: []etypes.Log{
				{BlockNumber: 100, BlockHash: blockHash1, Address: address1},
				{BlockNumber: 100, BlockHash: blockHash1, Address: address2},
			},
			queries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(90),
					ToBlock:   big.NewInt(110),
					Addresses: []common.Address{address1, address3}, // address2 not included
				},
			},
			expected: [][]etypes.Log{
				{{BlockNumber: 100, BlockHash: blockHash1, Address: address1}},
			},
		},
		{
			name: "Empty addresses in query (should match all)",
			allLogs: []etypes.Log{
				{BlockNumber: 100, BlockHash: blockHash1, Address: address1},
				{BlockNumber: 100, BlockHash: blockHash1, Address: address2},
				{BlockNumber: 100, BlockHash: blockHash1, Address: address3},
			},
			queries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(90),
					ToBlock:   big.NewInt(110),
					Addresses: []common.Address{}, // Empty addresses should match all
				},
			},
			expected: [][]etypes.Log{
				{
					{BlockNumber: 100, BlockHash: blockHash1, Address: address1},
					{BlockNumber: 100, BlockHash: blockHash1, Address: address2},
					{BlockNumber: 100, BlockHash: blockHash1, Address: address3},
				},
			},
		},
		{
			name: "Nil addresses in query (should match all)",
			allLogs: []etypes.Log{
				{BlockNumber: 100, BlockHash: blockHash1, Address: address1},
				{BlockNumber: 100, BlockHash: blockHash1, Address: address2},
			},
			queries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(90),
					ToBlock:   big.NewInt(110),
					Addresses: nil, // Nil addresses should match all
				},
			},
			expected: [][]etypes.Log{
				{
					{BlockNumber: 100, BlockHash: blockHash1, Address: address1},
					{BlockNumber: 100, BlockHash: blockHash1, Address: address2},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := distributeLogs(tt.allLogs, tt.queries)
			assert.Equal(t, len(tt.expected), len(result))
			for i := range result {
				assert.Equal(t, len(tt.expected[i]), len(result[i]))
				for j := range result[i] {
					assert.Equal(t, tt.expected[i][j].Address, result[i][j].Address)
				}
			}
		})
	}
}

// TestDistributeLogsTopicFiltering tests complex topic filtering scenarios
func TestDistributeLogsTopicFiltering(t *testing.T) {
	blockHash1 := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	address1 := common.HexToAddress("0x742d35Cc6634C893292Ce8bB6239C002Ad8e6b59")
	topic1 := common.HexToHash("0xabc123def456abc123def456abc123def456abc123def456abc123def456abc1")
	topic2 := common.HexToHash("0xdef456abc123def456abc123def456abc123def456abc123def456abc123def4")
	topic3 := common.HexToHash("0x123abc456def123abc456def123abc456def123abc456def123abc456def123a")
	topic4 := common.HexToHash("0x456def123abc456def123abc456def123abc456def123abc456def123abc456d")

	tests := []struct {
		name     string
		allLogs  []etypes.Log
		queries  []ethereum.FilterQuery
		expected [][]etypes.Log
	}{
		{
			name: "Multiple topic positions with complex conditions",
			allLogs: []etypes.Log{
				{
					BlockNumber: 100,
					BlockHash:   blockHash1,
					Address:     address1,
					Topics:      []common.Hash{topic1, topic2, topic3},
				},
				{
					BlockNumber: 100,
					BlockHash:   blockHash1,
					Address:     address1,
					Topics:      []common.Hash{topic1, topic2, topic4},
				},
				{
					BlockNumber: 100,
					BlockHash:   blockHash1,
					Address:     address1,
					Topics:      []common.Hash{topic1, topic3, topic3},
				},
			},
			queries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(90),
					ToBlock:   big.NewInt(110),
					Addresses: []common.Address{address1},
					Topics: [][]common.Hash{
						{topic1},         // First topic must be topic1
						{topic2, topic3}, // Second topic must be topic2 or topic3
						{topic3},         // Third topic must be topic3
					},
				},
			},
			expected: [][]etypes.Log{
				{
					{
						BlockNumber: 100,
						BlockHash:   blockHash1,
						Address:     address1,
						Topics:      []common.Hash{topic1, topic2, topic3},
					},
					{
						BlockNumber: 100,
						BlockHash:   blockHash1,
						Address:     address1,
						Topics:      []common.Hash{topic1, topic3, topic3},
					},
				},
			},
		},
		{
			name: "Topic wildcard at multiple positions",
			allLogs: []etypes.Log{
				{
					BlockNumber: 100,
					BlockHash:   blockHash1,
					Address:     address1,
					Topics:      []common.Hash{topic1, topic2, topic3},
				},
				{
					BlockNumber: 100,
					BlockHash:   blockHash1,
					Address:     address1,
					Topics:      []common.Hash{topic1, topic3, topic4},
				},
			},
			queries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(90),
					ToBlock:   big.NewInt(110),
					Addresses: []common.Address{address1},
					Topics: [][]common.Hash{
						{topic1}, // First topic must be topic1
						nil,      // Second topic is wildcard
						nil,      // Third topic is wildcard
					},
				},
			},
			expected: [][]etypes.Log{
				{
					{
						BlockNumber: 100,
						BlockHash:   blockHash1,
						Address:     address1,
						Topics:      []common.Hash{topic1, topic2, topic3},
					},
					{
						BlockNumber: 100,
						BlockHash:   blockHash1,
						Address:     address1,
						Topics:      []common.Hash{topic1, topic3, topic4},
					},
				},
			},
		},
		{
			name: "Log has fewer topics than query",
			allLogs: []etypes.Log{
				{
					BlockNumber: 100,
					BlockHash:   blockHash1,
					Address:     address1,
					Topics:      []common.Hash{topic1, topic2}, // Only 2 topics
				},
			},
			queries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(90),
					ToBlock:   big.NewInt(110),
					Addresses: []common.Address{address1},
					Topics: [][]common.Hash{
						{topic1}, // First topic must be topic1
						{topic2}, // Second topic must be topic2
						{topic3}, // Third topic must be topic3 (log doesn't have this)
					},
				},
			},
			expected: [][]etypes.Log{{}}, // Should not match
		},
		{
			name: "Query has fewer topics than log",
			allLogs: []etypes.Log{
				{
					BlockNumber: 100,
					BlockHash:   blockHash1,
					Address:     address1,
					Topics:      []common.Hash{topic1, topic2, topic3}, // 3 topics
				},
			},
			queries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(90),
					ToBlock:   big.NewInt(110),
					Addresses: []common.Address{address1},
					Topics: [][]common.Hash{
						{topic1}, // First topic must be topic1
						{topic2}, // Second topic must be topic2
						// No third topic specified
					},
				},
			},
			expected: [][]etypes.Log{
				{{
					BlockNumber: 100,
					BlockHash:   blockHash1,
					Address:     address1,
					Topics:      []common.Hash{topic1, topic2, topic3},
				}},
			},
		},
		{
			name: "Empty topics in log",
			allLogs: []etypes.Log{
				{
					BlockNumber: 100,
					BlockHash:   blockHash1,
					Address:     address1,
					Topics:      []common.Hash{}, // Empty topics
				},
			},
			queries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(90),
					ToBlock:   big.NewInt(110),
					Addresses: []common.Address{address1},
					Topics: [][]common.Hash{
						{topic1}, // Query expects topic1
					},
				},
			},
			expected: [][]etypes.Log{{}}, // Should not match
		},
		{
			name: "Empty topics in query",
			allLogs: []etypes.Log{
				{
					BlockNumber: 100,
					BlockHash:   blockHash1,
					Address:     address1,
					Topics:      []common.Hash{topic1, topic2},
				},
			},
			queries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(90),
					ToBlock:   big.NewInt(110),
					Addresses: []common.Address{address1},
					Topics:    [][]common.Hash{}, // Empty topics in query
				},
			},
			expected: [][]etypes.Log{
				{{
					BlockNumber: 100,
					BlockHash:   blockHash1,
					Address:     address1,
					Topics:      []common.Hash{topic1, topic2},
				}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := distributeLogs(tt.allLogs, tt.queries)
			assert.Equal(t, len(tt.expected), len(result))
			for i := range result {
				assert.Equal(t, len(tt.expected[i]), len(result[i]))
				for j := range result[i] {
					if len(tt.expected[i]) > 0 {
						assert.Equal(t, tt.expected[i][j].Topics, result[i][j].Topics)
					}
				}
			}
		})
	}
}

// TestDistributeLogsBlockConditions tests block condition scenarios
func TestDistributeLogsBlockConditions(t *testing.T) {
	blockHash1 := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	blockHash2 := common.HexToHash("0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321")
	address1 := common.HexToAddress("0x742d35Cc6634C893292Ce8bB6239C002Ad8e6b59")

	tests := []struct {
		name     string
		allLogs  []etypes.Log
		queries  []ethereum.FilterQuery
		expected [][]etypes.Log
	}{
		{
			name: "Block number equals boundary values",
			allLogs: []etypes.Log{
				{BlockNumber: 100, BlockHash: blockHash1, Address: address1},
				{BlockNumber: 200, BlockHash: blockHash1, Address: address1},
			},
			queries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(100), // Equal to first log
					ToBlock:   big.NewInt(200), // Equal to second log
					Addresses: []common.Address{address1},
				},
			},
			expected: [][]etypes.Log{
				{
					{BlockNumber: 100, BlockHash: blockHash1, Address: address1},
					{BlockNumber: 200, BlockHash: blockHash1, Address: address1},
				},
			},
		},
		{
			name: "BlockHash takes precedence over block range",
			allLogs: []etypes.Log{
				{BlockNumber: 100, BlockHash: blockHash1, Address: address1},
				{BlockNumber: 100, BlockHash: blockHash2, Address: address1},
			},
			queries: []ethereum.FilterQuery{
				{
					BlockHash: &blockHash1,
					FromBlock: big.NewInt(50),  // Should be ignored when BlockHash is set
					ToBlock:   big.NewInt(150), // Should be ignored when BlockHash is set
					Addresses: []common.Address{address1},
				},
			},
			expected: [][]etypes.Log{
				{{BlockNumber: 100, BlockHash: blockHash1, Address: address1}},
			},
		},
		{
			name: "Multiple queries with overlapping block ranges",
			allLogs: []etypes.Log{
				{BlockNumber: 100, BlockHash: blockHash1, Address: address1},
				{BlockNumber: 150, BlockHash: blockHash1, Address: address1},
				{BlockNumber: 200, BlockHash: blockHash1, Address: address1},
			},
			queries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(120),
					ToBlock:   big.NewInt(180),
					Addresses: []common.Address{address1},
				},
				{
					FromBlock: big.NewInt(140),
					ToBlock:   big.NewInt(160),
					Addresses: []common.Address{address1},
				},
			},
			expected: [][]etypes.Log{
				{
					{BlockNumber: 150, BlockHash: blockHash1, Address: address1},
				},
				{
					{BlockNumber: 150, BlockHash: blockHash1, Address: address1},
				},
			},
		},
		{
			name: "Nil FromBlock and ToBlock",
			allLogs: []etypes.Log{
				{BlockNumber: 100, BlockHash: blockHash1, Address: address1},
				{BlockNumber: 200, BlockHash: blockHash1, Address: address1},
				{BlockNumber: 300, BlockHash: blockHash1, Address: address1},
			},
			queries: []ethereum.FilterQuery{
				{
					FromBlock: nil, // No lower bound
					ToBlock:   nil, // No upper bound
					Addresses: []common.Address{address1},
				},
			},
			expected: [][]etypes.Log{
				{
					{BlockNumber: 100, BlockHash: blockHash1, Address: address1},
					{BlockNumber: 200, BlockHash: blockHash1, Address: address1},
					{BlockNumber: 300, BlockHash: blockHash1, Address: address1},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := distributeLogs(tt.allLogs, tt.queries)
			assert.Equal(t, len(tt.expected), len(result))
			for i := range result {
				assert.Equal(t, len(tt.expected[i]), len(result[i]))
				for j := range result[i] {
					assert.Equal(t, tt.expected[i][j].BlockNumber, result[i][j].BlockNumber)
					assert.Equal(t, tt.expected[i][j].BlockHash, result[i][j].BlockHash)
				}
			}
		})
	}
}

// TestDistributeLogsComplexScenarios tests complex combination scenarios
func TestDistributeLogsComplexScenarios(t *testing.T) {
	blockHash1 := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	blockHash2 := common.HexToHash("0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321")
	address1 := common.HexToAddress("0x742d35Cc6634C893292Ce8bB6239C002Ad8e6b59")
	address2 := common.HexToAddress("0x853d43Cc6634C893292Ce8bB6239C002Ad8e6b60")
	topic1 := common.HexToHash("0xabc123def456abc123def456abc123def456abc123def456abc123def456abc1")
	topic2 := common.HexToHash("0xdef456abc123def456abc123def456abc123def456abc123def456abc123def4")

	tests := []struct {
		name     string
		allLogs  []etypes.Log
		queries  []ethereum.FilterQuery
		expected [][]etypes.Log
	}{
		{
			name: "Complex combination of all conditions",
			allLogs: []etypes.Log{
				{
					BlockNumber: 100,
					BlockHash:   blockHash1,
					Address:     address1,
					Topics:      []common.Hash{topic1, topic2},
				},
				{
					BlockNumber: 100,
					BlockHash:   blockHash1,
					Address:     address2,
					Topics:      []common.Hash{topic1, topic2},
				},
				{
					BlockNumber: 150,
					BlockHash:   blockHash1,
					Address:     address1,
					Topics:      []common.Hash{topic1, topic2},
				},
				{
					BlockNumber: 150,
					BlockHash:   blockHash2,
					Address:     address1,
					Topics:      []common.Hash{topic1, topic2},
				},
			},
			queries: []ethereum.FilterQuery{
				{
					BlockHash: &blockHash1,
					Addresses: []common.Address{address1},
					Topics:    [][]common.Hash{{topic1}, {topic2}},
				},
				{
					FromBlock: big.NewInt(120),
					ToBlock:   big.NewInt(180),
					Addresses: []common.Address{address1, address2},
					Topics:    [][]common.Hash{{topic1}, nil},
				},
			},
			expected: [][]etypes.Log{
				{
					{
						BlockNumber: 100,
						BlockHash:   blockHash1,
						Address:     address1,
						Topics:      []common.Hash{topic1, topic2},
					},
					{
						BlockNumber: 150,
						BlockHash:   blockHash1,
						Address:     address1,
						Topics:      []common.Hash{topic1, topic2},
					},
				},
				{
					{
						BlockNumber: 150,
						BlockHash:   blockHash1,
						Address:     address1,
						Topics:      []common.Hash{topic1, topic2},
					},
					{
						BlockNumber: 150,
						BlockHash:   blockHash2,
						Address:     address1,
						Topics:      []common.Hash{topic1, topic2},
					},
				},
			},
		},
		{
			name: "Logs with incomplete data",
			allLogs: []etypes.Log{
				{
					BlockNumber: 100,
					BlockHash:   blockHash1,
					Address:     address1,
					// No topics
				},
				{
					BlockNumber: 100,
					BlockHash:   blockHash1,
					// No address
					Topics: []common.Hash{topic1},
				},
			},
			queries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(90),
					ToBlock:   big.NewInt(110),
					Addresses: []common.Address{address1},
					Topics:    [][]common.Hash{{topic1}},
				},
			},
			expected: [][]etypes.Log{{}}, // Should not match due to missing data
		},
		{
			name: "Multiple queries with different matching criteria",
			allLogs: []etypes.Log{
				{
					BlockNumber: 100,
					BlockHash:   blockHash1,
					Address:     address1,
					Topics:      []common.Hash{topic1, topic2},
				},
				{
					BlockNumber: 200,
					BlockHash:   blockHash2,
					Address:     address2,
					Topics:      []common.Hash{topic1},
				},
			},
			queries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(50),
					ToBlock:   big.NewInt(150),
					Addresses: []common.Address{address1},
					Topics:    [][]common.Hash{{topic1}},
				},
				{
					BlockHash: &blockHash2,
					Addresses: []common.Address{address2},
				},
			},
			expected: [][]etypes.Log{
				{{
					BlockNumber: 100,
					BlockHash:   blockHash1,
					Address:     address1,
					Topics:      []common.Hash{topic1, topic2},
				}},
				{{
					BlockNumber: 200,
					BlockHash:   blockHash2,
					Address:     address2,
					Topics:      []common.Hash{topic1},
				}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := distributeLogs(tt.allLogs, tt.queries)
			assert.Equal(t, len(tt.expected), len(result))
			for i := range result {
				assert.Equal(t, len(tt.expected[i]), len(result[i]))
				for j := range result[i] {
					if len(tt.expected[i]) > 0 {
						assert.Equal(t, tt.expected[i][j].BlockNumber, result[i][j].BlockNumber)
						assert.Equal(t, tt.expected[i][j].BlockHash, result[i][j].BlockHash)
						assert.Equal(t, tt.expected[i][j].Address, result[i][j].Address)
						if len(tt.expected[i][j].Topics) > 0 || len(result[i][j].Topics) > 0 {
							assert.Equal(t, tt.expected[i][j].Topics, result[i][j].Topics)
						}
					}
				}
			}
		})
	}
}

func TestSplitFilterQuery(t *testing.T) {
	testcases := []struct {
		query                     ethereum.FilterQuery
		maxAddressesPerQuery      int
		expectedQueriesAfterSplit []ethereum.FilterQuery
	}{
		{
			query: ethereum.FilterQuery{
				FromBlock: big.NewInt(1),
				ToBlock:   big.NewInt(2),
				Addresses: []common.Address{
					common.HexToAddress("0x1"),
					common.HexToAddress("0x2"),
					common.HexToAddress("0x3"),
					common.HexToAddress("0x4"),
					common.HexToAddress("0x5"),
				},
			},
			maxAddressesPerQuery: 2,
			expectedQueriesAfterSplit: []ethereum.FilterQuery{
				ethereum.FilterQuery{
					FromBlock: big.NewInt(1),
					ToBlock:   big.NewInt(2),
					Addresses: []common.Address{
						common.HexToAddress("0x1"),
						common.HexToAddress("0x2"),
					},
				},
				ethereum.FilterQuery{
					FromBlock: big.NewInt(1),
					ToBlock:   big.NewInt(2),
					Addresses: []common.Address{
						common.HexToAddress("0x3"),
						common.HexToAddress("0x4"),
					},
				},
				ethereum.FilterQuery{
					FromBlock: big.NewInt(1),
					ToBlock:   big.NewInt(2),
					Addresses: []common.Address{
						common.HexToAddress("0x5"),
					},
				},
			},
		},
		{
			query: ethereum.FilterQuery{
				FromBlock: big.NewInt(1),
				ToBlock:   big.NewInt(2),
				Addresses: []common.Address{
					common.HexToAddress("0x1"),
					common.HexToAddress("0x2"),
					common.HexToAddress("0x3"),
					common.HexToAddress("0x4"),
					common.HexToAddress("0x5"),
				},
			},
			maxAddressesPerQuery: 3,
			expectedQueriesAfterSplit: []ethereum.FilterQuery{
				ethereum.FilterQuery{
					FromBlock: big.NewInt(1),
					ToBlock:   big.NewInt(2),
					Addresses: []common.Address{
						common.HexToAddress("0x1"),
						common.HexToAddress("0x2"),
						common.HexToAddress("0x3"),
					},
				},
				ethereum.FilterQuery{
					FromBlock: big.NewInt(1),
					ToBlock:   big.NewInt(2),
					Addresses: []common.Address{
						common.HexToAddress("0x4"),
						common.HexToAddress("0x5"),
					},
				},
			},
		},
	}

	for i, tt := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			queriesAfterSplit, err := splitFilterQuery(tt.query, tt.maxAddressesPerQuery)
			if err != nil {
				t.Fatal(err)
			}

			if len(queriesAfterSplit) != len(tt.expectedQueriesAfterSplit) {
				t.Errorf("queriesAfterSplit length mismatch, queriesAfterSplit %v, but want %v", len(queriesAfterSplit), len(tt.expectedQueriesAfterSplit))
			}

			assert.Equal(t, tt.expectedQueriesAfterSplit, queriesAfterSplit, "queriesAfterSplit mismatch")
		})
	}
}

func TestMergeFilterQuery(t *testing.T) {
	testcases := []struct {
		name                       string
		inputQueries               []ethereum.FilterQuery
		mergeOverlappingBlockRange bool
		expectedMergedQueries      []ethereum.FilterQuery
	}{
		{
			name: "Merge same block range and topics",
			inputQueries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(1),
					ToBlock:   big.NewInt(10),
					Addresses: []common.Address{
						common.HexToAddress("0x1"),
					},
					Topics: [][]common.Hash{
						{common.HexToHash("0xA")},
					},
				},
				{
					FromBlock: big.NewInt(1),
					ToBlock:   big.NewInt(10),
					Addresses: []common.Address{
						common.HexToAddress("0x2"),
					},
					Topics: [][]common.Hash{
						{common.HexToHash("0xA")},
					},
				},
			},
			mergeOverlappingBlockRange: true,
			expectedMergedQueries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(1),
					ToBlock:   big.NewInt(10),
					Addresses: []common.Address{
						common.HexToAddress("0x1"),
						common.HexToAddress("0x2"),
					},
					Topics: [][]common.Hash{
						{common.HexToHash("0xA")},
					},
				},
			},
		},
		{
			name: "Merge with topic hierarchy",
			inputQueries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(100),
					ToBlock:   big.NewInt(200),
					Addresses: []common.Address{
						common.HexToAddress("0x3"),
					},
					Topics: [][]common.Hash{
						{common.HexToHash("0xB")},
						{common.HexToHash("0xC")},
					},
				},
				{
					FromBlock: big.NewInt(100),
					ToBlock:   big.NewInt(200),
					Addresses: []common.Address{
						common.HexToAddress("0x4"),
					},
					Topics: [][]common.Hash{
						{common.HexToHash("0xB")},
						nil, // Second topic is wildcard
					},
				},
			},
			mergeOverlappingBlockRange: true,
			expectedMergedQueries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(100),
					ToBlock:   big.NewInt(200),
					Addresses: []common.Address{
						common.HexToAddress("0x3"),
						common.HexToAddress("0x4"),
					},
					Topics: [][]common.Hash{
						{common.HexToHash("0xB")},
						nil, // Merged to wildcard
					},
				},
			},
		},
		{
			name: "Merge overlapping block ranges",
			inputQueries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(50),
					ToBlock:   big.NewInt(150),
					Addresses: []common.Address{
						common.HexToAddress("0x5"),
					},
				},
				{
					FromBlock: big.NewInt(100),
					ToBlock:   big.NewInt(200),
					Addresses: []common.Address{
						common.HexToAddress("0x6"),
					},
				},
			},
			mergeOverlappingBlockRange: true,
			expectedMergedQueries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(50),  // Expanded from
					ToBlock:   big.NewInt(200), // Expanded to
					Addresses: []common.Address{
						common.HexToAddress("0x5"),
						common.HexToAddress("0x6"),
					},
				},
			},
		},
		{
			name: "Don't merge different block hashes",
			inputQueries: []ethereum.FilterQuery{
				{
					BlockHash: ptrToHash("0x123"),
					Addresses: []common.Address{
						common.HexToAddress("0x7"),
					},
				},
				{
					BlockHash: ptrToHash("0x456"),
					Addresses: []common.Address{
						common.HexToAddress("0x8"),
					},
				},
			},
			mergeOverlappingBlockRange: true,
			expectedMergedQueries: []ethereum.FilterQuery{
				{
					BlockHash: ptrToHash("0x123"),
					Addresses: []common.Address{
						common.HexToAddress("0x7"),
					},
				},
				{
					BlockHash: ptrToHash("0x456"),
					Addresses: []common.Address{
						common.HexToAddress("0x8"),
					},
				},
			},
		},
		{
			name: "Don't merge overlapping block ranges when mergeOverlappingBlockRange is false",
			inputQueries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(50),
					ToBlock:   big.NewInt(150),
					Addresses: []common.Address{
						common.HexToAddress("0x5"),
					},
				},
				{
					FromBlock: big.NewInt(100),
					ToBlock:   big.NewInt(200),
					Addresses: []common.Address{
						common.HexToAddress("0x6"),
					},
				},
			},
			mergeOverlappingBlockRange: false,
			expectedMergedQueries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(50),
					ToBlock:   big.NewInt(150),
					Addresses: []common.Address{
						common.HexToAddress("0x5"),
					},
				},
				{
					FromBlock: big.NewInt(100),
					ToBlock:   big.NewInt(200),
					Addresses: []common.Address{
						common.HexToAddress("0x6"),
					},
				},
			},
		},
		{
			name: "Merge same block ranges when mergeOverlappingBlockRange is false",
			inputQueries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(100),
					ToBlock:   big.NewInt(200),
					Addresses: []common.Address{
						common.HexToAddress("0x5"),
					},
				},
				{
					FromBlock: big.NewInt(100),
					ToBlock:   big.NewInt(200),
					Addresses: []common.Address{
						common.HexToAddress("0x6"),
					},
				},
			},
			mergeOverlappingBlockRange: false,
			expectedMergedQueries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(100),
					ToBlock:   big.NewInt(200),
					Addresses: []common.Address{
						common.HexToAddress("0x5"),
						common.HexToAddress("0x6"),
					},
				},
			},
		},
		{
			name: "Complex overlapping block ranges with multiple queries",
			inputQueries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(100),
					ToBlock:   big.NewInt(200),
					Addresses: []common.Address{
						common.HexToAddress("0x1"),
					},
				},
				{
					FromBlock: big.NewInt(150),
					ToBlock:   big.NewInt(250),
					Addresses: []common.Address{
						common.HexToAddress("0x2"),
					},
				},
				{
					FromBlock: big.NewInt(200),
					ToBlock:   big.NewInt(300),
					Addresses: []common.Address{
						common.HexToAddress("0x3"),
					},
				},
				{
					FromBlock: big.NewInt(50),
					ToBlock:   big.NewInt(150),
					Addresses: []common.Address{
						common.HexToAddress("0x4"),
					},
				},
			},
			mergeOverlappingBlockRange: true,
			expectedMergedQueries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(50),  // Expanded from minimum
					ToBlock:   big.NewInt(300), // Expanded to maximum
					Addresses: []common.Address{
						common.HexToAddress("0x1"),
						common.HexToAddress("0x2"),
						common.HexToAddress("0x3"),
						common.HexToAddress("0x4"),
					},
				},
			},
		},
		{
			name: "Complex overlapping block ranges with mergeOverlappingBlockRange false",
			inputQueries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(100),
					ToBlock:   big.NewInt(200),
					Addresses: []common.Address{
						common.HexToAddress("0x1"),
					},
				},
				{
					FromBlock: big.NewInt(150),
					ToBlock:   big.NewInt(250),
					Addresses: []common.Address{
						common.HexToAddress("0x2"),
					},
				},
				{
					FromBlock: big.NewInt(100),
					ToBlock:   big.NewInt(200),
					Addresses: []common.Address{
						common.HexToAddress("0x3"),
					},
				},
			},
			mergeOverlappingBlockRange: false,
			expectedMergedQueries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(100),
					ToBlock:   big.NewInt(200),
					Addresses: []common.Address{
						common.HexToAddress("0x1"),
						common.HexToAddress("0x3"),
					},
				},
				{
					FromBlock: big.NewInt(150),
					ToBlock:   big.NewInt(250),
					Addresses: []common.Address{
						common.HexToAddress("0x2"),
					},
				},
			},
		},
		{
			name: "Complex topic merging with multiple levels",
			inputQueries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(100),
					ToBlock:   big.NewInt(200),
					Addresses: []common.Address{
						common.HexToAddress("0x1"),
					},
					Topics: [][]common.Hash{
						{common.HexToHash("0xA")},
						{common.HexToHash("0xB")},
						{common.HexToHash("0xC")},
					},
				},
				{
					FromBlock: big.NewInt(100),
					ToBlock:   big.NewInt(200),
					Addresses: []common.Address{
						common.HexToAddress("0x2"),
					},
					Topics: [][]common.Hash{
						{common.HexToHash("0xA")},
						nil, // Wildcard at second level
						{common.HexToHash("0xC")},
					},
				},
				{
					FromBlock: big.NewInt(100),
					ToBlock:   big.NewInt(200),
					Addresses: []common.Address{
						common.HexToAddress("0x3"),
					},
					Topics: [][]common.Hash{
						{common.HexToHash("0xA")},
						{common.HexToHash("0xB")},
						nil, // Wildcard at third level
					},
				},
			},
			mergeOverlappingBlockRange: true,
			expectedMergedQueries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(100),
					ToBlock:   big.NewInt(200),
					Addresses: []common.Address{
						common.HexToAddress("0x1"),
						common.HexToAddress("0x2"),
						common.HexToAddress("0x3"),
					},
					Topics: [][]common.Hash{
						{common.HexToHash("0xA")},
						nil, // Merged to wildcard due to mixed topics
						nil, // Merged to wildcard due to mixed topics
					},
				},
			},
		},
		{
			name: "Mixed BlockHash and block range queries",
			inputQueries: []ethereum.FilterQuery{
				{
					BlockHash: ptrToHash("0x123"),
					Addresses: []common.Address{
						common.HexToAddress("0x1"),
					},
					Topics: [][]common.Hash{
						{common.HexToHash("0xA")},
					},
				},
				{
					FromBlock: big.NewInt(100),
					ToBlock:   big.NewInt(200),
					Addresses: []common.Address{
						common.HexToAddress("0x2"),
					},
					Topics: [][]common.Hash{
						{common.HexToHash("0xA")},
					},
				},
				{
					BlockHash: ptrToHash("0x123"),
					Addresses: []common.Address{
						common.HexToAddress("0x3"),
					},
					Topics: [][]common.Hash{
						{common.HexToHash("0xB")},
					},
				},
			},
			mergeOverlappingBlockRange: true,
			expectedMergedQueries: []ethereum.FilterQuery{
				{
					BlockHash: ptrToHash("0x123"),
					Addresses: []common.Address{
						common.HexToAddress("0x1"),
						common.HexToAddress("0x3"),
					},
					Topics: [][]common.Hash{
						{common.HexToHash("0xA"), common.HexToHash("0xB")},
					},
				},
				{
					FromBlock: big.NewInt(100),
					ToBlock:   big.NewInt(200),
					Addresses: []common.Address{
						common.HexToAddress("0x2"),
					},
					Topics: [][]common.Hash{
						{common.HexToHash("0xA")},
					},
				},
			},
		},
		{
			name: "Nil block ranges handling",
			inputQueries: []ethereum.FilterQuery{
				{
					FromBlock: nil,
					ToBlock:   big.NewInt(200),
					Addresses: []common.Address{
						common.HexToAddress("0x1"),
					},
				},
				{
					FromBlock: big.NewInt(100),
					ToBlock:   nil,
					Addresses: []common.Address{
						common.HexToAddress("0x2"),
					},
				},
				{
					FromBlock: nil,
					ToBlock:   nil,
					Addresses: []common.Address{
						common.HexToAddress("0x3"),
					},
				},
			},
			mergeOverlappingBlockRange: true,
			expectedMergedQueries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(100), // Expanded from minimum non-nil
					ToBlock:   big.NewInt(200), // Expanded to maximum non-nil
					Addresses: []common.Address{
						common.HexToAddress("0x1"),
						common.HexToAddress("0x2"),
						common.HexToAddress("0x3"),
					},
				},
			},
		},
		{
			name: "Empty addresses and topics",
			inputQueries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(100),
					ToBlock:   big.NewInt(200),
					Addresses: []common.Address{},
					Topics:    [][]common.Hash{},
				},
				{
					FromBlock: big.NewInt(100),
					ToBlock:   big.NewInt(200),
					Addresses: []common.Address{
						common.HexToAddress("0x1"),
					},
					Topics: [][]common.Hash{},
				},
			},
			mergeOverlappingBlockRange: true,
			expectedMergedQueries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(100),
					ToBlock:   big.NewInt(200),
					Addresses: []common.Address{
						common.HexToAddress("0x1"),
					},
					Topics: [][]common.Hash(nil), // nil is the actual result
				},
			},
		},
		{
			name: "Single query should remain unchanged",
			inputQueries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(100),
					ToBlock:   big.NewInt(200),
					Addresses: []common.Address{
						common.HexToAddress("0x1"),
					},
					Topics: [][]common.Hash{
						{common.HexToHash("0xA")},
					},
				},
			},
			mergeOverlappingBlockRange: true,
			expectedMergedQueries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(100),
					ToBlock:   big.NewInt(200),
					Addresses: []common.Address{
						common.HexToAddress("0x1"),
					},
					Topics: [][]common.Hash{
						{common.HexToHash("0xA")},
					},
				},
			},
		},
		{
			name: "Adjacent block ranges with mergeOverlappingBlockRange true",
			inputQueries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(100),
					ToBlock:   big.NewInt(150),
					Addresses: []common.Address{
						common.HexToAddress("0x1"),
					},
				},
				{
					FromBlock: big.NewInt(151),
					ToBlock:   big.NewInt(200),
					Addresses: []common.Address{
						common.HexToAddress("0x2"),
					},
				},
			},
			mergeOverlappingBlockRange: true,
			expectedMergedQueries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(100),
					ToBlock:   big.NewInt(200),
					Addresses: []common.Address{
						common.HexToAddress("0x1"),
						common.HexToAddress("0x2"),
					},
				},
			},
		},
		{
			name: "Adjacent block ranges with mergeOverlappingBlockRange false",
			inputQueries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(100),
					ToBlock:   big.NewInt(150),
					Addresses: []common.Address{
						common.HexToAddress("0x1"),
					},
				},
				{
					FromBlock: big.NewInt(151),
					ToBlock:   big.NewInt(200),
					Addresses: []common.Address{
						common.HexToAddress("0x2"),
					},
				},
			},
			mergeOverlappingBlockRange: false,
			expectedMergedQueries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(100),
					ToBlock:   big.NewInt(150),
					Addresses: []common.Address{
						common.HexToAddress("0x1"),
					},
				},
				{
					FromBlock: big.NewInt(151),
					ToBlock:   big.NewInt(200),
					Addresses: []common.Address{
						common.HexToAddress("0x2"),
					},
				},
			},
		},
		{
			name: "Multiple BlockHash queries with same hash",
			inputQueries: []ethereum.FilterQuery{
				{
					BlockHash: ptrToHash("0x123"),
					Addresses: []common.Address{
						common.HexToAddress("0x1"),
					},
				},
				{
					BlockHash: ptrToHash("0x123"),
					Addresses: []common.Address{
						common.HexToAddress("0x2"),
					},
				},
				{
					BlockHash: ptrToHash("0x123"),
					Addresses: []common.Address{
						common.HexToAddress("0x3"),
					},
				},
			},
			mergeOverlappingBlockRange: true,
			expectedMergedQueries: []ethereum.FilterQuery{
				{
					BlockHash: ptrToHash("0x123"),
					Addresses: []common.Address{
						common.HexToAddress("0x1"),
						common.HexToAddress("0x2"),
						common.HexToAddress("0x3"),
					},
				},
			},
		},
		{
			name: "Complex topic wildcard scenarios",
			inputQueries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(100),
					ToBlock:   big.NewInt(200),
					Addresses: []common.Address{
						common.HexToAddress("0x1"),
					},
					Topics: [][]common.Hash{
						nil, // Wildcard at first level
						{common.HexToHash("0xB")},
					},
				},
				{
					FromBlock: big.NewInt(100),
					ToBlock:   big.NewInt(200),
					Addresses: []common.Address{
						common.HexToAddress("0x2"),
					},
					Topics: [][]common.Hash{
						{common.HexToHash("0xA")},
						nil, // Wildcard at second level
					},
				},
			},
			mergeOverlappingBlockRange: true,
			expectedMergedQueries: []ethereum.FilterQuery{
				{
					FromBlock: big.NewInt(100),
					ToBlock:   big.NewInt(200),
					Addresses: []common.Address{
						common.HexToAddress("0x1"),
						common.HexToAddress("0x2"),
					},
					Topics: [][]common.Hash{
						nil, // Merged to wildcard
						nil, // Merged to wildcard
					},
				},
			},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			mergedQueries, err := mergeFilterQueries(tt.inputQueries, tt.mergeOverlappingBlockRange)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, len(tt.expectedMergedQueries), len(mergedQueries),
				"merged queries count mismatch")

			for i := range mergedQueries {
				// Sort addresses for comparison
				sortAddresses(mergedQueries[i].Addresses)
				sortAddresses(tt.expectedMergedQueries[i].Addresses)
			}

			// Sort both expected and actual results for comparison
			sortFilterQueries(tt.expectedMergedQueries)
			sortFilterQueries(mergedQueries)

			assert.Equal(t, tt.expectedMergedQueries, mergedQueries,
				"merged queries content mismatch")
		})
	}
}

// Helper functions
func ptrToHash(s string) *common.Hash {
	h := common.HexToHash(s)
	return &h
}

func sortAddresses(addrs []common.Address) {
	if len(addrs) <= 1 {
		return
	}
	// Simple bubble sort for test purposes
	for i := 0; i < len(addrs)-1; i++ {
		for j := i + 1; j < len(addrs); j++ {
			if addrs[i].Hex() > addrs[j].Hex() {
				addrs[i], addrs[j] = addrs[j], addrs[i]
			}
		}
	}
}

func sortFilterQueries(queries []ethereum.FilterQuery) {
	if len(queries) <= 1 {
		return
	}
	// Simple bubble sort for test purposes
	for i := 0; i < len(queries)-1; i++ {
		for j := i + 1; j < len(queries); j++ {
			if shouldSwap(queries[i], queries[j]) {
				queries[i], queries[j] = queries[j], queries[i]
			}
		}
	}
}

func shouldSwap(a, b ethereum.FilterQuery) bool {
	// Compare BlockHash first
	if a.BlockHash != nil && b.BlockHash != nil {
		return a.BlockHash.Hex() > b.BlockHash.Hex()
	}
	if a.BlockHash != nil {
		return false // a has BlockHash, b doesn't, so a should come first
	}
	if b.BlockHash != nil {
		return true // b has BlockHash, a doesn't, so b should come first
	}

	// Compare FromBlock
	if a.FromBlock != nil && b.FromBlock != nil {
		if a.FromBlock.Cmp(b.FromBlock) != 0 {
			return a.FromBlock.Cmp(b.FromBlock) > 0
		}
	} else if a.FromBlock != nil {
		return false
	} else if b.FromBlock != nil {
		return true
	}

	// Compare ToBlock
	if a.ToBlock != nil && b.ToBlock != nil {
		if a.ToBlock.Cmp(b.ToBlock) != 0 {
			return a.ToBlock.Cmp(b.ToBlock) > 0
		}
	} else if a.ToBlock != nil {
		return false
	} else if b.ToBlock != nil {
		return true
	}

	// Compare first address
	if len(a.Addresses) > 0 && len(b.Addresses) > 0 {
		return a.Addresses[0].Hex() > b.Addresses[0].Hex()
	}

	return false
}

// TestGetQueryHash tests the GetQueryHash function with various scenarios
func TestGetQueryHash(t *testing.T) {
	// Test data
	chainId1 := big.NewInt(1)
	chainId2 := big.NewInt(137)
	address1 := common.HexToAddress("0x742d35Cc6634C893292Ce8bB6239C002Ad8e6b59")
	address2 := common.HexToAddress("0x853d43Cc6634C893292Ce8bB6239C002Ad8e6b60")
	address3 := common.HexToAddress("0x964d51Dd6634C893292Ce8bB6239C002Ad8e6b71")
	topic1 := common.HexToHash("0xabc123def456abc123def456abc123def456abc123def456abc123def456abc1")
	topic2 := common.HexToHash("0xdef456abc123def456abc123def456abc123def456abc123def456abc123def4")
	topic3 := common.HexToHash("0x123abc456def123abc456def123abc456def123abc456def123abc456def123a")
	blockHash1 := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")

	tests := []struct {
		name     string
		chainId  *big.Int
		query    ethereum.FilterQuery
		expected common.Hash
	}{
		{
			name:    "Basic query with single address and topic",
			chainId: chainId1,
			query: ethereum.FilterQuery{
				FromBlock: big.NewInt(100),
				ToBlock:   big.NewInt(200),
				Addresses: []common.Address{address1},
				Topics:    [][]common.Hash{{topic1}},
			},
		},
		{
			name:    "Query with multiple addresses",
			chainId: chainId1,
			query: ethereum.FilterQuery{
				FromBlock: big.NewInt(100),
				ToBlock:   big.NewInt(200),
				Addresses: []common.Address{address2, address1, address3}, // Will be sorted
				Topics:    [][]common.Hash{{topic1}},
			},
		},
		{
			name:    "Query with multiple topics",
			chainId: chainId1,
			query: ethereum.FilterQuery{
				FromBlock: big.NewInt(100),
				ToBlock:   big.NewInt(200),
				Addresses: []common.Address{address1},
				Topics:    [][]common.Hash{{topic2, topic1, topic3}}, // Will be sorted
			},
		},
		{
			name:    "Query with complex topic structure",
			chainId: chainId1,
			query: ethereum.FilterQuery{
				FromBlock: big.NewInt(100),
				ToBlock:   big.NewInt(200),
				Addresses: []common.Address{address1},
				Topics: [][]common.Hash{
					{topic1},
					{topic2, topic3},
					{topic1, topic2},
				},
			},
		},
		{
			name:    "Query with block hash",
			chainId: chainId1,
			query: ethereum.FilterQuery{
				BlockHash: &blockHash1,
				Addresses: []common.Address{address1},
				Topics:    [][]common.Hash{{topic1}},
			},
		},
		{
			name:    "Query with nil block range",
			chainId: chainId1,
			query: ethereum.FilterQuery{
				FromBlock: nil,
				ToBlock:   nil,
				Addresses: []common.Address{address1},
				Topics:    [][]common.Hash{{topic1}},
			},
		},
		{
			name:    "Query with empty addresses",
			chainId: chainId1,
			query: ethereum.FilterQuery{
				FromBlock: big.NewInt(100),
				ToBlock:   big.NewInt(200),
				Addresses: []common.Address{},
				Topics:    [][]common.Hash{{topic1}},
			},
		},
		{
			name:    "Query with nil addresses",
			chainId: chainId1,
			query: ethereum.FilterQuery{
				FromBlock: big.NewInt(100),
				ToBlock:   big.NewInt(200),
				Addresses: nil,
				Topics:    [][]common.Hash{{topic1}},
			},
		},
		{
			name:    "Query with empty topics",
			chainId: chainId1,
			query: ethereum.FilterQuery{
				FromBlock: big.NewInt(100),
				ToBlock:   big.NewInt(200),
				Addresses: []common.Address{address1},
				Topics:    [][]common.Hash{},
			},
		},
		{
			name:    "Query with nil topics",
			chainId: chainId1,
			query: ethereum.FilterQuery{
				FromBlock: big.NewInt(100),
				ToBlock:   big.NewInt(200),
				Addresses: []common.Address{address1},
				Topics:    nil,
			},
		},
		{
			name:    "Query with empty topic group",
			chainId: chainId1,
			query: ethereum.FilterQuery{
				FromBlock: big.NewInt(100),
				ToBlock:   big.NewInt(200),
				Addresses: []common.Address{address1},
				Topics:    [][]common.Hash{{}},
			},
		},
		{
			name:    "Query with wildcard topics",
			chainId: chainId1,
			query: ethereum.FilterQuery{
				FromBlock: big.NewInt(100),
				ToBlock:   big.NewInt(200),
				Addresses: []common.Address{address1},
				Topics: [][]common.Hash{
					{topic1},
					nil, // Wildcard
					{topic2},
				},
			},
		},
		{
			name:    "Different chain ID",
			chainId: chainId2,
			query: ethereum.FilterQuery{
				FromBlock: big.NewInt(100),
				ToBlock:   big.NewInt(200),
				Addresses: []common.Address{address1},
				Topics:    [][]common.Hash{{topic1}},
			},
		},
		{
			name:    "Minimal query",
			chainId: chainId1,
			query:   ethereum.FilterQuery{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := GetQueryHash(tt.chainId, tt.query)

			// Verify hash is not zero
			assert.NotEqual(t, common.Hash{}, hash, "Hash should not be zero")

			// Verify hash is consistent (same input should produce same hash)
			hash2 := GetQueryHash(tt.chainId, tt.query)
			assert.Equal(t, hash, hash2, "Hash should be consistent for same input")
		})
	}
}

// TestGetQueryHashConsistency tests that the hash is consistent regardless of input order
func TestGetQueryHashConsistency(t *testing.T) {
	chainId := big.NewInt(1)
	address1 := common.HexToAddress("0x742d35Cc6634C893292Ce8bB6239C002Ad8e6b59")
	address2 := common.HexToAddress("0x853d43Cc6634C893292Ce8bB6239C002Ad8e6b60")
	address3 := common.HexToAddress("0x964d51Dd6634C893292Ce8bB6239C002Ad8e6b71")
	topic1 := common.HexToHash("0xabc123def456abc123def456abc123def456abc123def456abc123def456abc1")
	topic2 := common.HexToHash("0xdef456abc123def456abc123def456abc123def456abc123def456abc123def4")
	topic3 := common.HexToHash("0x123abc456def123abc456def123abc456def123abc456def123abc456def123a")

	tests := []struct {
		name     string
		query1   ethereum.FilterQuery
		query2   ethereum.FilterQuery
		expected bool // true if hashes should be equal
	}{
		{
			name: "Same addresses in different order",
			query1: ethereum.FilterQuery{
				FromBlock: big.NewInt(100),
				ToBlock:   big.NewInt(200),
				Addresses: []common.Address{address1, address2, address3},
				Topics:    [][]common.Hash{{topic1}},
			},
			query2: ethereum.FilterQuery{
				FromBlock: big.NewInt(100),
				ToBlock:   big.NewInt(200),
				Addresses: []common.Address{address3, address1, address2}, // Different order
				Topics:    [][]common.Hash{{topic1}},
			},
			expected: true,
		},
		{
			name: "Same topics in different order",
			query1: ethereum.FilterQuery{
				FromBlock: big.NewInt(100),
				ToBlock:   big.NewInt(200),
				Addresses: []common.Address{address1},
				Topics:    [][]common.Hash{{topic1, topic2, topic3}},
			},
			query2: ethereum.FilterQuery{
				FromBlock: big.NewInt(100),
				ToBlock:   big.NewInt(200),
				Addresses: []common.Address{address1},
				Topics:    [][]common.Hash{{topic3, topic1, topic2}}, // Different order
			},
			expected: true,
		},
		{
			name: "Different addresses",
			query1: ethereum.FilterQuery{
				FromBlock: big.NewInt(100),
				ToBlock:   big.NewInt(200),
				Addresses: []common.Address{address1, address2},
				Topics:    [][]common.Hash{{topic1}},
			},
			query2: ethereum.FilterQuery{
				FromBlock: big.NewInt(100),
				ToBlock:   big.NewInt(200),
				Addresses: []common.Address{address1, address3}, // Different address
				Topics:    [][]common.Hash{{topic1}},
			},
			expected: false,
		},
		{
			name: "Different topics",
			query1: ethereum.FilterQuery{
				FromBlock: big.NewInt(100),
				ToBlock:   big.NewInt(200),
				Addresses: []common.Address{address1},
				Topics:    [][]common.Hash{{topic1, topic2}},
			},
			query2: ethereum.FilterQuery{
				FromBlock: big.NewInt(100),
				ToBlock:   big.NewInt(200),
				Addresses: []common.Address{address1},
				Topics:    [][]common.Hash{{topic1, topic3}}, // Different topic
			},
			expected: false,
		},
		{
			name: "Different block range",
			query1: ethereum.FilterQuery{
				FromBlock: big.NewInt(100),
				ToBlock:   big.NewInt(200),
				Addresses: []common.Address{address1},
				Topics:    [][]common.Hash{{topic1}},
			},
			query2: ethereum.FilterQuery{
				FromBlock: big.NewInt(100),
				ToBlock:   big.NewInt(300), // Different ToBlock
				Addresses: []common.Address{address1},
				Topics:    [][]common.Hash{{topic1}},
			},
			expected: false,
		},
		{
			name: "Empty vs nil addresses",
			query1: ethereum.FilterQuery{
				FromBlock: big.NewInt(100),
				ToBlock:   big.NewInt(200),
				Addresses: []common.Address{},
				Topics:    [][]common.Hash{{topic1}},
			},
			query2: ethereum.FilterQuery{
				FromBlock: big.NewInt(100),
				ToBlock:   big.NewInt(200),
				Addresses: nil,
				Topics:    [][]common.Hash{{topic1}},
			},
			expected: false, // Empty slice and nil are different in JSON
		},
		{
			name: "Empty vs nil topics",
			query1: ethereum.FilterQuery{
				FromBlock: big.NewInt(100),
				ToBlock:   big.NewInt(200),
				Addresses: []common.Address{address1},
				Topics:    [][]common.Hash{},
			},
			query2: ethereum.FilterQuery{
				FromBlock: big.NewInt(100),
				ToBlock:   big.NewInt(200),
				Addresses: []common.Address{address1},
				Topics:    nil,
			},
			expected: true, // Both empty and nil topics get normalized to [[]common.Hash{{}}]
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := GetQueryHash(chainId, tt.query1)
			hash2 := GetQueryHash(chainId, tt.query2)

			if tt.expected {
				assert.Equal(t, hash1, hash2, "Hashes should be equal")
			} else {
				assert.NotEqual(t, hash1, hash2, "Hashes should be different")
			}
		})
	}
}

// TestGetQueryHashEdgeCases tests edge cases and boundary conditions
func TestGetQueryHashEdgeCases(t *testing.T) {
	chainId := big.NewInt(1)
	address1 := common.HexToAddress("0x742d35Cc6634C893292Ce8bB6239C002Ad8e6b59")
	topic1 := common.HexToHash("0xabc123def456abc123def456abc123def456abc123def456abc123def456abc1")

	tests := []struct {
		name    string
		chainId *big.Int
		query   ethereum.FilterQuery
	}{
		{
			name:    "Nil chain ID",
			chainId: nil,
			query: ethereum.FilterQuery{
				FromBlock: big.NewInt(100),
				ToBlock:   big.NewInt(200),
				Addresses: []common.Address{address1},
				Topics:    [][]common.Hash{{topic1}},
			},
		},
		{
			name:    "Zero chain ID",
			chainId: big.NewInt(0),
			query: ethereum.FilterQuery{
				FromBlock: big.NewInt(100),
				ToBlock:   big.NewInt(200),
				Addresses: []common.Address{address1},
				Topics:    [][]common.Hash{{topic1}},
			},
		},
		{
			name:    "Very large chain ID",
			chainId: big.NewInt(0).Lsh(big.NewInt(1), 256), // 2^256
			query: ethereum.FilterQuery{
				FromBlock: big.NewInt(100),
				ToBlock:   big.NewInt(200),
				Addresses: []common.Address{address1},
				Topics:    [][]common.Hash{{topic1}},
			},
		},
		{
			name:    "Nil block pointers",
			chainId: chainId,
			query: ethereum.FilterQuery{
				FromBlock: nil,
				ToBlock:   nil,
				Addresses: []common.Address{address1},
				Topics:    [][]common.Hash{{topic1}},
			},
		},
		{
			name:    "Very large block numbers",
			chainId: chainId,
			query: ethereum.FilterQuery{
				FromBlock: big.NewInt(0).Lsh(big.NewInt(1), 64), // 2^64
				ToBlock:   big.NewInt(0).Lsh(big.NewInt(1), 64).Add(big.NewInt(0).Lsh(big.NewInt(1), 64), big.NewInt(1000)),
				Addresses: []common.Address{address1},
				Topics:    [][]common.Hash{{topic1}},
			},
		},
		{
			name:    "Empty query with nil chain ID",
			chainId: nil,
			query:   ethereum.FilterQuery{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := GetQueryHash(tt.chainId, tt.query)

			// Verify hash is not zero (even for edge cases)
			assert.NotEqual(t, common.Hash{}, hash, "Hash should not be zero")

			// Verify hash is consistent
			hash2 := GetQueryHash(tt.chainId, tt.query)
			assert.Equal(t, hash, hash2, "Hash should be consistent for same input")
		})
	}
}

// TestGetQueryKey tests the GetQueryKey function
func TestGetQueryKey(t *testing.T) {
	chainId := big.NewInt(1)
	address1 := common.HexToAddress("0x742d35Cc6634C893292Ce8bB6239C002Ad8e6b59")
	topic1 := common.HexToHash("0xabc123def456abc123def456abc123def456abc123def456abc123def456abc1")

	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(100),
		ToBlock:   big.NewInt(200),
		Addresses: []common.Address{address1},
		Topics:    [][]common.Hash{{topic1}},
	}

	key := GetQueryKey(chainId, query)

	// Verify key is not empty
	assert.NotEmpty(t, key, "Query key should not be empty")

	// Verify key starts with "0x" (hex format)
	assert.True(t, strings.HasPrefix(key, "0x"), "Query key should start with 0x")

	// Verify key is consistent
	key2 := GetQueryKey(chainId, query)
	assert.Equal(t, key, key2, "Query key should be consistent for same input")

	// Verify key is the hex representation of the hash
	hash := GetQueryHash(chainId, query)
	expectedKey := hash.Hex()
	assert.Equal(t, expectedKey, key, "Query key should be the hex representation of the hash")
}
