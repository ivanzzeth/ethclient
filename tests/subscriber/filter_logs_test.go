package subscriber_test

import (
	"context"
	"math/big"
	"os"
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
