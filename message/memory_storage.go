package message

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

var _ Storage = &MemoryStorage{}

type MemoryStorage struct {
	store sync.Map
}

func NewMemoryStorage() (*MemoryStorage, error) {
	return &MemoryStorage{}, nil
}

func (s *MemoryStorage) AddMsg(req Request) error {
	log.Debug("MemoryStorage AddMsg", "req", req)
	s.store.Store(req.id, Message{
		Req:    &req,
		Status: MessageStatusSubmitted,
	})
	return nil
}

func (s *MemoryStorage) HasMsg(msgId common.Hash) bool {
	_, ok := s.store.Load(msgId)
	return ok
}

func (s *MemoryStorage) GetMsg(msgId common.Hash) (Message, error) {
	msg, ok := s.store.Load(msgId)
	if !ok {
		return Message{}, fmt.Errorf("not found")
	}

	log.Debug("MemoryStorage GetMsg", "msgId", msgId.Hex(), "msg", msg.(Message))
	return msg.(Message), nil
}

func (s *MemoryStorage) UpdateMsg(msg Message) error {
	s.store.Store(msg.Req.id, msg)
	return nil
}

func (s *MemoryStorage) UpdateResponse(msgId common.Hash, resp Response) error {
	log.Debug("MemoryStorage UpdateResponse", "msgId", msgId.Hex(), "resp", resp)

	msg, err := s.GetMsg(msgId)
	if err != nil {
		return err
	}

	if msg.Resp != nil {
		panic("same msg not allowed updating response twice")
	}

	msg.Resp = &resp
	return s.UpdateMsg(msg)
}

func (s *MemoryStorage) UpdateReceipt(msgId common.Hash, receipt Receipt) error {
	log.Debug("MemoryStorage UpdateReceipt", "msgId", msgId.Hex(), "receipt", receipt)

	msg, err := s.GetMsg(msgId)
	if err != nil {
		return err
	}

	msg.Receipt = &receipt
	return s.UpdateMsg(msg)
}

func (s *MemoryStorage) UpdateMsgStatus(msgId common.Hash, status MessageStatus) error {
	log.Debug("MemoryStorage UpdateMsgStatus", "msgId", msgId.Hex(), "status", status)

	msg, err := s.GetMsg(msgId)
	if err != nil {
		return err
	}

	msg.Status = status

	return s.UpdateMsg(msg)
}

func (s *MemoryStorage) GetNonce(msgId common.Hash) (nonce uint64, err error) {
	msg, err := s.GetMsg(msgId)
	if err != nil {
		return
	}

	if msg.Resp == nil || msg.Resp.Tx == nil {
		return 0, fmt.Errorf("no nonce assigned")
	}

	return msg.Resp.Tx.Nonce(), nil
}
