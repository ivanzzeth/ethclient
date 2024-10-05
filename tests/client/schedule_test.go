package client_test

import (
	"context"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/ivanzzeth/ethclient"
	"github.com/ivanzzeth/ethclient/contracts"
	"github.com/ivanzzeth/ethclient/message"
	"github.com/ivanzzeth/ethclient/nonce"
	"github.com/ivanzzeth/ethclient/tests/helper"
	goredislib "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func Test_Schedule(t *testing.T) {
	client := helper.SetUpClient(t)

	test_Schedule(t, client)
}

func Test_ScheduleMsg_RandomlyReverted_WithRedis(t *testing.T) {
	client := helper.SetUpClient(t)

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

	test_ScheduleMsg_RandomlyReverted(t, client)
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

			t.Logf("ScheduleMsg#%v", i)
			client.ScheduleMsg(req)
			t.Log("Write MSG to channel")
		}

		time.Sleep(5 * time.Second)
		t.Log("Close client")
		client.Close()
	}()

	for resp := range client.Response() {
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
	contractAddr, _, _ := helper.DeployTestContract(t, ctx, client)

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

		t.Log("Close send channel")

		client.Close()
	}()

	for resp := range client.Response() {
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

func test_Schedule(t *testing.T, client *ethclient.Client) {
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

	// for resp := range client.Response() {
	// 	t.Log("execution resp: ", resp)
	// }
}
