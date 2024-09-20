package subscriber

import (
	"context"
	"math/big"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ivanzzeth/ethclient"
	"github.com/ivanzzeth/ethclient/contracts"
	"github.com/ivanzzeth/ethclient/message"
	"github.com/ivanzzeth/ethclient/subscriber"
	"github.com/ivanzzeth/ethclient/subscriber/handler"
	"github.com/ivanzzeth/ethclient/tests/helper"
	"github.com/stretchr/testify/assert"
)

func Test_QueryHandler(t *testing.T) {
	handler := log.NewTerminalHandlerWithLevel(os.Stdout, log.LevelInfo, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	client := helper.SetUpClient(t)
	defer client.Close()

	test_QueryHandler(t, client)
}

func test_QueryHandler(t *testing.T, client *ethclient.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	chainID, err := client.ChainID(ctx)
	if err != nil {
		t.Fatal(err)
	}

	storage := subscriber.NewMemoryStorage(chainID)
	handler := newTestQueryHandler(storage)

	client.SetQueryHandler(handler)

	// Deploy Test contract.
	contractAddr, txOfContractCreation, contract, err := helper.DeployTestContract(t, ctx, client)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("TestContract creation transaction", "txHex", txOfContractCreation.Hash().Hex(), "contract", contractAddr.Hex())

	_, contains := client.WaitTxReceipt(txOfContractCreation.Hash(), 2, 5*time.Second)
	assert.Equal(t, true, contains)

	startBlock, _ := client.BlockNumber(ctx)

	err = client.SubmitQuery(ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(startBlock)),
	})
	if err != nil {
		t.Fatal(err)
	}

	evmABI := contracts.GetTestContractABI()
	for _, event := range evmABI.Events {
		t.Logf("event %v: %v", event.Name, event.ID.Hex())
	}

	nonceBefore, _ := client.Client.PendingNonceAt(ctx, helper.Addr)
	callCount := 3
	for i := 0; i < callCount; i++ {
		// Method args
		arg1 := "hello"
		arg2 := big.NewInt(100)
		arg3 := []byte("world")

		// First transact.
		opts, err := client.MessageToTransactOpts(ctx, message.Request{
			From: helper.Addr,
		})
		if err != nil {
			t.Fatalf("TestFunc1 err: %v", err)
		}
		_, err = contract.TestFunc1(opts, arg1, arg2, arg3)
		if err != nil {
			t.Fatalf("TestFunc1 err: %v", err)
		}
	}

	time.Sleep(10 * time.Second)

	nonceAfter, _ := client.Client.PendingNonceAt(ctx, helper.Addr)

	t.Log("nonce comparison", nonceBefore, nonceAfter)
	assert.Equal(t, uint64(callCount), nonceAfter-nonceBefore)

	assert.Equal(t, callCount*2, int(handler.logsCounter.Load()))
}

var _ subscriber.QueryHandler = (*testQueryHandler)(nil)

type testQueryHandler struct {
	handler.SimpleQueryHandler
	logsCounter atomic.Int64
}

func newTestQueryHandler(storage subscriber.SubscriberStorage) *testQueryHandler {
	return &testQueryHandler{
		SimpleQueryHandler: *handler.NewSimpleQueryHandler(storage),
	}
}

func (h *testQueryHandler) HandleQuery(ctx context.Context, query ethereum.FilterQuery, log types.Log) error {
	err := h.SimpleQueryHandler.HandleQuery(ctx, query, log)
	if err != nil {
		return err
	}

	h.logsCounter.Add(1)

	return nil
}
