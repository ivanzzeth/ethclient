package subscriber

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

var _ SubscriberStorage = (*MemoryStorage)(nil)

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

func (s *MemoryStorage) FilterLogs(ctx context.Context, q ethereum.FilterQuery) (logs []types.Log, err error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *MemoryStorage) IsFilterLogsSupported(q ethereum.FilterQuery) bool {
	return false
}

func (s *MemoryStorage) FilterLogsBatch(ctx context.Context, queries []ethereum.FilterQuery) (logs [][]types.Log, err error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *MemoryStorage) SaveLatestBlockForQuery(ctx context.Context, query ethereum.FilterQuery, blockNum uint64) error {
	log.Debug("SaveLatestBlockForQuery in memory", "query", query, "blockNum", blockNum)
	s.blockMap.Store(GetQueryKey(s.chainId, query), blockNum)
	return nil
}

func (s *MemoryStorage) SaveLatestLogForQuery(ctx context.Context, query ethereum.FilterQuery, l types.Log) error {
	log.Debug("SaveLatestLogForQuery in memory", "query", query, "log", l)
	s.logMap.Store(GetQueryKey(s.chainId, query), l)
	return nil
}

func (s *MemoryStorage) SaveFilterLogs(q ethereum.FilterQuery, logs []types.Log) (err error) {
	return nil
}
