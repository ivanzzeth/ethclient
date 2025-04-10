package gnosissafe

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ivanzzeth/ethclient/message"
	"github.com/ivanzzeth/ethclient/nonce"
	"github.com/ivanzzeth/ethclient/tests/helper"
	"github.com/stretchr/testify/assert"
)

func TestSafeTxDelivererByEthClient(t *testing.T) {
	sim := helper.SetUpClient(t)

	deliverer := NewSafeTxDelivererByEthClient(sim.Client(), helper.Addr1, nil)

	safeAddr, safeContract := helper.DeploySafeContract(t, sim)

	safeOwnerKeys := []*ecdsa.PrivateKey{helper.PrivateKey2, helper.PrivateKey3, helper.PrivateKey4}
	signers := make(map[common.Address]Signer)
	for _, ownerKey := range safeOwnerKeys {
		addr := crypto.PubkeyToAddress(ownerKey.PublicKey)
		signers[addr] = NewPrivateKeySigner(ownerKey)
	}

	nonceStorage := nonce.NewMemoryStorage()

	safeNonce, err := safeContract.Nonce(nil)
	if err != nil {
		t.Error(err)
	}

	req1 := &message.Request{From: helper.Addr1, To: &safeAddr, Value: big.NewInt(0).Mul(big.NewInt(1000000000000000000), big.NewInt(10))}
	req1 = message.AssignMessageId(req1)
	err = deliverer.Deliver(req1, safeNonce.Uint64())
	if err != nil {
		t.Error(err)
	}
	time.Sleep(10 * time.Second)

	getReq1, err := sim.Client().GetMsg(req1.Id())
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, req1.Id(), getReq1.Id())

	sim.CommitAndExpectTx(getReq1.Resp.Tx.Hash())

	_, contains := sim.Client().WaitTxReceipt(getReq1.Resp.Tx.Hash(), 0, 5*time.Second)
	if !contains {
		t.Fatal("tx failed")
	}

	safeContractV1_3, err := NewSafeContractVersion1_3_0(safeAddr, sim.Client())
	if err != nil {
		t.Error(err)
	}

	builder, err := NewSafeTxBuilderByContract(safeContractV1_3, signers, nonceStorage)
	if err != nil {
		t.Error(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(3)

	for i := range 3 {
		ff := func(index int) {

			param := &SafeTxParamV1_3{
				To:             common.HexToAddress("0xa0Ee7A142d267C1f36714E4a8F75612F20a79720"),
				Value:          big.NewInt(100000000000000),
				Calldata:       []byte{},
				Operation:      0,
				SafeTxGas:      big.NewInt(32000),
				BaseGas:        big.NewInt(50000),
				GasPrice:       big.NewInt(2000),
				GasToken:       common.HexToAddress("0x00"),
				RefundReceiver: common.HexToAddress("0x00"),
			}

			callData, _, nonce, err := builder.Build(param)
			if err != nil {
				t.Error(err)
			}
			req2 := &message.Request{From: helper.Addr1, To: &safeAddr, Data: callData, Value: big.NewInt(0),
				Gas:      500000,
				GasPrice: big.NewInt(3000)}
			req2.SetId(*message.GenerateMessageIdByAddressAndNonce(safeAddr, int64(nonce)))
			log.Debug("safe nonce from builder", "nonce", nonce, "MSGID", req2.Id())
			err = deliverer.Deliver(req2, nonce)
			if err != nil {
				t.Error(err)
			}
			time.Sleep(5 * time.Second)

			getReq2, err := sim.Client().GetMsg(req2.Id())
			if err != nil {
				t.Error(err)
			}
			log.Debug("getReq2", "ID", getReq2.Id(), "afterMsg", getReq2.Req.AfterMsg)
			assert.Equal(t, req2.Id(), getReq2.Id())

		flag:
			resp, contains := sim.Client().WaitMsgResponse(getReq2.Id(), 10*time.Second)
			if !contains {
				log.Crit(fmt.Sprintf("wait resp failed %d", index))
			}
			if !(resp != nil && resp.Tx != nil) {
				log.Debug("tx resp", "ID", getReq2.Id(), "resp", resp)
				safeNonceLast, err := safeContractV1_3.GetNonce()
				if err != nil {
					t.Error(err)
				}
				log.Debug("safe nonce last in goto", "safe nonce last", safeNonceLast)
				time.Sleep(1 * time.Second)
				goto flag
			}
			safeNonceLast, err := safeContractV1_3.GetNonce()
			if err != nil {
				t.Error(err)
			}
			log.Debug("safe nonce last", "safe nonce last", safeNonceLast, "index", index)
			wg.Done()
			log.Debug("done after", "index", index)

			// _, isPending, err := sim.Client().TransactionByHash(context.Background(), resp.Tx.Hash())
			// if err != nil {
			// 	log.Crit("get tx by hash failed", index)
			// }
			// log.Info("isPending", "isPending", isPending)
			// if isPending {
			// 	time.Sleep(time.Duration(int64(rand.Int31n(5)) * int64(time.Second)))
			// 	_, isPending, err := sim.Client().TransactionByHash(context.Background(), resp.Tx.Hash())
			// 	if err != nil {
			// 		sim.CommitAndExpectTx(resp.Tx.Hash())
			// 		//log.Crit("get tx by hash failed 2", "index", index)
			// 	}
			// 	if isPending {

			// 	}
			// }

			//sim.Commit()
			// recepit, contains := sim.Client().WaitMsgReceipt(getReq2.Id(), 0, 5*time.Second)
			// if !contains {
			// 	//log.Crit("tx2 failed", index)
			// 	goto flag
			// }
			// log.Debug("recepit for req", "req ID", getReq2.Req.Id(), "recepit", recepit)

		}
		ff(i)
	}

	go func() {
		for i := 0; i < 100; i++ {
			block, err := sim.Client().BlockByNumber(context.Background(), big.NewInt(int64(i)))
			if err != nil {
				//log.Error(err.Error())
				i = i - 1
				time.Sleep(time.Duration(i * int(time.Second)))

				continue
			}
			for _, tx := range block.Transactions() {
				log.Info("tx in block", "block number", block.Number(), "TX hash", tx.Hash().Hex())
			}
		}

	}()

	//go func() {
	for res := range sim.Client().Response() {
		sim.Commit()
		log.Debug("resp after send", "resp", res)
	}
	//}()
	log.Debug("wait after")
	wg.Wait()
	//time.Sleep(10 * time.Second)

	req3 := &message.Request{From: helper.Addr1, To: &safeAddr}
	req3 = message.AssignMessageId(req3)
	err = deliverer.Deliver(req3, safeNonce.Uint64())

	assert.Equal(t, "safeNonce is invalid", err.Error())

	req4 := &message.Request{From: helper.Addr2, To: &safeAddr}
	req4 = message.AssignMessageId(req4)
	err = deliverer.Deliver(req4, safeNonce.Uint64()+100)
	assert.Equal(t, "from address do not match", err.Error())
}
