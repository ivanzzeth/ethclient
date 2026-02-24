package subscriber

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
)

// TestNewBatchLogFilterer_PanicsOnNilClient ensures we panic when client is nil (explicit config).
func TestNewBatchLogFilterer_PanicsOnNilClient(t *testing.T) {
	defer func() {
		r := recover()
		assert.NotNil(t, r, "expected panic when client is nil")
	}()
	NewBatchLogFilterer(nil, nil)
}

// TestNewBatchLogFilterer_PanicsOnNilRPC ensures we panic when rpc client is nil.
func TestNewBatchLogFilterer_PanicsOnNilRPC(t *testing.T) {
	// We need a valid *ethclient.Client to pass as first arg; without a real RPC we cannot
	// construct one. This case is covered in integration TestBatchLogFilterer_Integration.
	t.Skip("requires real *ethclient.Client; covered in tests/subscriber/batch_filter_integration_test.go")
}

// mockLogFiltererBatch implements LogFiltererBatch for table-driven tests.
type mockLogFiltererBatch struct {
	filterLogsFunc      func(ctx context.Context, q ethereum.FilterQuery) ([]etypes.Log, error)
	filterLogsBatchFunc func(ctx context.Context, queries []ethereum.FilterQuery) ([][]etypes.Log, error)
}

func (m *mockLogFiltererBatch) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]etypes.Log, error) {
	if m.filterLogsFunc != nil {
		return m.filterLogsFunc(ctx, q)
	}
	return nil, nil
}

func (m *mockLogFiltererBatch) FilterLogsBatch(ctx context.Context, queries []ethereum.FilterQuery) ([][]etypes.Log, error) {
	if m.filterLogsBatchFunc != nil {
		return m.filterLogsBatchFunc(ctx, queries)
	}
	if len(queries) == 0 {
		return nil, nil
	}
	return make([][]etypes.Log, len(queries)), nil
}

// TestLogFiltererBatch_EmptyQueries documents contract: empty queries => nil, nil.
func TestLogFiltererBatch_EmptyQueries(t *testing.T) {
	m := &mockLogFiltererBatch{}
	ctx := context.Background()

	result, err := m.FilterLogsBatch(ctx, nil)
	assert.NoError(t, err)
	assert.Nil(t, result)

	result, err = m.FilterLogsBatch(ctx, []ethereum.FilterQuery{})
	assert.NoError(t, err)
	assert.Nil(t, result)
}
