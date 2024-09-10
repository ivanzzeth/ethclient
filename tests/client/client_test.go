package client_test

import (
	"context"
	"math/big"
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/ivanzz/ethclient"
	"github.com/ivanzz/ethclient/contracts"
	"github.com/ivanzz/ethclient/message"
	"github.com/ivanzz/ethclient/nonce"
	"github.com/ivanzz/ethclient/tests/helper"
	goredislib "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestBatchSendMsg(t *testing.T) {
	client := helper.SetUpClient(t)
	defer client.Close()

	testScheduleMsg(t, client)
}

func Test_BatchSendMsg_RandomlyReverted(t *testing.T) {
	client := helper.SetUpClient(t)
	defer client.Close()

	test_ScheduleMsg_RandomlyReverted(t, client)
}

func Test_ScheduleMsg_RandomlyReverted_WithRedis(t *testing.T) {
	client := helper.SetUpClient(t)
	defer client.Close()

	// Create a pool with go-redis (or redigo) which is the pool redisync will
	// use while communicating with Redis. This can also be any pool that
	// implements the `redis.Pool` interface.
	redisClient := goredislib.NewClient(&goredislib.Options{
		Addr:     "localhost:16379",
		Password: "135683271d06e8",
	})
	pool := goredis.NewPool(redisClient)

	storage := nonce.NewRedisStorage(pool)
	nm, err := nonce.NewSimpleManager(client.Client, storage)
	if err != nil {
		t.Fatal(err)
	}

	client.SetNonceManager(nm)

	test_ScheduleMsg_RandomlyReverted(t, client)
}

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

func Test_DecodeJsonRpcError(t *testing.T) {
	client := helper.SetUpClient(t)
	defer client.Close()

	client.SetABI(contracts.GetTestContractABI())

	ctx := context.Background()
	// Deploy Test contract.
	contractAddr, txOfContractCreation, _, err := helper.DeployTestContract(t, ctx, client)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("TestContract creation transaction", "txHex", txOfContractCreation.Hash().Hex(), "contract", contractAddr.Hex())

	_, contains := client.WaitTxReceipt(txOfContractCreation.Hash(), 2, 5*time.Second)
	assert.Equal(t, true, contains)

	testContract, err := contracts.NewContracts(contractAddr, client)
	if err != nil {
		t.Fatal(err)
	}

	err = testContract.TestReverted(nil, true)
	t.Log("TestReverted err: ", err)

	err = testContract.TestRevertedString(nil, true)
	t.Log("TestRevertedString err: ", err)
}

func Test_Sequencer_Concurrent(t *testing.T) {
	client := helper.SetUpClient(t)
	defer client.Close()

	test_Sequencer_Concurrent(t, client)
}

func Test_Schedule(t *testing.T) {
	client := helper.SetUpClient(t)
	defer client.Close()

	test_Schedule(t, client)
}

func testScheduleMsg(t *testing.T, client *ethclient.Client) {
	buffer := 10
	go func() {
		for i := 0; i < 2*buffer; i++ {
			to := common.HexToAddress("0x06514D014e997bcd4A9381bF0C4Dc21bD32718D4")
			req := &message.Request{
				From: helper.Addr,
				To:   &to,
			}

			message.AssignMessageId(req)

			client.ScheduleMsg(*req)
			t.Log("Write MSG to channel")
		}

		time.Sleep(5 * time.Second)
		t.Log("Close send channel")
		client.CloseSendMsg()
	}()

	for resp := range client.ScheduleMsgResponse() {
		tx := resp.Tx
		err := resp.Err
		var js []byte
		if tx != nil {
			js, _ = tx.MarshalJSON()
		}

		log.Info("Get Transaction", "tx", string(js), "err", err)
		assert.Equal(t, nil, err)
	}
	t.Log("Exit")
}

func test_ScheduleMsg_RandomlyReverted(t *testing.T, client *ethclient.Client) {
	buffer := 1000

	client.SetMsgBuffer(buffer)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Deploy Test contract.
	contractAddr, txOfContractCreation, _, err := helper.DeployTestContract(t, ctx, client)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("TestContract creation transaction", "txHex", txOfContractCreation.Hash().Hex(), "contract", contractAddr.Hex())

	_, contains := client.WaitTxReceipt(txOfContractCreation.Hash(), 2, 5*time.Second)
	assert.Equal(t, true, contains)

	wantErrMap := make(map[common.Hash]bool, 0)

	go func() {
		contractAbi := contracts.GetTestContractABI()

		for i := 0; i < 2*buffer; i++ {
			number, _ := client.BlockNumber(context.Background())
			data, err := client.NewMethodData(contractAbi, "testRandomlyReverted")
			assert.Equal(t, nil, err)

			to := contractAddr
			msg := message.AssignMessageId(
				&message.Request{
					From: helper.Addr,
					To:   &to,
					Data: data,
					Gas:  1000000,
				},
			)

			client.ScheduleMsg(*msg)
			wantErrMap[msg.Id()] = number%4 == 0

			t.Logf("Write MSG to channel, block: %v, blockMod: %v, msgId: %v", number, number%4, msg.Id().Hex())
		}

		t.Log("Close send channel")

		client.CloseSendMsg()
	}()

	for resp := range client.ScheduleMsgResponse() {
		tx := resp.Tx
		err := resp.Err

		// wantErr := false
		// if wantErr {
		// 	assert.NotNil(t, err)
		// } else {
		// 	assert.Nil(t, err)
		// }

		if tx == nil {
			continue
		}

		js, _ := tx.MarshalJSON()

		log.Info("Get Transaction", "tx", string(js), "err", err)
		receipt, confirmed := client.WaitTxReceipt(tx.Hash(), 1, 4*time.Second)

		if !assert.True(t, confirmed) {
			t.Fatal("Confirmation failed")
		}

		wantExecutionFail := receipt.BlockNumber.Int64()%4 == 0
		if wantExecutionFail {
			assert.Equal(t, types.ReceiptStatusFailed, receipt.Status,
				"id=%v block=%v blockMod=%v", resp.Id.String(), receipt.BlockNumber.Int64(), receipt.BlockNumber.Int64()%4)
		} else {
			assert.Equal(t, types.ReceiptStatusSuccessful, receipt.Status,
				"id=%v block=%v blockMod=%v", resp.Id.String(), receipt.BlockNumber.Int64(), receipt.BlockNumber.Int64()%4)
		}
	}
	t.Log("Exit")
}

func testCallContract(t *testing.T, client *ethclient.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Deploy Test contract.
	contractAddr, txOfContractCreation, contract, err := helper.DeployTestContract(t, ctx, client)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("TestContract creation transaction", "txHex", txOfContractCreation.Hash().Hex(), "contract", contractAddr.Hex())

	_, contains := client.WaitTxReceipt(txOfContractCreation.Hash(), 2, 5*time.Second)
	if !contains {
		t.Fatalf("Deploy Contract err: %v", err)
	}

	assert.Equal(t, true, contains)

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

	contractCallTx, err := client.SendMsg(ctx, message.Request{
		From: helper.Addr,
		To:   &contractAddr,
		Data: data,
	})
	if err != nil {
		t.Fatalf("Send single Message err: %v", err)
	}

	t.Log("contractCallTx send sucessul", "txHash", contractCallTx.Hash().Hex())

	_, contains = client.WaitTxReceipt(contractCallTx.Hash(), 2, 20*time.Second)
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
	contractAddr, txOfContractCreation, _, err := helper.DeployTestContract(t, ctx, client)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("TestContract creation transaction", "txHex", txOfContractCreation.Hash().Hex(), "contract", contractAddr.Hex())

	_, contains := client.WaitTxReceipt(txOfContractCreation.Hash(), 2, 5*time.Second)
	assert.Equal(t, true, contains)

	// Call contract method `testFunc1` id -> 0x88655d98
	contractAbi := contracts.GetTestContractABI()
	data, err := client.NewMethodData(contractAbi, "testReverted", true)
	assert.Equal(t, nil, err)

	// Send successful, but executation failed.
	contractCallTx, err := client.SendMsg(ctx, message.Request{
		From:     helper.Addr,
		To:       &contractAddr,
		Data:     data,
		Gas:      210000,
		GasPrice: big.NewInt(10),
	})
	if err != nil {
		t.Fatalf("Send single Message, err: %v", err)
	}

	receipt, contains := client.WaitTxReceipt(contractCallTx.Hash(), 1, 3*time.Second)
	assert.Equal(t, true, contains)
	assert.NotNil(t, receipt)
	assert.Equal(t, types.ReceiptStatusFailed, receipt.Status)

	t.Log("contractCallTx send sucessul", "txHash", contractCallTx.Hash().Hex())

	// Send failed, because estimateGas faield.
	contractCallTx, err = client.SendMsg(ctx, message.Request{
		From: helper.Addr,
		To:   &contractAddr,
		Data: data,
	})
	t.Log("Send Message without specific gas and gasPrice, err: ", err)
	// Send Message without specific gas and gasPrice, err:  NewTransaction err: execution reverted: test reverted
	assert.NotEqual(t, nil, err, "expect revert transaction")

	// Call failed, because evm execution faield.
	returnData, err := client.CallMsg(ctx, message.Request{
		From: helper.Addr,
		To:   &contractAddr,
		Data: data,
	}, nil)
	t.Log("Call Message err: ", err)
	assert.Equal(t, 0, len(returnData))
	assert.NotEqual(t, nil, err, "expect revert transaction")
}

func test_CallContract_Concurrent(t *testing.T, client *ethclient.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Deploy Test contract.
	contractAddr, txOfContractCreation, contract, err := helper.DeployTestContract(t, ctx, client)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("TestContract creation transaction", "txHex", txOfContractCreation.Hash().Hex(), "contract", contractAddr.Hex())

	_, contains := client.WaitTxReceipt(txOfContractCreation.Hash(), 2, 5*time.Second)
	if !contains {
		t.Fatalf("Deploy Contract err: %v", err)
	}

	assert.Equal(t, true, contains)

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

func test_Sequencer_Concurrent(t *testing.T, client *ethclient.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Deploy Test contract.
	contractAddr, txOfContractCreation, contract, err := helper.DeployTestContract(t, ctx, client)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("TestContract creation transaction", "txHex", txOfContractCreation.Hash().Hex(), "contract", contractAddr.Hex())

	_, contains := client.WaitTxReceipt(txOfContractCreation.Hash(), 2, 5*time.Second)
	if !contains {
		t.Fatalf("Deploy Contract err: %v", err)
	}

	assert.Equal(t, true, contains)

	// Call contract method `testFunc1` id -> 0x88655d98
	contractAbi := contracts.GetTestContractABI()

	nonces := []int{2, 3, 1, 4, 7, 5, 6, 0, 8, 9}

	// shuffle

	rand.Shuffle(len(nonces), func(i, j int) {
		nonces[i], nonces[j] = nonces[j], nonces[i]
	})

	t.Logf("shuffled nonces: %v", nonces)

	blockNumber, err := client.BlockNumber(ctx)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		for _, nonce := range nonces {
			data, err := client.NewMethodData(contractAbi, "testSequence", big.NewInt(int64(nonce)))
			if err == nil {
				var afterMsgId *common.Hash
				if nonce > 0 {
					afterMsgId = message.GenerateMessageIdByNonce(int64(nonce) - 1)
				}
				msg := &message.Request{
					From:     helper.Addr,
					To:       &contractAddr,
					Data:     data,
					AfterMsg: afterMsgId,
				}
				msg.SetIdWithNonce(int64(nonce))

				client.ScheduleMsg(*msg)
			}
		}

		time.Sleep(5 * time.Second)
		// client.CloseSendMsg()
	}()

	// for resp := range client.ScheduleMsgResponse() {
	// 	t.Logf("resp: %+v", resp)
	// }

	time.Sleep(5 * time.Second)

	itr, err := contract.FilterExecution(&bind.FilterOpts{
		Start: blockNumber,
		End:   nil,
	})
	if err != nil {
		t.Fatal(err)
	}

	nonceRes := []int{}
	for itr.Next() {
		nonce := itr.Event.Nonce.Int64()
		t.Log("nonce:", nonce)
		nonceRes = append(nonceRes, int(nonce))
	}

	assert.True(t, sort.IsSorted(sort.IntSlice(nonceRes)))
}

func test_Schedule(t *testing.T, client *ethclient.Client) {
	go func() {
		client.ScheduleMsg(*message.AssignMessageId(&message.Request{
			From:      helper.Addr,
			To:        &helper.Addr,
			StartTime: time.Now().Add(5 * time.Second).UnixNano(),
		}))

		client.ScheduleMsg(*message.AssignMessageId(&message.Request{
			From: helper.Addr,
			To:   &helper.Addr,
			// StartTime:      time.Now().Add(5 * time.Second).UnixNano(),
			ExpirationTime: time.Now().UnixNano() - int64(5*time.Second),
		}))

		client.ScheduleMsg(*message.AssignMessageId(&message.Request{
			From:           helper.Addr,
			To:             &helper.Addr,
			ExpirationTime: time.Now().Add(10 * time.Second).UnixNano(),
			Interval:       2 * time.Second,
		}))

		time.Sleep(20 * time.Second)
		client.CloseSendMsg()
	}()

	for resp := range client.ScheduleMsgResponse() {
		t.Log("execution resp: ", resp)
	}
}
