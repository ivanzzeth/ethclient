// Temporary test to verify SubscribeFilterLogs (realtime, same as prediction-exchange)
// receives PositionsMerge events after RPC merge. Requires forked node with a Merge in range.
//
// Setup: see tests/fork/README.md — run anvil with fork, optionally trigger one Merge,
// then set SUBSCRIBE_MERGE_TEST_RPC, CT_ADDRESS, STAKEHOLDER_ADDRESS and run:
//
//	go test -v -count=1 -run TestSubscribeMergeRealtime ./tests/subscriber/
package subscriber_test

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ivanzzeth/ethclient/subscriber"
	"github.com/stretchr/testify/require"
)

// PositionsMerge event sig (ConditionalTokens): keccak256("PositionsMerge(address,address,bytes32,bytes32,uint256[],uint256)")
// From predict-go-contracts conditional_tokens.go binding.
const positionsMergeTopic0Hex = "0x6f13ca62553fcc2bcd2372180a43949c1e4cebba603901ede2f4e14f36b282ca"

// PositionSplit event sig: keccak256("PositionSplit(address,address,bytes32,bytes32,uint256[],uint256)")
const positionSplitTopic0Hex = "0x2e6bb91f8cbcda0c93623c54d0403a43514fabc40084ec96b6d5379a74786298"

// PayoutRedemption (ConditionalTokens): event PayoutRedemption(address indexed redeemer, ...)
const payoutRedemptionTopic0Hex = "0x2682012a4a4f1973119f1c9b90745d1bd91fa2bab387344f044cb3586864d18d"

// OrderFilled (Exchange): event OrderFilled(bytes32 indexed orderHash, address indexed maker, address indexed taker, ...)
const orderFilledTopic0Hex = "0xd0a08e8c493f9c94f29311604c9de1b4e8c8d4c06bd0c789af57f2d65bfec0f6"

// YBNegRiskAdapter (predict-go-contracts yb_neg_risk_adapter.go)
const ybNegRiskAdapterPositionSplitTopic0Hex = "0xbbed930dbfb7907ae2d60ddf78345610214f26419a0128df39b6cc3d9e5df9b0"
const ybNegRiskAdapterPositionsMergeTopic0Hex = "0xba33ac50d8894676597e6e35dc09cff59854708b642cd069d21eb9c7ca072a04"
const ybNegRiskAdapterPayoutRedemptionTopic0Hex = "0x9140a6a270ef945260c03894b3c6b3b2695e9d5101feef0ff24fec960cfd3224"

// Known Merge tx 0x79a5e872... (block 83318207): funder and adapter addresses on BSC.
var knownMergeFunderAddr = common.HexToAddress("0x53c68c954F85a29D2098E90AdDAf41bAF2fF0a50")
var knownMergeYBNegRiskAdapterAddr = common.HexToAddress("0x41dCe1A4B8FB5e6327701750aF6231B7CD0B2A40")
var knownMergeYBNegRiskCTAddr = common.HexToAddress("0xF64b0b318aaf83bd9071110af24d24445719a07f")

// Production-like exchange placeholder addresses (4 exchanges, distinct for 4 OrderFilled subscriptions).
var exchangePlaceholderAddrs = []common.Address{
	common.HexToAddress("0x0000000000000000000000000000000000000001"),
	common.HexToAddress("0x0000000000000000000000000000000000000002"),
	common.HexToAddress("0x0000000000000000000000000000000000000003"),
	common.HexToAddress("0x0000000000000000000000000000000000000004"),
}

// TestSubscribeMergeRealtime subscribes with the same FilterQuery shape as Predict's
// WatchPositionsMerge (address + topic0=event sig, topic1=stakeholder, topic2/3=nil),
// using SubscribeFilterLogs directly (realtime, ToBlock=nil). Asserts at least one log
// is received within timeout. Skip if SUBSCRIBE_MERGE_TEST_RPC is not set.
func TestSubscribeMergeRealtime(t *testing.T) {
	rpcURL := os.Getenv("SUBSCRIBE_MERGE_TEST_RPC")
	if rpcURL == "" {
		t.Skip("SUBSCRIBE_MERGE_TEST_RPC not set (e.g. http://127.0.0.1:8545)")
	}
	ctAddr := os.Getenv("CT_ADDRESS")
	stakeholder := os.Getenv("STAKEHOLDER_ADDRESS")
	if ctAddr == "" || stakeholder == "" {
		t.Skip("CT_ADDRESS and STAKEHOLDER_ADDRESS required")
	}

	handler := log.NewTerminalHandlerWithLevel(os.Stdout, log.LevelDebug, true)
	log.SetDefault(log.NewLogger(handler))

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	rpcCli, err := rpc.DialContext(ctx, rpcURL)
	require.NoError(t, err)
	defer rpcCli.Close()

	chainId, err := getChainID(rpcCli)
	require.NoError(t, err)
	storage := subscriber.NewMemoryStorage(chainId)
	cs, err := subscriber.NewChainSubscriber(rpcCli, storage)
	require.NoError(t, err)
	defer cs.Close()

	cs.SetRetryInterval(500 * time.Millisecond)
	cs.SetBlockConfirmationsOnSubscription(0)

	// Same FilterQuery shape as Predict's contract.WatchPositionsMerge(stakeholder, nil, nil):
	// Addresses = [CT], Topics[0] = event sig, Topics[1] = stakeholder (indexed), Topics[2/3] = nil
	topic0 := common.HexToHash(positionsMergeTopic0Hex)
	stakeholderAddr := common.HexToAddress(stakeholder)
	// Indexed address in topics is left-padded 32 bytes
	stakeholderTopic := common.BytesToHash(common.LeftPadBytes(stakeholderAddr.Bytes(), 32))

	q := ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(ctAddr)},
		Topics: [][]common.Hash{
			{topic0},
			{stakeholderTopic},
			nil,
			nil,
		},
		FromBlock: nil,
		ToBlock:   nil, // realtime, same as prediction-exchange
	}
	if fromBlockStr := os.Getenv("FROM_BLOCK"); fromBlockStr != "" {
		fromU64, err := strconv.ParseUint(fromBlockStr, 10, 64)
		if err == nil {
			q.FromBlock = big.NewInt(0).SetUint64(fromU64)
		}
	}

	ch := make(chan types.Log, 8)
	sub, err := cs.SubscribeFilterLogs(ctx, q, ch)
	require.NoError(t, err)
	defer sub.Unsubscribe()

	// Wait for at least one PositionsMerge log (or sentinel block log from realtime scanner)
	var received int
	deadline := time.After(60 * time.Second)
	for received < 1 {
		select {
		case <-ctx.Done():
			t.Fatalf("context done before receiving log: %v", ctx.Err())
		case <-deadline:
			t.Fatalf("timeout waiting for first log (received %d)", received)
		case l, ok := <-ch:
			if !ok {
				t.Fatalf("channel closed after %d logs", received)
			}
			// Realtime scanner can send sentinel log with only BlockNumber set
			if l.Address == (common.Address{}) && l.BlockNumber > 0 {
				t.Logf("realtime sentinel block: %d", l.BlockNumber)
				continue
			}
			if len(l.Topics) > 0 && l.Topics[0] == topic0 {
				received++
				t.Logf("received PositionsMerge log: block=%d tx=%s", l.BlockNumber, l.TxHash.Hex())
			}
		case err := <-sub.Err():
			t.Fatalf("subscription error: %v", err)
		}
	}
	require.GreaterOrEqual(t, received, 1, "expected at least one PositionsMerge log")
}

// TestSubscribeMergeRealtime_MultiSubscription reproduces prediction-exchange: multiple
// realtime subscriptions (Merge + Split) in the same partition → one merged eth_getLogs.
// Asserts the Merge subscription still receives its log when another subscription (Split)
// is merged with it. If Merge does not receive, we've reproduced the missed-dispatch bug.
func TestSubscribeMergeRealtime_MultiSubscription(t *testing.T) {
	rpcURL := os.Getenv("SUBSCRIBE_MERGE_TEST_RPC")
	if rpcURL == "" {
		t.Skip("SUBSCRIBE_MERGE_TEST_RPC not set")
	}
	ctAddr := os.Getenv("CT_ADDRESS")
	stakeholder := os.Getenv("STAKEHOLDER_ADDRESS")
	if ctAddr == "" || stakeholder == "" {
		t.Skip("CT_ADDRESS and STAKEHOLDER_ADDRESS required")
	}

	handler := log.NewTerminalHandlerWithLevel(os.Stdout, log.LevelDebug, true)
	log.SetDefault(log.NewLogger(handler))

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	rpcCli, err := rpc.DialContext(ctx, rpcURL)
	require.NoError(t, err)
	defer rpcCli.Close()

	chainId, err := getChainID(rpcCli)
	require.NoError(t, err)
	storage := subscriber.NewMemoryStorage(chainId)
	cs, err := subscriber.NewChainSubscriber(rpcCli, storage)
	require.NoError(t, err)
	defer cs.Close()

	cs.SetRetryInterval(500 * time.Millisecond)
	cs.SetBlockConfirmationsOnSubscription(0)

	addr := common.HexToAddress(ctAddr)
	stakeholderAddr := common.HexToAddress(stakeholder)
	stakeholderTopic := common.BytesToHash(common.LeftPadBytes(stakeholderAddr.Bytes(), 32))
	mergeTopic0 := common.HexToHash(positionsMergeTopic0Hex)
	splitTopic0 := common.HexToHash(positionSplitTopic0Hex)

	fromBlock := (*big.Int)(nil)
	if fromBlockStr := os.Getenv("FROM_BLOCK"); fromBlockStr != "" {
		fromU64, err := strconv.ParseUint(fromBlockStr, 10, 64)
		if err == nil {
			fromBlock = big.NewInt(0).SetUint64(fromU64)
		}
	}

	// Same shape as Predict: Merge subscription (will be merged with Split in one partition)
	qMerge := ethereum.FilterQuery{
		Addresses: []common.Address{addr},
		Topics:    [][]common.Hash{{mergeTopic0}, {stakeholderTopic}, nil, nil},
		FromBlock: fromBlock,
		ToBlock:   nil,
	}
	// Split subscription: same contract + stakeholder, different topic0 → same partition key
	qSplit := ethereum.FilterQuery{
		Addresses: []common.Address{addr},
		Topics:    [][]common.Hash{{splitTopic0}, {stakeholderTopic}, nil, nil},
		FromBlock: fromBlock,
		ToBlock:   nil,
	}

	chMerge := make(chan types.Log, 8)
	chSplit := make(chan types.Log, 8)
	subMerge, err := cs.SubscribeFilterLogs(ctx, qMerge, chMerge)
	require.NoError(t, err)
	defer subMerge.Unsubscribe()
	subSplit, err := cs.SubscribeFilterLogs(ctx, qSplit, chSplit)
	require.NoError(t, err)
	defer subSplit.Unsubscribe()

	var mergeReceived, splitReceived int
	deadline := time.After(60 * time.Second)
	for mergeReceived < 1 {
		select {
		case <-ctx.Done():
			t.Fatalf("context done: %v", ctx.Err())
		case <-deadline:
			t.Fatalf("timeout: Merge received %d (expected >= 1), Split received %d", mergeReceived, splitReceived)
		case l := <-chMerge:
			if l.Address == (common.Address{}) && l.BlockNumber > 0 {
				continue
			}
			if len(l.Topics) > 0 && l.Topics[0] == mergeTopic0 {
				mergeReceived++
				t.Logf("Merge log: block=%d tx=%s", l.BlockNumber, l.TxHash.Hex())
			}
		case l := <-chSplit:
			if l.Address == (common.Address{}) && l.BlockNumber > 0 {
				continue
			}
			if len(l.Topics) > 0 && l.Topics[0] == splitTopic0 {
				splitReceived++
			}
		case err := <-subMerge.Err():
			t.Fatalf("Merge sub err: %v", err)
		case err := <-subSplit.Err():
			t.Fatalf("Split sub err: %v", err)
		}
	}
	require.GreaterOrEqual(t, mergeReceived, 1, "Merge subscription must receive at least one log (reproduce missed-dispatch if 0)")
}

// TestSubscribeMergeRealtime_13SubscriptionsProductionLike reproduces prediction-exchange
// exactly: 13 realtime subscriptions in one partition (3 CT × Split/Merge/Redeem + 4 Exchange × OrderFilled),
// one merged eth_getLogs per cycle. Asserts the Merge subscription receives its log. Fails if missed dispatch.
func TestSubscribeMergeRealtime_13SubscriptionsProductionLike(t *testing.T) {
	rpcURL := os.Getenv("SUBSCRIBE_MERGE_TEST_RPC")
	if rpcURL == "" {
		t.Skip("SUBSCRIBE_MERGE_TEST_RPC not set")
	}
	ctAddr := os.Getenv("CT_ADDRESS")
	stakeholder := os.Getenv("STAKEHOLDER_ADDRESS")
	if ctAddr == "" || stakeholder == "" {
		t.Skip("CT_ADDRESS and STAKEHOLDER_ADDRESS required")
	}

	handler := log.NewTerminalHandlerWithLevel(os.Stdout, log.LevelDebug, true)
	log.SetDefault(log.NewLogger(handler))

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	rpcCli, err := rpc.DialContext(ctx, rpcURL)
	require.NoError(t, err)
	defer rpcCli.Close()

	chainId, err := getChainID(rpcCli)
	require.NoError(t, err)
	storage := subscriber.NewMemoryStorage(chainId)
	cs, err := subscriber.NewChainSubscriber(rpcCli, storage)
	require.NoError(t, err)
	defer cs.Close()

	cs.SetRetryInterval(500 * time.Millisecond)
	cs.SetBlockConfirmationsOnSubscription(0)

	ctAddrCommon := common.HexToAddress(ctAddr)
	stakeholderAddr := common.HexToAddress(stakeholder)
	stakeholderTopic := common.BytesToHash(common.LeftPadBytes(stakeholderAddr.Bytes(), 32))

	topic0Split := common.HexToHash(positionSplitTopic0Hex)
	topic0Merge := common.HexToHash(positionsMergeTopic0Hex)
	topic0Redeem := common.HexToHash(payoutRedemptionTopic0Hex)
	topic0OrderFilled := common.HexToHash(orderFilledTopic0Hex)

	fromBlock := (*big.Int)(nil)
	if fromBlockStr := os.Getenv("FROM_BLOCK"); fromBlockStr != "" {
		fromU64, err := strconv.ParseUint(fromBlockStr, 10, 64)
		if err == nil {
			fromBlock = big.NewInt(0).SetUint64(fromU64)
		}
	}

	// Build 13 FilterQueries exactly as production (predict_exchange.go startSubscriptions):
	// 3 CT contracts × (PositionSplit, PositionsMerge, PayoutRedemption) + 4 Exchange × WatchOrderFilled.
	// Use 3 distinct CT addresses so we get 13 distinct query groups → one merged eth_getLogs with mergedQueries=13.
	// (On BNB all 3 can be same address; here we force 13 groups to match "13 subscriptions merged".)
	ctAddrs := [3]common.Address{
		ctAddrCommon,
		common.HexToAddress("0x0000000000000000000000000000000000000010"), // CT2 placeholder
		common.HexToAddress("0x0000000000000000000000000000000000000011"), // CT3 placeholder
	}
	queries := make([]ethereum.FilterQuery, 0, 13)
	// 9 CT subscriptions: 3 contracts × 3 events (Split, Merge, Redeem)
	for i := 0; i < 3; i++ {
		ctAddr := ctAddrs[i]
		queries = append(queries,
			ethereum.FilterQuery{
				Addresses: []common.Address{ctAddr},
				Topics:    [][]common.Hash{{topic0Split}, {stakeholderTopic}, nil, nil},
				FromBlock: fromBlock, ToBlock: nil,
			},
			ethereum.FilterQuery{
				Addresses: []common.Address{ctAddr},
				Topics:    [][]common.Hash{{topic0Merge}, {stakeholderTopic}, nil, nil},
				FromBlock: fromBlock, ToBlock: nil,
			},
			ethereum.FilterQuery{
				Addresses: []common.Address{ctAddr},
				Topics:    [][]common.Hash{{topic0Redeem}, {stakeholderTopic}, nil, nil},
				FromBlock: fromBlock, ToBlock: nil,
			},
		)
	}
	// 4 Exchange subscriptions: OrderFilled (3 topics: topic0, maker, taker)
	for _, exAddr := range exchangePlaceholderAddrs {
		queries = append(queries, ethereum.FilterQuery{
			Addresses: []common.Address{exAddr},
			Topics:    [][]common.Hash{{topic0OrderFilled}, {stakeholderTopic}, nil},
			FromBlock: fromBlock, ToBlock: nil,
		})
	}
	require.Len(t, queries, 13, "must have exactly 13 production-like subscriptions")

	// First CT: sub 0=Split, 1=Merge, 2=Redeem. Assert Merge receives at least one log.
	const mergeSubIndex = 1
	chans := make([]chan types.Log, 13)
	subs := make([]ethereum.Subscription, 13)
	for i := range queries {
		chans[i] = make(chan types.Log, 16)
		sub, err := cs.SubscribeFilterLogs(ctx, queries[i], chans[i])
		require.NoError(t, err)
		subs[i] = sub
	}
	defer func() {
		for _, sub := range subs {
			sub.Unsubscribe()
		}
	}()

	// Single channel to collect events from all 13 so we can select once
	type subEvent struct {
		idx int
		log types.Log
	}
	events := make(chan subEvent, 64)
	for i := range chans {
		i := i
		go func() {
			for l := range chans[i] {
				select {
				case events <- subEvent{idx: i, log: l}:
				case <-ctx.Done():
					return
				}
			}
		}()
	}
	subErrCh := make(chan error, 1)
	for i := range subs {
		i := i
		go func() {
			if err := <-subs[i].Err(); err != nil {
				select {
				case subErrCh <- err:
				default:
				}
			}
		}()
	}

	var mergeReceived int
	deadline := time.After(60 * time.Second)
	for mergeReceived < 1 {
		select {
		case <-ctx.Done():
			t.Fatalf("context done: %v", ctx.Err())
		case err := <-subErrCh:
			t.Fatalf("subscription error: %v", err)
		case <-deadline:
			t.Fatalf("timeout: Merge subscription (index %d) received %d (expected >= 1) — missed dispatch reproduced", mergeSubIndex, mergeReceived)
		case e := <-events:
			if e.log.Address == (common.Address{}) && e.log.BlockNumber > 0 {
				continue
			}
			if e.idx == mergeSubIndex && len(e.log.Topics) > 0 && e.log.Topics[0] == topic0Merge {
				mergeReceived++
				t.Logf("Merge log on sub %d: block=%d tx=%s", e.idx, e.log.BlockNumber, e.log.TxHash.Hex())
			}
		}
	}
	require.GreaterOrEqual(t, mergeReceived, 1, "Merge subscription must receive at least one log (13-sub missed-dispatch reproduced if 0)")
}

// TestMergedFilterLogs_ProductionLike_AdapterMergeInResponse builds the exact same merged
// FilterQuery as production (19 queries: 3 CT × Split/Merge/Redeem + YBNegRiskAdapter × 3 + 4 Exchange × OrderFilled),
// calls FilterLogs(merged) for the block range that contains tx 0x79a5e872 (adapter PositionsMerge),
// and asserts the response includes the adapter's PositionsMerge log. Use to reproduce RPC/merge bugs locally.
//
// TestMergedFilterLogs_RPCAnalysis only calls RPC (eth_getLogs via FilterLogs) with merged
// filters of different sizes and logs request shape + response (log count, error). No assertions.
// Use to locate root cause: run with SUBSCRIBE_MERGE_TEST_RPC and inspect at which N the node returns 0 logs.
//
//	go test -v -count=1 -run TestMergedFilterLogs_RPCAnalysis ./tests/subscriber/
func TestMergedFilterLogs_RPCAnalysis(t *testing.T) {
	rpcURL := os.Getenv("SUBSCRIBE_MERGE_TEST_RPC")
	if rpcURL == "" {
		t.Skip("SUBSCRIBE_MERGE_TEST_RPC not set (e.g. https://learnverse.top/chain/evm/56)")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	rpcCli, err := rpc.DialContext(ctx, rpcURL)
	require.NoError(t, err)
	defer rpcCli.Close()
	ec := ethclient.NewClient(rpcCli)
	defer ec.Close()

	fromBlock := big.NewInt(83318201)
	toBlock := big.NewInt(83318217)
	funderTopic := common.BytesToHash(common.LeftPadBytes(knownMergeFunderAddr.Bytes(), 32))
	topic0Split := common.HexToHash(positionSplitTopic0Hex)
	topic0Merge := common.HexToHash(positionsMergeTopic0Hex)
	topic0Redeem := common.HexToHash(payoutRedemptionTopic0Hex)
	topic0OrderFilled := common.HexToHash(orderFilledTopic0Hex)
	adapterSplitT0 := common.HexToHash(ybNegRiskAdapterPositionSplitTopic0Hex)
	adapterMergeT0 := common.HexToHash(ybNegRiskAdapterPositionsMergeTopic0Hex)
	adapterRedeemT0 := common.HexToHash(ybNegRiskAdapterPayoutRedemptionTopic0Hex)
	ctAddrs := [3]common.Address{
		knownMergeYBNegRiskCTAddr,
		common.HexToAddress("0x0000000000000000000000000000000000000010"),
		common.HexToAddress("0x0000000000000000000000000000000000000011"),
	}
	var queries []ethereum.FilterQuery
	for i := 0; i < 3; i++ {
		addr := ctAddrs[i]
		queries = append(queries,
			ethereum.FilterQuery{FromBlock: fromBlock, ToBlock: toBlock, Addresses: []common.Address{addr}, Topics: [][]common.Hash{{topic0Split}, {funderTopic}, nil, nil}},
			ethereum.FilterQuery{FromBlock: fromBlock, ToBlock: toBlock, Addresses: []common.Address{addr}, Topics: [][]common.Hash{{topic0Merge}, {funderTopic}, nil, nil}},
			ethereum.FilterQuery{FromBlock: fromBlock, ToBlock: toBlock, Addresses: []common.Address{addr}, Topics: [][]common.Hash{{topic0Redeem}, {funderTopic}, nil, nil}},
		)
	}
	queries = append(queries,
		ethereum.FilterQuery{FromBlock: fromBlock, ToBlock: toBlock, Addresses: []common.Address{knownMergeYBNegRiskAdapterAddr}, Topics: [][]common.Hash{{adapterSplitT0}, {funderTopic}, nil}},
		ethereum.FilterQuery{FromBlock: fromBlock, ToBlock: toBlock, Addresses: []common.Address{knownMergeYBNegRiskAdapterAddr}, Topics: [][]common.Hash{{adapterMergeT0}, {funderTopic}, nil}},
		ethereum.FilterQuery{FromBlock: fromBlock, ToBlock: toBlock, Addresses: []common.Address{knownMergeYBNegRiskAdapterAddr}, Topics: [][]common.Hash{{adapterRedeemT0}, {funderTopic}, nil}},
	)
	negRiskAdapterAddr := common.HexToAddress("0x0000000000000000000000000000000000000012")
	queries = append(queries,
		ethereum.FilterQuery{FromBlock: fromBlock, ToBlock: toBlock, Addresses: []common.Address{negRiskAdapterAddr}, Topics: [][]common.Hash{{adapterSplitT0}, {funderTopic}, nil}},
		ethereum.FilterQuery{FromBlock: fromBlock, ToBlock: toBlock, Addresses: []common.Address{negRiskAdapterAddr}, Topics: [][]common.Hash{{adapterMergeT0}, {funderTopic}, nil}},
		ethereum.FilterQuery{FromBlock: fromBlock, ToBlock: toBlock, Addresses: []common.Address{negRiskAdapterAddr}, Topics: [][]common.Hash{{adapterRedeemT0}, {funderTopic}, nil}},
	)
	for _, exAddr := range exchangePlaceholderAddrs {
		queries = append(queries, ethereum.FilterQuery{
			FromBlock: fromBlock, ToBlock: toBlock,
			Addresses: []common.Address{exAddr},
			Topics:    [][]common.Hash{{topic0OrderFilled}, {funderTopic}, nil},
		})
	}
	require.Len(t, queries, 19)

	logFilterShape := func(name string, m ethereum.FilterQuery) {
		addrN := 0
		if m.Addresses != nil {
			addrN = len(m.Addresses)
		}
		topicLens := make([]interface{}, 0, 4)
		for i := 0; i < 4; i++ {
			if i >= len(m.Topics) || m.Topics[i] == nil {
				topicLens = append(topicLens, "nil")
			} else {
				topicLens = append(topicLens, len(m.Topics[i]))
			}
		}
		t.Logf("[RPC analysis] %s: addresses=%d topics=%v from=%s to=%s", name, addrN, topicLens, fromBlock.String(), toBlock.String())
	}

	doFilterLogs := func(name string, m ethereum.FilterQuery) (int, error) {
		logFilterShape(name, m)
		logs, err := ec.FilterLogs(ctx, m)
		if err != nil {
			t.Logf("[RPC analysis] %s: FilterLogs err=%v", name, err)
			return 0, err
		}
		t.Logf("[RPC analysis] %s: FilterLogs ok, log_count=%d", name, len(logs))
		return len(logs), nil
	}

	// Single query: only YBNegRiskAdapter PositionsMerge (block 83318207 has this log). If this returns 0, RPC is missing data for this range.
	adapterMergeOnly := ethereum.FilterQuery{
		FromBlock: fromBlock, ToBlock: toBlock,
		Addresses: []common.Address{knownMergeYBNegRiskAdapterAddr},
		Topics:    [][]common.Hash{{adapterMergeT0}, {funderTopic}, nil},
	}
	doFilterLogs("adapter_merge_only", adapterMergeOnly)

	// 2-query merge (CT Merge + Adapter Merge): 2 addrs, 2 topic0
	queriesMinimal := []ethereum.FilterQuery{
		{FromBlock: fromBlock, ToBlock: toBlock, Addresses: []common.Address{knownMergeYBNegRiskCTAddr}, Topics: [][]common.Hash{{topic0Merge}, {funderTopic}, nil, nil}},
		{FromBlock: fromBlock, ToBlock: toBlock, Addresses: []common.Address{knownMergeYBNegRiskAdapterAddr}, Topics: [][]common.Hash{{adapterMergeT0}, {funderTopic}, nil}},
	}
	mergedMinimal, _ := subscriber.MergeFilterQueries(queriesMinimal, fromBlock, toBlock)
	doFilterLogs("2_query_merge", mergedMinimal)

	// 3-query merge with exactly 2 topic0 (CT Merge + Adapter Merge + duplicate CT Merge so union topic0 stays 2)
	queries3TwoTopic0 := []ethereum.FilterQuery{
		{FromBlock: fromBlock, ToBlock: toBlock, Addresses: []common.Address{knownMergeYBNegRiskCTAddr}, Topics: [][]common.Hash{{topic0Merge}, {funderTopic}, nil, nil}},
		{FromBlock: fromBlock, ToBlock: toBlock, Addresses: []common.Address{knownMergeYBNegRiskAdapterAddr}, Topics: [][]common.Hash{{adapterMergeT0}, {funderTopic}, nil}},
		{FromBlock: fromBlock, ToBlock: toBlock, Addresses: []common.Address{knownMergeYBNegRiskCTAddr}, Topics: [][]common.Hash{{topic0Merge}, {funderTopic}, nil, nil}}, // duplicate topic0
	}
	merged3TwoT0, _ := subscriber.MergeFilterQueries(queries3TwoTopic0, fromBlock, toBlock)
	doFilterLogs("3_query_2_topic0_merge", merged3TwoT0)

	// 19-query merge
	merged19, _ := subscriber.MergeFilterQueries(queries, fromBlock, toBlock)
	doFilterLogs("19_query_merge", merged19)

	// Sweep N=4,6,8,...,18 to find threshold where RPC returns 0
	for _, n := range []int{4, 6, 8, 10, 12, 14, 16, 18} {
		chunk := queries[:n]
		mergedN, err := subscriber.MergeFilterQueries(chunk, fromBlock, toBlock)
		require.NoError(t, err)
		doFilterLogs(fmt.Sprintf("%d_query_merge", n), mergedN)
	}
}

// Transfer(address,address,uint256) topic0 on Polygon (and most EVM chains)
const transferTopic0Hex = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
// Approval(address,address,uint256)
const approvalTopic0Hex = "0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3c925"
// Deposit(address,uint256) WETH
const depositTopic0Hex = "0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c"

// Polygon mainnet WETH and USDC (used only to test eth_getLogs merge behavior on Polygon)
var polygonWETH = common.HexToAddress("0x7ceB23fD6bC0adD59E62ac25578270cFf1b9f619")
var polygonUSDC = common.HexToAddress("0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174")

// TestMergedFilterLogs_Polygon_MergeBehavior uses the same merge shapes as BSC (TestMergedFilterLogs_RPCAnalysis):
// single (1 addr, 1 topic0), 2_query (2 addr, 2 topic0), 3_query_2_topic0 (2 addr, 2 topic0), 4_query (2 addr, 3 topic0).
//
//	go test -v -count=1 -run TestMergedFilterLogs_Polygon_MergeBehavior ./tests/subscriber/
func TestMergedFilterLogs_Polygon_MergeBehavior(t *testing.T) {
	rpcURL := os.Getenv("POLYGON_RPC")
	if rpcURL == "" {
		t.Skip("POLYGON_RPC not set (e.g. https://learnverse.top/chain/evm/137)")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	rpcCli, err := rpc.DialContext(ctx, rpcURL)
	require.NoError(t, err)
	defer rpcCli.Close()
	ec := ethclient.NewClient(rpcCli)
	defer ec.Close()

	latest, err := ec.BlockNumber(ctx)
	require.NoError(t, err)
	window := 100
	if s := os.Getenv("POLYGON_BLOCK_WINDOW"); s != "" {
		if w, err := strconv.Atoi(s); err == nil && w > 0 && w <= 500 {
			window = w
		}
	}
	fromBlock := new(big.Int).SetUint64(latest - uint64(window*5))
	toBlock := new(big.Int).SetUint64(latest - uint64(window*4))
	transferT0 := common.HexToHash(transferTopic0Hex)
	approvalT0 := common.HexToHash(approvalTopic0Hex)
	depositT0 := common.HexToHash(depositTopic0Hex)

	logShape := func(name string, m ethereum.FilterQuery, logCount int, err error) {
		addrN := 0
		if m.Addresses != nil {
			addrN = len(m.Addresses)
		}
		topic0N := 0
		if len(m.Topics) > 0 && m.Topics[0] != nil {
			topic0N = len(m.Topics[0])
		}
		if err != nil {
			t.Logf("[Polygon] %s: addresses=%d topics0=%d log_count=err %v", name, addrN, topic0N, err)
		} else {
			t.Logf("[Polygon] %s: addresses=%d topics0=%d log_count=%d", name, addrN, topic0N, logCount)
		}
	}

	// Single: 1 addr, 1 topic0 (same shape as BSC adapter_merge_only)
	single := ethereum.FilterQuery{
		FromBlock: fromBlock, ToBlock: toBlock,
		Addresses: []common.Address{polygonWETH},
		Topics:    [][]common.Hash{{transferT0}},
	}
	logsSingle, err := ec.FilterLogs(ctx, single)
	logShape("single_1addr_1topic0", single, len(logsSingle), err)

	// 2_query: 2 addr, 2 topic0 (WETH Transfer + USDC Approval)
	twoQueries := []ethereum.FilterQuery{
		{FromBlock: fromBlock, ToBlock: toBlock, Addresses: []common.Address{polygonWETH}, Topics: [][]common.Hash{{transferT0}}},
		{FromBlock: fromBlock, ToBlock: toBlock, Addresses: []common.Address{polygonUSDC}, Topics: [][]common.Hash{{approvalT0}}},
	}
	merged2, err := subscriber.MergeFilterQueries(twoQueries, fromBlock, toBlock)
	require.NoError(t, err)
	logs2, err := ec.FilterLogs(ctx, merged2)
	logShape("2_query_merge", merged2, len(logs2), err)

	// 3_query with 2 topic0 (add duplicate so union topic0 still 2)
	threeQueries := []ethereum.FilterQuery{
		{FromBlock: fromBlock, ToBlock: toBlock, Addresses: []common.Address{polygonWETH}, Topics: [][]common.Hash{{transferT0}}},
		{FromBlock: fromBlock, ToBlock: toBlock, Addresses: []common.Address{polygonUSDC}, Topics: [][]common.Hash{{approvalT0}}},
		{FromBlock: fromBlock, ToBlock: toBlock, Addresses: []common.Address{polygonWETH}, Topics: [][]common.Hash{{transferT0}}},
	}
	merged3, err := subscriber.MergeFilterQueries(threeQueries, fromBlock, toBlock)
	require.NoError(t, err)
	logs3, err := ec.FilterLogs(ctx, merged3)
	logShape("3_query_2_topic0_merge", merged3, len(logs3), err)

	// 4_query: 2 addr, 3 topic0 (Transfer, Approval, Deposit)
	fourQueries := []ethereum.FilterQuery{
		{FromBlock: fromBlock, ToBlock: toBlock, Addresses: []common.Address{polygonWETH}, Topics: [][]common.Hash{{transferT0}}},
		{FromBlock: fromBlock, ToBlock: toBlock, Addresses: []common.Address{polygonUSDC}, Topics: [][]common.Hash{{transferT0}}},
		{FromBlock: fromBlock, ToBlock: toBlock, Addresses: []common.Address{polygonWETH}, Topics: [][]common.Hash{{approvalT0}}},
		{FromBlock: fromBlock, ToBlock: toBlock, Addresses: []common.Address{polygonWETH}, Topics: [][]common.Hash{{depositT0}}},
	}
	merged4, err := subscriber.MergeFilterQueries(fourQueries, fromBlock, toBlock)
	require.NoError(t, err)
	logs4, err := ec.FilterLogs(ctx, merged4)
	logShape("4_query_merge", merged4, len(logs4), err)
}

// Run: SUBSCRIBE_MERGE_TEST_RPC=https://learnverse.top/chain/evm/56 go test -v -count=1 -run TestMergedFilterLogs_ProductionLike_AdapterMergeInResponse ./tests/subscriber/
func TestMergedFilterLogs_ProductionLike_AdapterMergeInResponse(t *testing.T) {
	rpcURL := os.Getenv("SUBSCRIBE_MERGE_TEST_RPC")
	if rpcURL == "" {
		t.Skip("SUBSCRIBE_MERGE_TEST_RPC not set (e.g. https://learnverse.top/chain/evm/56)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rpcCli, err := rpc.DialContext(ctx, rpcURL)
	require.NoError(t, err)
	defer rpcCli.Close()

	ec := ethclient.NewClient(rpcCli)
	defer ec.Close()

	// Block range that contains the known Merge tx 0x79a5e872 (block 83318207)
	fromBlock := big.NewInt(83318201)
	toBlock := big.NewInt(83318217)
	if s := os.Getenv("FROM_BLOCK"); s != "" {
		u, err := strconv.ParseUint(s, 10, 64)
		if err == nil {
			fromBlock = big.NewInt(0).SetUint64(u)
		}
	}
	if s := os.Getenv("TO_BLOCK"); s != "" {
		u, err := strconv.ParseUint(s, 10, 64)
		if err == nil {
			toBlock = big.NewInt(0).SetUint64(u)
		}
	}

	funderTopic := common.BytesToHash(common.LeftPadBytes(knownMergeFunderAddr.Bytes(), 32))
	topic0Split := common.HexToHash(positionSplitTopic0Hex)
	topic0Merge := common.HexToHash(positionsMergeTopic0Hex)
	topic0Redeem := common.HexToHash(payoutRedemptionTopic0Hex)
	topic0OrderFilled := common.HexToHash(orderFilledTopic0Hex)
	adapterSplitT0 := common.HexToHash(ybNegRiskAdapterPositionSplitTopic0Hex)
	adapterMergeT0 := common.HexToHash(ybNegRiskAdapterPositionsMergeTopic0Hex)
	adapterRedeemT0 := common.HexToHash(ybNegRiskAdapterPayoutRedemptionTopic0Hex)

	// 3 CT addresses (one real YBNegRisk CT, two placeholders), 1 YBNegRiskAdapter, 4 Exchange placeholders
	ctAddrs := [3]common.Address{
		knownMergeYBNegRiskCTAddr,
		common.HexToAddress("0x0000000000000000000000000000000000000010"),
		common.HexToAddress("0x0000000000000000000000000000000000000011"),
	}

	var queries []ethereum.FilterQuery
	// 9 CT: 3 contracts × (Split, Merge, Redeem)
	for i := 0; i < 3; i++ {
		addr := ctAddrs[i]
		queries = append(queries,
			ethereum.FilterQuery{FromBlock: fromBlock, ToBlock: toBlock, Addresses: []common.Address{addr}, Topics: [][]common.Hash{{topic0Split}, {funderTopic}, nil, nil}},
			ethereum.FilterQuery{FromBlock: fromBlock, ToBlock: toBlock, Addresses: []common.Address{addr}, Topics: [][]common.Hash{{topic0Merge}, {funderTopic}, nil, nil}},
			ethereum.FilterQuery{FromBlock: fromBlock, ToBlock: toBlock, Addresses: []common.Address{addr}, Topics: [][]common.Hash{{topic0Redeem}, {funderTopic}, nil, nil}},
		)
	}
	// 3 YBNegRiskAdapter: Split, Merge, Redeem (stakeholder=funder, conditionId=nil → topic2 nil)
	queries = append(queries,
		ethereum.FilterQuery{FromBlock: fromBlock, ToBlock: toBlock, Addresses: []common.Address{knownMergeYBNegRiskAdapterAddr}, Topics: [][]common.Hash{{adapterSplitT0}, {funderTopic}, nil}},
		ethereum.FilterQuery{FromBlock: fromBlock, ToBlock: toBlock, Addresses: []common.Address{knownMergeYBNegRiskAdapterAddr}, Topics: [][]common.Hash{{adapterMergeT0}, {funderTopic}, nil}},
		ethereum.FilterQuery{FromBlock: fromBlock, ToBlock: toBlock, Addresses: []common.Address{knownMergeYBNegRiskAdapterAddr}, Topics: [][]common.Hash{{adapterRedeemT0}, {funderTopic}, nil}},
	)
	// 3 NegRiskAdapter (same topic0s as YBNegRiskAdapter, different contract address)
	negRiskAdapterAddr := common.HexToAddress("0x0000000000000000000000000000000000000012")
	queries = append(queries,
		ethereum.FilterQuery{FromBlock: fromBlock, ToBlock: toBlock, Addresses: []common.Address{negRiskAdapterAddr}, Topics: [][]common.Hash{{adapterSplitT0}, {funderTopic}, nil}},
		ethereum.FilterQuery{FromBlock: fromBlock, ToBlock: toBlock, Addresses: []common.Address{negRiskAdapterAddr}, Topics: [][]common.Hash{{adapterMergeT0}, {funderTopic}, nil}},
		ethereum.FilterQuery{FromBlock: fromBlock, ToBlock: toBlock, Addresses: []common.Address{negRiskAdapterAddr}, Topics: [][]common.Hash{{adapterRedeemT0}, {funderTopic}, nil}},
	)
	// 4 Exchange: OrderFilled (maker=funder)
	for _, exAddr := range exchangePlaceholderAddrs {
		queries = append(queries, ethereum.FilterQuery{
			FromBlock: fromBlock, ToBlock: toBlock,
			Addresses: []common.Address{exAddr},
			Topics:    [][]common.Hash{{topic0OrderFilled}, {funderTopic}, nil},
		})
	}
	require.Len(t, queries, 19, "production-like 19 queries")

	// Step 1: minimal merge (CT Merge + Adapter Merge only) must return adapter log — proves RPC and merge work for small request
	queriesMinimal := []ethereum.FilterQuery{
		{FromBlock: fromBlock, ToBlock: toBlock, Addresses: []common.Address{knownMergeYBNegRiskCTAddr}, Topics: [][]common.Hash{{topic0Merge}, {funderTopic}, nil, nil}},
		{FromBlock: fromBlock, ToBlock: toBlock, Addresses: []common.Address{knownMergeYBNegRiskAdapterAddr}, Topics: [][]common.Hash{{adapterMergeT0}, {funderTopic}, nil}},
	}
	mergedMinimal, err := subscriber.MergeFilterQueries(queriesMinimal, fromBlock, toBlock)
	require.NoError(t, err)
	logsMinimal, err := ec.FilterLogs(ctx, mergedMinimal)
	require.NoError(t, err)
	var foundMinimal bool
	for _, l := range logsMinimal {
		if l.Address == knownMergeYBNegRiskAdapterAddr && len(l.Topics) > 0 && l.Topics[0] == adapterMergeT0 {
			foundMinimal = true
			t.Logf("minimal merge (2 queries): adapter PositionsMerge log found, block=%d tx=%s", l.BlockNumber, l.TxHash.Hex())
			break
		}
	}
	require.True(t, foundMinimal, "minimal merged FilterLogs must include adapter PositionsMerge (sanity check); got %d logs", len(logsMinimal))

	// Step 2: full production-like 19-query merge — reproduce whether RPC returns adapter log or 0/subset
	merged, err := subscriber.MergeFilterQueries(queries, fromBlock, toBlock)
	require.NoError(t, err)

	logs, err := ec.FilterLogs(ctx, merged)
	require.NoError(t, err)

	var found bool
	for i := range logs {
		l := &logs[i]
		if l.Address == knownMergeYBNegRiskAdapterAddr && len(l.Topics) > 0 && l.Topics[0] == adapterMergeT0 {
			found = true
			t.Logf("19-query merge: adapter PositionsMerge log found: block=%d tx=%s logIndex=%d", l.BlockNumber, l.TxHash.Hex(), l.Index)
			break
		}
	}
	require.True(t, found, "merged FilterLogs (19 queries) response must include YBNegRiskAdapter PositionsMerge (block 83318207, tx 0x79a5e872); got %d logs total (if 0, RPC may limit large filter)", len(logs))
}

func getChainID(rpcCli *rpc.Client) (*big.Int, error) {
	var hex string
	if err := rpcCli.CallContext(context.Background(), &hex, "eth_chainId"); err != nil {
		return nil, err
	}
	s := hex
	if len(s) > 2 && (s[0] == '0' && (s[1] == 'x' || s[1] == 'X')) {
		s = s[2:]
	}
	var chainId big.Int
	if _, ok := chainId.SetString(s, 16); !ok {
		return nil, errors.New("invalid eth_chainId")
	}
	return &chainId, nil
}
