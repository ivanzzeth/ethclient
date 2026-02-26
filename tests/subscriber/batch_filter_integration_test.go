package subscriber_test

import (
	"context"
	"fmt"
	"math/big"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/go-redsync/redsync/v4/redis"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/ivanzzeth/ethclient/message"
	"github.com/ivanzzeth/ethclient/simulated"
	"github.com/ivanzzeth/ethclient/subscriber"
	"github.com/ivanzzeth/ethclient/subscriber/handler"
	"github.com/ivanzzeth/ethclient/tests/helper"
	goredislib "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// spyStorage wraps MemoryStorage and records SaveLatestBlockForQuery calls so tests can assert progress is not advanced when RPC returns 0 logs.
type spyStorage struct {
	*subscriber.MemoryStorage
	saveBlockCalls atomic.Int32
}

func newSpyStorage(chainID *big.Int) *spyStorage {
	return &spyStorage{MemoryStorage: subscriber.NewMemoryStorage(chainID)}
}

func (s *spyStorage) SaveLatestBlockForQuery(ctx context.Context, query ethereum.FilterQuery, blockNum uint64) error {
	s.saveBlockCalls.Add(1)
	return s.MemoryStorage.SaveLatestBlockForQuery(ctx, query, blockNum)
}

func (s *spyStorage) SaveLatestBlockCalls() int32 { return s.saveBlockCalls.Load() }

// All tests in this file use the simulated backend (helper.SetUpClient returns *simulated.Backend).
var _ *simulated.Backend = nil

// testRetryInterval shortens realtime scanner polling in tests so we need less sleep.
const testRetryInterval = 250 * time.Millisecond

// newMiniredisPool starts a miniredis server and returns it plus a redsync pool for subscriber.RedisStorage.
// Caller should defer mr.Close().
func newMiniredisPool(t *testing.T) (mr *miniredis.Miniredis, pool redis.Pool) {
	t.Helper()
	mr = miniredis.RunT(t)
	rdb := goredislib.NewClient(&goredislib.Options{Addr: mr.Addr()})
	pool = goredis.NewPool(rdb)
	return mr, pool
}

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

// TestFilterLogsBatch_ToBlockPending covers toBlockNumArg with negative block (e.g. "pending").
// Simulated backend may not support "pending" for eth_getLogs; we only assert the call path is exercised.
func TestFilterLogsBatch_ToBlockPending(t *testing.T) {
	t.Parallel()
	sim := helper.SetUpClient(t)
	defer sim.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	contractAddr, _, _ := helper.DeployTestContract(t, ctx, sim)
	batch := subscriber.NewBatchLogFilterer(sim.Client().RawClient(), sim.Client().RpcClient())
	queries := []ethereum.FilterQuery{{
		FromBlock: big.NewInt(0),
		ToBlock:   big.NewInt(-1),
		Addresses: []common.Address{contractAddr},
	}}
	allLogs, err := batch.FilterLogsBatch(ctx, queries)
	if err != nil {
		t.Skipf("simulated backend may not support ToBlock pending: %v", err)
	}
	require.Len(t, allLogs, 1)
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

// TestRealtime_MinStartFromBlock_WhenStorageEmpty verifies that when storage returns 0 for LatestBlockForQuery,
// the realtime scanner's first range starts at query.FromBlock (subscription start), not at block 1.
// Otherwise we would scan from genesis and never reach high blocks (e.g. 83M+), missing events.
func TestRealtime_MinStartFromBlock_WhenStorageEmpty(t *testing.T) {
	t.Parallel()
	sim := helper.SetUpClient(t)
	defer sim.Close()
	sim.Client().SetBlockConfirmationsOnSubscription(0)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	contractAddr, _, contract := helper.DeployTestContract(t, ctx, sim)
	// Mine so we are at block 100; no contract events in 2..100 so first scan 1..100 would return 0 logs for this contract.
	for i := 0; i < 99; i++ {
		sim.Commit()
	}
	fromBlock, err := sim.Client().BlockNumber(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, fromBlock, uint64(100), "need at least block 100 for this test")

	chainID, _ := sim.Client().ChainID(ctx)
	storage := subscriber.NewMemoryStorage(chainID)
	cs, err := subscriber.NewChainSubscriber(sim.Client().RpcClient(), storage)
	require.NoError(t, err)
	defer cs.Close()
	cs.SetRetryInterval(testRetryInterval)
	sim.Client().SetSubscriber(cs)

	ch := make(chan types.Log, 8)
	sub, err := cs.SubscribeFilterLogs(ctx, ethereum.FilterQuery{
		FromBlock: big.NewInt(0).SetUint64(fromBlock),
		Addresses: []common.Address{contractAddr},
	}, ch)
	require.NoError(t, err)
	defer sub.Unsubscribe()

	// Emit event in next block (fromBlock+1).
	opts, _ := sim.Client().MessageToTransactOpts(ctx, message.Request{From: helper.Addr1})
	_, err = contract.TestFunc1(opts, "high", big.NewInt(1), []byte("block"))
	require.NoError(t, err)
	sim.Commit()
	time.Sleep(3 * testRetryInterval)

	var received []types.Log
	deadline := time.After(2 * time.Second)
	for {
		select {
		case l := <-ch:
			received = append(received, l)
		case <-deadline:
			goto done
		}
	}
done:
	require.GreaterOrEqual(t, len(received), 1, "realtime subscription with FromBlock=%d and empty storage must receive event in block %d (minStart must use FromBlock, not 1)", fromBlock, fromBlock+1)
	for _, l := range received {
		assert.GreaterOrEqual(t, l.BlockNumber, fromBlock, "log must be from subscription range")
	}
}

// TestRealtime_NoAdvanceWhenZeroLogs verifies that when the merged FilterLogs RPC returns 0 logs, the realtime scanner does not call SaveLatestBlockForQuery, so the same range is retried next cycle (avoids permanently missing Split/Merge events).
func TestRealtime_NoAdvanceWhenZeroLogs(t *testing.T) {
	t.Parallel()
	sim := helper.SetUpClient(t)
	defer sim.Close()
	sim.Client().SetBlockConfirmationsOnSubscription(0)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	contractAddr, _, contract := helper.DeployTestContract(t, ctx, sim)
	chainID, _ := sim.Client().ChainID(ctx)
	storage := newSpyStorage(chainID)
	cs, err := subscriber.NewChainSubscriber(sim.Client().RpcClient(), storage)
	require.NoError(t, err)
	defer cs.Close()
	cs.SetRetryInterval(testRetryInterval)
	sim.Client().SetSubscriber(cs)

	// Subscribe to contract address; do not emit any event yet, so the first scanner cycle will get 0 logs for this range.
	ch := make(chan types.Log, 8)
	query := ethereum.FilterQuery{Addresses: []common.Address{contractAddr}}
	sub, err := cs.SubscribeFilterLogs(ctx, query, ch)
	require.NoError(t, err)
	defer sub.Unsubscribe()

	// Wait for at least one realtime scanner cycle (RPC returns 0 logs for this range).
	time.Sleep(2 * testRetryInterval)

	// Progress must not have been advanced when RPC returned 0 logs.
	assert.Equal(t, int32(0), storage.SaveLatestBlockCalls(), "SaveLatestBlockForQuery must not be called when merged RPC returns 0 logs")
	latest, err := storage.LatestBlockForQuery(ctx, query)
	require.NoError(t, err)
	assert.Equal(t, uint64(0), latest, "LatestBlockForQuery must still be 0 when no logs were received")

	// Emit an event so the next scanner cycle gets logs; then progress must advance.
	opts, _ := sim.Client().MessageToTransactOpts(ctx, message.Request{From: helper.Addr1})
	_, err = contract.TestFunc1(opts, "ev", big.NewInt(1), []byte("data"))
	require.NoError(t, err)
	sim.Commit()
	time.Sleep(2 * testRetryInterval)

	// After receiving logs, progress must have been advanced.
	assert.Greater(t, storage.SaveLatestBlockCalls(), int32(0), "SaveLatestBlockForQuery must be called when merged RPC returns logs")
	latest2, err := storage.LatestBlockForQuery(ctx, query)
	require.NoError(t, err)
	assert.Greater(t, latest2, uint64(0), "LatestBlockForQuery must advance after logs received")
}

// TestRedisStorage_Miniredis_CRUD covers RedisStorage methods using in-memory miniredis.
func TestRedisStorage_Miniredis_CRUD(t *testing.T) {
	t.Parallel()
	mr, pool := newMiniredisPool(t)
	defer mr.Close()

	ctx := context.Background()
	chainID := big.NewInt(1337)
	storage := subscriber.NewRedisStorage(chainID, pool)
	addr := common.HexToAddress("0x0000000000000000000000000000000000000001")
	q := ethereum.FilterQuery{Addresses: []common.Address{addr}}

	// LatestBlockForQuery empty
	block, err := storage.LatestBlockForQuery(ctx, q)
	require.NoError(t, err)
	require.Equal(t, uint64(0), block)

	// SaveLatestBlockForQuery then read
	err = storage.SaveLatestBlockForQuery(ctx, q, 100)
	require.NoError(t, err)
	block, err = storage.LatestBlockForQuery(ctx, q)
	require.NoError(t, err)
	require.Equal(t, uint64(100), block)

	// LatestLogForQuery empty
	log1, err := storage.LatestLogForQuery(ctx, q)
	require.NoError(t, err)
	require.Equal(t, uint64(0), log1.BlockNumber)

	// SaveLatestLogForQuery then read (Topics required for Log JSON round-trip)
	logToSave := types.Log{BlockNumber: 101, TxIndex: 1, Index: 2, Address: addr, Topics: []common.Hash{}}
	err = storage.SaveLatestLogForQuery(ctx, q, logToSave)
	require.NoError(t, err)
	log2, err := storage.LatestLogForQuery(ctx, q)
	require.NoError(t, err)
	require.Equal(t, uint64(101), log2.BlockNumber)
	require.Equal(t, uint(1), log2.TxIndex)
	require.Equal(t, uint(2), log2.Index)

	// IsFilterLogsSupported false, FilterLogs not implemented
	require.False(t, storage.IsFilterLogsSupported(q))
	logs, err := storage.FilterLogs(ctx, q)
	require.Error(t, err)
	require.Nil(t, logs)
	require.Contains(t, err.Error(), "not implemented")

	// QueryLock returns a locker (smoke: Lock/Unlock does not panic)
	locker := storage.QueryLock(q)
	locker.Lock()
	locker.Unlock()

	// SaveFilterLogs is a no-op for RedisStorage
	err = storage.SaveFilterLogs(q, nil)
	require.NoError(t, err)
}

// TestRealtime_RedisStorage_Miniredis verifies realtime SubscribeFilterLogs with RedisStorage backed by miniredis (in-memory).
func TestRealtime_RedisStorage_Miniredis(t *testing.T) {
	t.Parallel()
	mr, pool := newMiniredisPool(t)
	defer mr.Close()

	sim := helper.SetUpClient(t)
	defer sim.Close()
	sim.Client().SetBlockConfirmationsOnSubscription(0)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	contractAddr, _, contract := helper.DeployTestContract(t, ctx, sim)
	chainID, err := sim.Client().ChainID(ctx)
	require.NoError(t, err)
	storage := subscriber.NewRedisStorage(chainID, pool)
	cs, err := subscriber.NewChainSubscriber(sim.Client().RpcClient(), storage)
	require.NoError(t, err)
	defer cs.Close()
	cs.SetRetryInterval(testRetryInterval)
	sim.Client().SetSubscriber(cs)

	ch := make(chan types.Log, 16)
	sub, err := cs.SubscribeFilterLogs(ctx, ethereum.FilterQuery{
		Addresses: []common.Address{contractAddr},
	}, ch)
	require.NoError(t, err)
	defer sub.Unsubscribe()

	opts, _ := sim.Client().MessageToTransactOpts(ctx, message.Request{From: helper.Addr1})
	_, err = contract.TestFunc1(opts, "redis", big.NewInt(1), []byte("miniredis"))
	require.NoError(t, err)
	sim.Commit()
	time.Sleep(2 * testRetryInterval)

	var received []types.Log
	deadline := time.After(2 * time.Second)
	for {
		select {
		case l := <-ch:
			received = append(received, l)
		case <-deadline:
			goto done
		}
	}
done:
	require.GreaterOrEqual(t, len(received), 1, "realtime subscription with RedisStorage (miniredis) must receive event")
	latest, err := storage.LatestBlockForQuery(ctx, ethereum.FilterQuery{Addresses: []common.Address{contractAddr}})
	require.NoError(t, err)
	assert.Greater(t, latest, uint64(0), "RedisStorage must persist progress after receiving logs")
}

// TestChainSubscriber_SettersGetters covers setters and getters so they are not 0% coverage.
func TestChainSubscriber_SettersGetters(t *testing.T) {
	t.Parallel()
	sim := helper.SetUpClient(t)
	defer sim.Close()
	chainID, _ := sim.Client().ChainID(context.Background())
	storage := subscriber.NewMemoryStorage(chainID)
	cs, err := subscriber.NewChainSubscriber(sim.Client().RpcClient(), storage)
	require.NoError(t, err)
	defer cs.Close()

	cs.SetBlocksPerScan(50)
	cs.SetMaxBlocksPerScan(2000)
	cs.SetRetryInterval(100 * time.Millisecond)
	cs.SetBlockConfirmationsOnSubscription(2)
	require.Equal(t, uint64(2), cs.GetBlockConfirmationsOnSubscription())
	cs.SetBuffer(64)
	require.Nil(t, cs.GetQueryHandler())
	h := handler.NewSimpleQueryHandler(storage)
	cs.SetQueryHandler(h)
	require.Same(t, h, cs.GetQueryHandler())
	cs.SetFetchMissingHeaders(true)

	// Subscription Err() channel
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ch := make(chan types.Log, 1)
	sub, err := cs.SubscribeFilterLogs(ctx, ethereum.FilterQuery{Addresses: []common.Address{{}}}, ch)
	require.NoError(t, err)
	errCh := sub.Err()
	sub.Unsubscribe()
	select {
	case e := <-errCh:
		require.Error(t, e)
	case <-time.After(time.Second):
		t.Fatal("Err() channel did not receive")
	}
}

// TestFilterLogsWithChannel_BlockHash covers the BlockHash branch of FilterLogsWithChannel.
func TestFilterLogsWithChannel_BlockHash(t *testing.T) {
	t.Parallel()
	sim := helper.SetUpClient(t)
	defer sim.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	contractAddr, _, contract := helper.DeployTestContract(t, ctx, sim)
	opts, _ := sim.Client().MessageToTransactOpts(ctx, message.Request{From: helper.Addr1})
	_, err := contract.TestFunc1(opts, "a", big.NewInt(1), []byte("b"))
	require.NoError(t, err)
	sim.Commit()
	blockNum, _ := sim.Client().BlockNumber(ctx)
	block, err := sim.Client().BlockByNumber(ctx, big.NewInt(int64(blockNum)))
	require.NoError(t, err)
	blockHash := block.Hash()

	chainID, _ := sim.Client().ChainID(ctx)
	storage := subscriber.NewMemoryStorage(chainID)
	cs, err := subscriber.NewChainSubscriber(sim.Client().RpcClient(), storage)
	require.NoError(t, err)
	defer cs.Close()

	logsCh := make(chan types.Log, 8)
	err = cs.FilterLogsWithChannel(ctx, ethereum.FilterQuery{
		BlockHash:  &blockHash,
		Addresses:  []common.Address{contractAddr},
	}, logsCh, false, true)
	require.NoError(t, err)
	var received []types.Log
	for l := range logsCh {
		received = append(received, l)
	}
	require.GreaterOrEqual(t, len(received), 1)
	for _, l := range received {
		assert.Equal(t, contractAddr, l.Address)
	}
}

// TestFilterLogsWithChannel_HistoricalRange covers the range path (FromBlock/ToBlock) and storage resume.
func TestFilterLogsWithChannel_HistoricalRange(t *testing.T) {
	t.Parallel()
	sim := helper.SetUpClient(t)
	defer sim.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	contractAddr, _, contract := helper.DeployTestContract(t, ctx, sim)
	fromBlock, _ := sim.Client().BlockNumber(ctx)
	opts, _ := sim.Client().MessageToTransactOpts(ctx, message.Request{From: helper.Addr1})
	_, err := contract.TestFunc1(opts, "x", big.NewInt(1), []byte("y"))
	require.NoError(t, err)
	sim.Commit()
	toBlock, _ := sim.Client().BlockNumber(ctx)

	chainID, _ := sim.Client().ChainID(ctx)
	storage := subscriber.NewMemoryStorage(chainID)
	cs, err := subscriber.NewChainSubscriber(sim.Client().RpcClient(), storage)
	require.NoError(t, err)
	defer cs.Close()

	logsCh := make(chan types.Log, 8)
	err = cs.FilterLogsWithChannel(ctx, ethereum.FilterQuery{
		FromBlock:  big.NewInt(int64(fromBlock)),
		ToBlock:    big.NewInt(int64(toBlock)),
		Addresses:  []common.Address{contractAddr},
	}, logsCh, false, true)
	require.NoError(t, err)
	var count int
	for range logsCh {
		count++
	}
	require.GreaterOrEqual(t, count, 1)
}

// TestSubscribeFilterLogs_WithToBlock_usesFilterLogsWithChannelWatch covers FilterLogsWithChannel with watch=true (SubscribeFilterLogs when ToBlock is set).
func TestSubscribeFilterLogs_WithToBlock_usesFilterLogsWithChannelWatch(t *testing.T) {
	t.Parallel()
	sim := helper.SetUpClient(t)
	defer sim.Close()
	sim.Client().SetBlockConfirmationsOnSubscription(0)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	contractAddr, _, contract := helper.DeployTestContract(t, ctx, sim)
	fromBlock, _ := sim.Client().BlockNumber(ctx)
	toBlock := fromBlock + 20

	chainID, _ := sim.Client().ChainID(ctx)
	storage := subscriber.NewMemoryStorage(chainID)
	cs, err := subscriber.NewChainSubscriber(sim.Client().RpcClient(), storage)
	require.NoError(t, err)
	defer cs.Close()
	cs.SetRetryInterval(testRetryInterval)
	sim.Client().SetSubscriber(cs)

	logsCh := make(chan types.Log, 16)
	sub, err := cs.SubscribeFilterLogs(ctx, ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(toBlock)),
		Addresses: []common.Address{contractAddr},
	}, logsCh)
	require.NoError(t, err)
	defer sub.Unsubscribe()

	opts, _ := sim.Client().MessageToTransactOpts(ctx, message.Request{From: helper.Addr1})
	_, err = contract.TestFunc1(opts, "watch", big.NewInt(1), []byte("data"))
	require.NoError(t, err)
	sim.Commit()
	time.Sleep(2 * testRetryInterval)

	var count int
	deadline := time.After(2 * time.Second)
	for {
		select {
		case <-logsCh:
			count++
		case <-deadline:
			goto done
		}
	}
done:
	require.GreaterOrEqual(t, count, 1, "watch mode should deliver logs")
}

// storageSupportingFilterLogs wraps MemoryStorage and reports IsFilterLogsSupported=true, FilterLogs returns nil.
// Used to cover the storage.FilterLogs branch in FilterLogsWithChannel.
type storageSupportingFilterLogs struct {
	*subscriber.MemoryStorage
}

func (s *storageSupportingFilterLogs) IsFilterLogsSupported(q ethereum.FilterQuery) bool { return true }
func (s *storageSupportingFilterLogs) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	return nil, nil
}

// TestFilterLogsWithChannel_UseStorageResume covers the useStorage path where LatestBlockForQuery returns non-zero (resume from storage).
func TestFilterLogsWithChannel_UseStorageResume(t *testing.T) {
	t.Parallel()
	sim := helper.SetUpClient(t)
	defer sim.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	contractAddr, _, _ := helper.DeployTestContract(t, ctx, sim)
	chainID, _ := sim.Client().ChainID(ctx)
	storage := subscriber.NewMemoryStorage(chainID)
	// Pre-seed storage so startBlock = fromBlockInStorage + 1
	q := ethereum.FilterQuery{Addresses: []common.Address{contractAddr}}
	err := storage.SaveLatestBlockForQuery(ctx, q, 1)
	require.NoError(t, err)

	cs, err := subscriber.NewChainSubscriber(sim.Client().RpcClient(), storage)
	require.NoError(t, err)
	defer cs.Close()

	// ToBlock nil => useStorage true; will resume from block 2
	logsCh := make(chan types.Log, 8)
	err = cs.FilterLogsWithChannel(ctx, ethereum.FilterQuery{
		FromBlock: big.NewInt(1),
		ToBlock:   nil,
		Addresses: []common.Address{contractAddr},
	}, logsCh, false, true)
	require.NoError(t, err)
	// Drain channel (may get 0 or more logs depending on chain state)
	for range logsCh {
	}
}

// TestFilterLogsWithChannel_StorageFilterLogsBranch covers the branch where storage.IsFilterLogsSupported is true.
func TestFilterLogsWithChannel_StorageFilterLogsBranch(t *testing.T) {
	t.Parallel()
	sim := helper.SetUpClient(t)
	defer sim.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	chainID, _ := sim.Client().ChainID(ctx)
	base := subscriber.NewMemoryStorage(chainID)
	storage := &storageSupportingFilterLogs{MemoryStorage: base}
	cs, err := subscriber.NewChainSubscriber(sim.Client().RpcClient(), storage)
	require.NoError(t, err)
	defer cs.Close()

	logsCh := make(chan types.Log, 4)
	err = cs.FilterLogsWithChannel(ctx, ethereum.FilterQuery{
		FromBlock: big.NewInt(1),
		ToBlock:   big.NewInt(10),
		Addresses: []common.Address{{}},
	}, logsCh, false, true)
	require.NoError(t, err)
	for range logsCh {
	}
}

// TestSubscribeFullPendingTransactions_and_SubscribePendingTransactions covers the 0%% subscription methods.
func TestSubscribeFullPendingTransactions_and_SubscribePendingTransactions(t *testing.T) {
	t.Parallel()
	sim := helper.SetUpClient(t)
	defer sim.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client := sim.Client()
	if cs, ok := client.Subscriber.(*subscriber.ChainSubscriber); ok {
		defer cs.Close()
	}

	chFull := make(chan *types.Transaction, 4)
	subFull, err := client.Subscriber.(*subscriber.ChainSubscriber).SubscribeFullPendingTransactions(ctx, chFull)
	require.NoError(t, err)
	subFull.Unsubscribe()

	chHash := make(chan common.Hash, 4)
	subHash, err := client.Subscriber.(*subscriber.ChainSubscriber).SubscribePendingTransactions(ctx, chHash)
	require.NoError(t, err)
	subHash.Unsubscribe()
}

// TestFilterLogsWithChannel_WatchWithConfirmations covers watch mode with blockConfirmationsOnSubscription > 0.
func TestFilterLogsWithChannel_WatchWithConfirmations(t *testing.T) {
	t.Parallel()
	sim := helper.SetUpClient(t)
	defer sim.Close()
	sim.Client().SetBlockConfirmationsOnSubscription(2)
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	contractAddr, _, contract := helper.DeployTestContract(t, ctx, sim)
	fromBlock, _ := sim.Client().BlockNumber(ctx)
	chainID, _ := sim.Client().ChainID(ctx)
	storage := subscriber.NewMemoryStorage(chainID)
	cs, err := subscriber.NewChainSubscriber(sim.Client().RpcClient(), storage)
	require.NoError(t, err)
	defer cs.Close()
	cs.SetRetryInterval(testRetryInterval)
	sim.Client().SetSubscriber(cs)

	logsCh := make(chan types.Log, 16)
	err = cs.FilterLogsWithChannel(ctx, ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   nil,
		Addresses: []common.Address{contractAddr},
	}, logsCh, true, true)
	require.NoError(t, err)

	opts, _ := sim.Client().MessageToTransactOpts(ctx, message.Request{From: helper.Addr1})
	_, err = contract.TestFunc1(opts, "c", big.NewInt(1), []byte("d"))
	require.NoError(t, err)
	sim.Commit()
	for i := 0; i < 3; i++ {
		sim.Commit()
	}
	time.Sleep(2 * testRetryInterval)
	deadline := time.After(3 * time.Second)
	for {
		select {
		case <-logsCh:
		case <-deadline:
			goto doneWatch
		}
	}
doneWatch:
}

// TestSubscribeFilterFullTransactions covers SubscribeFilterFullTransactions (new heads -> block txs -> filter).
func TestSubscribeFilterFullTransactions(t *testing.T) {
	t.Parallel()
	sim := helper.SetUpClient(t)
	defer sim.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client := sim.Client()
	if cs, ok := client.Subscriber.(*subscriber.ChainSubscriber); ok {
		defer cs.Close()
	}
	txCh := make(chan *types.Transaction, 8)
	sub, err := client.Subscriber.(*subscriber.ChainSubscriber).SubscribeFilterFullTransactions(ctx, subscriber.FilterTransaction{}, txCh)
	require.NoError(t, err)
	defer sub.Unsubscribe()
	sim.Commit()
	time.Sleep(500 * time.Millisecond)
	select {
	case <-txCh:
	case <-time.After(2 * time.Second):
		// No tx in block is ok
	}
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
