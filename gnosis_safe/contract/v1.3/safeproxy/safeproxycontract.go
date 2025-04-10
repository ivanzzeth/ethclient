// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package safeproxycontract

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// SafeproxycontractMetaData contains all meta data concerning the Safeproxycontract contract.
var SafeproxycontractMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"_singleton\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"fallback\",\"stateMutability\":\"payable\"}]",
	Bin: "0x60803460c357601f61017838819003918201601f19168301916001600160401b0383118484101760c75780849260209460405283398101031260c357516001600160a01b0381169081900360c35780156073575f80546001600160a01b031916919091179055604051609c90816100dc8239f35b60405162461bcd60e51b815260206004820152602260248201527f496e76616c69642073696e676c65746f6e20616464726573732070726f766964604482015261195960f21b6064820152608490fd5b5f80fd5b634e487b7160e01b5f52604160045260245ffdfe608060405273ffffffffffffffffffffffffffffffffffffffff5f54167fa619486e000000000000000000000000000000000000000000000000000000005f3514605f575f8091368280378136915af43d5f803e15605b573d5ff35b3d5ffd5b5f5260205ff3fea2646970667358221220fdd848c09dc105936919288be03b6aeac98e422017319145b4b9e4c1a1ba66e764736f6c634300081c0033",
}

// SafeproxycontractABI is the input ABI used to generate the binding from.
// Deprecated: Use SafeproxycontractMetaData.ABI instead.
var SafeproxycontractABI = SafeproxycontractMetaData.ABI

// SafeproxycontractBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use SafeproxycontractMetaData.Bin instead.
var SafeproxycontractBin = SafeproxycontractMetaData.Bin

// DeploySafeproxycontract deploys a new Ethereum contract, binding an instance of Safeproxycontract to it.
func DeploySafeproxycontract(auth *bind.TransactOpts, backend bind.ContractBackend, _singleton common.Address) (common.Address, *types.Transaction, *Safeproxycontract, error) {
	parsed, err := SafeproxycontractMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(SafeproxycontractBin), backend, _singleton)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Safeproxycontract{SafeproxycontractCaller: SafeproxycontractCaller{contract: contract}, SafeproxycontractTransactor: SafeproxycontractTransactor{contract: contract}, SafeproxycontractFilterer: SafeproxycontractFilterer{contract: contract}}, nil
}

// Safeproxycontract is an auto generated Go binding around an Ethereum contract.
type Safeproxycontract struct {
	SafeproxycontractCaller     // Read-only binding to the contract
	SafeproxycontractTransactor // Write-only binding to the contract
	SafeproxycontractFilterer   // Log filterer for contract events
}

// SafeproxycontractCaller is an auto generated read-only Go binding around an Ethereum contract.
type SafeproxycontractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SafeproxycontractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SafeproxycontractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SafeproxycontractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SafeproxycontractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SafeproxycontractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SafeproxycontractSession struct {
	Contract     *Safeproxycontract // Generic contract binding to set the session for
	CallOpts     bind.CallOpts      // Call options to use throughout this session
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// SafeproxycontractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SafeproxycontractCallerSession struct {
	Contract *SafeproxycontractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts            // Call options to use throughout this session
}

// SafeproxycontractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SafeproxycontractTransactorSession struct {
	Contract     *SafeproxycontractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts            // Transaction auth options to use throughout this session
}

// SafeproxycontractRaw is an auto generated low-level Go binding around an Ethereum contract.
type SafeproxycontractRaw struct {
	Contract *Safeproxycontract // Generic contract binding to access the raw methods on
}

// SafeproxycontractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SafeproxycontractCallerRaw struct {
	Contract *SafeproxycontractCaller // Generic read-only contract binding to access the raw methods on
}

// SafeproxycontractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SafeproxycontractTransactorRaw struct {
	Contract *SafeproxycontractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSafeproxycontract creates a new instance of Safeproxycontract, bound to a specific deployed contract.
func NewSafeproxycontract(address common.Address, backend bind.ContractBackend) (*Safeproxycontract, error) {
	contract, err := bindSafeproxycontract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Safeproxycontract{SafeproxycontractCaller: SafeproxycontractCaller{contract: contract}, SafeproxycontractTransactor: SafeproxycontractTransactor{contract: contract}, SafeproxycontractFilterer: SafeproxycontractFilterer{contract: contract}}, nil
}

// NewSafeproxycontractCaller creates a new read-only instance of Safeproxycontract, bound to a specific deployed contract.
func NewSafeproxycontractCaller(address common.Address, caller bind.ContractCaller) (*SafeproxycontractCaller, error) {
	contract, err := bindSafeproxycontract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SafeproxycontractCaller{contract: contract}, nil
}

// NewSafeproxycontractTransactor creates a new write-only instance of Safeproxycontract, bound to a specific deployed contract.
func NewSafeproxycontractTransactor(address common.Address, transactor bind.ContractTransactor) (*SafeproxycontractTransactor, error) {
	contract, err := bindSafeproxycontract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SafeproxycontractTransactor{contract: contract}, nil
}

// NewSafeproxycontractFilterer creates a new log filterer instance of Safeproxycontract, bound to a specific deployed contract.
func NewSafeproxycontractFilterer(address common.Address, filterer bind.ContractFilterer) (*SafeproxycontractFilterer, error) {
	contract, err := bindSafeproxycontract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SafeproxycontractFilterer{contract: contract}, nil
}

// bindSafeproxycontract binds a generic wrapper to an already deployed contract.
func bindSafeproxycontract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := SafeproxycontractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Safeproxycontract *SafeproxycontractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Safeproxycontract.Contract.SafeproxycontractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Safeproxycontract *SafeproxycontractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Safeproxycontract.Contract.SafeproxycontractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Safeproxycontract *SafeproxycontractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Safeproxycontract.Contract.SafeproxycontractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Safeproxycontract *SafeproxycontractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Safeproxycontract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Safeproxycontract *SafeproxycontractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Safeproxycontract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Safeproxycontract *SafeproxycontractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Safeproxycontract.Contract.contract.Transact(opts, method, params...)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() payable returns()
func (_Safeproxycontract *SafeproxycontractTransactor) Fallback(opts *bind.TransactOpts, calldata []byte) (*types.Transaction, error) {
	return _Safeproxycontract.contract.RawTransact(opts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() payable returns()
func (_Safeproxycontract *SafeproxycontractSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _Safeproxycontract.Contract.Fallback(&_Safeproxycontract.TransactOpts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() payable returns()
func (_Safeproxycontract *SafeproxycontractTransactorSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _Safeproxycontract.Contract.Fallback(&_Safeproxycontract.TransactOpts, calldata)
}
