package message

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

var _ Storage = &MemoryStorage{}

type MemoryStorage struct {
	store sync.Map
}

func NewMemoryStorage() (*MemoryStorage, error) {
	return &MemoryStorage{}, nil
}

func (s *MemoryStorage) AddMsg(req Request) error {
	s.store.Store(req.id, Message{
		Req:    &req,
		Status: MessageStatusPending,
	})
	return nil
}

func (s *MemoryStorage) GetMsg(msgId common.Hash) (Message, error) {
	msg, ok := s.store.Load(msgId)
	if !ok {
		return Message{}, fmt.Errorf("not found")
	}

	return msg.(Message), nil
}

func (s *MemoryStorage) UpdateMsg(msg Message) error {
	s.store.Store(msg.Req.id, msg)
	return nil
}

func (s *MemoryStorage) UpdateResponse(msgId common.Hash, resp Response) error {
	msg, err := s.GetMsg(msgId)
	if err != nil {
		return err
	}

	msg.Resp = &resp
	return s.UpdateMsg(msg)
}

func (s *MemoryStorage) UpdateMsgStatus(msgId common.Hash, status MessageStatus) error {
	msg, err := s.GetMsg(msgId)
	if err != nil {
		return err
	}

	msg.Status = status

	return s.UpdateMsg(msg)
}
