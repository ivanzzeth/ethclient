package nonce

import (
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

var _ Storage = &MemoryStorage{}

type MemoryStorage struct {
	lockMap  sync.Map
	nonceMap map[common.Address]uint64
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		nonceMap: make(map[common.Address]uint64),
		lockMap:  sync.Map{},
	}
}

func (s *MemoryStorage) NonceLockFrom(from common.Address) sync.Locker {
	lock, _ := s.lockMap.LoadOrStore(from, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

func (s *MemoryStorage) GetNonce(account common.Address) (uint64, error) {
	nonce := s.nonceMap[account]

	return nonce, nil
}

func (s *MemoryStorage) SetNonce(account common.Address, nonce uint64) error {
	s.nonceMap[account] = nonce

	return nil
}
