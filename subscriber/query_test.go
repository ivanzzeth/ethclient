package subscriber

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

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
