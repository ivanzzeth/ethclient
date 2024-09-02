package ethclient

import (
	"context"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/ivanzz/ethclient/contracts"
	"github.com/ivanzz/ethclient/message"
	"github.com/ivanzz/ethclient/nonce"
	goredislib "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

var (
	privateKey, _ = crypto.HexToECDSA("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	addr          = crypto.PubkeyToAddress(privateKey.PublicKey)
)

func deployTestContract(t *testing.T, ctx context.Context, client *Client) (common.Address, *types.Transaction, *contracts.Contracts, error) {
	auth, err := client.MessageToTransactOpts(ctx, message.Request{From: addr})
	if err != nil {
		t.Fatal(err)
	}

	return contracts.DeployContracts(auth, client)
}

func newTestClient(t *testing.T) *Client {
	tmpDataDir := t.TempDir()
	t.Log("testAddr:", addr)
	backend, err := NewTestEthBackend(privateKey, types.GenesisAlloc{
		addr: types.Account{
			Balance: new(big.Int).Mul(big.NewInt(1000), big.NewInt(0).Exp(big.NewInt(10), big.NewInt(18), nil)),
		},
	}, tmpDataDir)
	if err != nil {
		t.Fatal("newTestClient err:", err)
	}
	// defer backend.Close()

	rpcClient := backend.Attach()
	if rpcClient == nil {
		panic("newTestClient attach failed")
	}

	client, err := NewClient(rpcClient)
	if err != nil {
		t.Fatal(err)
	}

	return client
}

func setUpClient(t *testing.T) *Client {
	handler := log.NewTerminalHandler(os.Stdout, true)
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	client, err := Dial("http://localhost:8545")
	if err != nil {
		t.Fatal(err)
	}

	err = client.RegisterPrivateKey(context.Background(), privateKey)
	if err != nil {
		t.Fatal(err)
	}

	client.SetABI(contracts.GetTestContractABI())

	return client
}

func TestBatchSendMsg(t *testing.T) {
	client := setUpClient(t)
	defer client.Close()

	testBatchSendMsg(t, client)
}

func Test_BatchSendMsg_RandomlyReverted(t *testing.T) {
	client := setUpClient(t)
	defer client.Close()

	test_BatchSendMsg_RandomlyReverted(t, client)
}

func Test_BatchSendMsg_RandomlyReverted_WithRedis(t *testing.T) {
	client := setUpClient(t)
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

	test_BatchSendMsg_RandomlyReverted(t, client)
}

func TestCallContract(t *testing.T) {
	client := setUpClient(t)
	defer client.Close()

	testCallContract(t, client)
}

func TestContractRevert(t *testing.T) {
	client := setUpClient(t)
	defer client.Close()

	testContractRevert(t, client)
}

func Test_CallContract_Concurrent(t *testing.T) {
	client := setUpClient(t)
	defer client.Close()

	test_CallContract_Concurrent(t, client)
}

func Test_DecodeJsonRpcError(t *testing.T) {
	client := setUpClient(t)
	defer client.Close()

	client.SetABI(contracts.GetTestContractABI())

	ctx := context.Background()
	// Deploy Test contract.
	contractAddr, txOfContractCreation, _, err := deployTestContract(t, ctx, client)
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

func testBatchSendMsg(t *testing.T, client *Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	buffer := 1000
	mesgsChan := make(chan message.Request, buffer)
	msgRespChan := client.BatchSendMsg(ctx, mesgsChan)
	go func() {
		for i := 0; i < 2*buffer; i++ {
			to := common.HexToAddress("0x06514D014e997bcd4A9381bF0C4Dc21bD32718D4")
			mesgsChan <- message.Request{
				From: addr,
				To:   &to,
			}
			t.Log("Write MSG to channel")
		}

		t.Log("Close send channel")
		close(mesgsChan)
	}()

	for resp := range msgRespChan {
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

func test_BatchSendMsg_RandomlyReverted(t *testing.T, client *Client) {
	buffer := 1000

	client.SetMsgBuffer(buffer)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Deploy Test contract.
	contractAddr, txOfContractCreation, _, err := deployTestContract(t, ctx, client)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("TestContract creation transaction", "txHex", txOfContractCreation.Hash().Hex(), "contract", contractAddr.Hex())

	_, contains := client.WaitTxReceipt(txOfContractCreation.Hash(), 2, 5*time.Second)
	assert.Equal(t, true, contains)

	wantErrMap := make(map[common.Hash]bool, 0)

	mesgsChan := make(chan message.Request, buffer)
	msgRespChan := client.BatchSendMsg(ctx, mesgsChan)
	go func() {
		contractAbi := contracts.GetTestContractABI()

		for i := 0; i < 2*buffer; i++ {
			number, _ := client.BlockNumber(context.Background())
			data, err := client.NewMethodData(contractAbi, "testRandomlyReverted")
			assert.Equal(t, nil, err)

			to := contractAddr
			msg := message.AssignMessageId(
				&message.Request{
					From: addr,
					To:   &to,
					Data: data,
					Gas:  1000000,
				},
			)
			mesgsChan <- *msg
			wantErrMap[msg.Id()] = number%4 == 0

			t.Logf("Write MSG to channel, block: %v, blockMod: %v, msgId: %v", number, number%4, msg.Id().Hex())
		}

		t.Log("Close send channel")
		close(mesgsChan)
	}()

	for resp := range msgRespChan {
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

func testCallContract(t *testing.T, client *Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Deploy Test contract.
	contractAddr, txOfContractCreation, contract, err := deployTestContract(t, ctx, client)
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
		From: crypto.PubkeyToAddress(privateKey.PublicKey),
		To:   &contractAddr,
		Data: data,
	}, nil)
	if err != nil {
		t.Fatalf("CallMsg err: %v", err)
	}

	contractCallTx, err := client.SendMsg(ctx, message.Request{
		From: addr,
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

func testContractRevert(t *testing.T, client *Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Deploy Test contract.
	contractAddr, txOfContractCreation, _, err := deployTestContract(t, ctx, client)
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
		From:     addr,
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
		From: addr,
		To:   &contractAddr,
		Data: data,
	})
	t.Log("Send Message without specific gas and gasPrice, err: ", err)
	// Send Message without specific gas and gasPrice, err:  NewTransaction err: execution reverted: test reverted
	assert.NotEqual(t, nil, err, "expect revert transaction")

	// Call failed, because evm execution faield.
	returnData, err := client.CallMsg(ctx, message.Request{
		From: addr,
		To:   &contractAddr,
		Data: data,
	}, nil)
	t.Log("Call Message err: ", err)
	assert.Equal(t, 0, len(returnData))
	assert.NotEqual(t, nil, err, "expect revert transaction")
}

func test_CallContract_Concurrent(t *testing.T, client *Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Deploy Test contract.
	contractAddr, txOfContractCreation, contract, err := deployTestContract(t, ctx, client)
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

			auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainId)
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
