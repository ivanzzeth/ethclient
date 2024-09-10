package subscriber

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type MemoryStorage struct {
	blockMap sync.Map
	logMap   sync.Map
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		blockMap: sync.Map{},
		logMap:   sync.Map{},
	}
}

func (s *MemoryStorage) LatestBlockForQuery(ctx context.Context, query ethereum.FilterQuery) (uint64, error) {
	ret, _ := s.blockMap.LoadOrStore(GetQueryKey(query), uint64(0))

	return ret.(uint64), nil
}

func (s *MemoryStorage) LatestLogForQuery(ctx context.Context, query ethereum.FilterQuery) (types.Log, error) {
	ret, _ := s.logMap.LoadOrStore(GetQueryKey(query), types.Log{})
	return ret.(types.Log), nil
}

func (s *MemoryStorage) SaveLatestBlockForQuery(ctx context.Context, query ethereum.FilterQuery, blockNum uint64) error {
	s.blockMap.Store(GetQueryKey(query), blockNum)
	return nil
}

func (s *MemoryStorage) SaveLatestLogForQuery(ctx context.Context, query ethereum.FilterQuery, log types.Log) error {
	s.logMap.Store(GetQueryKey(query), log)
	return nil
}

func GetQueryKey(query ethereum.FilterQuery) string {
	json, _ := json.Marshal(query)
	hash := crypto.Keccak256Hash(json)

	return hash.Hex()
}
