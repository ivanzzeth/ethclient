package nonce

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis"
)

var _ Storage = &RedisStorage{}

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

type redSyncMutexWrapper redsync.Mutex

func (l *redSyncMutexWrapper) Lock() {
	(*redsync.Mutex)(l).Lock()
}

func (l *redSyncMutexWrapper) Unlock() {
	(*redsync.Mutex)(l).Unlock()
}

func (s *RedisStorage) NonceLockFrom(from common.Address) sync.Locker {
	mutexKey := fmt.Sprintf("nonce-lock-%s", strings.ToLower(from.Hex()))
	mutex := s.rsync.NewMutex(mutexKey)

	m := redSyncMutexWrapper(*mutex)
	return &m
}

func (s *RedisStorage) GetNonce(account common.Address) (uint64, error) {
	conn, err := s.redisPool.Get(context.Background())
	if err != nil {
		return 0, err
	}

	nonceKey := fmt.Sprintf("nonce-account-%s", strings.ToLower(account.Hex()))
	nonceStr, err := conn.Get(nonceKey)
	if err != nil {
		return 0, err
	}

	if nonceStr == "" {
		nonceStr = "0"
	}

	nonce, err := strconv.Atoi(nonceStr)
	if err != nil {
		return 0, err
	}

	return uint64(nonce), nil
}

func (s *RedisStorage) SetNonce(account common.Address, nonce uint64) error {
	conn, err := s.redisPool.Get(context.Background())
	if err != nil {
		return err
	}

	nonceKey := fmt.Sprintf("nonce-account-%s", strings.ToLower(account.Hex()))

	nonceStr := strconv.Itoa(int(nonce))
	ok, err := conn.Set(nonceKey, nonceStr)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("set nonce failed")
	}

	return nil
}
