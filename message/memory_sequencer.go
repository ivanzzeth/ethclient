package message

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ivanzz/ethclient/ds/graph"
)

var _ Sequencer = &MemorySequencer{}

var ErrPendingChannelClosed = errors.New("pending channel was closed")

type MemorySequencer struct {
	client       *ethclient.Client
	msgStorage   Storage
	dag          *graph.DiGraph
	queuedReq    chan Request
	queuedCount  atomic.Int64
	pendingReq   chan Request
	pendingCount atomic.Int64
}

func NewMemorySequencer(client *ethclient.Client, msgStorage Storage, buffer int) *MemorySequencer {
	s := &MemorySequencer{
		client:     client,
		msgStorage: msgStorage,
		dag:        graph.NewDirectedGraph(buffer),
		queuedReq:  make(chan Request, buffer),
		pendingReq: make(chan Request, buffer),
	}

	go s.run()

	return s
}

func (s *MemorySequencer) PushMsg(msg Request) error {
	s.queuedReq <- msg
	s.queuedCount.Add(1)

	err := s.msgStorage.UpdateMsgStatus(msg.Id(), MessageStatusQueued)
	if err != nil {
		return err
	}

	return nil
}

func (s *MemorySequencer) PopMsg() (Request, error) {
	req, ok := <-s.pendingReq
	if !ok {
		return Request{}, ErrPendingChannelClosed
	}
	s.pendingCount.Add(-1)

	return req, nil
}

func (s *MemorySequencer) PeekMsg() (Request, error) {
	// TODO:
	return Request{}, nil
}

func (s *MemorySequencer) QueuedMsgCount() (int, error) {
	return int(s.queuedCount.Load()), nil
}

func (s *MemorySequencer) PendingMsgCount() (int, error) {
	return int(s.pendingCount.Load()), nil
}

func (s *MemorySequencer) Close() {
	close(s.queuedReq)
	close(s.pendingReq)
}

func (s *MemorySequencer) run() {
	go func() {
		for req := range s.queuedReq {
			s.queuedCount.Add(-1)
			if req.AfterMsg == nil {
				s.dag.AddVertex(req.Id())
			} else {
				tx, isPending, getTxErr := s.client.TransactionByHash(context.Background(), *req.AfterMsg)
				if getTxErr == nil && !isPending {
					log.Debug("tx was found, so push msg to dag", "reqId", req.Id().Hex(),
						"tx", tx.Hash().Hex(), "is_pending", isPending)
					s.dag.AddVertex(req.Id())
				}

				_, err := s.msgStorage.GetMsg(*req.AfterMsg)
				if err == nil {
					s.dag.AddEdge(*req.AfterMsg, req.Id())
				} else {
					// after message not ready, so push back
					log.Debug("after message not ready, so push back", "reqId", req.Id().Hex())
					s.queuedCount.Add(1)
					s.queuedReq <- req
				}
			}
		}
	}()

	for reqId := range s.dag.Pipeline() {
		msg, _ := s.msgStorage.GetMsg(reqId.(common.Hash))
		s.pendingCount.Add(1)
		s.pendingReq <- *msg.Req
	}
}
