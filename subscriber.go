package ethclient

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
)

var (
	retryInterval = 2 * time.Second
)

var _ Subscriber = (*ChainSubscriber)(nil)

// ChainSubscriber implements Subscriber interface
type ChainSubscriber struct {
	c             *ethclient.Client
	retryInterval time.Duration
	buffer        int
	blocksPerScan uint64
}

// NewChainSubscriber .
func NewChainSubscriber(c *ethclient.Client) (*ChainSubscriber, error) {
	return &ChainSubscriber{
		c:             c,
		buffer:        DefaultMsgBuffer,
		blocksPerScan: DefaultBlocksPerScan,
		retryInterval: retryInterval,
	}, nil
}

func (cs *ChainSubscriber) SubscribeFilterlogs(ctx context.Context, q ethereum.FilterQuery, ch chan<- types.Log) (err error) {
	log.Debug("SubscribeFilterlogs starts", "query", q)
	fromBlock := uint64(0)
	if q.FromBlock != nil {
		fromBlock = q.FromBlock.Uint64()
	}

	lastBlock, err := cs.c.BlockNumber(ctx)
	if err != nil {
		return err
	}

	toBlock := uint64(0)
	if q.ToBlock != nil {
		toBlock = q.ToBlock.Uint64()
	} else {
		toBlock = lastBlock
	}

	startBlock := fromBlock
	endBlock := toBlock + 1

	if startBlock >= endBlock {
		return fmt.Errorf("invalid block number")
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				close(ch)
				return
			default:
				log.Debug("SubscribeFilterlogs loops", "start", startBlock, "end", endBlock)
				lastBlock, err := cs.c.BlockNumber(ctx)
				if err != nil {
					log.Debug("", "curr", lastBlock, "wait", lastBlock, "sleep", cs.retryInterval)
					time.Sleep(cs.retryInterval)
					continue
				}

				if endBlock > lastBlock+1 {
					// wait for generation
					log.Debug("waitng for block generation", "curr", lastBlock, "wait", lastBlock, "sleep", cs.retryInterval)
					time.Sleep(cs.retryInterval)
					continue
				}

				q.FromBlock = big.NewInt(0).SetUint64(startBlock)
				q.ToBlock = big.NewInt(0).SetUint64(endBlock - 1)
				err = cs.FilterLogsWithChannel(ctx, q, ch, false)
				if err != nil {
					log.Debug("SubscribeFilterlogs call FilterLogsWithChannel failed", "err", err)
					time.Sleep(cs.retryInterval)
					continue
				}

				lastBlock, err = cs.c.BlockNumber(ctx)
				if err != nil {
					log.Debug("SubscribeFilterlogs call BlockNumber failed", "err", err)
					time.Sleep(cs.retryInterval)
					continue
				}

				startBlock = endBlock
				endBlock = lastBlock + 1

				time.Sleep(cs.retryInterval)
			}
		}
	}()

	return
}

func (cs *ChainSubscriber) FilterLogs(ctx context.Context, q ethereum.FilterQuery) (logs []types.Log, err error) {
	logsChan := make(chan types.Log, cs.buffer)
	err = cs.FilterLogsWithChannel(ctx, q, logsChan, true)
	if err != nil {
		return nil, err
	}

	for l := range logsChan {
		logs = append(logs, l)
	}

	return
}

func (cs *ChainSubscriber) FilterLogsWithChannel(ctx context.Context, q ethereum.FilterQuery, logsChan chan<- types.Log, closeOnExit bool) (err error) {
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

	startBlock := fromBlock
	endBlock := startBlock + cs.blocksPerScan
	log.Debug("FilterLogsWithChannel starts", "from", fromBlock, "to", toBlock, "startBlock", startBlock, "endBlock", endBlock)

	go func() {
		for startBlock <= toBlock {
			select {
			case <-ctx.Done():
				log.Debug("FilterLogsWithChannel exits", "err", ctx.Err(), "from", fromBlock, "to", toBlock, "startBlock", startBlock, "endBlock", endBlock)
				close(logsChan)
				return
			default:
				if endBlock > toBlock {
					endBlock = toBlock
				}

				lastBlock, err := cs.c.BlockNumber(ctx)
				if err != nil {
					time.Sleep(cs.retryInterval)
					continue
				}
				if endBlock > lastBlock {
					endBlock = lastBlock
				}

				log.Debug("FilterLogsWithChannel loops", "from", startBlock, "to", endBlock)

				lgs, err := cs.c.FilterLogs(ctx, ethereum.FilterQuery{
					BlockHash: nil,
					FromBlock: big.NewInt(0).SetUint64(startBlock),
					ToBlock:   big.NewInt(0).SetUint64(endBlock),
					Addresses: q.Addresses,
					Topics:    q.Topics,
				})
				if err != nil {
					log.Debug("FilterLogsWithChannel filter failed, waiting for retry...", "err", err)
					time.Sleep(cs.retryInterval)
					continue
				}

				for _, l := range lgs {
					logsChan <- l
				}

				startBlock, endBlock = endBlock+1, endBlock+cs.blocksPerScan
			}
		}

		if closeOnExit {
			close(logsChan)
		}
	}()

	return nil
}

// SubscribeNewHead .
func (cs *ChainSubscriber) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) error {
	checkChan := make(chan *types.Header)
	resubscribeFunc := func() (ethereum.Subscription, error) {
		return cs.c.SubscribeNewHead(ctx, checkChan)
	}

	return cs.subscribeNewHead(ctx, resubscribeFunc, checkChan, ch)
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
								time.Sleep(retryInterval)
								continue
							case nil:
								log.Debug("Client get missing header", "number", start)
								start.Add(start, big.NewInt(1))
								resultChan <- header
							default: // ! nil
								log.Warn("Client subscribeNewHead", "err", err)
								time.Sleep(retryInterval)
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
				time.Sleep(retryInterval)
				continue
			}

			select {
			case err := <-sub.Err():
				log.Warn("ChainClient subscribe head", "err", err)
				sub.Unsubscribe()
				time.Sleep(retryInterval)
			}
		}
	}()

	return nil
}
