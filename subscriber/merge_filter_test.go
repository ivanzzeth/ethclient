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

func TestTopicLayoutKey(t *testing.T) {
	// No topics → all zeros
	q := ethereum.FilterQuery{Addresses: []common.Address{{}}}
	assert.Equal(t, "0000", TopicLayoutKey(q))

	// topic0 set only (e.g. event sig)
	q.Topics = [][]common.Hash{{common.Hash{0x01}}}
	assert.Equal(t, "1000", TopicLayoutKey(q))

	// topic0 + topic1 set (e.g. CT PositionSplit: sig + stakeholder)
	q.Topics = [][]common.Hash{{common.Hash{0x01}}, {common.Hash{0x02}}}
	assert.Equal(t, "1100", TopicLayoutKey(q))

	// topic0 nil, topic1 set (e.g. OrderFilled: sig + orderHash nil + maker at topic2)
	q.Topics = [][]common.Hash{{common.Hash{0x01}}, nil, {common.Hash{0x03}}}
	assert.Equal(t, "1010", TopicLayoutKey(q))

	// Different layouts must get different keys so they are not merged (would lose address filter)
	ctLike := ethereum.FilterQuery{Topics: [][]common.Hash{{common.Hash{1}}, {common.Hash{2}}}} // topic1=stakeholder
	ofLike := ethereum.FilterQuery{Topics: [][]common.Hash{{common.Hash{1}}, nil, {common.Hash{2}}}} // topic2=maker
	assert.NotEqual(t, TopicLayoutKey(ctLike), TopicLayoutKey(ofLike), "CT (topic1=addr) and OrderFilled (topic2=addr) must not merge")
}

// TestPartitionByTopicLayout_PredictAndPolymarket verifies that after partitioning by TopicLayoutKey,
// Predict ends up with 2 merged groups (CT/Adapter topic1=stakeholder vs Exchange topic2=maker) and
// Polymarket with 2 merged groups, and that merging within each group preserves the address filter.
func TestPartitionByTopicLayout_PredictAndPolymarket(t *testing.T) {
	from, to := big.NewInt(1), big.NewInt(100)
	funder := common.HexToHash("0x00000000000000000000000088ed75e9ece373997221e3c0229e74007c1ad718")

	// Build CT/Adapter-like queries: topic0=eventSig, topic1=stakeholder (funder) → layout "1100"
	makeCTLike := func(addr common.Address, topic0 common.Hash) ethereum.FilterQuery {
		return ethereum.FilterQuery{
			Addresses: []common.Address{addr},
			Topics:    [][]common.Hash{{topic0}, {funder}},
		}
	}
	// Build OrderFilled-like: topic0=sig, topic1=nil (orderHash), topic2=maker → layout "1010"
	makeOrderFilledLike := func(addr common.Address, topic0 common.Hash) ethereum.FilterQuery {
		return ethereum.FilterQuery{
			Addresses: []common.Address{addr},
			Topics:    [][]common.Hash{{topic0}, nil, {funder}},
		}
	}

	// Predict: 2 CT contracts × 3 events = 6, NegRiskAdapter 3, YBNegRiskAdapter 3, Exchange 4 = 16 total
	addrCT := common.HexToAddress("0x4d97dcd97ec945f40cf65f87097ace5ea0476045")
	addrAdapter1 := common.HexToAddress("0x0000000000000000000000000000000000000001")
	addrAdapter2 := common.HexToAddress("0x0000000000000000000000000000000000000002")
	addrExchange := common.HexToAddress("0x0000000000000000000000000000000000000003")
	var predictQueries []ethereum.FilterQuery
	sigs := []common.Hash{{0x01}, {0x02}, {0x03}, {0x04}, {0x05}, {0x06}, {0x07}, {0x08}, {0x09}, {0x0a}, {0x0b}, {0x0c}}
	for i := 0; i < 12; i++ {
		addr := addrCT
		if i >= 6 {
			addr = addrAdapter1
		}
		if i >= 9 {
			addr = addrAdapter2
		}
		predictQueries = append(predictQueries, makeCTLike(addr, sigs[i]))
	}
	orderFilledSig := common.Hash{0xd0}
	for i := 0; i < 4; i++ {
		predictQueries = append(predictQueries, makeOrderFilledLike(addrExchange, orderFilledSig))
	}
	require.Len(t, predictQueries, 16, "Predict has 16 subscriptions")

	// Partition by layout (same as chain_subscriber)
	layoutToQueries := make(map[string][]ethereum.FilterQuery)
	for _, q := range predictQueries {
		key := TopicLayoutKey(q)
		layoutToQueries[key] = append(layoutToQueries[key], q)
	}
	assert.Len(t, layoutToQueries, 2, "Predict must split into 2 layout groups (CT/Adapter vs OrderFilled)")
	ctLayoutKey := TopicLayoutKey(predictQueries[0])
	ofLayoutKey := TopicLayoutKey(predictQueries[12])
	assert.Equal(t, "1100", ctLayoutKey)
	assert.Equal(t, "1010", ofLayoutKey)
	require.Len(t, layoutToQueries[ctLayoutKey], 12, "CT/Adapter group has 12 queries")
	require.Len(t, layoutToQueries[ofLayoutKey], 4, "OrderFilled group has 4 queries")

	// Merge each group and assert address filter preserved
	mergedCT, err := MergeFilterQueries(layoutToQueries[ctLayoutKey], from, to)
	require.NoError(t, err)
	require.Len(t, mergedCT.Topics, 2)
	assert.NotNil(t, mergedCT.Topics[1], "CT merged must keep topic1=stakeholder filter")
	assert.Len(t, mergedCT.Topics[1], 1)
	assert.Equal(t, funder, mergedCT.Topics[1][0])

	mergedOF, err := MergeFilterQueries(layoutToQueries[ofLayoutKey], from, to)
	require.NoError(t, err)
	require.Len(t, mergedOF.Topics, 3)
	assert.NotNil(t, mergedOF.Topics[2], "OrderFilled merged must keep topic2=maker filter")
	assert.Len(t, mergedOF.Topics[2], 1)
	assert.Equal(t, funder, mergedOF.Topics[2][0])

	// Polymarket: 1 CT × 3 events + 1 Exchange × OrderFilled = 4 total
	var polyQueries []ethereum.FilterQuery
	for i := 0; i < 3; i++ {
		polyQueries = append(polyQueries, makeCTLike(addrCT, sigs[i]))
	}
	polyQueries = append(polyQueries, makeOrderFilledLike(addrExchange, orderFilledSig))
	require.Len(t, polyQueries, 4, "Polymarket has 4 subscriptions")

	layoutToQueriesPoly := make(map[string][]ethereum.FilterQuery)
	for _, q := range polyQueries {
		key := TopicLayoutKey(q)
		layoutToQueriesPoly[key] = append(layoutToQueriesPoly[key], q)
	}
	assert.Len(t, layoutToQueriesPoly, 2, "Polymarket must split into 2 layout groups")
	require.Len(t, layoutToQueriesPoly[ctLayoutKey], 3, "Polymarket CT group has 3 queries")
	require.Len(t, layoutToQueriesPoly[ofLayoutKey], 1, "Polymarket OrderFilled group has 1 query")

	mergedPolyCT, err := MergeFilterQueries(layoutToQueriesPoly[ctLayoutKey], from, to)
	require.NoError(t, err)
	assert.NotNil(t, mergedPolyCT.Topics[1], "Polymarket CT merged keeps topic1")
	mergedPolyOF, err := MergeFilterQueries(layoutToQueriesPoly[ofLayoutKey], from, to)
	require.NoError(t, err)
	assert.NotNil(t, mergedPolyOF.Topics[2], "Polymarket OrderFilled merged keeps topic2")
}

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

// TestMergeFilterQueries_EmptyTopicUnionBecomesNil: when all queries have Topics[i] non-nil but
// the union is empty (e.g. all Topics[i]==[]), merged must get topics[i]=nil to avoid filtering
// out all logs (missed dispatch).
func TestMergeFilterQueries_EmptyTopicUnionBecomesNil(t *testing.T) {
	from, to := big.NewInt(1), big.NewInt(10)
	t0 := common.HexToHash("0x00")
	emptySlice := []common.Hash{} // non-nil, len 0
	queries := []ethereum.FilterQuery{
		{Addresses: []common.Address{{}}, Topics: [][]common.Hash{{t0}, emptySlice}},
		{Addresses: []common.Address{{}}, Topics: [][]common.Hash{{t0}, emptySlice}},
	}
	merged, err := MergeFilterQueries(queries, from, to)
	require.NoError(t, err)
	require.Len(t, merged.Topics, 2)
	assert.NotNil(t, merged.Topics[0])
	assert.Nil(t, merged.Topics[1], "empty union at index 1 must become nil so eth_getLogs returns logs")
}

func TestLogMatchesQuery_EmptyTopicSliceMeansAny(t *testing.T) {
	t0 := common.HexToHash("0x111")
	t1 := common.HexToHash("0x222")
	// Query has 4 topic slots; index 2 is empty slice (bind sometimes emits [] instead of nil).
	q := ethereum.FilterQuery{
		Addresses: []common.Address{{}},
		Topics:    [][]common.Hash{{t0}, {t1}, {}, {}},
	}
	logWithFourTopics := etypes.Log{Address: common.Address{}, Topics: []common.Hash{t0, t1, common.Hash{0x33}, common.Hash{0x44}}}
	assert.True(t, LogMatchesQuery(&logWithFourTopics, q), "empty filterTopics must mean any (avoid missed dispatch)")
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

func TestMergeFilterQueriesByBlockHash_AnyAddress(t *testing.T) {
	h := common.Hash{0x02}
	queries := []ethereum.FilterQuery{
		{BlockHash: &h, Addresses: nil},
		{BlockHash: &h, Addresses: []common.Address{}},
	}
	merged, err := MergeFilterQueriesByBlockHash(queries, h)
	require.NoError(t, err)
	assert.Nil(t, merged.Addresses, "any address in input => merged addresses nil")
	assert.Equal(t, h, *merged.BlockHash)
}

func TestMergeFilterQueriesByBlockHash_TopicsWithNil(t *testing.T) {
	h := common.Hash{0x03}
	t0 := common.HexToHash("0x111")
	queries := []ethereum.FilterQuery{
		{BlockHash: &h, Addresses: []common.Address{{}}, Topics: [][]common.Hash{{t0}, {t0}}},
		{BlockHash: &h, Addresses: []common.Address{{}}, Topics: [][]common.Hash{{t0}, nil}},
	}
	merged, err := MergeFilterQueriesByBlockHash(queries, h)
	require.NoError(t, err)
	require.Len(t, merged.Topics, 2)
	assert.NotNil(t, merged.Topics[0])
	assert.Nil(t, merged.Topics[1], "nil at index 1 => merged topic nil")
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
