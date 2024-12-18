package subscriber

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ivanzzeth/ethclient/common/consts"
)

var _ Subscriber = (*ChainSubscriber)(nil)

var _ ethereum.LogFilterer = (*ChainSubscriber)(nil)

// ChainSubscriber implements Subscriber interface
type ChainSubscriber struct {
	c                                *ethclient.Client
	chainId                          *big.Int
	retryInterval                    time.Duration
	buffer                           int
	blocksPerScan                    uint64
	currBlocksPerScan                uint64 // adjust dynamiclly
	maxBlocksPerScan                 uint64
	blockConfirmationsOnSubscription uint64
	storage                          SubscriberStorage
	queryHandler                     QueryHandler
	queryMap                         sync.Map
	globalLogsChannels               sync.Map
}

// NewChainSubscriber .
func NewChainSubscriber(c *ethclient.Client, storage SubscriberStorage) (*ChainSubscriber, error) {
	chainId, err := c.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	subscriber := &ChainSubscriber{
		c:                 c,
		chainId:           chainId,
		buffer:            consts.DefaultMsgBuffer,
		blocksPerScan:     consts.DefaultBlocksPerScan,
		currBlocksPerScan: consts.DefaultBlocksPerScan,
		maxBlocksPerScan:  consts.MaxBlocksPerScan,
		retryInterval:     consts.RetryInterval,
		storage:           storage,
	}

	return subscriber, nil
}

func (s *ChainSubscriber) Close() {
	log.Debug("close subscriber...")
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

func (cs *ChainSubscriber) SubmitQuery(query ethereum.FilterQuery) error {
	queryHash := GetQueryHash(cs.chainId, query)
	log.Info("submit query", "queryHash", queryHash, "query", query)

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
			err := cs.FilterLogsWithChannel(context.Background(), query, globalLogsChannel, true, true)
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

func (cs *ChainSubscriber) handleQueryLogsChannel(query ethereum.FilterQuery, ch <-chan types.Log) {
	for l := range ch {
		err := cs.queryHandler.HandleQuery(context.Background(), NewQuery(cs.chainId, query), l)
		if err != nil {
			log.Warn("handle query failed", "err", err, "queryHash", GetQueryHash(cs.chainId, query))
		}
	}
}

func (cs *ChainSubscriber) getQueryLogChannel(queryHash common.Hash) chan types.Log {
	ch, _ := cs.globalLogsChannels.LoadOrStore(queryHash, make(chan types.Log, cs.buffer))
	globalLogsChannel := ch.(chan types.Log)

	return globalLogsChannel
}

func (cs *ChainSubscriber) SubscribeFilterLogs(ctx context.Context, q ethereum.FilterQuery, ch chan<- types.Log) (sub ethereum.Subscription, err error) {
	log.Debug("SubscribeFilterlogs starts", "query", q)

	ctx, cancel := context.WithCancel(ctx)
	err = cs.FilterLogsWithChannel(ctx, q, ch, true, true)

	sub = &subscription{ctx, cancel}
	return
}

func (cs *ChainSubscriber) FilterLogs(ctx context.Context, q ethereum.FilterQuery) (logs []types.Log, err error) {
	logsChan := make(chan types.Log, cs.buffer)
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
// 1. reduce eth_blockNumber calls
// 2. cache eth_chainId
// 3. cache all of finalized historical data, e.g., blockByHash, txByHash
func (cs *ChainSubscriber) FilterLogsWithChannel(ctx context.Context, q ethereum.FilterQuery, logsChan chan<- types.Log, watch bool, closeOnExit bool) (err error) {
	if q.BlockHash != nil {
		logs, err := cs.c.FilterLogs(ctx, q)
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
		queryStateWriter = nil
	}

	startBlock := fromBlock

	if useStorage {
		fromBlockInStorage, err := queryStateReader.LatestBlockForQuery(ctx, q)
		if err != nil {
			return err
		}

		if fromBlockInStorage != 0 {
			startBlock = fromBlockInStorage + 1
		}
	}

	endBlock := startBlock + cs.currBlocksPerScan

	query := NewQuery(cs.chainId, q)
	log.Debug("FilterLogsWithChannel starts", "queryHash", query.Hash(),
		"blocksPerScan", cs.blocksPerScan, "currBlocksPerScan", cs.currBlocksPerScan,
		"from", fromBlock, "to", toBlock, "startBlock", startBlock, "endBlock", endBlock)

	go func() {
	Scan:
		for {
			lastBlock, err := cs.c.BlockNumber(ctx)
			if err != nil {
				// More time to avoid rate-limit
				log.Warn("get block number failed", "err", err)
				time.Sleep(5 * cs.retryInterval)
				continue
			}

			if watch {
				if lastBlock >= cs.blockConfirmationsOnSubscription {
					log.Info("filtering logs decreases lastBlock for confirmations", "queryHash", query.Hash(),
						"lastBlock", lastBlock, "after", lastBlock-cs.blockConfirmationsOnSubscription)
					lastBlock -= cs.blockConfirmationsOnSubscription
				} else {
					time.Sleep(cs.retryInterval)
					continue
				}
			}

			if startBlock > toBlock || (watch && startBlock > lastBlock) {
				if !watch {
					break
				}

				// Update toBlock
				if lastBlock >= startBlock {
					toBlock = lastBlock
				} else {
					log.Debug("FilterLogsWithChannel waits for new block generated", "queryHash", query.Hash(), "lastBlock", lastBlock, "confirmations", cs.blockConfirmationsOnSubscription)
					time.Sleep(cs.retryInterval)
					continue
				}
			}

			select {
			case <-ctx.Done():
				log.Debug("FilterLogsWithChannel exits", "err", ctx.Err(), "queryHash", query.Hash(), "from", fromBlock, "to", toBlock, "startBlock", startBlock, "endBlock", endBlock)
				close(logsChan)
				return
			default:
				if endBlock < startBlock {
					endBlock = startBlock
				}

				if endBlock > toBlock {
					endBlock = toBlock
				}

				if endBlock > lastBlock {
					endBlock = lastBlock
				}

				log.Info("start filtering logs", "queryHash", query.Hash(),
					"currBlocksPerScan", cs.currBlocksPerScan, "blocksPerScan", cs.blocksPerScan,
					"from", startBlock, "to", endBlock, "latest", lastBlock)

				filterQuery := ethereum.FilterQuery{
					BlockHash: nil,
					FromBlock: big.NewInt(0).SetUint64(startBlock),
					ToBlock:   big.NewInt(0).SetUint64(endBlock),
					Addresses: q.Addresses,
					Topics:    q.Topics,
				}
				var lgs []types.Log

				if cs.storage.IsFilterLogsSupported(filterQuery) {
					lgs, err = cs.storage.FilterLogs(ctx, filterQuery)
				} else {
					lgs, err = cs.c.FilterLogs(ctx, filterQuery)

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
						cs.currBlocksPerScan *= 2
						if cs.currBlocksPerScan > cs.maxBlocksPerScan {
							cs.currBlocksPerScan = cs.maxBlocksPerScan
						}
						saveFilterLogsErr := cs.storage.SaveFilterLogs(filterQuery, lgs)
						if saveFilterLogsErr != nil {
							log.Error("save filter logs failed", "err", saveFilterLogsErr, "query", query)
							time.Sleep(cs.retryInterval)
							continue Scan
						}
					} else {
						err = consts.DecodeJsonRpcError(err, abi.ABI{})
						jsonRpcErr := err.(*consts.JsonRpcError)
						if jsonRpcErr.Code == consts.JsonRpcErrorCodeLimitExceeded {
							blocksPerScanToDebase := cs.currBlocksPerScan - cs.blocksPerScan
							cs.currBlocksPerScan = cs.blocksPerScan
							endBlock -= blocksPerScanToDebase
						}
					}
				}

				if err != nil {
					log.Warn("filtering logs, waiting for retry...", "err", err, "queryHash", query.Hash())
					time.Sleep(cs.retryInterval)
					continue Scan
				}

				var latestHandledLog types.Log
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
					log.Debug("FilterLogsWithChannel sending log", "queryHash", query.Hash(), "log", l, "latest_log", latestHandledLog)

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
					if useStorage && queryStateWriter != nil {
						log.Debug("SaveLatestLogForQuery", "queryHash", query.Hash(), "query", q, "log", l)
						err := queryStateWriter.SaveLatestLogForQuery(ctx, q, l)
						if err != nil {
							log.Error("SaveLatestLogForQuery failed", "err", err, "queryHash", query.Hash(), "query", q, "block", endBlock)
							time.Sleep(consts.RetryInterval)
							continue Scan
						}
					}
				}

				// TODO: fix SaveLatestBlockForQuery not called when there's no logs.
				if useStorage && queryStateWriter != nil {
					log.Debug("SaveLatestBlockForQuery", "queryHash", query.Hash(), "query", q, "block", endBlock)
					err = queryStateWriter.SaveLatestBlockForQuery(ctx, q, endBlock)
					if err != nil {
						log.Error("SaveLatestBlockForQuery failed", "err", err, "queryHash", query.Hash(), "query", q, "block", endBlock)
						time.Sleep(consts.RetryInterval)
						continue Scan
					}
				}

				startBlock, endBlock = endBlock+1, endBlock+cs.currBlocksPerScan
			}
		}

		if closeOnExit {
			log.Debug("FilterLogsWithChannel was closed...", "queryHash", query.Hash())
			close(logsChan)
		}
	}()

	return nil
}

// SubscribeNewHead .
func (cs *ChainSubscriber) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (sub ethereum.Subscription, err error) {
	checkChan := make(chan *types.Header)
	resubscribeFunc := func() (ethereum.Subscription, error) {
		return cs.c.SubscribeNewHead(ctx, checkChan)
	}

	ctx, cancel := context.WithCancel(ctx)

	sub = &subscription{ctx, cancel}

	return sub, cs.subscribeNewHead(ctx, resubscribeFunc, checkChan, ch)
}

// subscribeNewHead subscribes new header and auto reconnect if the connection lost.
func (cs *ChainSubscriber) subscribeNewHead(ctx context.Context, fn resubscribeFunc, checkChan <-chan *types.Header, resultChan chan<- *types.Header) error {
	// The goroutine for geting missing header and sending header to result channel.
	go func() {
		var lastHeader *types.Header
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

			select {
			case err := <-sub.Err():
				log.Warn("ChainClient subscribe head", "err", err)
				sub.Unsubscribe()
				time.Sleep(consts.RetryInterval)
			}
		}
	}()

	return nil
}
