package subscriber

import (
	"context"
	"fmt"
	"math/big"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ivanzzeth/ethclient/common/consts"
	"github.com/ivanzzeth/ethclient/types"
)

var _ Subscriber = (*ChainSubscriber)(nil)

var _ ethereum.LogFilterer = (*ChainSubscriber)(nil)

// realtimeEntry holds one realtime subscription for the merged scanner.
type realtimeEntry struct {
	query ethereum.FilterQuery
	ch    chan<- etypes.Log
}

// ChainSubscriber implements Subscriber interface
type ChainSubscriber struct {
	c                                *ethclient.Client
	logFilterer                       LogFiltererBatch
	geth                             *gethclient.Client
	chainId                          *big.Int
	retryInterval                    time.Duration
	buffer                           int
	fetchMissingHeaders              bool
	blocksPerScan                    uint64
	currBlocksPerScan                uint64 // adjust dynamiclly
	maxBlocksPerScan                 uint64
	blockConfirmationsOnSubscription uint64
	storage                          SubscriberStorage

	queryCtx           context.Context
	cancelQueryCtx     context.CancelFunc
	queryHandler       QueryHandler
	queryMap           sync.Map
	globalLogsChannels sync.Map

	realtimeMu           sync.Mutex
	realtimeQueries      map[common.Hash][]*realtimeEntry // same query can have multiple subscribers (channels)
	realtimeScannerStart sync.Once
	maxQueriesPerMerge   int // if > 0, split partition into batches of this size to avoid RPC returning 0 (e.g. BSC); 0 = no limit
}

// NewChainSubscriber .
func NewChainSubscriber(rpcCli *rpc.Client, storage SubscriberStorage) (*ChainSubscriber, error) {
	c := ethclient.NewClient(rpcCli)
	geth := gethclient.New(rpcCli)
	chainId, err := c.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	queryCtx, cancel := context.WithCancel(context.Background())

	subscriber := &ChainSubscriber{
		c:                 c,
		logFilterer:       NewBatchLogFilterer(c, rpcCli),
		geth:              geth,
		chainId:           chainId,
		buffer:            consts.DefaultMsgBuffer,
		blocksPerScan:     consts.DefaultBlocksPerScan,
		currBlocksPerScan: consts.DefaultBlocksPerScan,
		maxBlocksPerScan:  consts.MaxBlocksPerScan,
		retryInterval:     consts.RetryInterval,
		storage:           storage,
		queryCtx:          queryCtx,
		cancelQueryCtx:    cancel,
		realtimeQueries:   make(map[common.Hash][]*realtimeEntry),
	}

	return subscriber, nil
}

func (s *ChainSubscriber) Close() {
	log.Debug("close subscriber...")
	s.cancelQueryCtx()

	// s.queryMap.Range(func(key, _ any) bool {
	// 	queryHash := key.(common.Hash)
	// 	ch := s.getQueryLogChannel(queryHash)
	// 	close(ch)

	// 	return true
	// })
}

func (s *ChainSubscriber) SetBlocksPerScan(blocksPerScan uint64) {
	s.blocksPerScan = blocksPerScan
}

func (s *ChainSubscriber) SetMaxBlocksPerScan(maxBlocksPerScan uint64) {
	s.maxBlocksPerScan = maxBlocksPerScan
}

func (s *ChainSubscriber) SetRetryInterval(retryInterval time.Duration) {
	s.retryInterval = retryInterval
}

// SetMaxQueriesPerMerge sets the maximum number of queries to merge into one eth_getLogs.
// When > 0 and a partition has more queries, they are split into batches; each batch is
// requested separately and results are merged and sorted. Use 2 on BSC to avoid nodes
// returning 0 for large merged filters. 0 = no limit (default).
func (cs *ChainSubscriber) SetMaxQueriesPerMerge(n int) {
	cs.realtimeMu.Lock()
	defer cs.realtimeMu.Unlock()
	cs.maxQueriesPerMerge = n
}

type subscription struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func (s *subscription) Unsubscribe() {
	s.cancel()
}

func (s *subscription) Err() <-chan error {
	errChan := make(chan error)
	go func() {
		<-s.ctx.Done()
		errChan <- s.ctx.Err()
	}()

	return errChan
}

func (cs *ChainSubscriber) GetBlockConfirmationsOnSubscription() uint64 {
	return cs.blockConfirmationsOnSubscription
}

func (cs *ChainSubscriber) SetBlockConfirmationsOnSubscription(confirmations uint64) {
	cs.blockConfirmationsOnSubscription = confirmations
}

func (cs *ChainSubscriber) SetBuffer(buffer int) {
	cs.buffer = buffer
}

func (cs *ChainSubscriber) SetQueryHandler(handler QueryHandler) {
	cs.queryHandler = handler
}

func (cs *ChainSubscriber) GetQueryHandler() QueryHandler {
	return cs.queryHandler
}

func (cs *ChainSubscriber) isQueryHandlerSet() bool {
	return cs.queryHandler != nil
}

func (cs *ChainSubscriber) SetFetchMissingHeaders(enable bool) {
	cs.fetchMissingHeaders = enable
}

func (cs *ChainSubscriber) SubmitQuery(query ethereum.FilterQuery) error {
	queryHash := GetQueryHash(cs.chainId, query)
	log.Info("submit query", "queryHash", queryHash, "query", query, "client", fmt.Sprintf("%p", cs.c))

	if !cs.isQueryHandlerSet() {
		return fmt.Errorf("setup queryHandler before calling it")
	}

	once, loaded := cs.queryMap.LoadOrStore(queryHash, &sync.Once{})
	queryOnce := once.(*sync.Once)
	if loaded {
		return fmt.Errorf("query already submitted")
	}

	globalLogsChannel := cs.getQueryLogChannel(queryHash)

	queryOnce.Do(func() {
		if query.ToBlock == nil {
			// Realtime: register and use merged scanner (one eth_getLogs per cycle for mergeable queries).
			cs.realtimeMu.Lock()
			cs.realtimeQueries[queryHash] = append(cs.realtimeQueries[queryHash], &realtimeEntry{query: query, ch: globalLogsChannel})
			cs.realtimeMu.Unlock()
			cs.startRealtimeScanner()
			go cs.handleQueryLogsChannel(query, globalLogsChannel)
			return
		}
		for {
			err := cs.FilterLogsWithChannel(cs.queryCtx, query, globalLogsChannel, true, true)
			if err != nil {
				log.Warn("submit query subscription failed, waiting for retrying", "err", err,
					"queryHash", queryHash)
				time.Sleep(cs.retryInterval)
				continue
			}
			break
		}
		go cs.handleQueryLogsChannel(query, globalLogsChannel)
	})

	return nil
}

func (cs *ChainSubscriber) handleQueryLogsChannel(query ethereum.FilterQuery, ch <-chan etypes.Log) {
	for l := range ch {
		err := cs.queryHandler.HandleQuery(context.Background(), NewQuery(cs.chainId, query), l)
		if err != nil {
			log.Warn("handle query failed", "err", err, "queryHash", GetQueryHash(cs.chainId, query))
		}
	}
}

func (cs *ChainSubscriber) getQueryLogChannel(queryHash common.Hash) chan etypes.Log {
	ch, _ := cs.globalLogsChannels.LoadOrStore(queryHash, make(chan etypes.Log, cs.buffer))
	globalLogsChannel := ch.(chan etypes.Log)

	return globalLogsChannel
}

func (cs *ChainSubscriber) startRealtimeScanner() {
	cs.realtimeScannerStart.Do(func() {
		go cs.runRealtimeScanner()
	})
}

// runRealtimeScanner runs one loop per chain: collect all realtime queries, one FilterLogsBatch, then dispatch.
// realtimeAdvanceProgress returns whether to advance persisted block progress after a merged FilterLogs call.
// When the node returns 0 logs we must not advance, so the same range is retried and we do not skip events (e.g. Split/Merge).
func realtimeAdvanceProgress(mergedLogCount int) bool {
	return mergedLogCount > 0
}

// nextBlockToScanForRealtime returns the next block number to scan (inclusive) for a realtime query.
// When storage returns 0 for LatestBlockForQuery, next would be 1; this enforces that we never scan
// before query.FromBlock (subscription start), avoiding scanning from chain genesis and missing events.
func nextBlockToScanForRealtime(latest uint64, q ethereum.FilterQuery) uint64 {
	next := latest + 1
	if q.FromBlock != nil {
		from := q.FromBlock.Uint64()
		if next < from {
			next = from
		}
	}
	return next
}

func (cs *ChainSubscriber) runRealtimeScanner() {
	ctx := cs.queryCtx
	var lastBlockAtomic atomic.Uint64
	go func() {
		for {
			lastBlock, err := cs.c.BlockNumber(ctx)
			if err != nil {
				log.Warn("realtime scanner BlockNumber failed", "err", err)
				time.Sleep(cs.retryInterval)
				continue
			}
			lastBlockAtomic.Store(lastBlock)
			time.Sleep(cs.retryInterval)
		}
	}()

	reduceBlocksPerScan := false
	currBlocks := cs.currBlocksPerScan

	updateScan := func() {
		if reduceBlocksPerScan {
			reduceBlocksPerScan = false
			if currBlocks > cs.blocksPerScan {
				currBlocks = cs.blocksPerScan
			}
		} else {
			if currBlocks < cs.maxBlocksPerScan {
				currBlocks *= 2
				if currBlocks > cs.maxBlocksPerScan {
					currBlocks = cs.maxBlocksPerScan
				}
			}
		}
	}

	queryStateReader := cs.storage
	queryStateWriter := cs.storage
	if cs.isQueryHandlerSet() {
		queryStateReader = cs.queryHandler
		queryStateWriter = cs.queryHandler
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		cs.realtimeMu.Lock()
		// Group by queryHash so same query has one batch request and one storage update, but multiple channels.
		type queryGroup struct {
			hash   common.Hash
			query  ethereum.FilterQuery
			entries []*realtimeEntry
		}
		groups := make([]queryGroup, 0, len(cs.realtimeQueries))
		for hash, entries := range cs.realtimeQueries {
			if len(entries) == 0 {
				continue
			}
			groups = append(groups, queryGroup{hash: hash, query: entries[0].query, entries: entries})
		}
		cs.realtimeMu.Unlock()

		if len(groups) == 0 {
			time.Sleep(cs.retryInterval)
			continue
		}

		lastBlock := lastBlockAtomic.Load()
		if lastBlock < cs.blockConfirmationsOnSubscription {
			time.Sleep(cs.retryInterval)
			continue
		}
		lastBlock -= cs.blockConfirmationsOnSubscription

		minStart := uint64(0)
		for _, g := range groups {
			latest, err := queryStateReader.LatestBlockForQuery(ctx, g.query)
			if err != nil {
				log.Warn("realtime scanner LatestBlockForQuery failed", "err", err, "queryHash", g.hash)
				time.Sleep(cs.retryInterval)
				continue
			}
			next := nextBlockToScanForRealtime(latest, g.query)
			if minStart == 0 || next < minStart {
				minStart = next
			}
		}
		if minStart > lastBlock {
			time.Sleep(cs.retryInterval)
			continue
		}

		endBlock := minStart + currBlocks - 1
		if endBlock > lastBlock {
			endBlock = lastBlock
		}
		if minStart > endBlock {
			time.Sleep(cs.retryInterval)
			continue
		}
		startBlock := minStart
		fromBlock := big.NewInt(0).SetUint64(startBlock)
		toBlock := big.NewInt(0).SetUint64(endBlock)

		// Greedy partition by (BlockHash, FromBlock, ToBlock): same key → one merged eth_getLogs.
		// Realtime queries have nil range; fallback fromBlock/toBlock gives one range partition.
		// Ensures no extra/missing logs per original query (union filter + LogMatchesQuery dispatch).
		partition := make(map[string][]queryGroup)
		for _, g := range groups {
			key := GetPartitionKey(g.query, fromBlock, toBlock)
			partition[key] = append(partition[key], g)
		}

		for partKey, groupList := range partition {
			if len(groupList) == 0 {
				continue
			}
			var merged ethereum.FilterQuery
			var mergeErr error
			if strings.HasPrefix(partKey, "H:") {
				blockHash := common.HexToHash(strings.TrimPrefix(partKey, "H:"))
				queries := make([]ethereum.FilterQuery, len(groupList))
				for i, g := range groupList {
					queries[i] = g.query
				}
				merged, mergeErr = MergeFilterQueriesByBlockHash(queries, blockHash)
			} else {
				// Range partition (key "R:from:to"); use cycle fromBlock/toBlock.
				queries := make([]ethereum.FilterQuery, len(groupList))
				for i, g := range groupList {
					queries[i] = ethereum.FilterQuery{
						FromBlock: fromBlock, ToBlock: toBlock,
						Addresses: g.query.Addresses, Topics: g.query.Topics,
					}
				}
				merged, mergeErr = MergeFilterQueries(queries, fromBlock, toBlock)
			}
			if mergeErr != nil {
				log.Warn("realtime scanner merge failed", "err", mergeErr, "partitionKey", partKey)
				time.Sleep(cs.retryInterval)
				continue
			}
			cs.realtimeMu.Lock()
			maxMerge := cs.maxQueriesPerMerge
			cs.realtimeMu.Unlock()

			var mergedLogs []etypes.Log
			useBatch := !strings.HasPrefix(partKey, "H:") && maxMerge > 0 && len(groupList) > maxMerge
			if useBatch {
				var allLogs []etypes.Log
				var batchErr error
				for start := 0; start < len(groupList); start += maxMerge {
					end := start + maxMerge
					if end > len(groupList) {
						end = len(groupList)
					}
					chunk := groupList[start:end]
					qChunk := make([]ethereum.FilterQuery, len(chunk))
					for i, g := range chunk {
						qChunk[i] = ethereum.FilterQuery{
							FromBlock: fromBlock, ToBlock: toBlock,
							Addresses: g.query.Addresses, Topics: g.query.Topics,
						}
					}
					m, mergeErr := MergeFilterQueries(qChunk, fromBlock, toBlock)
					if mergeErr != nil {
						log.Warn("realtime scanner batch merge failed", "err", mergeErr, "partitionKey", partKey)
						batchErr = mergeErr
						break
					}
					logs, err := cs.logFilterer.FilterLogs(ctx, m)
					if err != nil {
						log.Warn("realtime scanner FilterLogs failed (batch)", "err", err, "partitionKey", partKey)
						reduceBlocksPerScan = true
						batchErr = err
						break
					}
					allLogs = append(allLogs, logs...)
				}
				if batchErr != nil {
					time.Sleep(cs.retryInterval)
					continue
				}
				mergedLogs = allLogs
				slices.SortFunc(mergedLogs, func(a, b etypes.Log) int {
					if a.BlockNumber != b.BlockNumber {
						if a.BlockNumber < b.BlockNumber {
							return -1
						}
						return 1
					}
					if a.TxIndex != b.TxIndex {
						if a.TxIndex < b.TxIndex {
							return -1
						}
						return 1
					}
					if a.Index != b.Index {
						if a.Index < b.Index {
							return -1
						}
						return 1
					}
					return 0
				})
				log.Debug("realtime scanner FilterLogs (batched)", "partitionKey", partKey, "mergedQueries", len(groupList), "batches", (len(groupList)+maxMerge-1)/maxMerge, "logs", len(mergedLogs))
			} else {
				fromStr, toStr := "nil", "nil"
				if merged.FromBlock != nil {
					fromStr = merged.FromBlock.String()
				}
				if merged.ToBlock != nil {
					toStr = merged.ToBlock.String()
				}
				addrStrs := make([]string, 0, len(merged.Addresses))
				for _, a := range merged.Addresses {
					addrStrs = append(addrStrs, a.Hex())
				}
				topic0Strs := []string(nil)
				if len(merged.Topics) > 0 && merged.Topics[0] != nil {
					topic0Strs = make([]string, 0, len(merged.Topics[0]))
					for _, h := range merged.Topics[0] {
						topic0Strs = append(topic0Strs, h.Hex())
					}
				}
				topic1Strs := []string(nil)
				if len(merged.Topics) > 1 && merged.Topics[1] != nil {
					topic1Strs = make([]string, 0, len(merged.Topics[1]))
					for _, h := range merged.Topics[1] {
						topic1Strs = append(topic1Strs, h.Hex())
					}
				}
				topic2Strs := []string(nil)
				if len(merged.Topics) > 2 && merged.Topics[2] != nil {
					topic2Strs = make([]string, 0, len(merged.Topics[2]))
					for _, h := range merged.Topics[2] {
						topic2Strs = append(topic2Strs, h.Hex())
					}
				}
				topic3Strs := []string(nil)
				if len(merged.Topics) > 3 && merged.Topics[3] != nil {
					topic3Strs = make([]string, 0, len(merged.Topics[3]))
					for _, h := range merged.Topics[3] {
						topic3Strs = append(topic3Strs, h.Hex())
					}
				}
				log.Debug("realtime scanner FilterLogs request (merged query)",
					"partitionKey", partKey,
					"fromBlock", fromStr, "toBlock", toStr,
					"addresses", addrStrs,
					"topics0", topic0Strs, "topics1", topic1Strs, "topics2", topic2Strs, "topics3", topic3Strs,
					"mergedQueries", len(groupList))
				var err error
				mergedLogs, err = cs.logFilterer.FilterLogs(ctx, merged)
				if err != nil {
					log.Warn("realtime scanner FilterLogs failed", "err", err, "partitionKey", partKey)
					reduceBlocksPerScan = true
					time.Sleep(cs.retryInterval)
					continue
				}
				log.Debug("realtime scanner FilterLogs (merged)", "partitionKey", partKey, "mergedQueries", len(groupList), "logs", len(mergedLogs))
			}
			for j := range mergedLogs {
				l := &mergedLogs[j]
				t0 := common.Hash{}
				if len(l.Topics) > 0 {
					t0 = l.Topics[0]
				}
				log.Info("realtime scanner log received", "partitionKey", partKey, "block", l.BlockNumber, "txHash", l.TxHash.Hex(), "address", l.Address.Hex(), "topic0", t0.Hex())
			}
			// When RPC returns 0 logs do not advance progress: avoid permanently skipping blocks (Split/Merge events) on transient empty response or node drop.
			advanceProgress := realtimeAdvanceProgress(len(mergedLogs))
			// delivered[i] true if mergedLogs[i] was sent to at least one subscription (for missed-event diagnosis).
			delivered := make([]bool, len(mergedLogs))
			for _, g := range groupList {
				q := g.query
				latestHandledLog, err := queryStateReader.LatestLogForQuery(ctx, q)
				if err != nil {
					log.Warn("realtime scanner LatestLogForQuery failed", "err", err, "queryHash", g.hash)
					continue
				}
				var logsForGroup []etypes.Log
				for i := range mergedLogs {
					l := &mergedLogs[i]
					if !LogMatchesQuery(l, q) {
						continue
					}
					delivered[i] = true // matched this group (dedup skip still counts as "claimed")
					if latestHandledLog.BlockNumber > l.BlockNumber {
						continue
					}
					if latestHandledLog.BlockNumber == l.BlockNumber {
						if latestHandledLog.TxIndex > l.TxIndex {
							continue
						}
						if latestHandledLog.TxIndex == l.TxIndex && latestHandledLog.Index >= l.Index {
							continue
						}
					}
					logsForGroup = append(logsForGroup, *l)
					for _, e := range g.entries {
						select {
						case e.ch <- *l:
						case <-ctx.Done():
							return
						}
					}
					if queryStateWriter != nil && !cs.isQueryHandlerSet() {
						if err := queryStateWriter.SaveLatestLogForQuery(ctx, q, *l); err != nil {
							log.Error("realtime scanner SaveLatestLogForQuery failed", "err", err, "queryHash", g.hash)
						}
					}
				}
				// Diagnose missed dispatch: only warn when this group's filter (topic0/address) matches a log in the batch but got 0 → LogMatchesQuery bug. Applies to any event type (Split/Merge/Redeem/OrderFilled).
				if len(mergedLogs) > 0 && len(logsForGroup) == 0 {
					var queryTopic0 common.Hash
					if len(q.Topics) > 0 && len(q.Topics[0]) > 0 {
						queryTopic0 = q.Topics[0][0]
					}
					addrMatch := len(q.Addresses) == 0 || (len(q.Addresses) > 0 && slices.Contains(q.Addresses, mergedLogs[0].Address))
					topic0Match := len(mergedLogs[0].Topics) > 0 && mergedLogs[0].Topics[0] == queryTopic0
					if addrMatch && topic0Match {
						l0 := &mergedLogs[0]
						q1, q2, q3 := common.Hash{}, common.Hash{}, common.Hash{}
						if len(q.Topics) > 1 && len(q.Topics[1]) > 0 {
							q1 = q.Topics[1][0]
						}
						if len(q.Topics) > 2 && len(q.Topics[2]) > 0 {
							q2 = q.Topics[2][0]
						}
						if len(q.Topics) > 3 && len(q.Topics[3]) > 0 {
							q3 = q.Topics[3][0]
						}
						l1, l2, l3 := common.Hash{}, common.Hash{}, common.Hash{}
						if len(l0.Topics) > 1 {
							l1 = l0.Topics[1]
						}
						if len(l0.Topics) > 2 {
							l2 = l0.Topics[2]
						}
						if len(l0.Topics) > 3 {
							l3 = l0.Topics[3]
						}
						log.Warn("realtime scanner group got 0 logs but a log in batch matches query (missed dispatch)",
							"partitionKey", partKey, "queryHash", g.hash, "queryTopic0", queryTopic0.Hex(), "queryTopic1", q1.Hex(), "queryTopic2", q2.Hex(), "queryTopic3", q3.Hex(),
							"logBlock", l0.BlockNumber, "logTx", l0.TxHash.Hex(), "logTopic1", l1.Hex(), "logTopic2", l2.Hex(), "logTopic3", l3.Hex())
					}
				}
				if advanceProgress && queryStateWriter != nil && (!cs.isQueryHandlerSet() || len(logsForGroup) == 0) {
					if cs.isQueryHandlerSet() && len(logsForGroup) == 0 {
						for _, e := range g.entries {
							select {
							case e.ch <- etypes.Log{BlockNumber: endBlock}:
							case <-ctx.Done():
								return
							}
						}
					} else if err := queryStateWriter.SaveLatestBlockForQuery(ctx, q, endBlock); err != nil {
						log.Error("realtime scanner SaveLatestBlockForQuery failed", "err", err, "queryHash", g.hash)
					}
				}
			}
			// Log any log returned by RPC that matched no subscription (dispatch miss). Applies to all event types (Split/Merge/Redeem/OrderFilled).
			for i := range mergedLogs {
				if delivered[i] {
					continue
				}
				l := &mergedLogs[i]
				topic0 := common.Hash{}
				if len(l.Topics) > 0 {
					topic0 = l.Topics[0]
				}
				topic1, topic2, topic3 := common.Hash{}, common.Hash{}, common.Hash{}
				if len(l.Topics) > 1 {
					topic1 = l.Topics[1]
				}
				if len(l.Topics) > 2 {
					topic2 = l.Topics[2]
				}
				if len(l.Topics) > 3 {
					topic3 = l.Topics[3]
				}
				log.Warn("realtime scanner log matched no subscription (missed dispatch)",
					"partitionKey", partKey, "block", l.BlockNumber, "txHash", l.TxHash.Hex(),
					"address", l.Address.Hex(), "topic0", topic0.Hex(), "topic1", topic1.Hex(), "topic2", topic2.Hex(), "topic3", topic3.Hex(), "mergedQueries", len(groupList))
			}
		}

		updateScan()
		time.Sleep(cs.retryInterval)
	}
}

func (cs *ChainSubscriber) SubscribeFilterLogs(ctx context.Context, q ethereum.FilterQuery, ch chan<- etypes.Log) (sub ethereum.Subscription, err error) {
	log.Debug("SubscribeFilterlogs starts", "query", q)

	ctx, cancel := context.WithCancel(ctx)
	if q.ToBlock == nil {
		// Realtime: register and use merged scanner (one eth_getLogs per cycle for mergeable queries).
		// Same query can have multiple subscribers (different channels).
		queryHash := GetQueryHash(cs.chainId, q)
		entry := &realtimeEntry{query: q, ch: ch}
		cs.realtimeMu.Lock()
		cs.realtimeQueries[queryHash] = append(cs.realtimeQueries[queryHash], entry)
		cs.realtimeMu.Unlock()
		cs.startRealtimeScanner()
		go func() {
			<-ctx.Done()
			cs.realtimeMu.Lock()
			entries := cs.realtimeQueries[queryHash]
			for i, e := range entries {
				if e == entry {
					cs.realtimeQueries[queryHash] = append(entries[:i], entries[i+1:]...)
					if len(cs.realtimeQueries[queryHash]) == 0 {
						delete(cs.realtimeQueries, queryHash)
					}
					break
				}
			}
			cs.realtimeMu.Unlock()
		}()
		sub = &subscription{ctx, cancel}
		return sub, nil
	}
	err = cs.FilterLogsWithChannel(ctx, q, ch, true, true)
	sub = &subscription{ctx, cancel}
	return sub, err
}

func (cs *ChainSubscriber) FilterLogs(ctx context.Context, q ethereum.FilterQuery) (logs []etypes.Log, err error) {
	logsChan := make(chan etypes.Log, cs.buffer)
	err = cs.FilterLogsWithChannel(ctx, q, logsChan, false, true)
	if err != nil {
		return nil, err
	}

	for l := range logsChan {
		log.Debug("FilterLogs receiving log", "log", l)

		logs = append(logs, l)
	}

	return
}

// TODO:
// 3. cache all of finalized historical data, e.g., blockByHash, txByHash
func (cs *ChainSubscriber) FilterLogsWithChannel(ctx context.Context, q ethereum.FilterQuery, logsChan chan<- etypes.Log, watch bool, closeOnExit bool) (err error) {
	if q.BlockHash != nil {
		logs, err := cs.logFilterer.FilterLogs(ctx, q)
		if err != nil {
			return err
		}

		go func() {
			for _, l := range logs {
				logsChan <- l
			}
			if closeOnExit {
				close(logsChan)
			}
		}()

		return nil
	}

	fromBlock := uint64(0)
	if q.FromBlock != nil {
		fromBlock = q.FromBlock.Uint64()
	}

	toBlock := uint64(0)
	if q.ToBlock != nil {
		toBlock = q.ToBlock.Uint64()
	} else {
		toBlock, err = cs.c.BlockNumber(ctx)
		if err != nil {
			return err
		}
	}

	useStorage := q.ToBlock == nil
	var queryStateReader QueryStateReader = cs.storage
	var queryStateWriter QueryStateWriter = cs.storage
	if cs.isQueryHandlerSet() {
		queryStateReader = cs.queryHandler
		queryStateWriter = cs.queryHandler
	}

	startBlock := fromBlock

	if useStorage {
		fromBlockInStorage, err := queryStateReader.LatestBlockForQuery(ctx, q)
		if err != nil {
			return err
		}

		if fromBlockInStorage != 0 {
			log.Info("fromBlock in storage was found, then use it", "fromBlockInStorage", fromBlockInStorage)
			startBlock = fromBlockInStorage + 1
		}
	}

	endBlock := startBlock + cs.currBlocksPerScan

	query := NewQuery(cs.chainId, q)
	log.Debug("Subscriber FilterLogs starts", "queryHash", query.Hash(), "client", fmt.Sprintf("%p", cs.c),
		"blocksPerScan", cs.blocksPerScan, "currBlocksPerScan", cs.currBlocksPerScan,
		"from", fromBlock, "to", toBlock, "startBlock", startBlock, "endBlock", endBlock)

	var lastBlockAtomic atomic.Uint64

	go func() {
		for {
			lastBlock, err := cs.c.BlockNumber(ctx)
			if err != nil {
				log.Warn("Subscriber gets block number failed", "err", err)
				time.Sleep(cs.retryInterval)
				continue
			}

			lastBlockAtomic.Store(lastBlock)
			time.Sleep(cs.retryInterval)
		}
	}()

	reduceBlocksPerScan := false

	updateScan := func() {
		// startBlock=1000 endBlock=1200 currBlocksPerScan=200 blocksPerScan=10
		// 1. If reset
		// blocksPerScanToDebase = 190
		// currBlocksPerScan => 200
		// endBlock => 1200-190=1000+10 => 1010

		// 2, If not reset
		// startBlock => 1201
		// currBlocksPerScan => 200*2 = 400
		// endBlock => 1200+400 = 1600
		if reduceBlocksPerScan {
			reduceBlocksPerScan = false
			blocksPerScanToDebase := cs.currBlocksPerScan - cs.blocksPerScan
			cs.currBlocksPerScan = cs.blocksPerScan
			endBlock -= blocksPerScanToDebase
		} else {
			cs.currBlocksPerScan *= 2
			if cs.currBlocksPerScan > cs.maxBlocksPerScan {
				cs.currBlocksPerScan = cs.maxBlocksPerScan
			}
			startBlock, endBlock = endBlock+1, endBlock+cs.currBlocksPerScan
		}
	}

	go func() {
	Scan:
		for {
			select {
			case <-ctx.Done():
				log.Debug("Subscriber FilterLogs exits", "err", ctx.Err(), "client", fmt.Sprintf("%p", cs.c), "queryHash", query.Hash(), "from", fromBlock, "to", toBlock, "startBlock", startBlock, "endBlock", endBlock)
				close(logsChan)
				return
			default:
				lastBlock := lastBlockAtomic.Load()
				if lastBlock == 0 {
					time.Sleep(cs.retryInterval)
					continue Scan
				}

				if watch {
					if lastBlock >= cs.blockConfirmationsOnSubscription {
						log.Debug("Subscriber FilterLogs decreases lastBlock for confirmations", "client", fmt.Sprintf("%p", cs.c),
							"queryHash", query.Hash(),
							"lastBlock", lastBlock, "after", lastBlock-cs.blockConfirmationsOnSubscription)
						lastBlock -= cs.blockConfirmationsOnSubscription
					} else {
						time.Sleep(cs.retryInterval)
						continue Scan
					}
				}

				if startBlock > toBlock || (watch && startBlock > lastBlock) {
					if !watch {
						break Scan
					}

					// Update toBlock
					if lastBlock >= startBlock {
						toBlock = lastBlock
					} else {
						log.Debug("Subscriber FilterLogs waits for new block generated", "client", fmt.Sprintf("%p", cs.c),
							"queryHash", query.Hash(), "lastBlock", lastBlock,
							"confirmations", cs.blockConfirmationsOnSubscription)
						time.Sleep(cs.retryInterval)
						continue Scan
					}
				}
				if endBlock < startBlock {
					endBlock = startBlock
				}

				if endBlock > toBlock {
					endBlock = toBlock
				}

				if endBlock > lastBlock {
					endBlock = lastBlock
				}

				log.Info("Subscriber FilterLogs starts filtering logs", "client", fmt.Sprintf("%p", cs.c), "queryHash", query.Hash(),
					"currBlocksPerScan", cs.currBlocksPerScan, "blocksPerScan", cs.blocksPerScan,
					"from", startBlock, "to", endBlock, "latest", lastBlock)

				filterQuery := ethereum.FilterQuery{
					BlockHash: nil,
					FromBlock: big.NewInt(0).SetUint64(startBlock),
					ToBlock:   big.NewInt(0).SetUint64(endBlock),
					Addresses: q.Addresses,
					Topics:    q.Topics,
				}
				var lgs []etypes.Log

				if cs.storage.IsFilterLogsSupported(filterQuery) {
					lgs, err = cs.storage.FilterLogs(ctx, filterQuery)
				} else {
					lgs, err = cs.logFilterer.FilterLogs(ctx, filterQuery)

					/*
						If a query returns too many results or exceeds the max query duration,
						the following error is returned like below:
						{
							"jsonrpc": "2.0",
							"id": 1,
							"error": {
								"code": -32005,
								"message": "query returned more than 10000 results"
							}
						}

						So, we can adjust block range or reduce count of addresses being monitored.
					*/
					if err == nil {
						saveFilterLogsErr := cs.storage.SaveFilterLogs(filterQuery, lgs)
						if saveFilterLogsErr != nil {
							log.Error("save filter logs failed", "err", saveFilterLogsErr, "query", query)
							time.Sleep(cs.retryInterval)
							continue Scan
						}
					} else {
						err = consts.DecodeJsonRpcError(err, abi.ABI{})
						jsonRpcErr := err.(*consts.JsonRpcError)
						// Do not use JsonRpcErrorCodeLimitExceeded any more, becasuse
						// actual rpc that ethclient used may be behind jsonrpc gateway.

						// if jsonRpcErr.Code == consts.JsonRpcErrorCodeLimitExceeded {

						// if any error encountered, just reset currBlocksPerScan
						if jsonRpcErr.Code != 0 {
							log.Warn("Query filterLogs failed, so reducing blocks per scan", "startBlock", startBlock, "err", err)
							reduceBlocksPerScan = true
						}
					}
				}

				if err != nil {
					log.Warn("Subscriber FilterLogs is waiting for retry...", "err", err, "queryHash", query.Hash())
					updateScan()
					time.Sleep(cs.retryInterval)
					continue Scan
				}

				var latestHandledLog etypes.Log
				if useStorage {
					latestHandledLog, err = queryStateReader.LatestLogForQuery(ctx, q)
					if err != nil {
						log.Error("LatestLogForQuery failed", "err", err, "queryHash", query.Hash(), "query", q, "block", endBlock)
						time.Sleep(consts.RetryInterval)
						continue Scan
					}

					log.Debug("latestHandledLog", "queryHash", query.Hash(), "query", q, "log", latestHandledLog)
				}

				for _, l := range lgs {
					log.Debug("Subscriber FilterLogs is sending log", "queryHash", query.Hash(), "log", l, "latest_log", latestHandledLog)

					if latestHandledLog.BlockNumber > l.BlockNumber {
						continue
					}

					if latestHandledLog.BlockNumber == l.BlockNumber {
						if latestHandledLog.TxIndex > l.TxIndex {
							continue
						}
						if latestHandledLog.TxIndex == l.TxIndex && latestHandledLog.Index >= l.Index {
							continue
						}
					}

					logsChan <- l

					// if query handler is set, HandleQuery will be called.
					// so we do not update latestLog twice after sending log.
					if useStorage && queryStateWriter != nil && !cs.isQueryHandlerSet() {
						log.Debug("Subscriber FilterLogs SaveLatestLogForQuery", "queryHash", query.Hash(), "query", q, "log", l)
						err := queryStateWriter.SaveLatestLogForQuery(ctx, q, l)
						if err != nil {
							log.Error("Subscriber FilterLogs SaveLatestLogForQuery failed", "err", err, "queryHash", query.Hash(), "query", q, "block", endBlock)
							time.Sleep(consts.RetryInterval)
							continue Scan
						}
					}
				}

				// if there's no logs emitted during block range, enforcely update latestBlock
				// if query handler is not set, update latestBlock.
				if useStorage && queryStateWriter != nil && (!cs.isQueryHandlerSet() || len(lgs) == 0) {
					log.Debug("Subscriber FilterLogs SaveLatestBlockForQuery", "queryHash", query.Hash(), "query", q, "block", endBlock)
					if cs.isQueryHandlerSet() && len(lgs) == 0 {
						// Just notify latest block at which there's no logs emitted.
						log.Debug("notify latest block at which there's no logs emitted", "latestBlock", endBlock)
						logsChan <- etypes.Log{BlockNumber: endBlock}
					} else {
						err = queryStateWriter.SaveLatestBlockForQuery(ctx, q, endBlock)
						if err != nil {
							log.Error("SaveLatestBlockForQuery failed", "err", err, "queryHash", query.Hash(), "query", q, "block", endBlock)
							time.Sleep(consts.RetryInterval)
							continue Scan
						}
					}
				}

				updateScan()
			}
		}

		if closeOnExit {
			log.Debug("Subscriber FilterLogs was closed...", "queryHash", query.Hash())
			close(logsChan)
		}
	}()

	return nil
}

// SubscribeNewHead .
func (cs *ChainSubscriber) SubscribeNewHead(ctx context.Context, ch chan<- *etypes.Header) (sub ethereum.Subscription, err error) {
	checkChan := make(chan *etypes.Header)
	resubscribeFunc := func() (ethereum.Subscription, error) {
		return cs.c.SubscribeNewHead(ctx, checkChan)
	}

	ctx, cancel := context.WithCancel(ctx)

	sub = &subscription{ctx, cancel}

	return sub, cs.subscribeNewHead(ctx, resubscribeFunc, checkChan, ch)
}

// subscribeNewHead subscribes new header and auto reconnect if the connection lost.
func (cs *ChainSubscriber) subscribeNewHead(ctx context.Context, fn resubscribeFunc, checkChan <-chan *etypes.Header, resultChan chan<- *etypes.Header) error {
	// The goroutine for geting missing header and sending header to result channel.
	go func() {
		var lastHeader *etypes.Header
		for {
			select {
			case <-ctx.Done():
				log.Debug("SubscribeNewHead exit...")
				return
			case result := <-checkChan:
				if lastHeader != nil {
					if lastHeader.Number.Cmp(result.Number) >= 0 {
						// Ignore duplicate
						continue
					} else {
						// Get missing headers
						if cs.fetchMissingHeaders {
							start, end := new(big.Int).Add(lastHeader.Number, big.NewInt(1)), result.Number
							for start.Cmp(end) < 0 {
								header, err := cs.c.HeaderByNumber(ctx, start)
								switch err {
								case context.DeadlineExceeded, context.Canceled:
									log.Debug("SubscribeNewHead HeaderByNumber exit...")
									return
								case ethereum.NotFound:
									log.Warn("Client subscribeNewHead err: header not found")
									time.Sleep(consts.RetryInterval)
									continue
								case nil:
									log.Debug("Client get missing header", "number", start)
									start.Add(start, big.NewInt(1))
									resultChan <- header
								default: // ! nil
									log.Warn("Client subscribeNewHead", "err", err)
									time.Sleep(consts.RetryInterval)
									continue
								}
							}
						}
					}
				}
				lastHeader = result
				resultChan <- result
			}
		}
	}()

	// The goroutine to subscribe new header and send header to check channel.
	go func() {
		for {
			log.Debug("Client resubscribe...")
			sub, err := fn()
			if err != nil {
				if err == context.Canceled || err == context.DeadlineExceeded {
					log.Debug("SubscribeNewHead exit...")
					return
				}
				log.Warn("ChainClient resubscribeHeadFunc", "err", err)
				time.Sleep(consts.RetryInterval)
				continue
			}

			err = <-sub.Err()
			log.Warn("ChainClient subscribe head", "err", err)
			if err != nil {
				time.Sleep(consts.RetryInterval)
			}
		}
	}()

	return nil
}

func (cs *ChainSubscriber) SubscribeFilterFullTransactions(ctx context.Context, filter FilterTransaction, ch chan<- *etypes.Transaction) (ethereum.Subscription, error) {
	headerCh := make(chan *etypes.Header, cs.buffer)
	headerSub, err := cs.SubscribeNewHead(ctx, headerCh)
	if err != nil {
		return nil, err
	}
	defer headerSub.Unsubscribe()

	ctx, cancel := context.WithCancel(ctx)
	sub := &subscription{ctx, cancel}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case header := <-headerCh:
				for {
					block, err := cs.c.BlockByHash(ctx, header.Hash())
					if err != nil {
						log.Warn("SubscribeFullTransactions get block by hash failed", "err", err, "blockHash", header.Hash().Hex())
						time.Sleep(cs.retryInterval)
						continue
					}

					for _, tx := range block.Transactions() {
						cs.filterTransactions(filter, tx, ch)
					}
					break
				}
			}
		}
	}()

	return sub, nil
}

// SubscribeFullPendingTransactions subscribes to new pending transactions.
func (cs *ChainSubscriber) SubscribeFullPendingTransactions(ctx context.Context, ch chan<- *etypes.Transaction) (*rpc.ClientSubscription, error) {
	return cs.geth.SubscribeFullPendingTransactions(ctx, ch)
}

// SubscribePendingTransactions subscribes to new pending transaction hashes.
func (cs *ChainSubscriber) SubscribePendingTransactions(ctx context.Context, ch chan<- common.Hash) (*rpc.ClientSubscription, error) {
	return cs.geth.SubscribePendingTransactions(ctx, ch)
}

func (cs *ChainSubscriber) SubscribeFilterFullPendingTransactions(ctx context.Context, filter FilterTransaction, ch chan<- *etypes.Transaction) (*rpc.ClientSubscription, error) {
	fullIncomingsCh := make(chan *etypes.Transaction, cs.buffer)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case tx := <-fullIncomingsCh:
				cs.filterTransactions(filter, tx, ch)
			}
		}
	}()

	return cs.geth.SubscribeFullPendingTransactions(ctx, fullIncomingsCh)
}

func (cs *ChainSubscriber) filterTransactions(filter FilterTransaction, tx *etypes.Transaction, output chan<- *etypes.Transaction) {
	log.Debug("filterTransactions", "txHash", tx.Hash().Hex(), "filter", filter)
	signer := etypes.LatestSignerForChainID(cs.chainId)
	sender, err := signer.Sender(tx)
	if err != nil {
		log.Warn("Decode signer failed for the pending transaction", "err", err, "tx", tx)
		return
	}
	log.Debug("filterTransactions sender", "txHash", tx.Hash().Hex(), "sender", sender.Hex())

	filtered := false
	if slices.Contains(filter.From, sender) {
		filtered = true
	}

	if tx.To() != nil {
		log.Debug("filterTransactions to", "txHash", tx.Hash().Hex(), "to", tx.To().Hex())
		if slices.Contains(filter.To, *tx.To()) {
			filtered = true
		}
	}

	if len(tx.Data()) >= types.MethodSelectorLength {
		selector := types.MethodSelector(tx.Data()[:types.MethodSelectorLength])
		log.Debug("filterTransactions selector", "txHash", tx.Hash().Hex(), "selector", selector.Hex())

		if slices.Contains(filter.MethodSelector, selector) {
			filtered = true
		}
	}

	if filtered {
		log.Debug("filterTransactions filtered", "txHash", tx.Hash().Hex())
		output <- tx
	}
}
