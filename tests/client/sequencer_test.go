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
	"github.com/ivanzzeth/ethclient/contracts"
	"github.com/ivanzzeth/ethclient/message"
	"github.com/ivanzzeth/ethclient/simulated"
	"github.com/ivanzzeth/ethclient/tests/helper"
	"github.com/stretchr/testify/assert"
)

func Test_Sequencer_Concurrent(t *testing.T) {
	sim := helper.SetUpClient(t)
	defer sim.Close()

	test_Sequencer_Concurrent(t, sim)
}

func test_Sequencer_Concurrent(t *testing.T, sim *simulated.Backend) {
	client := sim.Client()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Deploy Test contract.
	contractAddr, _, contract := helper.DeployTestContract(t, ctx, sim)

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
					From:     helper.Addr1,
					To:       &contractAddr,
					Data:     data,
					AfterMsg: afterMsgId,
				}
				msg.SetIdWithNonce(int64(nonce))

				client.ScheduleMsg(msg)
			}
		}

		time.Sleep(5 * time.Second)
		// client.CloseSendMsg()
	}()

	go func() {
		for range client.Response() {
			sim.Commit()
		}
	}()

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
