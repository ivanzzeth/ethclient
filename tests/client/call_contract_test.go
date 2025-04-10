package client_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ivanzzeth/ethclient/contracts"
	"github.com/ivanzzeth/ethclient/message"
	"github.com/ivanzzeth/ethclient/simulated"
	"github.com/ivanzzeth/ethclient/tests/helper"
	"github.com/stretchr/testify/assert"
)

func TestCallContract(t *testing.T) {
	sim := helper.SetUpClient(t)
	defer sim.Close()

	testCallContract(t, sim)
}

func TestContractRevert(t *testing.T) {
	sim := helper.SetUpClient(t)
	defer sim.Close()

	testContractRevert(t, sim)
}

func Test_CallContract_Concurrent(t *testing.T) {
	sim := helper.SetUpClient(t)
	defer sim.Close()

	test_CallContract_Concurrent(t, sim)
}

func testCallContract(t *testing.T, sim *simulated.Backend) {
	client := sim.Client()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Deploy Test contract.
	contractAddr, _, contract := helper.DeployTestContract(t, ctx, sim)

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
		From: helper.Addr1,
		To:   &contractAddr,
		Data: data,
	}, nil)
	if err != nil {
		t.Fatalf("CallMsg err: %v", err)
	}

	msg := &message.Request{
		From: helper.Addr1,
		To:   &contractAddr,
		Data: data,
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

	counter, err := contract.Counter(nil)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, uint64(1), counter.Uint64())
}

func testContractRevert(t *testing.T, sim *simulated.Backend) {
	client := sim.Client()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Deploy Test contract.
	contractAddr, txOfContractCreation, _ := helper.DeployTestContract(t, ctx, sim)

	t.Log("TestContract creation transaction", "txHex", txOfContractCreation.Hash().Hex(), "contract", contractAddr.Hex())

	// Call contract method `testFunc1` id -> 0x88655d98
	contractAbi := contracts.GetTestContractABI()
	data, err := client.NewMethodData(contractAbi, "testReverted", true)
	assert.Equal(t, nil, err)

	// Send successful, but executation failed.
	msg := &message.Request{
		From: helper.Addr1,
		To:   &contractAddr,
		Data: data,
		Gas:  210000,
	}
	msg.SetRandomId()
	client.ScheduleMsg(msg)

	resp, ok := client.WaitMsgResponse(msg.Id(), 3*time.Second)
	if !ok {
		t.Fatal("wait msg resp1")
	}

	if resp.Err != nil {
		t.Fatal("send msg failed", resp.Err)
	}

	time.Sleep(1 * time.Second)
	sim.CommitAndExpectTx(resp.Tx.Hash())

	receipt, contains := client.WaitMsgReceipt(msg.Id(), 0, 5*time.Second)
	if !contains {
		t.Fatal("contains1")
	}
	if receipt == nil {
		t.Fatal("receipt1")
	}
	if receipt.TxReceipt.Status != types.ReceiptStatusFailed {
		t.Fatal("receipt status1")
	}

	t.Log("contractCallTx send sucessul", "txHash", receipt.TxReceipt.TxHash.Hex())

	// Send failed, because estimateGas faield.
	msg = &message.Request{
		From: helper.Addr1,
		To:   &contractAddr,
		Data: data,
	}
	msg.SetRandomId()
	client.ScheduleMsg(msg)
	t.Log("Send Message without specific gas and gasPrice")

	time.Sleep(3 * time.Second)
	sim.Commit()

	// Send Message without specific gas and gasPrice, err:  NewTransaction err: execution reverted: test reverted
	resp, contains = client.WaitMsgResponse(msg.Id(), 5*time.Second)
	if !contains {
		t.Fatal("contains2")
	}
	if resp == nil {
		t.Fatal("resp2")
	}
	if resp.Err == nil {
		t.Fatal("expect revert transaction2")
	}

	// Call failed, because evm execution faield.
	returnData, err := client.CallMsg(ctx, message.Request{
		From: helper.Addr1,
		To:   &contractAddr,
		Data: data,
	}, nil)
	t.Log("Call Message err: ", err)
	assert.Equal(t, 0, len(returnData), "returndata3")
	assert.NotNil(t, err, "expect revert transaction3")
}

func test_CallContract_Concurrent(t *testing.T, sim *simulated.Backend) {
	client := sim.Client()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Deploy Test contract.
	contractAddr, _, contract := helper.DeployTestContract(t, ctx, sim)

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

			auth, err := bind.NewKeyedTransactorWithChainID(helper.PrivateKey1, chainId)
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
			sim.CommitAndExpectTx(tx.Hash())
		}
	}()

	time.Sleep(5 * time.Second)

	counter, err := contract.Counter(nil)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, uint64(batch), counter.Uint64())
}
