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
	rawEthclient "github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/ivanzzeth/ethclient/message"
	"github.com/ivanzzeth/ethclient/simulated"
	"github.com/ivanzzeth/ethclient/subscriber"
	"github.com/ivanzzeth/ethclient/tests/helper"
	goredislib "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func Test_SubscribeFilterLogs(t *testing.T) {
	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	sim := helper.SetUpClient(t)
	defer sim.Close()

	testSubscribeFilterLogs(t, sim, 3)
}

func test_SubscribeFilterLogsRealTime(t *testing.T) {
	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// wss://opbnb-rpc.publicnode.com
	// ws://localhost:3005/ws/8453
	// client, err := rawEthclient.Dial("wss://opbnb-rpc.publicnode.com")
	client, err := rawEthclient.Dial("ws://localhost:3005/ws/8453")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("dial successful")

	ch := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(ctx, ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress("0x34aa5631bdad51583845e5e82e2caf6ce63ba64d")},
	}, ch)
	if err != nil {
		t.Fatal(err)
	}

	defer sub.Unsubscribe()
	go func() {
		for l := range ch {
			t.Logf("===> log: %v", l)
		}
	}()

	time.Sleep(10 * time.Minute)
}

func test_SubscribeFilterLogs_UsingRedisStorage(t *testing.T) {
	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	sim := helper.SetUpClient(t)
	defer sim.Close()

	client := sim.Client()

	chainId, err := client.ChainID(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	redisClient := goredislib.NewClient(&goredislib.Options{
		Addr:     "localhost:16379",
		Password: "135683271d06e8",
	})

	pool := goredis.NewPool(redisClient)

	storage := subscriber.NewRedisStorage(chainId, pool)
	subscriber, err := subscriber.NewChainSubscriber(client.RpcClient(), storage)
	if err != nil {
		t.Fatal(err)
	}

	client.SetSubscriber(subscriber)

	testSubscribeFilterLogs(t, sim, 0)
}

func testSubscribeFilterLogs(t *testing.T, sim *simulated.Backend, confirmations uint64) {
	client := sim.Client()
	client.SetBlockConfirmationsOnSubscription(confirmations)
	if cs, ok := client.Subscriber.(*subscriber.ChainSubscriber); ok {
		defer cs.Close()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Deploy Test contract.
	_, _, contract := helper.DeployTestContract(t, ctx, sim)

	// Subscribe logs

	fromBlock, err := client.BlockNumber(ctx)
	if err != nil {
		t.Fatal(err)
	}

	logs := make(chan types.Log)
	sub, err := client.Subscriber.SubscribeFilterLogs(ctx, ethereum.FilterQuery{
		FromBlock: big.NewInt(0).SetUint64(fromBlock),
	}, logs)
	if err != nil {
		t.Fatal("Subscribe logs err: ", err)
	}
	defer sub.Unsubscribe()
	logCount := 0

	// Method args
	arg1 := "hello"
	arg2 := big.NewInt(100)
	arg3 := []byte("world")

	// First transact.
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
	t.Log("contractCallTx send sucessul", "txHash", contractCallTx.Hash().Hex(), "block", receipt.BlockNumber.Uint64())

	go func() {
		for {
			select {
			case l := <-logs:
				logCount++
				t.Log("Get log", "block", l.BlockNumber, "tx", l.TxHash.Hex(),
					"txIndex", l.TxIndex, "index", l.Index)
			case <-ctx.Done():
				t.Log("Context done.")
				return
			}
		}
	}()

	// Second transact.
	opts, err = client.MessageToTransactOpts(ctx, message.Request{
		From: helper.Addr1,
	})
	if err != nil {
		t.Fatal(err)
	}
	contractCallTx, err = contract.TestFunc1(opts, arg1, arg2, arg3)
	if err != nil {
		t.Fatalf("TestFunc1 err: %v", err)
	}

	sim.CommitAndExpectTx(contractCallTx.Hash())

	receipt, contains = client.WaitTxReceipt(contractCallTx.Hash(), 0, 5*time.Second)
	assert.Equal(t, true, contains)
	t.Log("contractCallTx send sucessul", "txHash", contractCallTx.Hash().Hex(), "block", receipt.BlockNumber.Uint64())

	for i := 0; i < int(confirmations); i++ {
		sim.Commit()
	}

	time.Sleep(5 * time.Second)

	t.Log("FilterLogs...")

	toBlock, err := client.BlockNumber(ctx)
	if err != nil {
		t.Fatal(err)
	}

	filteredLogs, err := client.FilterLogs(ctx, ethereum.FilterQuery{
		FromBlock: big.NewInt(0).SetUint64(fromBlock),
		ToBlock:   big.NewInt(0).SetUint64(toBlock),
	})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 4, len(filteredLogs), "filter logs failed")
	assert.Equal(t, 4, logCount, "subscribe logs failed")

	t.Log("Exit")
}
