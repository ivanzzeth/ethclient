package client_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ivanzzeth/ethclient"
	"github.com/ivanzzeth/ethclient/contracts"
	"github.com/ivanzzeth/ethclient/message"
	"github.com/ivanzzeth/ethclient/tests/helper"
	"github.com/stretchr/testify/assert"
)

func TestCallContract(t *testing.T) {
	client := helper.SetUpClient(t)
	defer client.Close()

	testCallContract(t, client)
}

func TestContractRevert(t *testing.T) {
	client := helper.SetUpClient(t)
	defer client.Close()

	testContractRevert(t, client)
}

func Test_CallContract_Concurrent(t *testing.T) {
	client := helper.SetUpClient(t)
	defer client.Close()

	test_CallContract_Concurrent(t, client)
}

func testCallContract(t *testing.T, client *ethclient.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Deploy Test contract.
	contractAddr, _, contract := helper.DeployTestContract(t, ctx, client)

	// Call contract method `testFunc1` id -> 0x88655d98
	contractAbi := contracts.GetTestContractABI()

	arg1 := "hello"
	arg2 := big.NewInt(100)
	arg3 := []byte("world")
	data, err := client.NewMethodData(contractAbi, "testFunc1", arg1, arg2, arg3)
	if err != nil {
		t.Fatal(err)
	}

	if code, err := client.RawClient().CodeAt(ctx, contractAddr, nil); err != nil || len(code) == 0 {
		t.Fatal("no code or has err: ", err)
	}

	// contract.TestFunc1(nil)
	_, err = client.CallMsg(ctx, message.Request{
		From: helper.Addr,
		To:   &contractAddr,
		Data: data,
	}, nil)
	if err != nil {
		t.Fatalf("CallMsg err: %v", err)
	}

	msg := &message.Request{
		From: helper.Addr,
		To:   &contractAddr,
		Data: data,
	}
	msg.SetRandomId()
	client.ScheduleMsg(msg)

	resp, contains := client.WaitMsgResponse(msg.Id(), 10*time.Second)

	assert.True(t, contains)
	assert.Nil(t, resp.Err)
	t.Log("contractCallTx send sucessul", "txHash", resp.Tx.Hash().Hex())

	_, contains = client.WaitMsgReceipt(msg.Id(), 2, 20*time.Second)
	assert.Equal(t, true, contains)

	counter, err := contract.Counter(nil)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, uint64(1), counter.Uint64())
}

func testContractRevert(t *testing.T, client *ethclient.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Deploy Test contract.
	contractAddr, txOfContractCreation, _ := helper.DeployTestContract(t, ctx, client)

	t.Log("TestContract creation transaction", "txHex", txOfContractCreation.Hash().Hex(), "contract", contractAddr.Hex())

	_, contains := client.WaitTxReceipt(txOfContractCreation.Hash(), 1, 6*time.Second)
	assert.Equal(t, true, contains, "contains1")

	// Call contract method `testFunc1` id -> 0x88655d98
	contractAbi := contracts.GetTestContractABI()
	data, err := client.NewMethodData(contractAbi, "testReverted", true)
	assert.Equal(t, nil, err)

	// Send successful, but executation failed.
	msg := &message.Request{
		From:     helper.Addr,
		To:       &contractAddr,
		Data:     data,
		Gas:      210000,
		GasPrice: big.NewInt(10),
	}
	msg.SetRandomId()
	client.ScheduleMsg(msg)

	receipt, contains := client.WaitMsgReceipt(msg.Id(), 1, 6*time.Second)
	assert.Equal(t, true, contains, "contains2")
	assert.NotNil(t, receipt, "receipt")
	assert.Equal(t, types.ReceiptStatusFailed, receipt.TxReceipt.Status, "receipt status")

	t.Log("contractCallTx send sucessul", "txHash", receipt.TxReceipt.TxHash.Hex())

	// Send failed, because estimateGas faield.
	msg = &message.Request{
		From: helper.Addr,
		To:   &contractAddr,
		Data: data,
	}
	msg.SetRandomId()
	client.ScheduleMsg(msg)
	t.Log("Send Message without specific gas and gasPrice")
	// Send Message without specific gas and gasPrice, err:  NewTransaction err: execution reverted: test reverted
	resp, contains := client.WaitMsgResponse(msg.Id(), 10*time.Second)
	assert.True(t, contains, "contains2")
	assert.NotNil(t, resp.Err, "expect revert transaction")

	// Call failed, because evm execution faield.
	returnData, err := client.CallMsg(ctx, message.Request{
		From: helper.Addr,
		To:   &contractAddr,
		Data: data,
	}, nil)
	t.Log("Call Message err: ", err)
	assert.Equal(t, 0, len(returnData), "returndata1")
	assert.NotNil(t, err, "expect revert transaction")
}

func test_CallContract_Concurrent(t *testing.T, client *ethclient.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Deploy Test contract.
	contractAddr, _, contract := helper.DeployTestContract(t, ctx, client)

	if code, err := client.RawClient().CodeAt(ctx, contractAddr, nil); err != nil || len(code) == 0 {
		t.Fatal("no code or has err: ", err)
	}

	// Call contract method `testFunc1` id -> 0x88655d98
	testContract, err := contracts.NewContracts(contractAddr, client)
	if err != nil {
		t.Fatal(err)
	}

	batch := 100

	go func() {
		chainId, _ := client.ChainID(context.Background())
		for i := 0; i < batch; i++ {
			arg1 := "hello"
			arg2 := big.NewInt(100)
			arg3 := []byte("world")

			auth, err := bind.NewKeyedTransactorWithChainID(helper.PrivateKey, chainId)
			if err != nil {
				t.Error(err)
				return
			}

			tx, err := testContract.TestFunc1(auth, arg1, arg2, arg3)
			if err != nil {
				t.Error(err)
				return
			}
			t.Log("contractCallTx send sucessul", "txHash", tx.Hash().Hex())
		}
	}()

	time.Sleep(5 * time.Second)

	counter, err := contract.Counter(nil)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, uint64(batch), counter.Uint64())
}
