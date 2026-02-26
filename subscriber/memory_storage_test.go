package subscriber

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryStorage_FilterLogs_ReturnsNotImplemented(t *testing.T) {
	ctx := context.Background()
	chainID := big.NewInt(1)
	s := NewMemoryStorage(chainID)
	q := ethereum.FilterQuery{Addresses: []common.Address{{}}}
	logs, err := s.FilterLogs(ctx, q)
	require.Error(t, err)
	assert.Nil(t, logs)
	assert.Contains(t, err.Error(), "not implemented")
}
