package handler

import (
	"context"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ivanzzeth/ethclient/subscriber"
)

var _ subscriber.QueryHandler = (*SimpleQueryHandler)(nil)

type SimpleQueryHandler struct {
	subscriber.SubscriberStorage
	latestBlock atomic.Uint64
}

func NewSimpleQueryHandler(storage subscriber.SubscriberStorage) *SimpleQueryHandler {
	return &SimpleQueryHandler{SubscriberStorage: storage}
}

func (h *SimpleQueryHandler) HandleQuery(ctx context.Context, query ethereum.FilterQuery, l types.Log) error {
	log.Debug("handle query", "topic", l.Topics[0], "block", l.BlockNumber, "txIndex", l.TxIndex, "index", l.Index)

	err := h.SaveLatestLogForQuery(ctx, query, l)
	if err != nil {
		return err
	}

	if l.BlockNumber > h.latestBlock.Load() {
		err = h.SaveLatestBlockForQuery(ctx, query, l.BlockNumber)
		if err != nil {
			return err
		}
	}

	return nil
}
