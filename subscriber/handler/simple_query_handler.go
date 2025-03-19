package handler

import (
	"context"
	"sync/atomic"

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

func (h *SimpleQueryHandler) HandleQuery(ctx context.Context, query subscriber.Query, l types.Log) error {
	err := h.SaveLatestLogForQuery(ctx, query.FilterQuery, l)
	if err != nil {
		return err
	}

	if l.BlockNumber > h.latestBlock.Load() {
		err = h.SaveLatestBlockForQuery(ctx, query.FilterQuery, l.BlockNumber)
		if err != nil {
			return err
		}
	}

	return nil
}
