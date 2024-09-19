package subscriber_test

import (
	"context"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/ivanzzeth/ethclient"
	"github.com/ivanzzeth/ethclient/message"
	"github.com/ivanzzeth/ethclient/subscriber"
	"github.com/ivanzzeth/ethclient/tests/helper"
	goredislib "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func Test_Subscriber(t *testing.T) {
	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	client := helper.SetUpClient(t)
	defer client.Close()

	testSubscriber(t, client, 3)
}

func Test_Subscriber_UsingRedisStorage(t *testing.T) {
	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	client := helper.SetUpClient(t)
	defer client.Close()

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
	subscriber, err := subscriber.NewChainSubscriber(client.Client, storage)
	if err != nil {
		t.Fatal(err)
	}

	client.SetSubscriber(subscriber)

	testSubscriber(t, client, 0)
}

func testSubscriber(t *testing.T, client *ethclient.Client, confirmations uint64) {
	client.SetBlockConfirmationsOnSubscription(confirmations)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Deploy Test contract.
	contractAddr, txOfContractCreation, contract, err := helper.DeployTestContract(t, ctx, client)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("TestContract creation transaction", "txHex", txOfContractCreation.Hash().Hex(), "contract", contractAddr.Hex())

	_, contains := client.WaitTxReceipt(txOfContractCreation.Hash(), 2, 5*time.Second)
	assert.Equal(t, true, contains)

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
		From: helper.Addr,
	})
	contractCallTx, err := contract.TestFunc1(opts, arg1, arg2, arg3)
	if err != nil {
		t.Fatalf("TestFunc1 err: %v", err)
	}

	receipt, contains := client.WaitTxReceipt(contractCallTx.Hash(), 2, 5*time.Second)
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
		From: helper.Addr,
	})
	contractCallTx, err = contract.TestFunc1(opts, arg1, arg2, arg3)
	if err != nil {
		t.Fatalf("TestFunc1 err: %v", err)
	}

	t.Log("contractCallTx send sucessul", "txHash", contractCallTx.Hash().Hex())

	receipt, contains = client.WaitTxReceipt(contractCallTx.Hash(), 2, 5*time.Second)
	t.Log("contractCallTx send sucessul", "txHash", contractCallTx.Hash().Hex(), "block", receipt.BlockNumber.Uint64())

	toBlock, err := client.BlockNumber(ctx)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("FilterLogs...")
	filteredLogs, err := client.FilterLogs(ctx, ethereum.FilterQuery{
		FromBlock: big.NewInt(0).SetUint64(fromBlock),
		ToBlock:   big.NewInt(0).SetUint64(toBlock),
	})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 4, len(filteredLogs))

	time.Sleep(5 * time.Second)
	assert.Equal(t, true, contains)
	assert.Equal(t, 4, logCount)

	t.Log("Exit")
}
