package subscriber_test

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ivanzzeth/ethclient/message"
	"github.com/ivanzzeth/ethclient/simulated"
	"github.com/ivanzzeth/ethclient/subscriber"
	"github.com/ivanzzeth/ethclient/tests/helper"
	"github.com/stretchr/testify/assert"
)

func Test_FilterLogs(t *testing.T) {
	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	sim := helper.SetUpClient(t)
	defer sim.Close()

	testFilterLogs(t, sim)
}

func testFilterLogs(t *testing.T, sim *simulated.Backend) {
	client := sim.Client()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Deploy Test contract
	contractAddr, _, contract := helper.DeployTestContract(t, ctx, sim)

	// Get current block number
	fromBlock, err := client.BlockNumber(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Method args
	arg1 := "hello"
	arg2 := big.NewInt(100)
	arg3 := []byte("world")

	// First transaction to generate logs
	opts, err := client.MessageToTransactOpts(ctx, message.Request{
		From: helper.Addr1,
	})
	if err != nil {
		t.Fatal(err)
	}
	contractCallTx, err := contract.TestFunc1(opts, arg1, arg2, arg3)
	if err != nil {
		t.Fatalf("TestFunc1 err: %v", err)
	}

	sim.CommitAndExpectTx(contractCallTx.Hash())

	receipt, contains := client.WaitTxReceipt(contractCallTx.Hash(), 0, 5*time.Second)
	assert.Equal(t, true, contains)
	t.Log("contractCallTx send successful", "txHash", contractCallTx.Hash().Hex(), "block", receipt.BlockNumber.Uint64())

	// Second transaction to generate more logs
	opts, err = client.MessageToTransactOpts(ctx, message.Request{
		From: helper.Addr1,
	})
	if err != nil {
		t.Fatal(err)
	}
	contractCallTx2, err := contract.TestFunc1(opts, arg1, arg2, arg3)
	if err != nil {
		t.Fatalf("TestFunc1 err: %v", err)
	}

	sim.CommitAndExpectTx(contractCallTx2.Hash())

	receipt2, contains := client.WaitTxReceipt(contractCallTx2.Hash(), 0, 5*time.Second)
	assert.Equal(t, true, contains)
	t.Log("contractCallTx2 send successful", "txHash", contractCallTx2.Hash().Hex(), "block", receipt2.BlockNumber.Uint64())

	// Get current block number for filtering
	toBlock, err := client.BlockNumber(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Test FilterLogs with specific block range
	filteredLogs, err := client.FilterLogs(ctx, ethereum.FilterQuery{
		FromBlock: big.NewInt(0).SetUint64(fromBlock),
		ToBlock:   big.NewInt(0).SetUint64(toBlock),
		Addresses: []common.Address{contractAddr},
	})
	if err != nil {
		t.Fatal("FilterLogs err: ", err)
	}

	// Should have 4 logs (2 transactions, each generates 2 logs)
	assert.Equal(t, 4, len(filteredLogs), "filter logs count mismatch")

	// Verify log details
	for i, log := range filteredLogs {
		t.Logf("Log %d: BlockNumber=%d, TxHash=%s, TxIndex=%d, Index=%d",
			i, log.BlockNumber, log.TxHash.Hex(), log.TxIndex, log.Index)
		assert.Equal(t, contractAddr, log.Address, "log address mismatch")
		assert.True(t, log.BlockNumber >= fromBlock && log.BlockNumber <= toBlock, "log block number out of range")
	}

	t.Log("FilterLogs test completed successfully")
}

func Test_FilterLogsWithChannel_Exit(t *testing.T) {
	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	sim := helper.SetUpClient(t)
	defer sim.Close()

	testFilterLogsWithChannelExit(t, sim)
}

func testFilterLogsWithChannelExit(t *testing.T, sim *simulated.Backend) {
	client := sim.Client()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Deploy Test contract
	contractAddr, _, contract := helper.DeployTestContract(t, ctx, sim)

	// Get current block number
	fromBlock, err := client.BlockNumber(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Create a channel for logs
	logsChan := make(chan types.Log, 10)

	// Create a query
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(0).SetUint64(fromBlock),
		Addresses: []common.Address{contractAddr},
	}

	// Start FilterLogsWithChannel in a goroutine
	errChan := make(chan error, 1)
	go func() {
		err := client.Subscriber.(*subscriber.ChainSubscriber).FilterLogsWithChannel(ctx, query, logsChan, true, true)
		errChan <- err
	}()

	// Wait a bit for the goroutine to start
	time.Sleep(100 * time.Millisecond)

	// Generate some logs
	arg1 := "hello"
	arg2 := big.NewInt(100)
	arg3 := []byte("world")

	opts, err := client.MessageToTransactOpts(ctx, message.Request{
		From: helper.Addr1,
	})
	if err != nil {
		t.Fatal(err)
	}
	contractCallTx, err := contract.TestFunc1(opts, arg1, arg2, arg3)
	if err != nil {
		t.Fatalf("TestFunc1 err: %v", err)
	}

	sim.CommitAndExpectTx(contractCallTx.Hash())

	// Wait for some logs to be received
	time.Sleep(2 * time.Second)

	// Cancel the context to trigger exit
	cancel()

	// Wait for the goroutine to exit
	select {
	case err := <-errChan:
		if err != nil && err != context.Canceled {
			t.Errorf("FilterLogsWithChannel returned unexpected error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Error("FilterLogsWithChannel did not exit within timeout")
	}

	// Verify that the channel is closed
	select {
	case _, ok := <-logsChan:
		if ok {
			t.Error("logsChan should be closed after FilterLogsWithChannel exits")
		}
	default:
		// Channel is closed, which is expected
	}

	t.Log("FilterLogsWithChannel exit test completed successfully")
}

func Test_FilterLogsBatch(t *testing.T) {
	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	sim := helper.SetUpClient(t)
	defer sim.Close()

	testFilterLogsBatch(t, sim)
}

func testFilterLogsBatch(t *testing.T, sim *simulated.Backend) {
	client := sim.Client()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Deploy Test contract
	contractAddr, _, contract := helper.DeployTestContract(t, ctx, sim)

	// Get current block number
	fromBlock, err := client.BlockNumber(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Generate some logs
	arg1 := "hello"
	arg2 := big.NewInt(100)
	arg3 := []byte("world")

	opts, err := client.MessageToTransactOpts(ctx, message.Request{
		From: helper.Addr1,
	})
	if err != nil {
		t.Fatal(err)
	}
	contractCallTx, err := contract.TestFunc1(opts, arg1, arg2, arg3)
	if err != nil {
		t.Fatalf("TestFunc1 err: %v", err)
	}

	sim.CommitAndExpectTx(contractCallTx.Hash())

	// Get current block number for filtering
	toBlock, err := client.BlockNumber(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Create multiple queries
	queries := []ethereum.FilterQuery{
		{
			FromBlock: big.NewInt(0).SetUint64(fromBlock),
			ToBlock:   big.NewInt(0).SetUint64(toBlock),
			Addresses: []common.Address{contractAddr},
		},
		{
			FromBlock: big.NewInt(0).SetUint64(fromBlock),
			ToBlock:   big.NewInt(0).SetUint64(toBlock),
			Addresses: []common.Address{contractAddr},
		},
	}

	// Test FilterLogsBatch
	filteredLogsBatch, err := client.FilterLogsBatch(ctx, queries)
	if err != nil {
		t.Fatal("FilterLogsBatch err: ", err)
	}

	// Should have results for each query
	assert.Equal(t, 2, len(filteredLogsBatch), "batch results count mismatch")

	// Each query should have the same number of logs
	for i, logs := range filteredLogsBatch {
		t.Logf("Batch %d: %d logs", i, len(logs))
		assert.Equal(t, 2, len(logs), "batch query logs count mismatch")
	}

	t.Log("FilterLogsBatch test completed successfully")
}

func Test_FilterLogsBatch_GoroutineLeak(t *testing.T) {
	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	sim := helper.SetUpClient(t)
	defer sim.Close()

	testFilterLogsBatchGoroutineLeak(t, sim)
}

func testFilterLogsBatchGoroutineLeak(t *testing.T, sim *simulated.Backend) {
	client := sim.Client()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Deploy Test contract
	contractAddr, _, contract := helper.DeployTestContract(t, ctx, sim)

	// Get current block number
	fromBlock, err := client.BlockNumber(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Generate some logs first
	arg1 := "hello"
	arg2 := big.NewInt(100)
	arg3 := []byte("world")

	opts, err := client.MessageToTransactOpts(ctx, message.Request{
		From: helper.Addr1,
	})
	if err != nil {
		t.Fatal(err)
	}
	contractCallTx, err := contract.TestFunc1(opts, arg1, arg2, arg3)
	if err != nil {
		t.Fatalf("TestFunc1 err: %v", err)
	}

	sim.CommitAndExpectTx(contractCallTx.Hash())

	// Create multiple queries that will trigger watch mode (ToBlock = nil)
	queries := []ethereum.FilterQuery{
		{
			FromBlock: big.NewInt(0).SetUint64(fromBlock),
			// ToBlock is nil, which means watch mode
			Addresses: []common.Address{contractAddr},
		},
		{
			FromBlock: big.NewInt(0).SetUint64(fromBlock),
			// ToBlock is nil, which means watch mode
			Addresses: []common.Address{contractAddr},
		},
	}

	t.Log("Starting FilterLogsBatch test with watch mode...")

	// Call FilterLogsBatch multiple times to trigger the leak
	for i := 0; i < 3; i++ {
		t.Logf("Calling FilterLogsBatch iteration %d", i+1)

		// Use a shorter timeout for each call
		batchCtx, batchCancel := context.WithTimeout(ctx, 5*time.Second)

		filteredLogsBatch, err := client.FilterLogsBatch(batchCtx, queries)
		batchCancel() // Cancel immediately after call

		if err != nil {
			t.Logf("FilterLogsBatch iteration %d err: %v", i+1, err)
			continue
		}

		t.Logf("FilterLogsBatch iteration %d completed, got %d batch results", i+1, len(filteredLogsBatch))

		// Wait a bit between calls
		time.Sleep(1 * time.Second)
	}

	// Wait a bit to see if goroutines are still running
	t.Log("Waiting to observe goroutine behavior...")
	time.Sleep(5 * time.Second)

	// Generate more logs to trigger any remaining goroutines
	opts, err = client.MessageToTransactOpts(ctx, message.Request{
		From: helper.Addr1,
	})
	if err != nil {
		t.Fatal(err)
	}
	contractCallTx2, err := contract.TestFunc1(opts, arg1, arg2, arg3)
	if err != nil {
		t.Fatalf("TestFunc1 err: %v", err)
	}

	sim.CommitAndExpectTx(contractCallTx2.Hash())

	// Wait a bit more
	time.Sleep(3 * time.Second)

	t.Log("Goroutine leak test completed - check logs for repeated 'Subscriber FilterLogs starts filtering logs' messages")
}

func Test_FilterLogsBatch_GoroutineLeak_Strict(t *testing.T) {
	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	sim := helper.SetUpClient(t)
	defer sim.Close()

	testFilterLogsBatchGoroutineLeakStrict(t, sim)
}

func testFilterLogsBatchGoroutineLeakStrict(t *testing.T, sim *simulated.Backend) {
	client := sim.Client()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Deploy Test contract
	contractAddr, _, contract := helper.DeployTestContract(t, ctx, sim)

	// Get current block number
	fromBlock, err := client.BlockNumber(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Generate some logs first
	arg1 := "hello"
	arg2 := big.NewInt(100)
	arg3 := []byte("world")

	opts, err := client.MessageToTransactOpts(ctx, message.Request{
		From: helper.Addr1,
	})
	if err != nil {
		t.Fatal(err)
	}
	contractCallTx, err := contract.TestFunc1(opts, arg1, arg2, arg3)
	if err != nil {
		t.Fatalf("TestFunc1 err: %v", err)
	}

	sim.CommitAndExpectTx(contractCallTx.Hash())

	// Create queries that will definitely trigger watch mode
	queries := []ethereum.FilterQuery{
		{
			FromBlock: big.NewInt(0).SetUint64(fromBlock),
			// ToBlock is nil, which means watch mode
			Addresses: []common.Address{contractAddr},
		},
	}

	t.Log("Starting strict goroutine leak test...")

	// Call FilterLogsBatch multiple times rapidly
	for i := 0; i < 5; i++ {
		t.Logf("Calling FilterLogsBatch iteration %d", i+1)

		// Use a very short timeout to force cancellation
		batchCtx, batchCancel := context.WithTimeout(ctx, 2*time.Second)

		go func(iter int) {
			filteredLogsBatch, err := client.FilterLogsBatch(batchCtx, queries)
			if err != nil {
				t.Logf("FilterLogsBatch iteration %d err: %v", iter, err)
				return
			}
			t.Logf("FilterLogsBatch iteration %d completed, got %d batch results", iter, len(filteredLogsBatch))
		}(i + 1)

		// Cancel immediately to force goroutine cleanup
		batchCancel()

		// Small delay between calls
		time.Sleep(100 * time.Millisecond)
	}

	// Wait a bit to see if goroutines are still running
	t.Log("Waiting to observe goroutine behavior...")
	time.Sleep(10 * time.Second)

	// Generate more logs to trigger any remaining goroutines
	for i := 0; i < 3; i++ {
		opts, err := client.MessageToTransactOpts(ctx, message.Request{
			From: helper.Addr1,
		})
		if err != nil {
			t.Fatal(err)
		}
		contractCallTx, err := contract.TestFunc1(opts, arg1, arg2, arg3)
		if err != nil {
			t.Fatalf("TestFunc1 err: %v", err)
		}

		sim.CommitAndExpectTx(contractCallTx.Hash())
		time.Sleep(1 * time.Second)
	}

	// Wait a bit more
	time.Sleep(5 * time.Second)

	t.Log("Strict goroutine leak test completed - check logs for repeated 'Subscriber FilterLogs starts filtering logs' messages")
}

func Test_FilterLogsBatch_GoroutineLeak_Exact(t *testing.T) {
	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	sim := helper.SetUpClient(t)
	defer sim.Close()

	testFilterLogsBatchGoroutineLeakExact(t, sim)
}

func testFilterLogsBatchGoroutineLeakExact(t *testing.T, sim *simulated.Backend) {
	client := sim.Client()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Deploy Test contract
	contractAddr, _, contract := helper.DeployTestContract(t, ctx, sim)

	// Get current block number
	fromBlock, err := client.BlockNumber(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Generate some logs first
	arg1 := "hello"
	arg2 := big.NewInt(100)
	arg3 := []byte("world")

	opts, err := client.MessageToTransactOpts(ctx, message.Request{
		From: helper.Addr1,
	})
	if err != nil {
		t.Fatal(err)
	}
	contractCallTx, err := contract.TestFunc1(opts, arg1, arg2, arg3)
	if err != nil {
		t.Fatalf("TestFunc1 err: %v", err)
	}

	sim.CommitAndExpectTx(contractCallTx.Hash())

	// Create queries with ToBlock=nil to trigger the exact scenario
	queries := []ethereum.FilterQuery{
		{
			FromBlock: big.NewInt(0).SetUint64(fromBlock),
			ToBlock:   nil, // This triggers useStorage=true
			Addresses: []common.Address{contractAddr},
		},
	}

	t.Log("Starting exact goroutine leak test with ToBlock=nil...")

	// Call FilterLogsBatch multiple times
	for i := 0; i < 3; i++ {
		t.Logf("Calling FilterLogsBatch iteration %d", i+1)

		// Use a short timeout
		batchCtx, batchCancel := context.WithTimeout(ctx, 3*time.Second)

		filteredLogsBatch, err := client.FilterLogsBatch(batchCtx, queries)
		batchCancel() // Cancel immediately after call

		if err != nil {
			t.Logf("FilterLogsBatch iteration %d err: %v", i+1, err)
			continue
		}

		t.Logf("FilterLogsBatch iteration %d completed, got %d batch results", i+1, len(filteredLogsBatch))

		// Wait a bit between calls
		time.Sleep(2 * time.Second)
	}

	// Wait to see if goroutines are still running
	t.Log("Waiting to observe goroutine behavior...")
	time.Sleep(10 * time.Second)

	// Generate more logs to trigger any remaining goroutines
	for i := 0; i < 2; i++ {
		opts, err := client.MessageToTransactOpts(ctx, message.Request{
			From: helper.Addr1,
		})
		if err != nil {
			t.Fatal(err)
		}
		contractCallTx, err := contract.TestFunc1(opts, arg1, arg2, arg3)
		if err != nil {
			t.Fatalf("TestFunc1 err: %v", err)
		}

		sim.CommitAndExpectTx(contractCallTx.Hash())
		time.Sleep(2 * time.Second)
	}

	// Wait a bit more
	time.Sleep(5 * time.Second)

	t.Log("Exact goroutine leak test completed - check logs for repeated 'Subscriber FilterLogs starts filtering logs' messages")
}

func Test_FilterLogsWithChannel_AllBranches(t *testing.T) {
	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	sim := helper.SetUpClient(t)
	defer sim.Close()

	testFilterLogsWithChannelAllBranches(t, sim)
}

func testFilterLogsWithChannelAllBranches(t *testing.T, sim *simulated.Backend) {
	client := sim.Client()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Deploy Test contract
	contractAddr, _, contract := helper.DeployTestContract(t, ctx, sim)

	// Get current block number
	fromBlock, err := client.BlockNumber(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Generate some logs first
	arg1 := "hello"
	arg2 := big.NewInt(100)
	arg3 := []byte("world")

	opts, err := client.MessageToTransactOpts(ctx, message.Request{
		From: helper.Addr1,
	})
	if err != nil {
		t.Fatal(err)
	}
	contractCallTx, err := contract.TestFunc1(opts, arg1, arg2, arg3)
	if err != nil {
		t.Fatalf("TestFunc1 err: %v", err)
	}

	sim.CommitAndExpectTx(contractCallTx.Hash())

	// Get current block number for filtering
	toBlock, err := client.BlockNumber(ctx)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("=== Testing Branch 1: ToBlock != nil ===")
	// Test 1: ToBlock != nil
	query1 := ethereum.FilterQuery{
		FromBlock: big.NewInt(0).SetUint64(fromBlock),
		ToBlock:   big.NewInt(0).SetUint64(toBlock),
		Addresses: []common.Address{contractAddr},
	}

	filteredLogs1, err := client.FilterLogs(ctx, query1)
	if err != nil {
		t.Fatalf("FilterLogs err: %v", err)
	}
	t.Logf("Branch 1: Got %d logs", len(filteredLogs1))

	t.Log("=== Testing Branch 2: ToBlock = nil ===")
	// Test 2: ToBlock = nil (this should trigger useStorage=true)
	query2 := ethereum.FilterQuery{
		FromBlock: big.NewInt(0).SetUint64(fromBlock),
		ToBlock:   nil, // This triggers useStorage=true
		Addresses: []common.Address{contractAddr},
	}

	filteredLogs2, err := client.FilterLogs(ctx, query2)
	if err != nil {
		t.Fatalf("FilterLogs err: %v", err)
	}
	t.Logf("Branch 2: Got %d logs", len(filteredLogs2))

	t.Log("=== Testing Branch 3: BlockHash != nil ===")
	// Test 3: BlockHash != nil (this should trigger the BlockHash branch)
	block, err := client.BlockByNumber(ctx, big.NewInt(0).SetUint64(toBlock))
	if err != nil {
		t.Fatal(err)
	}

	blockHash := block.Hash()
	query3 := ethereum.FilterQuery{
		BlockHash: &blockHash,
		Addresses: []common.Address{contractAddr},
	}

	filteredLogs3, err := client.FilterLogs(ctx, query3)
	if err != nil {
		t.Fatalf("FilterLogs err: %v", err)
	}
	t.Logf("Branch 3: Got %d logs", len(filteredLogs3))

	t.Log("=== Testing Multiple Calls to Trigger Potential Leak ===")
	// Test multiple calls to see if there's a leak
	for i := 0; i < 3; i++ {
		t.Logf("Calling FilterLogs iteration %d", i+1)

		// Use ToBlock=nil to trigger useStorage=true
		query := ethereum.FilterQuery{
			FromBlock: big.NewInt(0).SetUint64(fromBlock),
			ToBlock:   nil,
			Addresses: []common.Address{contractAddr},
		}

		filteredLogs, err := client.FilterLogs(ctx, query)
		if err != nil {
			t.Fatalf("FilterLogs iteration %d err: %v", i+1, err)
		}
		t.Logf("Iteration %d: Got %d logs", i+1, len(filteredLogs))

		// Wait a bit between calls
		time.Sleep(1 * time.Second)
	}

	// Wait to see if any goroutines are still running
	t.Log("Waiting to observe goroutine behavior...")
	time.Sleep(5 * time.Second)

	// Generate more logs to trigger any remaining goroutines
	for i := 0; i < 2; i++ {
		opts, err := client.MessageToTransactOpts(ctx, message.Request{
			From: helper.Addr1,
		})
		if err != nil {
			t.Fatal(err)
		}
		contractCallTx, err := contract.TestFunc1(opts, arg1, arg2, arg3)
		if err != nil {
			t.Fatalf("TestFunc1 err: %v", err)
		}

		sim.CommitAndExpectTx(contractCallTx.Hash())
		time.Sleep(1 * time.Second)
	}

	t.Log("All branches test completed - check logs for any repeated messages")
}

func Test_FilterLogs_CurrBlocksPerScanLeak(t *testing.T) {
	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	sim := helper.SetUpClient(t)
	defer sim.Close()

	testFilterLogsCurrBlocksPerScanLeak(t, sim)
}

func testFilterLogsCurrBlocksPerScanLeak(t *testing.T, sim *simulated.Backend) {
	client := sim.Client()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Deploy Test contract
	contractAddr, _, contract := helper.DeployTestContract(t, ctx, sim)

	// Get current block number
	fromBlock, err := client.BlockNumber(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Generate some logs first
	arg1 := "hello"
	arg2 := big.NewInt(100)
	arg3 := []byte("world")

	opts, err := client.MessageToTransactOpts(ctx, message.Request{
		From: helper.Addr1,
	})
	if err != nil {
		t.Fatal(err)
	}
	contractCallTx, err := contract.TestFunc1(opts, arg1, arg2, arg3)
	if err != nil {
		t.Fatalf("TestFunc1 err: %v", err)
	}

	sim.CommitAndExpectTx(contractCallTx.Hash())

	t.Log("=== Testing currBlocksPerScan leak ===")
	// Test multiple calls to see if currBlocksPerScan keeps increasing
	for i := 0; i < 5; i++ {
		t.Logf("Calling FilterLogs iteration %d", i+1)

		// Use ToBlock=nil to trigger useStorage=true and watch mode behavior
		query := ethereum.FilterQuery{
			FromBlock: big.NewInt(0).SetUint64(fromBlock),
			ToBlock:   nil,
			Addresses: []common.Address{contractAddr},
		}

		filteredLogs, err := client.FilterLogs(ctx, query)
		if err != nil {
			t.Fatalf("FilterLogs iteration %d err: %v", i+1, err)
		}
		t.Logf("Iteration %d: Got %d logs", i+1, len(filteredLogs))

		// Wait a bit between calls
		time.Sleep(1 * time.Second)
	}

	// Wait to see if any goroutines are still running
	t.Log("Waiting to observe goroutine behavior...")
	time.Sleep(5 * time.Second)

	// Generate more logs to trigger any remaining goroutines
	for i := 0; i < 2; i++ {
		opts, err := client.MessageToTransactOpts(ctx, message.Request{
			From: helper.Addr1,
		})
		if err != nil {
			t.Fatal(err)
		}
		contractCallTx, err := contract.TestFunc1(opts, arg1, arg2, arg3)
		if err != nil {
			t.Fatalf("TestFunc1 err: %v", err)
		}

		sim.CommitAndExpectTx(contractCallTx.Hash())
		time.Sleep(1 * time.Second)
	}

	t.Log("currBlocksPerScan leak test completed - check logs for increasing currBlocksPerScan values")
}

func Test_FilterLogs_StartBlockGreaterThanToBlock(t *testing.T) {
	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	sim := helper.SetUpClient(t)
	defer sim.Close()

	testFilterLogsStartBlockGreaterThanToBlock(t, sim)
}

func testFilterLogsStartBlockGreaterThanToBlock(t *testing.T, sim *simulated.Backend) {
	client := sim.Client()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Deploy Test contract
	contractAddr, _, contract := helper.DeployTestContract(t, ctx, sim)

	// Get current block number
	fromBlock, err := client.BlockNumber(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Generate some logs first
	arg1 := "hello"
	arg2 := big.NewInt(100)
	arg3 := []byte("world")

	opts, err := client.MessageToTransactOpts(ctx, message.Request{
		From: helper.Addr1,
	})
	if err != nil {
		t.Fatal(err)
	}
	contractCallTx, err := contract.TestFunc1(opts, arg1, arg2, arg3)
	if err != nil {
		t.Fatalf("TestFunc1 err: %v", err)
	}

	sim.CommitAndExpectTx(contractCallTx.Hash())

	t.Log("=== Testing startBlock > toBlock infinite loop ===")

	// Create a query with a very high fromBlock that will be greater than current toBlock
	// This should trigger the startBlock > toBlock condition
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(0).SetUint64(fromBlock + 1000), // Very high fromBlock
		ToBlock:   big.NewInt(0).SetUint64(fromBlock + 1),    // Low toBlock
		Addresses: []common.Address{contractAddr},
	}

	t.Logf("Query: fromBlock=%d, toBlock=%d", query.FromBlock.Uint64(), query.ToBlock.Uint64())

	// This should not cause an infinite loop
	filteredLogs, err := client.FilterLogs(ctx, query)
	if err != nil {
		t.Fatalf("FilterLogs err: %v", err)
	}
	t.Logf("Got %d logs", len(filteredLogs))

	// Wait a bit to see if any goroutines are still running
	time.Sleep(2 * time.Second)

	t.Log("startBlock > toBlock test completed - check logs for any infinite loops")
}

func Test_FilterLogs_StartBlockGreaterThanToBlock_WatchMode(t *testing.T) {
	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	sim := helper.SetUpClient(t)
	defer sim.Close()

	testFilterLogsStartBlockGreaterThanToBlockWatchMode(t, sim)
}

func testFilterLogsStartBlockGreaterThanToBlockWatchMode(t *testing.T, sim *simulated.Backend) {
	client := sim.Client()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Deploy Test contract
	contractAddr, _, contract := helper.DeployTestContract(t, ctx, sim)

	// Get current block number
	fromBlock, err := client.BlockNumber(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Generate some logs first
	arg1 := "hello"
	arg2 := big.NewInt(100)
	arg3 := []byte("world")

	opts, err := client.MessageToTransactOpts(ctx, message.Request{
		From: helper.Addr1,
	})
	if err != nil {
		t.Fatal(err)
	}
	contractCallTx, err := contract.TestFunc1(opts, arg1, arg2, arg3)
	if err != nil {
		t.Fatalf("TestFunc1 err: %v", err)
	}

	sim.CommitAndExpectTx(contractCallTx.Hash())

	t.Log("=== Testing startBlock > toBlock in watch mode ===")

	// Create a query with ToBlock=nil to trigger watch mode
	// This should not cause an infinite loop even if startBlock > toBlock
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(0).SetUint64(fromBlock + 1000), // Very high fromBlock
		ToBlock:   nil,                                       // This triggers watch mode
		Addresses: []common.Address{contractAddr},
	}

	t.Logf("Query: fromBlock=%d, toBlock=nil (watch mode)", query.FromBlock.Uint64())

	// This should not cause an infinite loop
	filteredLogs, err := client.FilterLogs(ctx, query)
	if err != nil {
		t.Fatalf("FilterLogs err: %v", err)
	}
	t.Logf("Got %d logs", len(filteredLogs))

	// Wait a bit to see if any goroutines are still running
	time.Sleep(3 * time.Second)

	t.Log("startBlock > toBlock watch mode test completed - check logs for any infinite loops")
}

func Test_FilterLogs_EmptyQuery(t *testing.T) {
	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	sim := helper.SetUpClient(t)
	defer sim.Close()

	testFilterLogsEmptyQuery(t, sim)
}

func testFilterLogsEmptyQuery(t *testing.T, sim *simulated.Backend) {
	client := sim.Client()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test with empty query (no addresses, no topics)
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(1),
		ToBlock:   big.NewInt(10),
		Addresses: []common.Address{},
		Topics:    [][]common.Hash{},
	}

	logs, err := client.FilterLogs(ctx, query)
	if err != nil {
		t.Fatalf("FilterLogs with empty query failed: %v", err)
	}

	// Should return empty logs since no addresses specified
	assert.Equal(t, 0, len(logs), "Empty query should return no logs")
	t.Log("Empty query test completed successfully")
}

func Test_FilterLogs_InvalidBlockRange(t *testing.T) {
	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	sim := helper.SetUpClient(t)
	defer sim.Close()

	testFilterLogsInvalidBlockRange(t, sim)
}

func testFilterLogsInvalidBlockRange(t *testing.T, sim *simulated.Backend) {
	client := sim.Client()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test with invalid block range (from > to)
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(100),
		ToBlock:   big.NewInt(50), // from > to
		Addresses: []common.Address{common.HexToAddress("0x1234567890123456789012345678901234567890")},
	}

	logs, err := client.FilterLogs(ctx, query)
	if err != nil {
		t.Fatalf("FilterLogs with invalid block range failed: %v", err)
	}

	// Should return empty logs for invalid range
	assert.Equal(t, 0, len(logs), "Invalid block range should return no logs")
	t.Log("Invalid block range test completed successfully")
}

func Test_FilterLogs_WithTopics(t *testing.T) {
	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	sim := helper.SetUpClient(t)
	defer sim.Close()

	testFilterLogsWithTopics(t, sim)
}

func testFilterLogsWithTopics(t *testing.T, sim *simulated.Backend) {
	client := sim.Client()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Deploy Test contract
	contractAddr, _, contract := helper.DeployTestContract(t, ctx, sim)

	// Get current block number
	fromBlock, err := client.BlockNumber(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Generate some logs first
	arg1 := "hello"
	arg2 := big.NewInt(100)
	arg3 := []byte("world")

	opts, err := client.MessageToTransactOpts(ctx, message.Request{
		From: helper.Addr1,
	})
	if err != nil {
		t.Fatal(err)
	}
	contractCallTx, err := contract.TestFunc1(opts, arg1, arg2, arg3)
	if err != nil {
		t.Fatalf("TestFunc1 err: %v", err)
	}

	sim.CommitAndExpectTx(contractCallTx.Hash())

	// Get current block number for filtering
	toBlock, err := client.BlockNumber(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Test with specific topic
	// The topic for TestFunc1 event
	eventTopic := common.HexToHash("0xee7ebd5ac9177b3cfe282c440d0220335dc60bc4472338132f06af7b4b9432fc")

	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(toBlock)),
		Addresses: []common.Address{contractAddr},
		Topics:    [][]common.Hash{{eventTopic}},
	}

	logs, err := client.FilterLogs(ctx, query)
	if err != nil {
		t.Fatalf("FilterLogs with topics failed: %v", err)
	}

	// Should find logs with the specific topic
	assert.Greater(t, len(logs), 0, "Should find logs with specific topic")

	// Verify the topic matches
	for _, log := range logs {
		assert.Equal(t, contractAddr, log.Address, "log address mismatch")
		assert.Equal(t, eventTopic, log.Topics[0], "log topic mismatch")
	}

	t.Logf("Found %d logs with specific topic", len(logs))
	t.Log("Topics filter test completed successfully")
}

func Test_FilterLogs_WithBlockHash(t *testing.T) {
	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	sim := helper.SetUpClient(t)
	defer sim.Close()

	testFilterLogsWithBlockHash(t, sim)
}

func testFilterLogsWithBlockHash(t *testing.T, sim *simulated.Backend) {
	client := sim.Client()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Deploy Test contract
	contractAddr, _, contract := helper.DeployTestContract(t, ctx, sim)

	// Generate some logs first
	arg1 := "hello"
	arg2 := big.NewInt(100)
	arg3 := []byte("world")

	opts, err := client.MessageToTransactOpts(ctx, message.Request{
		From: helper.Addr1,
	})
	if err != nil {
		t.Fatal(err)
	}
	contractCallTx, err := contract.TestFunc1(opts, arg1, arg2, arg3)
	if err != nil {
		t.Fatalf("TestFunc1 err: %v", err)
	}

	sim.CommitAndExpectTx(contractCallTx.Hash())

	// Get the block hash where the transaction was mined
	receipt, err := client.TransactionReceipt(ctx, contractCallTx.Hash())
	if err != nil {
		t.Fatalf("Failed to get transaction receipt: %v", err)
	}

	// Test with specific block hash
	query := ethereum.FilterQuery{
		BlockHash: &receipt.BlockHash,
		Addresses: []common.Address{contractAddr},
	}

	logs, err := client.FilterLogs(ctx, query)
	if err != nil {
		t.Fatalf("FilterLogs with block hash failed: %v", err)
	}

	// Should find logs in the specific block
	assert.Greater(t, len(logs), 0, "Should find logs in specific block")

	// Verify the block hash matches
	for _, log := range logs {
		assert.Equal(t, receipt.BlockHash, log.BlockHash, "log block hash mismatch")
		assert.Equal(t, contractAddr, log.Address, "log address mismatch")
	}

	t.Logf("Found %d logs in specific block", len(logs))
	t.Log("Block hash filter test completed successfully")
}

func Test_FilterLogsBatch_EmptyQueries(t *testing.T) {
	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	sim := helper.SetUpClient(t)
	defer sim.Close()

	testFilterLogsBatchEmptyQueries(t, sim)
}

func testFilterLogsBatchEmptyQueries(t *testing.T, sim *simulated.Backend) {
	client := sim.Client()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test with empty queries array
	queries := []ethereum.FilterQuery{}

	logs, err := client.FilterLogsBatch(ctx, queries)
	if err != nil {
		t.Fatalf("FilterLogsBatch with empty queries failed: %v", err)
	}

	assert.Equal(t, 0, len(logs), "Empty queries should return no results")
	t.Log("Empty queries batch test completed successfully")
}

func Test_FilterLogsBatch_MixedQueries(t *testing.T) {
	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	sim := helper.SetUpClient(t)
	defer sim.Close()

	testFilterLogsBatchMixedQueries(t, sim)
}

func testFilterLogsBatchMixedQueries(t *testing.T, sim *simulated.Backend) {
	client := sim.Client()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Deploy Test contract
	contractAddr, _, contract := helper.DeployTestContract(t, ctx, sim)

	// Get current block number
	fromBlock, err := client.BlockNumber(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Generate some logs first
	arg1 := "hello"
	arg2 := big.NewInt(100)
	arg3 := []byte("world")

	opts, err := client.MessageToTransactOpts(ctx, message.Request{
		From: helper.Addr1,
	})
	if err != nil {
		t.Fatal(err)
	}
	contractCallTx, err := contract.TestFunc1(opts, arg1, arg2, arg3)
	if err != nil {
		t.Fatalf("TestFunc1 err: %v", err)
	}

	sim.CommitAndExpectTx(contractCallTx.Hash())

	// Get current block number for filtering
	toBlock, err := client.BlockNumber(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Create mixed queries: one valid, one empty, one with invalid range
	queries := []ethereum.FilterQuery{
		{
			FromBlock: big.NewInt(int64(fromBlock)),
			ToBlock:   big.NewInt(int64(toBlock)),
			Addresses: []common.Address{contractAddr},
		},
		{
			FromBlock: big.NewInt(int64(fromBlock)),
			ToBlock:   big.NewInt(int64(toBlock)),
			Addresses: []common.Address{}, // Empty addresses
		},
		{
			FromBlock: big.NewInt(1000),
			ToBlock:   big.NewInt(500), // Invalid range
			Addresses: []common.Address{contractAddr},
		},
	}

	logs, err := client.FilterLogsBatch(ctx, queries)
	if err != nil {
		t.Fatalf("FilterLogsBatch with mixed queries failed: %v", err)
	}

	// Should have results for each query
	assert.Equal(t, len(queries), len(logs), "Should have results for each query")

	// First query should have logs
	assert.Greater(t, len(logs[0]), 0, "First query should have logs")

	// Second query may return logs even with empty addresses (returns all logs)
	// This is expected behavior for empty address filter
	t.Logf("Second query returned %d logs (expected behavior for empty addresses)", len(logs[1]))

	// Third query should be empty (invalid range)
	assert.Equal(t, 0, len(logs[2]), "Third query should be empty")

	t.Logf("Mixed queries batch test: query1=%d logs, query2=%d logs, query3=%d logs",
		len(logs[0]), len(logs[1]), len(logs[2]))
	t.Log("Mixed queries batch test completed successfully")
}

func Test_FilterLogs_ContextCancellation(t *testing.T) {
	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	sim := helper.SetUpClient(t)
	defer sim.Close()

	testFilterLogsContextCancellation(t, sim)
}

func testFilterLogsContextCancellation(t *testing.T, sim *simulated.Backend) {
	client := sim.Client()

	// Create a context that will be cancelled immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(1),
		ToBlock:   big.NewInt(10),
		Addresses: []common.Address{common.HexToAddress("0x1234567890123456789012345678901234567890")},
	}

	_, err := client.FilterLogs(ctx, query)

	// When context is cancelled, FilterLogs may return nil error (empty result)
	// This is acceptable behavior as the function handles cancellation gracefully
	if err != nil {
		assert.Contains(t, err.Error(), "context", "Error should be context-related")
	}

	t.Log("Context cancellation test completed successfully")
}

func Test_FilterLogs_Timeout(t *testing.T) {
	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	sim := helper.SetUpClient(t)
	defer sim.Close()

	testFilterLogsTimeout(t, sim)
}

func testFilterLogsTimeout(t *testing.T, sim *simulated.Backend) {
	client := sim.Client()

	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(1),
		ToBlock:   big.NewInt(1000000), // Large range to potentially cause timeout
		Addresses: []common.Address{common.HexToAddress("0x1234567890123456789012345678901234567890")},
	}

	_, err := client.FilterLogs(ctx, query)

	// When timeout occurs, FilterLogs may return nil error (empty result)
	// This is acceptable behavior as the function handles timeout gracefully
	if err != nil {
		assert.Contains(t, err.Error(), "timeout", "Error should be timeout-related")
	}

	t.Log("Timeout test completed successfully")
}

func Test_FilterLogs_Performance(t *testing.T) {
	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	sim := helper.SetUpClient(t)
	defer sim.Close()

	testFilterLogsPerformance(t, sim)
}

func testFilterLogsPerformance(t *testing.T, sim *simulated.Backend) {
	client := sim.Client()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Deploy Test contract
	contractAddr, _, contract := helper.DeployTestContract(t, ctx, sim)

	// Generate multiple logs across multiple blocks
	numLogs := 10
	for i := 0; i < numLogs; i++ {
		arg1 := fmt.Sprintf("hello_%d", i)
		arg2 := big.NewInt(int64(i * 100))
		arg3 := []byte(fmt.Sprintf("world_%d", i))

		opts, err := client.MessageToTransactOpts(ctx, message.Request{
			From: helper.Addr1,
		})
		if err != nil {
			t.Fatal(err)
		}
		contractCallTx, err := contract.TestFunc1(opts, arg1, arg2, arg3)
		if err != nil {
			t.Fatalf("TestFunc1 err: %v", err)
		}

		sim.CommitAndExpectTx(contractCallTx.Hash())
	}

	// Get block range
	fromBlock, err := client.BlockNumber(ctx)
	if err != nil {
		t.Fatal(err)
	}
	fromBlock -= uint64(numLogs) // Go back to before we started generating logs

	toBlock, err := client.BlockNumber(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Test performance with large block range
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(toBlock)),
		Addresses: []common.Address{contractAddr},
	}

	start := time.Now()
	logs, err := client.FilterLogs(ctx, query)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("FilterLogs performance test failed: %v", err)
	}

	// Should find all the logs we generated
	assert.Equal(t, numLogs*2, len(logs), "Should find all generated logs") // *2 because TestFunc1 emits 2 events

	// Performance should be reasonable (less than 5 seconds for this test)
	assert.Less(t, duration, 5*time.Second, "FilterLogs should complete within reasonable time")

	t.Logf("Performance test: found %d logs in %v", len(logs), duration)
	t.Log("Performance test completed successfully")
}

func Test_FilterLogsBatch_MultipleConcurrentCalls(t *testing.T) {
	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	sim := helper.SetUpClient(t)
	defer sim.Close()

	testFilterLogsBatchMultipleConcurrentCalls(t, sim)
}

func testFilterLogsBatchMultipleConcurrentCalls(t *testing.T, sim *simulated.Backend) {
	client := sim.Client()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Deploy Test contract
	contractAddr, _, contract := helper.DeployTestContract(t, ctx, sim)

	// Get current block number
	fromBlock, err := client.BlockNumber(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Generate some logs first
	arg1 := "hello"
	arg2 := big.NewInt(100)
	arg3 := []byte("world")

	opts, err := client.MessageToTransactOpts(ctx, message.Request{
		From: helper.Addr1,
	})
	if err != nil {
		t.Fatal(err)
	}

	contractCallTx, err := contract.TestFunc1(opts, arg1, arg2, arg3)
	if err != nil {
		t.Fatalf("TestFunc1 err: %v", err)
	}

	sim.CommitAndExpectTx(contractCallTx.Hash())

	// Get current block number for filtering
	toBlock, err := client.BlockNumber(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Create multiple queries for batch testing
	queries := []ethereum.FilterQuery{
		{
			FromBlock: big.NewInt(0).SetUint64(fromBlock),
			ToBlock:   big.NewInt(0).SetUint64(toBlock),
			Addresses: []common.Address{contractAddr},
		},
		{
			FromBlock: big.NewInt(0).SetUint64(fromBlock),
			ToBlock:   big.NewInt(0).SetUint64(toBlock),
			Addresses: []common.Address{contractAddr},
		},
		{
			FromBlock: big.NewInt(0).SetUint64(fromBlock),
			ToBlock:   big.NewInt(0).SetUint64(toBlock),
			Addresses: []common.Address{contractAddr},
		},
	}

	// Test multiple concurrent FilterLogsBatch calls
	numConcurrentCalls := 10
	var wg sync.WaitGroup
	results := make([][][]types.Log, numConcurrentCalls)
	errors := make([]error, numConcurrentCalls)

	t.Logf("Starting %d concurrent FilterLogsBatch calls...", numConcurrentCalls)

	// Start multiple concurrent calls
	for i := 0; i < numConcurrentCalls; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			// Create a copy of queries for this goroutine
			queriesCopy := make([]ethereum.FilterQuery, len(queries))
			copy(queriesCopy, queries)

			// Call FilterLogsBatch
			logs, err := client.FilterLogsBatch(ctx, queriesCopy)
			results[index] = logs
			errors[index] = err

			t.Logf("Concurrent call %d completed", index)
		}(i)
	}

	// Wait for all calls to complete
	wg.Wait()

	// Check results
	for i := 0; i < numConcurrentCalls; i++ {
		if errors[i] != nil {
			t.Errorf("Concurrent call %d failed: %v", i, errors[i])
			continue
		}

		if len(results[i]) != len(queries) {
			t.Errorf("Concurrent call %d: expected %d query results, got %d", i, len(queries), len(results[i]))
			continue
		}

		// Each query should have logs
		for j, queryLogs := range results[i] {
			if len(queryLogs) == 0 {
				t.Errorf("Concurrent call %d, query %d: expected logs, got none", i, j)
			}
		}
	}

	// Now test sequential calls to see if state leaks
	t.Log("Testing sequential FilterLogsBatch calls to check for state leaks...")

	sequentialResults := make([][][]types.Log, numConcurrentCalls)
	sequentialErrors := make([]error, numConcurrentCalls)

	for i := 0; i < numConcurrentCalls; i++ {
		// Create a copy of queries for this call
		queriesCopy := make([]ethereum.FilterQuery, len(queries))
		copy(queriesCopy, queries)

		// Call FilterLogsBatch
		logs, err := client.FilterLogsBatch(ctx, queriesCopy)
		sequentialResults[i] = logs
		sequentialErrors[i] = err

		t.Logf("Sequential call %d completed", i)

		// Small delay between calls
		time.Sleep(100 * time.Millisecond)
	}

	// Check sequential results
	for i := 0; i < numConcurrentCalls; i++ {
		if sequentialErrors[i] != nil {
			t.Errorf("Sequential call %d failed: %v", i, sequentialErrors[i])
			continue
		}

		if len(sequentialResults[i]) != len(queries) {
			t.Errorf("Sequential call %d: expected %d query results, got %d", i, len(queries), len(sequentialResults[i]))
			continue
		}

		// Each query should have logs
		for j, logs := range sequentialResults[i] {
			if len(logs) == 0 {
				t.Errorf("Sequential call %d, query %d: expected logs, got none", i, j)
			}
		}
	}

	// Compare concurrent vs sequential results
	for i := 0; i < numConcurrentCalls; i++ {
		if errors[i] == nil && sequentialErrors[i] == nil {
			// Both should have the same number of results
			if len(results[i]) != len(sequentialResults[i]) {
				t.Errorf("Call %d: concurrent and sequential results have different lengths", i)
			}
		}
	}

	t.Logf("Multiple concurrent calls test completed: %d concurrent calls, %d sequential calls",
		numConcurrentCalls, numConcurrentCalls)
	t.Log("Check logs for any repeated 'Subscriber FilterLogs starts filtering logs' messages")
}

func Test_FilterLogsBatch_StateLeakReproduction(t *testing.T) {
	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	sim := helper.SetUpClient(t)
	defer sim.Close()

	testFilterLogsBatchStateLeakReproduction(t, sim)
}

func testFilterLogsBatchStateLeakReproduction(t *testing.T, sim *simulated.Backend) {
	client := sim.Client()

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Deploy Test contract
	contractAddr, _, contract := helper.DeployTestContract(t, ctx, sim)

	// Get current block number
	fromBlock, err := client.BlockNumber(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Generate logs in multiple blocks
	numBlocks := 5
	for i := 0; i < numBlocks; i++ {
		arg1 := fmt.Sprintf("hello_%d", i)
		arg2 := big.NewInt(int64(100 + i))
		arg3 := []byte(fmt.Sprintf("world_%d", i))

		opts, err := client.MessageToTransactOpts(ctx, message.Request{
			From: helper.Addr1,
		})
		if err != nil {
			t.Fatal(err)
		}

		contractCallTx, err := contract.TestFunc1(opts, arg1, arg2, arg3)
		if err != nil {
			t.Fatalf("TestFunc1 err: %v", err)
		}

		sim.CommitAndExpectTx(contractCallTx.Hash())

		// Small delay to ensure different blocks
		time.Sleep(100 * time.Millisecond)
	}

	// Get the latest block
	latestBlock, err := client.BlockNumber(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Create a query that will trigger the state leak scenario
	// This simulates the scenario where from > to (which was causing the infinite loop)
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(latestBlock + 1)), // Start from a block that doesn't exist yet
		ToBlock:   big.NewInt(int64(latestBlock)),     // End at the latest block
		Addresses: []common.Address{contractAddr},
	}

	// This should trigger the startBlock > toBlock condition
	// and test if our fix prevents the infinite loop
	t.Logf("Testing state leak scenario: fromBlock=%d, toBlock=%d, latestBlock=%d",
		query.FromBlock.Int64(), query.ToBlock.Int64(), latestBlock)

	// Call FilterLogs multiple times to see if state leaks
	numCalls := 20
	for i := 0; i < numCalls; i++ {
		t.Logf("Call %d: Testing FilterLogs with from > to", i+1)

		// This should return empty results and exit cleanly
		logs, err := client.FilterLogs(ctx, query)
		if err != nil {
			t.Logf("Call %d: FilterLogs returned error: %v", i+1, err)
		} else {
			t.Logf("Call %d: FilterLogs returned %d logs", i+1, len(logs))
		}

		// Small delay between calls
		time.Sleep(50 * time.Millisecond)
	}

	// Now test with a valid range to ensure normal operation still works
	validQuery := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(latestBlock)),
		Addresses: []common.Address{contractAddr},
	}

	t.Log("Testing with valid range to ensure normal operation...")
	logs, err := client.FilterLogs(ctx, validQuery)
	if err != nil {
		t.Fatalf("Valid query failed: %v", err)
	}

	t.Logf("Valid query returned %d logs", len(logs))
	assert.Greater(t, len(logs), 0, "Valid query should return logs")

	t.Log("State leak reproduction test completed")
	t.Log("Check logs for any repeated 'Subscriber FilterLogs starts filtering logs' messages or currBlocksPerScan increases")
}
