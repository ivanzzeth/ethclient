package subscriber

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"slices"
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

// ChainSubscriber implements Subscriber interface
type ChainSubscriber struct {
	c                                *ethclient.Client
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

	queryParentCtx       context.Context
	cancelQueryParentCtx context.CancelFunc

	queryContextMu        sync.Mutex
	queryContextMap       sync.Map
	queryCancelContextMap sync.Map

	queryHandler       QueryHandler
	queryMap           sync.Map
	globalLogsChannels sync.Map

	lastBlockAtomic atomic.Uint64
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
		c:                    c,
		geth:                 geth,
		chainId:              chainId,
		buffer:               consts.DefaultMsgBuffer,
		blocksPerScan:        consts.DefaultBlocksPerScan,
		currBlocksPerScan:    consts.DefaultBlocksPerScan,
		maxBlocksPerScan:     consts.MaxBlocksPerScan,
		retryInterval:        consts.RetryInterval,
		storage:              storage,
		queryParentCtx:       queryCtx,
		cancelQueryParentCtx: cancel,
	}

	go func() {
		for {
			lastBlock, err := subscriber.c.BlockNumber(context.Background())
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || errors.Is(err, rpc.ErrClientQuit) {
					return
				}
				log.Warn("Subscriber gets block number failed", "err", err)
				time.Sleep(subscriber.retryInterval)
				continue
			}

			subscriber.lastBlockAtomic.Store(lastBlock)
			time.Sleep(subscriber.retryInterval)
		}
	}()

	return subscriber, nil
}

func (s *ChainSubscriber) Close() {
	log.Debug("close subscriber...")
	s.cancelQueryParentCtx()

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
		for {
			queryCtx, _ := cs.getQueryContext(queryHash)

			err := cs.FilterLogsWithChannel(queryCtx, query, globalLogsChannel, true, true)
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

func (cs *ChainSubscriber) DeleteQuery(query ethereum.FilterQuery) error {
	queryHash := GetQueryHash(cs.chainId, query)
	log.Info("delete query", "queryHash", queryHash, "query", query, "client", fmt.Sprintf("%p", cs.c))
	if !cs.isQueryHandlerSet() {
		return fmt.Errorf("setup queryHandler before calling it")
	}

	_, loaded := cs.queryMap.LoadOrStore(queryHash, &sync.Once{})
	if !loaded {
		return fmt.Errorf("query has not submitted yet")
	}

	// Notify go routines exit
	_, cancelCtx := cs.getQueryContext(queryHash)
	cancelCtx()

	cs.queryMap.Delete(queryHash)

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

func (cs *ChainSubscriber) getQueryContext(queryHash common.Hash) (context.Context, context.CancelFunc) {
	cs.queryContextMu.Lock()
	defer cs.queryContextMu.Unlock()

	ctx, cancel := context.WithCancel(cs.queryParentCtx)
	ctxVal, _ := cs.queryContextMap.LoadOrStore(queryHash, ctx)
	cancelVal, _ := cs.queryCancelContextMap.LoadOrStore(queryHash, cancel)

	ctx = ctxVal.(context.Context)
	cancel = cancelVal.(context.CancelFunc)

	return ctx, cancel
}

func (cs *ChainSubscriber) SubscribeFilterLogs(ctx context.Context, q ethereum.FilterQuery, ch chan<- etypes.Log) (sub ethereum.Subscription, err error) {
	log.Debug("SubscribeFilterlogs starts", "query", q)

	ctx, cancel := context.WithCancel(ctx)
	err = cs.FilterLogsWithChannel(ctx, q, ch, true, true)

	sub = &subscription{ctx, cancel}
	return
}

func (cs *ChainSubscriber) FilterLogs(ctx context.Context, q ethereum.FilterQuery) (logs []etypes.Log, err error) {
	logsChan := make(chan etypes.Log, cs.buffer)
	filterCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	err = cs.FilterLogsWithChannel(filterCtx, q, logsChan, false, true)
	if err != nil {
		return nil, err
	}

	for l := range logsChan {
		log.Debug("FilterLogs receiving log", "log", l)

		logs = append(logs, l)
	}

	return
}

func (cs *ChainSubscriber) FilterLogsBatch(ctx context.Context, queries []ethereum.FilterQuery) (logs [][]etypes.Log, err error) {
	// Merge queries to reduce the number of rpc calls.
	mergedQueries, err := mergeFilterQueriesWithMaxAddressesPerQuery(queries, 5, true)
	if err != nil {
		return nil, err
	}

	// Filter logs for each merged query.
	allLogs := []etypes.Log{}
	for _, q := range mergedQueries {
		ls, err := cs.FilterLogs(ctx, q)
		if err != nil {
			return nil, err
		}

		allLogs = append(allLogs, ls...)
	}

	// Distribute logs to each query.
	logs = distributeLogs(allLogs, queries)
	return
}

// TODO:
// 3. cache all of finalized historical data, e.g., blockByHash, txByHash
func (cs *ChainSubscriber) FilterLogsWithChannel(ctx context.Context, q ethereum.FilterQuery, logsChan chan<- etypes.Log, watch bool, closeOnExit bool) (err error) {
	if q.BlockHash != nil {
		logs, err := cs.filterLogsWithAutoSplit(ctx, q)
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
				log.Info("Subscriber FilterLogs exits", "err", ctx.Err(), "client", fmt.Sprintf("%p", cs.c), "queryHash", query.Hash(), "from", fromBlock, "to", toBlock, "startBlock", startBlock, "endBlock", endBlock)
				close(logsChan)
				return
			default:
				lastBlock := cs.lastBlockAtomic.Load()
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
					// We need to call rpc nodes so that splitting query as needed.
					lgs, err = cs.filterLogsWithAutoSplit(ctx, filterQuery)
					// lgs, err = cs.c.FilterLogs(ctx, filterQuery)

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

func (cs *ChainSubscriber) filterLogsWithAutoSplit(ctx context.Context, q ethereum.FilterQuery) (logs []etypes.Log, err error) {
	// TODO: Configuration
	queries, err := splitFilterQuery(q, 5)
	if err != nil {
		return
	}

	for _, query := range queries {
		var ls []etypes.Log
		ls, err = cs.c.FilterLogs(ctx, query)
		if err != nil {
			logs = nil
			return
		}

		logs = append(logs, ls...)
	}

	slices.SortFunc(logs, func(a, b etypes.Log) int {
		if a.BlockNumber < b.BlockNumber {
			return -1
		} else if a.BlockNumber > b.BlockNumber {
			return 1
		}

		if a.TxIndex < b.TxIndex {
			return -1
		} else if a.TxIndex > b.TxIndex {
			return 1
		}

		if a.Index < b.Index {
			return -1
		} else if a.Index > b.Index {
			return 1
		}

		return 0
	})

	return
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
