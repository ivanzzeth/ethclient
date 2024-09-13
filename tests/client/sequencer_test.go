package client_test

import (
	"context"
	"math/big"
	"math/rand/v2"
	"sort"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ivanzz/ethclient"
	"github.com/ivanzz/ethclient/contracts"
	"github.com/ivanzz/ethclient/message"
	"github.com/ivanzz/ethclient/tests/helper"
	"github.com/stretchr/testify/assert"
)

func Test_Sequencer_Concurrent(t *testing.T) {
	client := helper.SetUpClient(t)
	defer client.Close()

	test_Sequencer_Concurrent(t, client)
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