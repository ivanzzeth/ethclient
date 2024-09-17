package subscriber

import (
	"context"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
)

type MemoryStorage struct {
	chainId  *big.Int
	blockMap sync.Map
	logMap   sync.Map
}

func NewMemoryStorage(chainId *big.Int) *MemoryStorage {
	return &MemoryStorage{
		chainId:  chainId,
		blockMap: sync.Map{},
		logMap:   sync.Map{},
	}
}

func (s *MemoryStorage) LatestBlockForQuery(ctx context.Context, query ethereum.FilterQuery) (uint64, error) {
	ret, _ := s.blockMap.LoadOrStore(GetQueryKey(s.chainId, query), uint64(0))

	return ret.(uint64), nil
}

func (s *MemoryStorage) LatestLogForQuery(ctx context.Context, query ethereum.FilterQuery) (types.Log, error) {
	ret, _ := s.logMap.LoadOrStore(GetQueryKey(s.chainId, query), types.Log{})
	return ret.(types.Log), nil
}

func (s *MemoryStorage) SaveLatestBlockForQuery(ctx context.Context, query ethereum.FilterQuery, blockNum uint64) error {
	s.blockMap.Store(GetQueryKey(s.chainId, query), blockNum)
	return nil
}

func (s *MemoryStorage) SaveLatestLogForQuery(ctx context.Context, query ethereum.FilterQuery, log types.Log) error {
	s.logMap.Store(GetQueryKey(s.chainId, query), log)
	return nil
}
