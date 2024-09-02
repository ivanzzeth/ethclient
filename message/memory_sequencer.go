package message

import (
	"errors"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ivanzz/ethclient/ds/graph"
)

var _ Sequencer = &MemorySequencer{}

var ErrPendingChannelClosed = errors.New("pending channel was closed")

type MemorySequencer struct {
	msgStorage   Storage
	dag          *graph.DiGraph
	queuedReq    chan Request
	queuedCount  atomic.Int64
	pendingReq   chan Request
	pendingCount atomic.Int64
}

func NewMemorySequencer(msgStorage Storage, buffer int) *MemorySequencer {
	s := &MemorySequencer{
		msgStorage: msgStorage,
		dag:        graph.NewDirectedGraph(),
		queuedReq:  make(chan Request, buffer),
		pendingReq: make(chan Request, buffer),
	}

	go s.run()

	return s
}

func (s *MemorySequencer) PushMsg(msg Request) error {
	s.msgStorage.AddMsg(msg)
	s.queuedReq <- msg
	s.queuedCount.Add(1)
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

func (s *MemorySequencer) run() {
	go func() {
		for req := range s.queuedReq {
			s.queuedCount.Add(-1)
			if req.AfterMsg == nil {
				s.dag.AddVertex(req.Id())
			} else {
				s.dag.AddEdge(*req.AfterMsg, req.Id())
			}
		}
	}()

	for reqId := range s.dag.Pipeline() {
		msg, _ := s.msgStorage.GetMsg(reqId.(common.Hash))
		s.pendingCount.Add(1)
		s.pendingReq <- *msg.Req
	}
}
