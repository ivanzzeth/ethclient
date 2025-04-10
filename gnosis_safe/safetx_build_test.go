package gnosissafe

import (
	"crypto/ecdsa"
	"sort"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ivanzzeth/ethclient/nonce"
	"github.com/stretchr/testify/assert"
)

func TestSafeTxBuilderByContract(t *testing.T) {

	safeOwnerKey1, _ := crypto.HexToECDSA("59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d")
	safeOwnerKey2, _ := crypto.HexToECDSA("5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a")
	safeOwnerKey3, _ := crypto.HexToECDSA("7c852118294e51e653712a81e05800f419141751be58f605c371e15141b007a6")

	safeOwnerKeys := []*ecdsa.PrivateKey{safeOwnerKey1, safeOwnerKey2, safeOwnerKey3}
	sortAddresses := make([]common.Address, 0, 3)
	signers := make(map[common.Address]Signer)
	for _, ownerKey := range safeOwnerKeys {
		addr := crypto.PubkeyToAddress(ownerKey.PublicKey)
		signers[addr] = NewPrivateKeySigner(ownerKey)
		sortAddresses = append(sortAddresses, addr)
	}

	sort.Slice(sortAddresses, func(i, j int) bool {
		return sortAddresses[i].Cmp(sortAddresses[j]) < 0
	})

	wantContractAddress := common.HexToAddress("0x98765abcde98765abcde98765abcde98765abcde")

	nonceStorage := nonce.NewMemoryStorage()

	fakeSafeContract := NewFakeSafeContract(wantContractAddress)

	threshold := uint64(3)
	fakeSafeContract.NextReturn <- threshold
	nonce := uint64(0)
	fakeSafeContract.NextReturn <- nonce

	builder, err := NewSafeTxBuilderByContract(fakeSafeContract, signers, nonceStorage)
	if err != nil {
		t.Error(err)
	}

	getContractAddress, err := builder.GetContractAddress()
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, wantContractAddress, getContractAddress)

	wantCallData := []byte{1, 2, 3, 4, 5, 6, 7, 8,
		1, 2, 3, 4, 5, 6, 7, 8,
		1, 2, 3, 4, 5, 6, 7, 8,
		1, 2, 3, 4, 5, 6, 7, 8}

	fakeSafeContract.NextReturn <- wantCallData
	fakeSafeContract.NextReturn <- wantCallData

	param := SafeTxParamV1_3{}
	getCallData, getSignatures, getNonce, err := builder.Build(&param)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, wantCallData, getCallData)

	wantNonce, err := nonceStorage.GetNonce(fakeSafeContract.GetAddress())
	if err != nil {
		t.Error(err)
	}
	wantNonce -= 1
	assert.Equal(t, wantNonce, getNonce)

	wantSignatures := make([]byte, 0)
	for _, addr := range sortAddresses {

		signer, ok := signers[addr]
		if !ok {
			t.Error("unknown signer address")
		}

		signerFn := signer.GetSignerFn()
		signature, err := signerFn(wantCallData, addr)
		if err != nil {
			t.Error(err)
		}
		wantSignatures = append(wantSignatures, signature...)
	}

	assert.Equal(t, wantSignatures, getSignatures)
}

type FakeSafeContract struct {
	Addr       common.Address
	NextReturn chan any
}

func NewFakeSafeContract(addr common.Address) *FakeSafeContract {
	return &FakeSafeContract{
		Addr:       addr,
		NextReturn: make(chan any, 100),
	}
}

func (contract *FakeSafeContract) GetNonce() (uint64, error) {
	nonce := <-contract.NextReturn

	return nonce.(uint64), nil
}

func (contract *FakeSafeContract) GetThreshold() (uint64, error) {
	threshold := <-contract.NextReturn

	return threshold.(uint64), nil
}

func (contract *FakeSafeContract) GetAddress() common.Address {
	return contract.Addr
}

func (contract *FakeSafeContract) GetOwners() ([]common.Address, error) {
	owners := <-contract.NextReturn

	return owners.([]common.Address), nil
}

func (contract *FakeSafeContract) EncodeExecTransactionData(signatures []byte, txParams SafeTxParam) ([]byte, error) {
	callData := <-contract.NextReturn

	return callData.([]byte), nil
}

func (contract *FakeSafeContract) GetTransactionHash(nonce uint64, txParams SafeTxParam) ([]byte, error) {
	hash := <-contract.NextReturn

	return hash.([]byte), nil
}

func (contract *FakeSafeContract) EncodeTransactionData(nonce uint64, txParams SafeTxParam) ([]byte, error) {
	data := <-contract.NextReturn

	return data.([]byte), nil
}

func (contract *FakeSafeContract) GetVersion() (string, error) {
	return "1.3.0", nil
}
