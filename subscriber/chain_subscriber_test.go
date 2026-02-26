// Package subscriber tests. runRealtimeScanner was previously untested (no tests for
// chain_subscriber realtime path), so the minStart/advanceProgress logic was not covered.
// We now test the extracted helpers nextBlockToScanForRealtime and realtimeAdvanceProgress
// so regressions (e.g. scanning from block 1 when storage returns 0, or advancing on 0 logs)
// are caught by unit tests. Full runRealtimeScanner coverage would require integration tests
// with mock RPC and storage.
package subscriber

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNextBlockToScanForRealtime(t *testing.T) {
	subStart := uint64(83278370)

	tests := []struct {
		name     string
		latest   uint64
		query    ethereum.FilterQuery
		wantNext uint64
	}{
		{
			name:     "storage_zero_no_from_block_returns_1",
			latest:   0,
			query:    ethereum.FilterQuery{},
			wantNext: 1,
		},
		{
			name:     "storage_zero_with_from_block_uses_subscription_start",
			latest:   0,
			query:    ethereum.FilterQuery{FromBlock: big.NewInt(0).SetUint64(subStart)},
			wantNext: subStart,
		},
		{
			name:     "storage_just_before_sub_start_uses_subscription_start",
			latest:   subStart - 1,
			query:    ethereum.FilterQuery{FromBlock: big.NewInt(0).SetUint64(subStart)},
			wantNext: subStart,
		},
		{
			name:     "storage_at_sub_start_returns_next_block",
			latest:   subStart,
			query:    ethereum.FilterQuery{FromBlock: big.NewInt(0).SetUint64(subStart)},
			wantNext: subStart + 1,
		},
		{
			name:     "storage_ahead_of_sub_start_returns_latest_plus_one",
			latest:   subStart + 100,
			query:    ethereum.FilterQuery{FromBlock: big.NewInt(0).SetUint64(subStart)},
			wantNext: subStart + 101,
		},
		{
			name:     "storage_nonzero_no_from_block_returns_latest_plus_one",
			latest:   1000,
			query:    ethereum.FilterQuery{},
			wantNext: 1001,
		},
		{
			name:     "from_block_one_storage_zero_returns_one",
			latest:   0,
			query:    ethereum.FilterQuery{FromBlock: big.NewInt(1)},
			wantNext: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nextBlockToScanForRealtime(tt.latest, tt.query)
			require.Equal(t, tt.wantNext, got, "nextBlockToScanForRealtime(%d, query)", tt.latest)
		})
	}
}

// TestRealtimeAdvanceProgress ensures we only advance progress when the merged FilterLogs
// response has at least one log (avoids permanently skipping blocks when node returns 0 logs).
func TestRealtimeAdvanceProgress_ZeroLogsDoesNotAdvance(t *testing.T) {
	assert.False(t, realtimeAdvanceProgress(0), "0 logs must not advance progress")
	assert.True(t, realtimeAdvanceProgress(1), "1 log must advance progress")
	assert.True(t, realtimeAdvanceProgress(10), "10 logs must advance progress")
}
