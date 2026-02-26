package handler

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ivanzzeth/ethclient/subscriber"
	"github.com/stretchr/testify/require"
)

func TestNewSimpleQueryHandler(t *testing.T) {
	chainID := big.NewInt(1337)
	storage := subscriber.NewMemoryStorage(chainID)
	h := NewSimpleQueryHandler(storage)
	require.NotNil(t, h)
	require.Same(t, storage, h.SubscriberStorage)
}

func TestSimpleQueryHandler_HandleQuery(t *testing.T) {
	ctx := context.Background()
	chainID := big.NewInt(1337)
	storage := subscriber.NewMemoryStorage(chainID)
	h := NewSimpleQueryHandler(storage)

	addr := common.HexToAddress("0x0000000000000000000000000000000000000001")
	q := subscriber.NewQuery(chainID, ethereum.FilterQuery{Addresses: []common.Address{addr}})
	log1 := types.Log{BlockNumber: 10, TxIndex: 0, Index: 0, Address: addr, Topics: []common.Hash{}}

	err := h.HandleQuery(ctx, q, log1)
	require.NoError(t, err)

	latest, err := storage.LatestBlockForQuery(ctx, q.FilterQuery)
	require.NoError(t, err)
	require.Equal(t, uint64(10), latest)

	log2 := types.Log{BlockNumber: 5, TxIndex: 0, Index: 0, Address: addr, Topics: []common.Hash{}}
	err = h.HandleQuery(ctx, q, log2)
	require.NoError(t, err)
	// Handler compares against in-memory latestBlock (starts 0); 5 > 0 so block 5 is saved
	latest, err = storage.LatestBlockForQuery(ctx, q.FilterQuery)
	require.NoError(t, err)
	require.Equal(t, uint64(5), latest)

	log3 := types.Log{BlockNumber: 15, TxIndex: 0, Index: 0, Address: addr, Topics: []common.Hash{}}
	err = h.HandleQuery(ctx, q, log3)
	require.NoError(t, err)
	latest, err = storage.LatestBlockForQuery(ctx, q.FilterQuery)
	require.NoError(t, err)
	require.Equal(t, uint64(15), latest)
}
