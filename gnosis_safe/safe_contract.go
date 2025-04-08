package gnosissafe

import (
	"errors"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	safel2contract "github.com/ivanzzeth/ethclient/gnosis_safe/gnosissafel2contract/v1.3"
)

var _ SafeContract = &SafeContractFor1_3_0{}
var _ SafelContractCaller = &SafeContractFor1_3_0{}

type SafeContract interface {
	GetNonce() (uint64, error)
	GetThreshold() (uint64, error)
	GetAddress() common.Address
	GetOwners() ([]common.Address, error)

	EncodeExecTransactionData(signatures []byte, txParams ...any) ([]byte, error)
	GetTransactionHash(nonce uint64, txParams ...any) ([]byte, error)
	EncodeTransactionData(nonce uint64, txParams ...any) ([]byte, error)
}

type SafelContractCaller interface {
	GetNonce() (uint64, error)
}

func NewDefaultSafelContractCallerByAddress(address common.Address, backend bind.ContractBackend) (SafelContractCaller, error) {
	safel2contract, err := safel2contract.NewSafel2contract(address, backend)
	if err != nil {
		return nil, err
	}
	return NewSafeContractFor1_3_0(address, safel2contract), nil
}

type SafeContractFor1_3_0 struct {
	Address        common.Address
	safel2contract safel2contract.Safel2contract
	CountTxParams  int // 10
}

func NewSafeContractFor1_3_0(contractAdress common.Address, safel2contractV1_3 *safel2contract.Safel2contract) SafeContract {

	return &SafeContractFor1_3_0{
		Address:        contractAdress,
		safel2contract: *safel2contractV1_3,
		CountTxParams:  10,
	}
}

func (contract *SafeContractFor1_3_0) GetNonce() (uint64, error) {
	nonce, err := contract.safel2contract.Nonce(nil)
	if err != nil {
		return 0, nil
	}
	return nonce.Uint64(), nil
}

func (contract *SafeContractFor1_3_0) GetThreshold() (uint64, error) {
	threshold, err := contract.safel2contract.GetThreshold(nil)
	if err != nil {
		return 0, nil
	}
	return threshold.Uint64(), nil
}

func (contract *SafeContractFor1_3_0) GetAddress() common.Address {
	return contract.Address
}
func (contract *SafeContractFor1_3_0) GetOwners() ([]common.Address, error) {
	owners, err := contract.safel2contract.GetOwners(nil)
	if err != nil {
		return nil, nil
	}
	return owners, nil
}

func (contract *SafeContractFor1_3_0) EncodeExecTransactionData(signatures []byte, txParams ...any) ([]byte, error) {
	//TODO
	return nil, nil
}
func (contract *SafeContractFor1_3_0) GetTransactionHash(nonce uint64, txParams ...any) ([]byte, error) {

	if contract.CountTxParams != len(txParams) {
		return nil, errors.New("txParams len do not match")
	}

	//contract.safel2contract.GetTransactionHash(nil, txParams[0].(common.Address),)
	//TODO
	return nil, nil

}
func (contract *SafeContractFor1_3_0) EncodeTransactionData(nonce uint64, txParams ...any) ([]byte, error) {
	//TODO
	return nil, nil
}
