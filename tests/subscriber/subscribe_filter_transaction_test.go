package subscriber_test

import (
	"context"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ivanzzeth/ethclient/message"
	"github.com/ivanzzeth/ethclient/simulated"
	"github.com/ivanzzeth/ethclient/subscriber"
	"github.com/ivanzzeth/ethclient/tests/helper"
	"github.com/ivanzzeth/ethclient/types"
	"github.com/stretchr/testify/assert"
)

func Test_SubscribeFilterTransaction(t *testing.T) {
	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	sim := helper.SetUpClient(t)
	defer sim.Close()

	testSubscribeFilterTransaction(t, sim)
}

func testSubscribeFilterTransaction(t *testing.T, sim *simulated.Backend) {
	client := sim.Client()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Deploy Test contract.
	contractAddr, _, contract := helper.DeployTestContract(t, ctx, sim)

	t.Logf("helper.Addr: %v", helper.Addr.Hex())

	// Subscribe txs
	txCh := make(chan *etypes.Transaction)
	sub, err := client.Subscriber.SubscribeFilterFullPendingTransactions(ctx, subscriber.FilterTransaction{
		// To: []common.Address{contractAddr},
		MethodSelector: []types.MethodSelector{types.NewMethodSelector("0x88655d98")}, // TestFunc1
	}, txCh)
	if err != nil {
		t.Fatal("Subscribe txs err: ", err)
	}
	defer sub.Unsubscribe()
	txCount := 0

	// Method args
	arg1 := "hello"
	arg2 := big.NewInt(100)
	arg3 := []byte("world")

	// First transact.
	opts, err := client.MessageToTransactOpts(ctx, message.Request{
		From: helper.Addr,
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
			case tx := <-txCh:
				txCount++
				t.Log("Got tx", "tx", tx,
					"data", hexutil.Encode(tx.Data()))
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

	// Just sending regular message
	msg := &message.Request{
		From: helper.Addr,
		To:   &contractAddr,
		Gas:  3000000,
	}
	msg.SetRandomId()
	client.ScheduleMsg(msg)

	resp, ok := client.WaitMsgResponse(msg.Id(), 3*time.Second)
	if !ok {
		t.Fatal("!ok")
	}

	assert.True(t, ok)
	if resp.Err != nil {
		t.Fatal(resp.Err)
	}

	sim.CommitAndExpectTx(resp.Tx.Hash())

	_, ok = client.WaitMsgReceipt(msg.Id(), 0, 3*time.Second)
	assert.Equal(t, true, ok)

	time.Sleep(4 * time.Second)

	assert.Equal(t, 2, txCount, "subscribe txs failed")

	t.Log("Exit")
}
