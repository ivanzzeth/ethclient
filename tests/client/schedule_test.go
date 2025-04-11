package client_test

import (
	"context"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/ivanzzeth/ethclient/contracts"
	"github.com/ivanzzeth/ethclient/message"
	"github.com/ivanzzeth/ethclient/nonce"
	"github.com/ivanzzeth/ethclient/simulated"
	"github.com/ivanzzeth/ethclient/tests/helper"
	goredislib "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func Test_Schedule(t *testing.T) {
	sim := helper.SetUpClient(t)

	test_Schedule(t, sim)
}

func test_ScheduleMsg_RandomlyReverted_WithRedis(t *testing.T) {
	sim := helper.SetUpClient(t)
	client := sim.Client()
	chainId, err := client.ChainID(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// Create a pool with go-redis (or redigo) which is the pool redisync will
	// use while communicating with Redis. This can also be any pool that
	// implements the `redis.Pool` interface.
	redisClient := goredislib.NewClient(&goredislib.Options{
		Addr:     "localhost:16379",
		Password: "135683271d06e8",
	})
	pool := goredis.NewPool(redisClient)

	storage := nonce.NewRedisStorage(chainId, pool)
	nm, err := nonce.NewSimpleManager(client.Client, storage)
	if err != nil {
		t.Fatal(err)
	}

	client.SetNonceManager(nm)

	test_ScheduleMsg_RandomlyReverted(t, sim)
}

func testScheduleMsg(t *testing.T, sim *simulated.Backend, concurrent bool) {
	client := sim.Client()
	buffer := 3
	go func() {
		for i := 0; i < buffer; i++ {
			schedule := func() {
				to := common.HexToAddress("0x06514D014e997bcd4A9381bF0C4Dc21bD32718D4")
				req := &message.Request{
					From: helper.Addr,
					To:   &to,
				}

				message.AssignMessageId(req)

				t.Logf("ScheduleMsg#%v", i)
				client.ScheduleMsg(req)
				t.Log("Write MSG to channel")
			}
			if concurrent {
				go schedule()
			} else {
				schedule()
			}
		}

		time.Sleep(10 * time.Second)
		t.Log("Close client")
		client.Close()
	}()

	respCount := 0
	for resp := range client.Response() {
		tx := resp.Tx
		err := resp.Err

		if err != nil {
			t.Fatal("unexpected err", err)
		}

		if tx == nil {
			t.Fatal("tx must be not nil")
		}

		storedResp, ok := client.WaitMsgResponse(resp.Id, 1*time.Second)
		if !ok {
			t.Fatal("wait msg response failed")
		}

		assert.Equal(t, resp, *storedResp)

		msg, err := client.GetMsg(resp.Id)
		if err != nil {
			t.Fatal("get msg failed: ", err)
		}

		if msg.Status != message.MessageStatusInflight {
			t.Fatal("unexpected msg status: ", msg.Status)
		}

		sim.Commit()

		_, ok = client.WaitTxReceipt(tx.Hash(), 0, 1*time.Second)
		if !ok {
			t.Fatalf("wait tx %v receipt failed", tx.Hash().Hex())
		}

		_, ok = client.WaitMsgReceipt(resp.Id, 0, 2*time.Second)
		if !ok {
			t.Fatalf("wait msg %v receipt failed", resp.Id.Hex())
		}

		msg, err = client.GetMsg(resp.Id)
		if err != nil {
			t.Fatal("get msg failed: ", err)
		}

		// if msg.Status != message.MessageStatusOnChain
		if msg.Receipt == nil {
			t.Fatalf("get msg %v receipt failed", msg.Id())
		}
		respCount++
	}

	assert.Equal(t, buffer, respCount)
	t.Log("Exit")
}

func test_ScheduleMsg_RandomlyReverted(t *testing.T, sim *simulated.Backend) {
	client := sim.Client()
	buffer := 3

	client.SetMsgBuffer(buffer)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Deploy Test contract.
	contractAddr, _, _ := helper.DeployTestContract(t, ctx, sim)

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

			client.ScheduleMsg(msg)
			wantErrMap[msg.Id()] = number%4 == 0

			t.Logf("Write MSG to channel, block: %v, blockMod: %v, msgId: %v", number, number%4, msg.Id().Hex())
		}

		time.Sleep(10 * time.Second)
		t.Log("Close send channel")
		client.Close()
	}()

	t.Log("listening responses")

	for resp := range client.Response() {
		tx := resp.Tx
		if tx == nil {
			continue
		}

		sim.Commit()

		receipt, confirmed := client.WaitTxReceipt(tx.Hash(), 0, 4*time.Second)

		if !confirmed {
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

func test_Schedule(t *testing.T, sim *simulated.Backend) {
	client := sim.Client()
	go func() {
		client.ScheduleMsg(message.AssignMessageId(&message.Request{
			From:      helper.Addr,
			To:        &helper.Addr,
			StartTime: time.Now().Add(5 * time.Second).UnixNano(),
		}))

		client.ScheduleMsg(message.AssignMessageId(&message.Request{
			From: helper.Addr,
			To:   &helper.Addr,
			// StartTime:      time.Now().Add(5 * time.Second).UnixNano(),
			ExpirationTime: time.Now().UnixNano() - int64(5*time.Second),
		}))

		client.ScheduleMsg(message.AssignMessageId(&message.Request{
			From:           helper.Addr,
			To:             &helper.Addr,
			ExpirationTime: time.Now().Add(10 * time.Second).UnixNano(),
			Interval:       2 * time.Second,
		}))

		time.Sleep(20 * time.Second)
		client.CloseSendMsg()
	}()

	for resp := range client.Response() {
		t.Log("execution resp: ", resp)
	}
}
