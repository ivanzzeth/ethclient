package gnosissafe

import (
	"crypto/ecdsa"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ivanzzeth/ethclient/message"
	"github.com/ivanzzeth/ethclient/nonce"
	"github.com/ivanzzeth/ethclient/tests/helper"
	"github.com/stretchr/testify/assert"
)

func TestSafeTxDelivererByEthClient(t *testing.T) {
	sim := helper.SetUpClient(t)

	deliverer := NewSafeTxDelivererByEthClient(sim.Client(), helper.Addr1, nil)

	safeAddr, safeContract := helper.DeploySafeContract(t, sim)

	safeOwnerKey1, _ := crypto.HexToECDSA("59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d")
	safeOwnerKey2, _ := crypto.HexToECDSA("5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a")
	safeOwnerKey3, _ := crypto.HexToECDSA("7c852118294e51e653712a81e05800f419141751be58f605c371e15141b007a6")

	safeOwnerKeys := []*ecdsa.PrivateKey{safeOwnerKey1, safeOwnerKey2, safeOwnerKey3}
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

	req1 := &message.Request{From: helper.Addr1, To: &safeAddr, Value: big.NewInt(1000000000000000000)}
	req1 = message.AssignMessageId(req1)
	err = deliverer.Deliverer(req1, safeNonce.Uint64())
	if err != nil {
		t.Error(err)
	}
	time.Sleep(5 * time.Second)

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

	param := &SafeTxParamV1_3{
		To:             helper.Addr4,
		Value:          big.NewInt(100000000000000000),
		Operation:      0,
		SafeTxGas:      big.NewInt(300000),
		BaseGas:        big.NewInt(300000),
		GasPrice:       big.NewInt(10000),
		GasToken:       helper.AddrZero,
		RefundReceiver: helper.AddrZero,
	}

	callData, _, nonce, err := builder.Build(param)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, safeNonce.Uint64(), nonce)

	req2 := &message.Request{From: helper.Addr1, To: &safeAddr, Data: callData}
	req2 = message.AssignMessageId(req2)
	err = deliverer.Deliverer(req2, nonce)
	if err != nil {
		t.Error(err)
	}
	time.Sleep(5 * time.Second)

	getReq2, err := sim.Client().GetMsg(req2.Id())
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, req2.Id(), getReq2.Id())

	sim.CommitAndExpectTx(getReq2.Resp.Tx.Hash())

	_, contains = sim.Client().WaitTxReceipt(getReq2.Resp.Tx.Hash(), 0, 5*time.Second)
	if !contains {
		t.Fatal("tx2 failed")
	}

	req3 := &message.Request{From: helper.Addr1, To: &safeAddr}
	req3 = message.AssignMessageId(req3)
	err = deliverer.Deliverer(req3, nonce)

	assert.Equal(t, "safeNonce is invalid", err.Error())

	req4 := &message.Request{From: helper.Addr2, To: &safeAddr}
	req4 = message.AssignMessageId(req4)
	err = deliverer.Deliverer(req4, nonce)
	assert.Equal(t, "from address do not match", err.Error())
}
