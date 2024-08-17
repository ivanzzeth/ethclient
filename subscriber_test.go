package ethclient

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
)

func TestSubscriber(t *testing.T) {
	// log.SetDefault(log.NewLogger())
	// log.Root().SetHandler(log.StdoutHandler)
	// log.Root().SetHandler(log.LvlFilterHandler(log.LvlInfo, log.Root().GetHandler()))
	client := newTestClient(t)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Deploy Test contract.
	contractAddr, txOfContractCreation, contract, err := deployTestContract(t, ctx, client)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("TestContract creation transaction", "txHex", txOfContractCreation.Hash().Hex(), "contract", contractAddr.Hex())

	_, contains := client.WaitTxReceipt(txOfContractCreation.Hash(), 2, 5*time.Second)
	assert.Equal(t, true, contains)

	// Method args
	arg1 := "hello"
	arg2 := big.NewInt(100)
	arg3 := []byte("world")

	// First transact.
	opts, err := client.MessageToTransactOpts(ctx, Message{
		PrivateKey: privateKey,
	})
	contractCallTx, err := contract.TestFunc1(opts, arg1, arg2, arg3)
	if err != nil {
		t.Fatalf("TestFunc1 err: %v", err)
	}

	t.Log("contractCallTx send sucessul", "txHash", contractCallTx.Hash().Hex())

	_, contains = client.WaitTxReceipt(contractCallTx.Hash(), 2, 20*time.Second)
	assert.Equal(t, true, contains)

	// Subscribe logs
	logs := make(chan types.Log)
	err = client.Subscriber.SubscribeFilterlogs(ctx, ethereum.FilterQuery{}, logs)
	if err != nil {
		t.Fatal("Subscribe logs err: ", err)
	}
	logCount := 0

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
	opts, err = client.MessageToTransactOpts(ctx, Message{
		PrivateKey: privateKey,
	})
	contractCallTx, err = contract.TestFunc1(opts, arg1, arg2, arg3)
	if err != nil {
		t.Fatalf("TestFunc1 err: %v", err)
	}

	t.Log("contractCallTx send sucessul", "txHash", contractCallTx.Hash().Hex())

	_, contains = client.WaitTxReceipt(contractCallTx.Hash(), 2, 20*time.Second)

	assert.Equal(t, true, contains)
	assert.Equal(t, 4, logCount)
}
