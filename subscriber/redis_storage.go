package subscriber

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis"
	"github.com/ivanzz/ethclient/ds/locker"
)

type RedisStorage struct {
	redisPool redis.Pool
	rsync     *redsync.Redsync
}

func NewRedisStorage(pool redis.Pool) *RedisStorage {
	return &RedisStorage{
		redisPool: pool,
		rsync:     redsync.New(pool),
	}
}

func (s *RedisStorage) QueryLock(q ethereum.FilterQuery) sync.Locker {
	key := fmt.Sprintf("query_%v", GetQueryKey(q))

	mutext := s.rsync.NewMutex(key)
	m := locker.RedSyncMutexWrapper(*mutext)
	return &m
}

func (s *RedisStorage) LatestBlockForQuery(ctx context.Context, query ethereum.FilterQuery) (uint64, error) {
	conn, err := s.redisPool.Get(ctx)
	if err != nil {
		return 0, err
	}

	locker := s.QueryLock(query)
	locker.Lock()
	defer locker.Unlock()

	key := fmt.Sprintf("latest_block_of_query_%v", GetQueryKey(query))
	blockStr, err := conn.Get(key)
	if err != nil {
		return 0, err
	}
	if blockStr == "" {
		return 0, nil
	}
	block, err := strconv.Atoi(blockStr)
	if err != nil {
		return 0, err
	}

	return uint64(block), nil
}

func (s *RedisStorage) LatestLogForQuery(ctx context.Context, query ethereum.FilterQuery) (types.Log, error) {
	l := types.Log{}
	conn, err := s.redisPool.Get(ctx)
	if err != nil {
		return l, err
	}

	locker := s.QueryLock(query)
	locker.Lock()
	defer locker.Unlock()

	key := fmt.Sprintf("latest_log_of_query_%v", GetQueryKey(query))
	logStr, err := conn.Get(key)
	if err != nil {
		return l, err
	}

	if logStr == "" {
		return l, nil
	}

	err = l.UnmarshalJSON([]byte(logStr))
	if err != nil {
		return l, err
	}

	return l, nil
}

func (s *RedisStorage) SaveLatestBlockForQuery(ctx context.Context, query ethereum.FilterQuery, blockNum uint64) error {
	conn, err := s.redisPool.Get(ctx)
	if err != nil {
		return err
	}

	locker := s.QueryLock(query)
	locker.Lock()
	defer locker.Unlock()

	key := fmt.Sprintf("latest_block_of_query_%v", GetQueryKey(query))
	blockNumStr := strconv.Itoa(int(blockNum))
	_, err = conn.Set(key, blockNumStr)
	if err != nil {
		return err
	}

	return nil
}

func (s *RedisStorage) SaveLatestLogForQuery(ctx context.Context, query ethereum.FilterQuery, log types.Log) error {
	conn, err := s.redisPool.Get(ctx)
	if err != nil {
		return err
	}

	locker := s.QueryLock(query)
	locker.Lock()
	defer locker.Unlock()

	key := fmt.Sprintf("latest_log_of_query_%v", GetQueryKey(query))
	logBytes, err := log.MarshalJSON()
	if err != nil {
		return err
	}

	_, err = conn.Set(key, string(logBytes))
	if err != nil {
		return err
	}

	return nil
}
