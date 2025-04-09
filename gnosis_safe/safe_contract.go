package gnosissafe

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	safel2contract "github.com/ivanzzeth/ethclient/gnosis_safe/gnosissafel2contract/v1.3"
)

var _ SafeContract = &SafeContractVersion1_3_0{}
var _ SafelContractCaller = &SafeContractVersion1_3_0{}

var ErrSafeParamVersionNotMatch = errors.New("safe param version do not match")

type SafeContract interface {
	SafelContractCaller
	EncodeExecTransactionData(signatures []byte, txParams SafeTxParam) ([]byte, error)
}

type SafelContractCaller interface {
	GetNonce() (uint64, error)
	GetThreshold() (uint64, error)
	GetAddress() common.Address
	GetOwners() ([]common.Address, error)
	GetVersion() (string, error)
	GetTransactionHash(nonce uint64, txParams SafeTxParam) ([]byte, error)
	EncodeTransactionData(nonce uint64, txParams SafeTxParam) ([]byte, error)
}

type SafeContractVersion1_3_0 struct {
	Address        common.Address
	safel2contract safel2contract.Safel2contract
}

func NewSafeContractVersion1_3_0(contractAdress common.Address, safel2contractV1_3 *safel2contract.Safel2contract) SafeContract {

	return &SafeContractVersion1_3_0{
		Address:        contractAdress,
		safel2contract: *safel2contractV1_3,
	}
}

func (contract *SafeContractVersion1_3_0) GetNonce() (uint64, error) {
	nonce, err := contract.safel2contract.Nonce(nil)
	if err != nil {
		return 0, nil
	}
	return nonce.Uint64(), nil
}

func (contract *SafeContractVersion1_3_0) GetThreshold() (uint64, error) {
	threshold, err := contract.safel2contract.GetThreshold(nil)
	if err != nil {
		return 0, nil
	}
	return threshold.Uint64(), nil
}

func (contract *SafeContractVersion1_3_0) GetAddress() common.Address {
	return contract.Address
}

func (contract *SafeContractVersion1_3_0) GetOwners() ([]common.Address, error) {
	owners, err := contract.safel2contract.GetOwners(nil)
	if err != nil {
		return nil, nil
	}
	return owners, nil
}

func (contract *SafeContractVersion1_3_0) EncodeExecTransactionData(signatures []byte, txParams SafeTxParam) ([]byte, error) {
	param, ok := txParams.(*SafeTxParamV1_3)
	if !ok {
		return nil, ErrSafeParamVersionNotMatch
	}

	safel2Abi, err := safel2contract.Safel2contractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	callData, err := safel2Abi.Pack("execTransaction", param.To, param.Value, param.Calldata, param.Operation, param.SafeTxGas, param.BaseGas, param.GasPrice, param.GasToken, param.RefundReceiver, signatures)
	if err != nil {
		return nil, err
	}

	return callData, nil
}

func (contract *SafeContractVersion1_3_0) GetTransactionHash(nonce uint64, txParams SafeTxParam) ([]byte, error) {
	param, ok := txParams.(*SafeTxParamV1_3)
	if !ok {
		return nil, ErrSafeParamVersionNotMatch
	}

	hash, err := contract.safel2contract.GetTransactionHash(nil,
		param.To, param.Value, param.Calldata, param.Operation, param.SafeTxGas, param.BaseGas, param.GasPrice, param.GasToken, param.RefundReceiver, big.NewInt(int64(nonce)))
	if err != nil {
		return nil, err
	}

	return []byte(hash[:]), nil
}

func (contract *SafeContractVersion1_3_0) EncodeTransactionData(nonce uint64, txParams SafeTxParam) ([]byte, error) {
	param, ok := txParams.(*SafeTxParamV1_3)
	if !ok {
		return nil, ErrSafeParamVersionNotMatch
	}

	contract.safel2contract.EncodeTransactionData(nil,
		param.To, param.Value, param.Calldata, param.Operation, param.SafeTxGas, param.BaseGas, param.GasPrice, param.GasToken, param.RefundReceiver, big.NewInt(int64(nonce)))

	return nil, nil
}

func (contract *SafeContractVersion1_3_0) GetVersion() (string, error) {
	return contract.safel2contract.VERSION(nil)
}

func NewSafelContractCallerByAddress(address common.Address, backend bind.ContractBackend) (SafelContractCaller, error) {
	return NewDefaultSafelContractCallerByAddress(address, backend)
}

func NewDefaultSafelContractCallerByAddress(address common.Address, backend bind.ContractBackend) (SafelContractCaller, error) {
	safel2contract, err := safel2contract.NewSafel2contract(address, backend)
	if err != nil {
		return nil, err
	}
	return NewSafeContractVersion1_3_0(address, safel2contract), nil
}
