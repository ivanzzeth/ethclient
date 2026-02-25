package subscriber

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeFilterQueries_Empty(t *testing.T) {
	from, to := big.NewInt(1), big.NewInt(10)
	merged, err := MergeFilterQueries(nil, from, to)
	require.NoError(t, err)
	assert.Nil(t, merged.Addresses)
	assert.Nil(t, merged.Topics)
	// Empty input returns zero FilterQuery; caller should not use for FilterLogs.

	merged, err = MergeFilterQueries([]ethereum.FilterQuery{}, from, to)
	require.NoError(t, err)
	assert.Nil(t, merged.Addresses)
}

func TestMergeFilterQueries_RequiresBlockRange(t *testing.T) {
	_, err := MergeFilterQueries([]ethereum.FilterQuery{{Addresses: []common.Address{{}}}}, nil, big.NewInt(1))
	assert.ErrorIs(t, err, errMergeBlockRange)
	_, err = MergeFilterQueries([]ethereum.FilterQuery{{Addresses: []common.Address{{}}}}, big.NewInt(1), nil)
	assert.ErrorIs(t, err, errMergeBlockRange)
}

func TestMergeFilterQueries_RejectsBlockHash(t *testing.T) {
	h := common.Hash{0x01}
	_, err := MergeFilterQueries([]ethereum.FilterQuery{{BlockHash: &h}}, big.NewInt(1), big.NewInt(10))
	assert.ErrorIs(t, err, errMergeBlockHash)
}

func TestMergeFilterQueries_SingleQuery(t *testing.T) {
	from, to := big.NewInt(1), big.NewInt(10)
	addr := common.HexToAddress("0xabc")
	topic0 := common.HexToHash("0xdef")
	q := ethereum.FilterQuery{
		FromBlock: from, ToBlock: to,
		Addresses: []common.Address{addr},
		Topics:    [][]common.Hash{{topic0}},
	}
	merged, err := MergeFilterQueries([]ethereum.FilterQuery{q}, from, to)
	require.NoError(t, err)
	assert.Len(t, merged.Addresses, 1)
	assert.Equal(t, addr, merged.Addresses[0])
	require.Len(t, merged.Topics, 1)
	assert.Len(t, merged.Topics[0], 1)
	assert.Equal(t, topic0, merged.Topics[0][0])
}

func TestMergeFilterQueries_UnionAddressesAndTopics(t *testing.T) {
	from, to := big.NewInt(1), big.NewInt(10)
	a1 := common.HexToAddress("0xaaa")
	a2 := common.HexToAddress("0xbbb")
	t1 := common.HexToHash("0x111")
	t2 := common.HexToHash("0x222")
	queries := []ethereum.FilterQuery{
		{Addresses: []common.Address{a1}, Topics: [][]common.Hash{{t1}}},
		{Addresses: []common.Address{a2}, Topics: [][]common.Hash{{t2}}},
		{Addresses: []common.Address{a1}, Topics: [][]common.Hash{{t1}}}, // duplicate
	}
	merged, err := MergeFilterQueries(queries, from, to)
	require.NoError(t, err)
	require.Len(t, merged.Addresses, 2)
	assert.Contains(t, merged.Addresses, a1)
	assert.Contains(t, merged.Addresses, a2)
	require.Len(t, merged.Topics, 1)
	assert.Len(t, merged.Topics[0], 2)
	assert.Contains(t, merged.Topics[0], t1)
	assert.Contains(t, merged.Topics[0], t2)
}

func TestMergeFilterQueries_AnyAddressMakesMergedNilAddresses(t *testing.T) {
	from, to := big.NewInt(1), big.NewInt(10)
	addr := common.HexToAddress("0xabc")
	queries := []ethereum.FilterQuery{
		{Addresses: []common.Address{addr}, Topics: [][]common.Hash{{common.Hash{}}}},
		{Addresses: nil, Topics: [][]common.Hash{{common.Hash{}}}}, // any address
	}
	merged, err := MergeFilterQueries(queries, from, to)
	require.NoError(t, err)
	// Merged must match any address so we don't miss logs from the second query.
	assert.Nil(t, merged.Addresses)
}

func TestMergeFilterQueries_AnyTopicAtIndexMakesMergedNilAtThatIndex(t *testing.T) {
	from, to := big.NewInt(1), big.NewInt(10)
	t0 := common.HexToHash("0x00")
	queries := []ethereum.FilterQuery{
		{Addresses: []common.Address{{}}, Topics: [][]common.Hash{{t0}, {common.Hash{0x01}}}},
		{Addresses: []common.Address{{}}, Topics: [][]common.Hash{{t0}, nil}}, // any at index 1
	}
	merged, err := MergeFilterQueries(queries, from, to)
	require.NoError(t, err)
	require.Len(t, merged.Topics, 2)
	assert.NotNil(t, merged.Topics[0])
	assert.Nil(t, merged.Topics[1])
}

func TestLogMatchesQuery_AddressFilter(t *testing.T) {
	addr1 := common.HexToAddress("0xaaa")
	addr2 := common.HexToAddress("0xbbb")
	logAddr1 := etypes.Log{Address: addr1, Topics: []common.Hash{{}}}
	logAddr2 := etypes.Log{Address: addr2, Topics: []common.Hash{{}}}
	q := ethereum.FilterQuery{Addresses: []common.Address{addr1}}

	assert.True(t, LogMatchesQuery(&logAddr1, q))
	assert.False(t, LogMatchesQuery(&logAddr2, q))
}

func TestLogMatchesQuery_AnyAddress(t *testing.T) {
	q := ethereum.FilterQuery{Addresses: nil}
	log := etypes.Log{Address: common.HexToAddress("0xccc"), Topics: []common.Hash{}}
	assert.True(t, LogMatchesQuery(&log, q))

	q2 := ethereum.FilterQuery{Addresses: []common.Address{}}
	assert.True(t, LogMatchesQuery(&log, q2))
}

func TestLogMatchesQuery_TopicFilter(t *testing.T) {
	t0 := common.HexToHash("0x111")
	t1 := common.HexToHash("0x222")
	q := ethereum.FilterQuery{
		Addresses: []common.Address{{}},
		Topics:    [][]common.Hash{{t0}, {t1}},
	}
	logMatch := etypes.Log{Address: common.Address{}, Topics: []common.Hash{t0, t1}}
	logWrongTopic0 := etypes.Log{Address: common.Address{}, Topics: []common.Hash{common.Hash{0x99}, t1}}
	logWrongTopic1 := etypes.Log{Address: common.Address{}, Topics: []common.Hash{t0, common.Hash{0x99}}}
	logTooFewTopics := etypes.Log{Address: common.Address{}, Topics: []common.Hash{t0}}

	assert.True(t, LogMatchesQuery(&logMatch, q))
	assert.False(t, LogMatchesQuery(&logWrongTopic0, q))
	assert.False(t, LogMatchesQuery(&logWrongTopic1, q))
	assert.False(t, LogMatchesQuery(&logTooFewTopics, q))
}

func TestLogMatchesQuery_TopicNilMeansAny(t *testing.T) {
	t0 := common.HexToHash("0x111")
	q := ethereum.FilterQuery{
		Addresses: []common.Address{{}},
		Topics:    [][]common.Hash{{t0}, nil}, // any at index 1
	}
	logAnySecond := etypes.Log{Address: common.Address{}, Topics: []common.Hash{t0, common.Hash{0x99}}}
	assert.True(t, LogMatchesQuery(&logAnySecond, q))
}

// TestMergeFilterQueries_ThenLogMatchesQuery_NoExtraNoMissing simulates: run N queries individually
// and run one merged query + dispatch by LogMatchesQuery; the set of logs per original query must match.
func TestMergeFilterQueries_ThenLogMatchesQuery_NoExtraNoMissing(t *testing.T) {
	from, to := big.NewInt(1), big.NewInt(2)
	addrA := common.HexToAddress("0xaaa")
	addrB := common.HexToAddress("0xbbb")
	tA := common.HexToHash("0xa")
	tB := common.HexToHash("0xb")
	queries := []ethereum.FilterQuery{
		{Addresses: []common.Address{addrA}, Topics: [][]common.Hash{{tA}}},
		{Addresses: []common.Address{addrB}, Topics: [][]common.Hash{{tB}}},
		{Addresses: []common.Address{addrA, addrB}, Topics: [][]common.Hash{{tA}}},
	}
	_, err := MergeFilterQueries(queries, from, to)
	require.NoError(t, err)

	// Simulated logs that would be returned by merged eth_getLogs (union of all).
	logs := []etypes.Log{
		{Address: addrA, Topics: []common.Hash{tA}},
		{Address: addrB, Topics: []common.Hash{tB}},
		{Address: addrA, Topics: []common.Hash{tA}}, // same as first, e.g. different block
		{Address: addrB, Topics: []common.Hash{tA}},  // matches query 3 only
	}
	for i, q := range queries {
		var matched []etypes.Log
		for j := range logs {
			if LogMatchesQuery(&logs[j], q) {
				matched = append(matched, logs[j])
			}
		}
		// Query 0: addrA + tA → 2 logs. Query 1: addrB + tB → 1 log. Query 2: addrA or addrB + tA → 3 logs.
		switch i {
		case 0:
			assert.Len(t, matched, 2, "query 0 must get exactly logs matching addrA and tA")
		case 1:
			assert.Len(t, matched, 1, "query 1 must get exactly logs matching addrB and tB")
		case 2:
			assert.Len(t, matched, 3, "query 2 must get logs matching (addrA|addrB) and tA")
		}
	}
}

func TestMergeFilterQueriesByBlockHash_Empty(t *testing.T) {
	h := common.Hash{0x01}
	out, err := MergeFilterQueriesByBlockHash(nil, h)
	require.NoError(t, err)
	// Empty queries returns zero FilterQuery (caller should not use for FilterLogs).
	assert.Nil(t, out.BlockHash)
}

func TestMergeFilterQueriesByBlockHash_Mismatch(t *testing.T) {
	h1 := common.Hash{0x01}
	h2 := common.Hash{0x02}
	queries := []ethereum.FilterQuery{
		{BlockHash: &h1, Addresses: []common.Address{{}}},
		{BlockHash: &h2, Addresses: []common.Address{{}}},
	}
	_, err := MergeFilterQueriesByBlockHash(queries, h1)
	assert.ErrorIs(t, err, errMergeBlockHashMismatch)
}

func TestMergeFilterQueriesByBlockHash_Union(t *testing.T) {
	h := common.Hash{0x01}
	a1 := common.HexToAddress("0xaaa")
	a2 := common.HexToAddress("0xbbb")
	queries := []ethereum.FilterQuery{
		{BlockHash: &h, Addresses: []common.Address{a1}},
		{BlockHash: &h, Addresses: []common.Address{a2}},
	}
	merged, err := MergeFilterQueriesByBlockHash(queries, h)
	require.NoError(t, err)
	require.Len(t, merged.Addresses, 2)
	assert.Contains(t, merged.Addresses, a1)
	assert.Contains(t, merged.Addresses, a2)
	assert.Equal(t, h, *merged.BlockHash)
}

// --- Partition key and partition-by-merge-key tests (TDD: edge cases first) ---

func TestGetPartitionKey_SameRangeSameKey(t *testing.T) {
	from, to := big.NewInt(1), big.NewInt(10)
	q := ethereum.FilterQuery{FromBlock: from, ToBlock: to}
	k := GetPartitionKey(q, nil, nil)
	assert.Equal(t, "R:1:10", k)
	assert.Equal(t, GetPartitionKey(q, big.NewInt(0), big.NewInt(5)), "R:1:10", "fallback ignored when query has range")
}

func TestGetPartitionKey_DifferentRangeDifferentKey(t *testing.T) {
	q1 := ethereum.FilterQuery{FromBlock: big.NewInt(1), ToBlock: big.NewInt(10)}
	q2 := ethereum.FilterQuery{FromBlock: big.NewInt(5), ToBlock: big.NewInt(15)}
	assert.NotEqual(t, GetPartitionKey(q1, nil, nil), GetPartitionKey(q2, nil, nil))
}

func TestGetPartitionKey_BlockHashKey(t *testing.T) {
	h := common.Hash{0x01}
	q := ethereum.FilterQuery{BlockHash: &h}
	k := GetPartitionKey(q, nil, nil)
	assert.Equal(t, "H:"+h.Hex(), k)
}

func TestGetPartitionKey_RangeVsBlockHashDifferentKey(t *testing.T) {
	h := common.Hash{0x01}
	qH := ethereum.FilterQuery{BlockHash: &h}
	qR := ethereum.FilterQuery{FromBlock: big.NewInt(1), ToBlock: big.NewInt(10)}
	assert.NotEqual(t, GetPartitionKey(qH, nil, nil), GetPartitionKey(qR, nil, nil))
}

func TestGetPartitionKey_NilRangeUsesFallback(t *testing.T) {
	q := ethereum.FilterQuery{FromBlock: nil, ToBlock: nil}
	k := GetPartitionKey(q, big.NewInt(2), big.NewInt(20))
	assert.Equal(t, "R:2:20", k)
}

func TestGetPartitionKey_NilRangeNoFallbackUsesNilLiteral(t *testing.T) {
	q := ethereum.FilterQuery{FromBlock: nil, ToBlock: nil}
	k := GetPartitionKey(q, nil, nil)
	assert.Equal(t, "R:nil:nil", k)
}

func TestPartitionQueriesByMergeKey_Empty(t *testing.T) {
	groups := PartitionQueriesByMergeKey(nil, nil, nil)
	assert.Len(t, groups, 0)
	groups = PartitionQueriesByMergeKey([]ethereum.FilterQuery{}, nil, nil)
	assert.Len(t, groups, 0)
}

func TestPartitionQueriesByMergeKey_SameRangeOneGroup(t *testing.T) {
	from, to := big.NewInt(1), big.NewInt(10)
	queries := []ethereum.FilterQuery{
		{FromBlock: from, ToBlock: to, Addresses: []common.Address{{0x01}}},
		{FromBlock: from, ToBlock: to, Addresses: []common.Address{{0x02}}},
		{FromBlock: from, ToBlock: to, Addresses: []common.Address{{0x03}}},
	}
	groups := PartitionQueriesByMergeKey(queries, from, to)
	require.Len(t, groups, 1)
	assert.Len(t, groups[0], 3)
}

func TestPartitionQueriesByMergeKey_DifferentRangesTwoGroups(t *testing.T) {
	queries := []ethereum.FilterQuery{
		{FromBlock: big.NewInt(1), ToBlock: big.NewInt(10), Addresses: []common.Address{{0x01}}},
		{FromBlock: big.NewInt(1), ToBlock: big.NewInt(10), Addresses: []common.Address{{0x02}}},
		{FromBlock: big.NewInt(5), ToBlock: big.NewInt(15), Addresses: []common.Address{{0x03}}},
	}
	groups := PartitionQueriesByMergeKey(queries, nil, nil)
	require.Len(t, groups, 2)
	// Order of groups undefined; find the size-2 and size-1.
	var size2, size1 []ethereum.FilterQuery
	for _, g := range groups {
		if len(g) == 2 {
			size2 = g
		} else {
			size1 = g
		}
	}
	require.Len(t, size2, 2)
	require.Len(t, size1, 1)
	assert.Equal(t, uint64(10), size2[0].ToBlock.Uint64())
	assert.Equal(t, uint64(15), size1[0].ToBlock.Uint64())
}

func TestPartitionQueriesByMergeKey_ThreeRangesThreeGroups(t *testing.T) {
	queries := []ethereum.FilterQuery{
		{FromBlock: big.NewInt(1), ToBlock: big.NewInt(10)},
		{FromBlock: big.NewInt(5), ToBlock: big.NewInt(15)},
		{FromBlock: big.NewInt(20), ToBlock: big.NewInt(30)},
	}
	groups := PartitionQueriesByMergeKey(queries, nil, nil)
	require.Len(t, groups, 3)
	for _, g := range groups {
		require.Len(t, g, 1)
	}
}

func TestPartitionQueriesByMergeKey_SameBlockHashOneGroup(t *testing.T) {
	h := common.Hash{0x01}
	queries := []ethereum.FilterQuery{
		{BlockHash: &h, Addresses: []common.Address{{0x01}}},
		{BlockHash: &h, Addresses: []common.Address{{0x02}}},
	}
	groups := PartitionQueriesByMergeKey(queries, nil, nil)
	require.Len(t, groups, 1)
	assert.Len(t, groups[0], 2)
}

func TestPartitionQueriesByMergeKey_DifferentBlockHashTwoGroups(t *testing.T) {
	h1, h2 := common.Hash{0x01}, common.Hash{0x02}
	queries := []ethereum.FilterQuery{
		{BlockHash: &h1}, {BlockHash: &h1}, {BlockHash: &h2},
	}
	groups := PartitionQueriesByMergeKey(queries, nil, nil)
	require.Len(t, groups, 2)
	var g2, g1 []ethereum.FilterQuery
	for _, g := range groups {
		if len(g) == 2 {
			g2 = g
		} else {
			g1 = g
		}
	}
	require.Len(t, g2, 2)
	require.Len(t, g1, 1)
}

func TestPartitionQueriesByMergeKey_MixRangeAndBlockHash(t *testing.T) {
	h := common.Hash{0x01}
	queries := []ethereum.FilterQuery{
		{FromBlock: big.NewInt(1), ToBlock: big.NewInt(10)},
		{BlockHash: &h},
	}
	groups := PartitionQueriesByMergeKey(queries, nil, nil)
	require.Len(t, groups, 2)
	assert.Len(t, groups[0], 1)
	assert.Len(t, groups[1], 1)
}

func TestPartitionQueriesByMergeKey_NilRangeUsesFallback(t *testing.T) {
	fallbackFrom, fallbackTo := big.NewInt(1), big.NewInt(10)
	queries := []ethereum.FilterQuery{
		{FromBlock: nil, ToBlock: nil, Addresses: []common.Address{{0x01}}},
		{FromBlock: nil, ToBlock: nil, Addresses: []common.Address{{0x02}}},
	}
	groups := PartitionQueriesByMergeKey(queries, fallbackFrom, fallbackTo)
	require.Len(t, groups, 1)
	assert.Len(t, groups[0], 2)
}

func TestPartitionQueriesByMergeKey_NilRangeNoFallbackEachOwnPartition(t *testing.T) {
	queries := []ethereum.FilterQuery{
		{FromBlock: nil, ToBlock: nil},
		{FromBlock: nil, ToBlock: nil},
	}
	groups := PartitionQueriesByMergeKey(queries, nil, nil)
	// Key becomes "R:nil:nil" for both → same partition, 1 group of 2.
	require.Len(t, groups, 1)
	assert.Len(t, groups[0], 2)
}
