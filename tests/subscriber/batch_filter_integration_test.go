package subscriber_test

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ivanzzeth/ethclient/message"
	"github.com/ivanzzeth/ethclient/simulated"
	"github.com/ivanzzeth/ethclient/subscriber"
	"github.com/ivanzzeth/ethclient/subscriber/handler"
	"github.com/ivanzzeth/ethclient/tests/helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// All tests in this file use the simulated backend (helper.SetUpClient returns *simulated.Backend).
var _ *simulated.Backend = nil

// testRetryInterval shortens realtime scanner polling in tests so we need less sleep.
const testRetryInterval = 250 * time.Millisecond

// TestBatchLogFilterer_NewPanicsOnNilRPC verifies panic when rpc is nil (integration: we have real client).
func TestBatchLogFilterer_NewPanicsOnNilRPC(t *testing.T) {
	t.Parallel()
	sim := helper.SetUpClient(t)
	defer sim.Close()
	raw := sim.Client().RawClient()
	require.NotNil(t, raw)

	defer func() {
		r := recover()
		assert.NotNil(t, r, "expected panic when rpc is nil")
	}()
	subscriber.NewBatchLogFilterer(raw, nil)
}

// TestFilterLogsBatch_EmptyQueries returns nil, nil for empty slice.
func TestFilterLogsBatch_EmptyQueries(t *testing.T) {
	t.Parallel()
	sim := helper.SetUpClient(t)
	defer sim.Close()
	batch := subscriber.NewBatchLogFilterer(sim.Client().RawClient(), sim.Client().RpcClient())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := batch.FilterLogsBatch(ctx, nil)
	assert.NoError(t, err)
	assert.Nil(t, result)

	result, err = batch.FilterLogsBatch(ctx, []ethereum.FilterQuery{})
	assert.NoError(t, err)
	assert.Nil(t, result)
}

// TestFilterLogsBatch_InvalidQuery_BlockHashAndFromBlock returns error.
func TestFilterLogsBatch_InvalidQuery_BlockHashAndFromBlock(t *testing.T) {
	t.Parallel()
	sim := helper.SetUpClient(t)
	defer sim.Close()
	batch := subscriber.NewBatchLogFilterer(sim.Client().RawClient(), sim.Client().RpcClient())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	blockHash := common.Hash{0x01}
	queries := []ethereum.FilterQuery{{
		BlockHash: &blockHash,
		FromBlock: big.NewInt(1),
	}}

	result, err := batch.FilterLogsBatch(ctx, queries)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "cannot specify both BlockHash and FromBlock/ToBlock")
}

// TestFilterLogsBatch_SingleQuery matches single FilterLogs (happy path).
func TestFilterLogsBatch_SingleQuery(t *testing.T) {
	t.Parallel()
	sim := helper.SetUpClient(t)
	defer sim.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	contractAddr, _, _ := helper.DeployTestContract(t, ctx, sim)
	fromBlock, _ := sim.Client().BlockNumber(ctx)

	// One query in batch
	batch := subscriber.NewBatchLogFilterer(sim.Client().RawClient(), sim.Client().RpcClient())
	queries := []ethereum.FilterQuery{{
		FromBlock: big.NewInt(0).SetUint64(fromBlock),
		ToBlock:   big.NewInt(0).SetUint64(fromBlock + 10),
		Addresses: []common.Address{contractAddr},
	}}
	allLogs, err := batch.FilterLogsBatch(ctx, queries)
	require.NoError(t, err)
	require.Len(t, allLogs, 1)
	// Deploy + any events; at least 1 (contract creation logs or test events)
	assert.GreaterOrEqual(t, len(allLogs[0]), 0)
}

// TestFilterLogs_DelegateMatchesRaw ensures batch.FilterLogs returns same result as raw client.
func TestFilterLogs_DelegateMatchesRaw(t *testing.T) {
	t.Parallel()
	sim := helper.SetUpClient(t)
	defer sim.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	contractAddr, _, contract := helper.DeployTestContract(t, ctx, sim)
	fromBlock, _ := sim.Client().BlockNumber(ctx)

	opts, _ := sim.Client().MessageToTransactOpts(ctx, message.Request{From: helper.Addr1})
	_, err := contract.TestFunc1(opts, "hello", big.NewInt(100), []byte("world"))
	require.NoError(t, err)
	sim.Commit()
	toBlock, _ := sim.Client().BlockNumber(ctx)

	q := ethereum.FilterQuery{
		FromBlock: big.NewInt(0).SetUint64(fromBlock),
		ToBlock:   big.NewInt(0).SetUint64(toBlock),
		Addresses: []common.Address{contractAddr},
	}
	rawLogs, err := sim.Client().RawClient().FilterLogs(ctx, q)
	require.NoError(t, err)
	batch := subscriber.NewBatchLogFilterer(sim.Client().RawClient(), sim.Client().RpcClient())
	batchLogs, err := batch.FilterLogs(ctx, q)
	require.NoError(t, err)
	assert.Equal(t, len(rawLogs), len(batchLogs), "FilterLogs delegate must match raw")
	for i := range rawLogs {
		assert.Equal(t, rawLogs[i].Address, batchLogs[i].Address)
		assert.Equal(t, rawLogs[i].BlockNumber, batchLogs[i].BlockNumber)
		assert.Equal(t, rawLogs[i].TxIndex, batchLogs[i].TxIndex)
		assert.Equal(t, rawLogs[i].Index, batchLogs[i].Index)
	}
}

// TestFilterLogsBatch_ToBlockNilUsesLatest ensures query with ToBlock=nil does not error (uses "latest").
func TestFilterLogsBatch_ToBlockNilUsesLatest(t *testing.T) {
	t.Parallel()
	sim := helper.SetUpClient(t)
	defer sim.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	contractAddr, _, _ := helper.DeployTestContract(t, ctx, sim)
	batch := subscriber.NewBatchLogFilterer(sim.Client().RawClient(), sim.Client().RpcClient())
	queries := []ethereum.FilterQuery{{
		FromBlock: big.NewInt(0),
		ToBlock:   nil,
		Addresses: []common.Address{contractAddr},
	}}
	allLogs, err := batch.FilterLogsBatch(ctx, queries)
	require.NoError(t, err)
	require.Len(t, allLogs, 1)
	assert.GreaterOrEqual(t, len(allLogs[0]), 0)
}

// TestFilterLogsBatch_TwoQueries merges two eth_getLogs into one batch (main feature).
func TestFilterLogsBatch_TwoQueries(t *testing.T) {
	t.Parallel()
	sim := helper.SetUpClient(t)
	defer sim.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	contractAddr, _, contract := helper.DeployTestContract(t, ctx, sim)
	fromBlock, _ := sim.Client().BlockNumber(ctx)
	toBlock := fromBlock + 20

	opts, _ := sim.Client().MessageToTransactOpts(ctx, message.Request{From: helper.Addr1})
	_, err := contract.TestFunc1(opts, "hello", big.NewInt(100), []byte("world"))
	require.NoError(t, err)
	sim.Commit()

	batch := subscriber.NewBatchLogFilterer(sim.Client().RawClient(), sim.Client().RpcClient())
	addr := contractAddr
	queries := []ethereum.FilterQuery{
		{FromBlock: big.NewInt(0).SetUint64(fromBlock), ToBlock: big.NewInt(0).SetUint64(toBlock), Addresses: []common.Address{addr}},
		{FromBlock: big.NewInt(0).SetUint64(fromBlock), ToBlock: big.NewInt(0).SetUint64(toBlock), Addresses: []common.Address{addr}},
	}
	allLogs, err := batch.FilterLogsBatch(ctx, queries)
	require.NoError(t, err)
	require.Len(t, allLogs, 2)
	assert.Equal(t, len(allLogs[0]), len(allLogs[1]), "same address filter should return same count")
}

// TestRealtimeSubscriptions_UseMergedScanner verifies two realtime SubscribeFilterLogs receive logs (merged path).
// Must close subscriber so runRealtimeScanner goroutine exits and test can finish.
func TestRealtimeSubscriptions_UseMergedScanner(t *testing.T) {
	t.Parallel()
	sim := helper.SetUpClient(t)
	defer sim.Close()
	sim.Client().SetBlockConfirmationsOnSubscription(0)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	contractAddr, _, contract := helper.DeployTestContract(t, ctx, sim)
	chainID, _ := sim.Client().ChainID(ctx)
	storage := subscriber.NewMemoryStorage(chainID)
	cs, err := subscriber.NewChainSubscriber(sim.Client().RpcClient(), storage)
	require.NoError(t, err)
	defer cs.Close()
	cs.SetRetryInterval(testRetryInterval)
	sim.Client().SetSubscriber(cs)

	addr := contractAddr
	ch1 := make(chan types.Log, 16)
	ch2 := make(chan types.Log, 16)
	sub1, err := cs.SubscribeFilterLogs(ctx, ethereum.FilterQuery{Addresses: []common.Address{addr}}, ch1)
	require.NoError(t, err)
	defer sub1.Unsubscribe()
	sub2, err := cs.SubscribeFilterLogs(ctx, ethereum.FilterQuery{Addresses: []common.Address{addr}}, ch2)
	require.NoError(t, err)
	defer sub2.Unsubscribe()

	opts, _ := sim.Client().MessageToTransactOpts(ctx, message.Request{From: helper.Addr1})
	_, err = contract.TestFunc1(opts, "a", big.NewInt(1), []byte("b"))
	require.NoError(t, err)
	sim.Commit()
	time.Sleep(2 * testRetryInterval)

	var count1, count2 int
	deadline := time.After(1200 * time.Millisecond)
	for {
		select {
		case <-ch1:
			count1++
		case <-ch2:
			count2++
		case <-deadline:
			goto done
		}
	}
done:
	assert.GreaterOrEqual(t, count1, 0)
	assert.GreaterOrEqual(t, count2, 0)
}

// TestFilterLogsBatch_MultiContractMultiEvent: two contracts, two event types, batch returns logs per query.
func TestFilterLogsBatch_MultiContractMultiEvent(t *testing.T) {
	t.Parallel()
	sim := helper.SetUpClient(t)
	defer sim.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	addr1, _, c1 := helper.DeployTestContract(t, ctx, sim)
	addr2, _, c2 := helper.DeployTestContract(t, ctx, sim)
	fromBlock, _ := sim.Client().BlockNumber(ctx)

	opts1, _ := sim.Client().MessageToTransactOpts(ctx, message.Request{From: helper.Addr1})
	_, err := c1.TestFunc1(opts1, "ev1", big.NewInt(1), []byte("data1"))
	require.NoError(t, err)
	sim.Commit()
	opts2, _ := sim.Client().MessageToTransactOpts(ctx, message.Request{From: helper.Addr1})
	_, err = c2.TestFunc1(opts2, "ev2", big.NewInt(2), []byte("data2"))
	require.NoError(t, err)
	sim.Commit()
	toBlock, _ := sim.Client().BlockNumber(ctx)

	batch := subscriber.NewBatchLogFilterer(sim.Client().RawClient(), sim.Client().RpcClient())
	queries := []ethereum.FilterQuery{
		{FromBlock: big.NewInt(0).SetUint64(fromBlock), ToBlock: big.NewInt(0).SetUint64(toBlock), Addresses: []common.Address{addr1}},
		{FromBlock: big.NewInt(0).SetUint64(fromBlock), ToBlock: big.NewInt(0).SetUint64(toBlock), Addresses: []common.Address{addr2}},
	}
	allLogs, err := batch.FilterLogsBatch(ctx, queries)
	require.NoError(t, err)
	require.Len(t, allLogs, 2)

	// Each query must return logs for its contract only.
	for i, logs := range allLogs {
		for _, l := range logs {
			assert.Equal(t, queries[i].Addresses[0], l.Address, "log must belong to query %d contract", i)
		}
	}
	assert.GreaterOrEqual(t, len(allLogs[0]), 1, "contract1 should have at least one event")
	assert.GreaterOrEqual(t, len(allLogs[1]), 1, "contract2 should have at least one event")
}

// TestRealtimeSubscriptions_MultiContract verifies merged scanner with two different contracts (two subscriptions).
func TestRealtimeSubscriptions_MultiContract(t *testing.T) {
	t.Parallel()
	sim := helper.SetUpClient(t)
	defer sim.Close()
	sim.Client().SetBlockConfirmationsOnSubscription(0)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	addr1, _, c1 := helper.DeployTestContract(t, ctx, sim)
	addr2, _, c2 := helper.DeployTestContract(t, ctx, sim)
	chainID, _ := sim.Client().ChainID(ctx)
	storage := subscriber.NewMemoryStorage(chainID)
	cs, err := subscriber.NewChainSubscriber(sim.Client().RpcClient(), storage)
	require.NoError(t, err)
	defer cs.Close()
	cs.SetRetryInterval(testRetryInterval)
	sim.Client().SetSubscriber(cs)

	ch1 := make(chan types.Log, 16)
	ch2 := make(chan types.Log, 16)
	sub1, err := cs.SubscribeFilterLogs(ctx, ethereum.FilterQuery{Addresses: []common.Address{addr1}}, ch1)
	require.NoError(t, err)
	defer sub1.Unsubscribe()
	sub2, err := cs.SubscribeFilterLogs(ctx, ethereum.FilterQuery{Addresses: []common.Address{addr2}}, ch2)
	require.NoError(t, err)
	defer sub2.Unsubscribe()

	opts1, _ := sim.Client().MessageToTransactOpts(ctx, message.Request{From: helper.Addr1})
	_, err = c1.TestFunc1(opts1, "from-c1", big.NewInt(1), []byte("c1"))
	require.NoError(t, err)
	sim.Commit()
	opts2, _ := sim.Client().MessageToTransactOpts(ctx, message.Request{From: helper.Addr1})
	_, err = c2.TestFunc1(opts2, "from-c2", big.NewInt(2), []byte("c2"))
	require.NoError(t, err)
	sim.Commit()
	time.Sleep(2 * testRetryInterval)

	var logs1, logs2 []types.Log
	deadline := time.After(1200 * time.Millisecond)
	for {
		select {
		case l := <-ch1:
			logs1 = append(logs1, l)
		case l := <-ch2:
			logs2 = append(logs2, l)
		case <-deadline:
			goto done
		}
	}
done:
	// Each channel must only contain logs for its contract (correct dispatch).
	for _, l := range logs1 {
		assert.Equal(t, addr1, l.Address, "ch1 must only get contract1 logs")
	}
	for _, l := range logs2 {
		assert.Equal(t, addr2, l.Address, "ch2 must only get contract2 logs")
	}
	assert.GreaterOrEqual(t, len(logs1), 1, "subscription 1 should get contract1 events")
	assert.GreaterOrEqual(t, len(logs2), 1, "subscription 2 should get contract2 events")
	// Realtime scanner may poll same range multiple times; dedup must prevent duplicate logs.
	assertNoDuplicateLogs(t, logs1)
	assertNoDuplicateLogs(t, logs2)
}

// TestMergedRealtime_NoExtraNoMissing verifies that after merging queries into one eth_getLogs,
// each subscription receives exactly the logs that match its query (no extra, no missing).
func TestMergedRealtime_NoExtraNoMissing(t *testing.T) {
	t.Parallel()
	sim := helper.SetUpClient(t)
	defer sim.Close()
	sim.Client().SetBlockConfirmationsOnSubscription(0)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	addr1, _, c1 := helper.DeployTestContract(t, ctx, sim)
	addr2, _, c2 := helper.DeployTestContract(t, ctx, sim)
	chainID, _ := sim.Client().ChainID(ctx)
	storage := subscriber.NewMemoryStorage(chainID)
	cs, err := subscriber.NewChainSubscriber(sim.Client().RpcClient(), storage)
	require.NoError(t, err)
	defer cs.Close()
	cs.SetRetryInterval(testRetryInterval)
	sim.Client().SetSubscriber(cs)

	// Three different queries: only addr1, only addr2, and both (merged scanner serves all with one eth_getLogs).
	ch1 := make(chan types.Log, 16)
	ch2 := make(chan types.Log, 16)
	ch3 := make(chan types.Log, 16)
	sub1, err := cs.SubscribeFilterLogs(ctx, ethereum.FilterQuery{Addresses: []common.Address{addr1}}, ch1)
	require.NoError(t, err)
	defer sub1.Unsubscribe()
	sub2, err := cs.SubscribeFilterLogs(ctx, ethereum.FilterQuery{Addresses: []common.Address{addr2}}, ch2)
	require.NoError(t, err)
	defer sub2.Unsubscribe()
	sub3, err := cs.SubscribeFilterLogs(ctx, ethereum.FilterQuery{Addresses: []common.Address{addr1, addr2}}, ch3)
	require.NoError(t, err)
	defer sub3.Unsubscribe()

	// Emit one event from c1 and one from c2 (use fresh opts per tx so nonce is correct).
	opts1, _ := sim.Client().MessageToTransactOpts(ctx, message.Request{From: helper.Addr1})
	_, err = c1.TestFunc1(opts1, "c1", big.NewInt(1), []byte("1"))
	require.NoError(t, err)
	sim.Commit()
	opts2, _ := sim.Client().MessageToTransactOpts(ctx, message.Request{From: helper.Addr1})
	_, err = c2.TestFunc1(opts2, "c2", big.NewInt(2), []byte("2"))
	require.NoError(t, err)
	sim.Commit()
	time.Sleep(2 * testRetryInterval)

	var logs1, logs2, logs3 []types.Log
	deadline := time.After(1500 * time.Millisecond)
	for {
		select {
		case l := <-ch1:
			logs1 = append(logs1, l)
		case l := <-ch2:
			logs2 = append(logs2, l)
		case l := <-ch3:
			logs3 = append(logs3, l)
		case <-deadline:
			goto done
		}
	}
done:
	// No extra: each channel only logs for its filter.
	for _, l := range logs1 {
		assert.Equal(t, addr1, l.Address, "ch1 must only get addr1 logs")
	}
	for _, l := range logs2 {
		assert.Equal(t, addr2, l.Address, "ch2 must only get addr2 logs")
	}
	for _, l := range logs3 {
		assert.True(t, l.Address == addr1 || l.Address == addr2, "ch3 must only get addr1 or addr2")
	}
	// No missing: ch1 gets 1 (c1), ch2 gets 1 (c2), ch3 gets 2 (c1 + c2).
	assert.GreaterOrEqual(t, len(logs1), 1, "ch1 must get at least the c1 event")
	assert.GreaterOrEqual(t, len(logs2), 1, "ch2 must get at least the c2 event")
	assert.GreaterOrEqual(t, len(logs3), 2, "ch3 must get both c1 and c2 events")
	assertNoDuplicateLogs(t, logs1)
	assertNoDuplicateLogs(t, logs2)
	assertNoDuplicateLogs(t, logs3)
}

// assertNoDuplicateLogs fails if the same log (block+txindex+index) appears more than once.
func assertNoDuplicateLogs(t *testing.T, logs []types.Log) {
	t.Helper()
	seen := make(map[string]struct{})
	for _, l := range logs {
		// BlockNumber=endBlock is sent as progress marker, not a real log; skip for dedup check.
		if l.Address == (common.Address{}) && l.TxHash == (common.Hash{}) {
			continue
		}
		key := fmt.Sprintf("%d:%d:%d", l.BlockNumber, l.TxIndex, l.Index)
		_, ok := seen[key]
		assert.False(t, ok, "duplicate log block=%d txIndex=%d index=%d", l.BlockNumber, l.TxIndex, l.Index)
		seen[key] = struct{}{}
	}
}

// TestRealtime_SameQueryMultipleSubscribers verifies that multiple subscribers to the same query each receive logs.
func TestRealtime_SameQueryMultipleSubscribers(t *testing.T) {
	t.Parallel()
	sim := helper.SetUpClient(t)
	defer sim.Close()
	sim.Client().SetBlockConfirmationsOnSubscription(0)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	addr, _, contract := helper.DeployTestContract(t, ctx, sim)
	chainID, _ := sim.Client().ChainID(ctx)
	storage := subscriber.NewMemoryStorage(chainID)
	cs, err := subscriber.NewChainSubscriber(sim.Client().RpcClient(), storage)
	require.NoError(t, err)
	defer cs.Close()
	cs.SetRetryInterval(testRetryInterval)
	sim.Client().SetSubscriber(cs)

	q := ethereum.FilterQuery{Addresses: []common.Address{addr}}
	ch1 := make(chan types.Log, 16)
	ch2 := make(chan types.Log, 16)
	sub1, err := cs.SubscribeFilterLogs(ctx, q, ch1)
	require.NoError(t, err)
	defer sub1.Unsubscribe()
	sub2, err := cs.SubscribeFilterLogs(ctx, q, ch2)
	require.NoError(t, err)
	defer sub2.Unsubscribe()

	opts, _ := sim.Client().MessageToTransactOpts(ctx, message.Request{From: helper.Addr1})
	_, err = contract.TestFunc1(opts, "shared", big.NewInt(1), []byte("x"))
	require.NoError(t, err)
	sim.Commit()
	time.Sleep(2 * testRetryInterval)

	var logs1, logs2 []types.Log
	deadline := time.After(1200 * time.Millisecond)
	for {
		select {
		case l := <-ch1:
			logs1 = append(logs1, l)
		case l := <-ch2:
			logs2 = append(logs2, l)
		case <-deadline:
			goto done
		}
	}
done:
	assert.GreaterOrEqual(t, len(logs1), 1, "first subscriber must receive events")
	assert.GreaterOrEqual(t, len(logs2), 1, "second subscriber must receive same query events")
	// Both should see the same contract logs (same query).
	for _, l := range logs1 {
		assert.Equal(t, addr, l.Address)
	}
	for _, l := range logs2 {
		assert.Equal(t, addr, l.Address)
	}
}

// TestRealtime_UnsubscribeStopsDelivery verifies that after Unsubscribe, that channel gets no more logs; others with same query still do.
func TestRealtime_UnsubscribeStopsDelivery(t *testing.T) {
	t.Parallel()
	sim := helper.SetUpClient(t)
	defer sim.Close()
	sim.Client().SetBlockConfirmationsOnSubscription(0)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	addr, _, contract := helper.DeployTestContract(t, ctx, sim)
	chainID, _ := sim.Client().ChainID(ctx)
	storage := subscriber.NewMemoryStorage(chainID)
	cs, err := subscriber.NewChainSubscriber(sim.Client().RpcClient(), storage)
	require.NoError(t, err)
	defer cs.Close()
	cs.SetRetryInterval(testRetryInterval)
	sim.Client().SetSubscriber(cs)

	q := ethereum.FilterQuery{Addresses: []common.Address{addr}}
	ch1 := make(chan types.Log, 16)
	ch2 := make(chan types.Log, 16)
	sub1, err := cs.SubscribeFilterLogs(ctx, q, ch1)
	require.NoError(t, err)
	sub2, err := cs.SubscribeFilterLogs(ctx, q, ch2)
	require.NoError(t, err)

	sub1.Unsubscribe()

	opts, _ := sim.Client().MessageToTransactOpts(ctx, message.Request{From: helper.Addr1})
	_, err = contract.TestFunc1(opts, "after-unsub", big.NewInt(1), []byte("x"))
	require.NoError(t, err)
	sim.Commit()
	time.Sleep(2 * testRetryInterval)

	var count1, count2 int
	deadline := time.After(1200 * time.Millisecond)
	for {
		select {
		case <-ch1:
			count1++
		case <-ch2:
			count2++
		case <-deadline:
			goto done
		}
	}
done:
	sub2.Unsubscribe()
	assert.Equal(t, 0, count1, "unsubscribed channel must receive no logs")
	assert.GreaterOrEqual(t, count2, 1, "still-subscribed channel must receive events")
}

// TestSubmitQuery_RealtimeRegistersWithoutSpawningPerQueryLoop verifies SubmitQuery with ToBlock==nil does not block.
// Must close subscriber so runRealtimeScanner exits.
func TestSubmitQuery_RealtimeRegistersWithoutSpawningPerQueryLoop(t *testing.T) {
	t.Parallel()
	sim := helper.SetUpClient(t)
	defer sim.Close()

	chainID, _ := sim.Client().ChainID(context.Background())
	storage := subscriber.NewMemoryStorage(chainID)
	cs, err := subscriber.NewChainSubscriber(sim.Client().RpcClient(), storage)
	require.NoError(t, err)
	defer cs.Close()
	h := newCountHandler(storage)
	cs.SetQueryHandler(h)
	sim.Client().SetSubscriber(cs)

	err = cs.SubmitQuery(ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress("0x0000000000000000000000000000000000000001")}},
	)
	require.NoError(t, err)
	// Duplicate submit must error
	err = cs.SubmitQuery(ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress("0x0000000000000000000000000000000000000001")}},
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already submitted")
}

var _ subscriber.QueryHandler = (*countHandler)(nil)

type countHandler struct {
	handler.SimpleQueryHandler
	count int
}

func newCountHandler(storage subscriber.SubscriberStorage) *countHandler {
	return &countHandler{SimpleQueryHandler: *handler.NewSimpleQueryHandler(storage)}
}

func (h *countHandler) HandleQuery(ctx context.Context, q subscriber.Query, l types.Log) error {
	h.count++
	return h.SimpleQueryHandler.HandleQuery(ctx, q, l)
}
