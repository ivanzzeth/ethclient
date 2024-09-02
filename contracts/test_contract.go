// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contracts

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

// ContractsMetaData contains all meta data concerning the Contracts contract.
var ContractsMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"a\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"b\",\"type\":\"uint256\"}],\"name\":\"TestRevert\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"counter\",\"type\":\"uint256\"}],\"name\":\"CounterUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"name\":\"Execution\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"arg1\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"arg2\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"arg3\",\"type\":\"bytes\"}],\"name\":\"FuncEvent1\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"counter\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"arg1\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"arg2\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"arg3\",\"type\":\"bytes\"}],\"name\":\"testFunc1\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"testRandomlyReverted\",\"outputs\":[],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"r\",\"type\":\"bool\"}],\"name\":\"testReverted\",\"outputs\":[],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"r\",\"type\":\"bool\"}],\"name\":\"testRevertedString\",\"outputs\":[],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"name\":\"testSequence\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561001057600080fd5b50610981806100206000396000f3fe608060405234801561001057600080fd5b50600436106100625760003560e01c806325e25eed146100675780634293dafd1461008357806352df67681461008d57806361bc221a146100a95780637237a3bd146100c757806388655d98146100e3575b600080fd5b610081600480360381019061007c9190610303565b6100ff565b005b61008b610149565b005b6100a760048036038101906100a29190610366565b6101a1565b005b6100b16101db565b6040516100be91906103a2565b60405180910390f35b6100e160048036038101906100dc9190610303565b6101e1565b005b6100fd60048036038101906100f891906105a4565b610225565b005b801561014657600160026040517fbff0f7d700000000000000000000000000000000000000000000000000000000815260040161013d9291906106af565b60405180910390fd5b50565b60006004436101589190610707565b141561019f57600160026040517fbff0f7d70000000000000000000000000000000000000000000000000000000081526004016101969291906106af565b60405180910390fd5b565b7f33e13ecb54c3076d8e8bb8c2881800a4d972b792045ffae98fdf46df365fed75816040516101d091906103a2565b60405180910390a150565b60005481565b8015610222576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161021990610795565b60405180910390fd5b50565b7fee7ebd5ac9177b3cfe282c440d0220335dc60bc4472338132f06af7b4b9432fc83838360405161025893929190610881565b60405180910390a1600160008082825461027291906108f5565b925050819055507f4785d80d2593e2cb7a3331d31eb5106408bdde2aab0db9e9b616b036a1b6039d6000546040516102aa91906103a2565b60405180910390a1505050565b6000604051905090565b600080fd5b600080fd5b60008115159050919050565b6102e0816102cb565b81146102eb57600080fd5b50565b6000813590506102fd816102d7565b92915050565b600060208284031215610319576103186102c1565b5b6000610327848285016102ee565b91505092915050565b6000819050919050565b61034381610330565b811461034e57600080fd5b50565b6000813590506103608161033a565b92915050565b60006020828403121561037c5761037b6102c1565b5b600061038a84828501610351565b91505092915050565b61039c81610330565b82525050565b60006020820190506103b76000830184610393565b92915050565b600080fd5b600080fd5b6000601f19601f8301169050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b610410826103c7565b810181811067ffffffffffffffff8211171561042f5761042e6103d8565b5b80604052505050565b60006104426102b7565b905061044e8282610407565b919050565b600067ffffffffffffffff82111561046e5761046d6103d8565b5b610477826103c7565b9050602081019050919050565b82818337600083830152505050565b60006104a66104a184610453565b610438565b9050828152602081018484840111156104c2576104c16103c2565b5b6104cd848285610484565b509392505050565b600082601f8301126104ea576104e96103bd565b5b81356104fa848260208601610493565b91505092915050565b600067ffffffffffffffff82111561051e5761051d6103d8565b5b610527826103c7565b9050602081019050919050565b600061054761054284610503565b610438565b905082815260208101848484011115610563576105626103c2565b5b61056e848285610484565b509392505050565b600082601f83011261058b5761058a6103bd565b5b813561059b848260208601610534565b91505092915050565b6000806000606084860312156105bd576105bc6102c1565b5b600084013567ffffffffffffffff8111156105db576105da6102c6565b5b6105e7868287016104d5565b93505060206105f886828701610351565b925050604084013567ffffffffffffffff811115610619576106186102c6565b5b61062586828701610576565b9150509250925092565b6000819050919050565b6000819050919050565b600061065e6106596106548461062f565b610639565b610330565b9050919050565b61066e81610643565b82525050565b6000819050919050565b600061069961069461068f84610674565b610639565b610330565b9050919050565b6106a98161067e565b82525050565b60006040820190506106c46000830185610665565b6106d160208301846106a0565b9392505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b600061071282610330565b915061071d83610330565b92508261072d5761072c6106d8565b5b828206905092915050565b600082825260208201905092915050565b7f72657665727420737472696e6700000000000000000000000000000000000000600082015250565b600061077f600d83610738565b915061078a82610749565b602082019050919050565b600060208201905081810360008301526107ae81610772565b9050919050565b600081519050919050565b60005b838110156107de5780820151818401526020810190506107c3565b838111156107ed576000848401525b50505050565b60006107fe826107b5565b6108088185610738565b93506108188185602086016107c0565b610821816103c7565b840191505092915050565b600081519050919050565b600082825260208201905092915050565b60006108538261082c565b61085d8185610837565b935061086d8185602086016107c0565b610876816103c7565b840191505092915050565b6000606082019050818103600083015261089b81866107f3565b90506108aa6020830185610393565b81810360408301526108bc8184610848565b9050949350505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b600061090082610330565b915061090b83610330565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff038211156109405761093f6108c6565b5b82820190509291505056fea2646970667358221220df151aada969f4e8bbe3ad1776ce7791ffe05b25115c2589c3f078430ec4534464736f6c634300080c0033",
}

// ContractsABI is the input ABI used to generate the binding from.
// Deprecated: Use ContractsMetaData.ABI instead.
var ContractsABI = ContractsMetaData.ABI

// ContractsBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ContractsMetaData.Bin instead.
var ContractsBin = ContractsMetaData.Bin

// DeployContracts deploys a new Ethereum contract, binding an instance of Contracts to it.
func DeployContracts(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Contracts, error) {
	parsed, err := ContractsMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ContractsBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Contracts{ContractsCaller: ContractsCaller{contract: contract}, ContractsTransactor: ContractsTransactor{contract: contract}, ContractsFilterer: ContractsFilterer{contract: contract}}, nil
}

// Contracts is an auto generated Go binding around an Ethereum contract.
type Contracts struct {
	ContractsCaller     // Read-only binding to the contract
	ContractsTransactor // Write-only binding to the contract
	ContractsFilterer   // Log filterer for contract events
}

// ContractsCaller is an auto generated read-only Go binding around an Ethereum contract.
type ContractsCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractsTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ContractsTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractsFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ContractsFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractsSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ContractsSession struct {
	Contract     *Contracts        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ContractsCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ContractsCallerSession struct {
	Contract *ContractsCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// ContractsTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ContractsTransactorSession struct {
	Contract     *ContractsTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// ContractsRaw is an auto generated low-level Go binding around an Ethereum contract.
type ContractsRaw struct {
	Contract *Contracts // Generic contract binding to access the raw methods on
}

// ContractsCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ContractsCallerRaw struct {
	Contract *ContractsCaller // Generic read-only contract binding to access the raw methods on
}

// ContractsTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ContractsTransactorRaw struct {
	Contract *ContractsTransactor // Generic write-only contract binding to access the raw methods on
}

// NewContracts creates a new instance of Contracts, bound to a specific deployed contract.
func NewContracts(address common.Address, backend bind.ContractBackend) (*Contracts, error) {
	contract, err := bindContracts(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Contracts{ContractsCaller: ContractsCaller{contract: contract}, ContractsTransactor: ContractsTransactor{contract: contract}, ContractsFilterer: ContractsFilterer{contract: contract}}, nil
}

// NewContractsCaller creates a new read-only instance of Contracts, bound to a specific deployed contract.
func NewContractsCaller(address common.Address, caller bind.ContractCaller) (*ContractsCaller, error) {
	contract, err := bindContracts(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ContractsCaller{contract: contract}, nil
}

// NewContractsTransactor creates a new write-only instance of Contracts, bound to a specific deployed contract.
func NewContractsTransactor(address common.Address, transactor bind.ContractTransactor) (*ContractsTransactor, error) {
	contract, err := bindContracts(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ContractsTransactor{contract: contract}, nil
}

// NewContractsFilterer creates a new log filterer instance of Contracts, bound to a specific deployed contract.
func NewContractsFilterer(address common.Address, filterer bind.ContractFilterer) (*ContractsFilterer, error) {
	contract, err := bindContracts(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ContractsFilterer{contract: contract}, nil
}

// bindContracts binds a generic wrapper to an already deployed contract.
func bindContracts(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ContractsMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Contracts *ContractsRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Contracts.Contract.ContractsCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Contracts *ContractsRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contracts.Contract.ContractsTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Contracts *ContractsRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Contracts.Contract.ContractsTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Contracts *ContractsCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Contracts.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Contracts *ContractsTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contracts.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Contracts *ContractsTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Contracts.Contract.contract.Transact(opts, method, params...)
}

// Counter is a free data retrieval call binding the contract method 0x61bc221a.
//
// Solidity: function counter() view returns(uint256)
func (_Contracts *ContractsCaller) Counter(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Contracts.contract.Call(opts, &out, "counter")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Counter is a free data retrieval call binding the contract method 0x61bc221a.
//
// Solidity: function counter() view returns(uint256)
func (_Contracts *ContractsSession) Counter() (*big.Int, error) {
	return _Contracts.Contract.Counter(&_Contracts.CallOpts)
}

// Counter is a free data retrieval call binding the contract method 0x61bc221a.
//
// Solidity: function counter() view returns(uint256)
func (_Contracts *ContractsCallerSession) Counter() (*big.Int, error) {
	return _Contracts.Contract.Counter(&_Contracts.CallOpts)
}

// TestRandomlyReverted is a free data retrieval call binding the contract method 0x4293dafd.
//
// Solidity: function testRandomlyReverted() view returns()
func (_Contracts *ContractsCaller) TestRandomlyReverted(opts *bind.CallOpts) error {
	var out []interface{}
	err := _Contracts.contract.Call(opts, &out, "testRandomlyReverted")

	if err != nil {
		return err
	}

	return err

}

// TestRandomlyReverted is a free data retrieval call binding the contract method 0x4293dafd.
//
// Solidity: function testRandomlyReverted() view returns()
func (_Contracts *ContractsSession) TestRandomlyReverted() error {
	return _Contracts.Contract.TestRandomlyReverted(&_Contracts.CallOpts)
}

// TestRandomlyReverted is a free data retrieval call binding the contract method 0x4293dafd.
//
// Solidity: function testRandomlyReverted() view returns()
func (_Contracts *ContractsCallerSession) TestRandomlyReverted() error {
	return _Contracts.Contract.TestRandomlyReverted(&_Contracts.CallOpts)
}

// TestReverted is a free data retrieval call binding the contract method 0x25e25eed.
//
// Solidity: function testReverted(bool r) pure returns()
func (_Contracts *ContractsCaller) TestReverted(opts *bind.CallOpts, r bool) error {
	var out []interface{}
	err := _Contracts.contract.Call(opts, &out, "testReverted", r)

	if err != nil {
		return err
	}

	return err

}

// TestReverted is a free data retrieval call binding the contract method 0x25e25eed.
//
// Solidity: function testReverted(bool r) pure returns()
func (_Contracts *ContractsSession) TestReverted(r bool) error {
	return _Contracts.Contract.TestReverted(&_Contracts.CallOpts, r)
}

// TestReverted is a free data retrieval call binding the contract method 0x25e25eed.
//
// Solidity: function testReverted(bool r) pure returns()
func (_Contracts *ContractsCallerSession) TestReverted(r bool) error {
	return _Contracts.Contract.TestReverted(&_Contracts.CallOpts, r)
}

// TestRevertedString is a free data retrieval call binding the contract method 0x7237a3bd.
//
// Solidity: function testRevertedString(bool r) pure returns()
func (_Contracts *ContractsCaller) TestRevertedString(opts *bind.CallOpts, r bool) error {
	var out []interface{}
	err := _Contracts.contract.Call(opts, &out, "testRevertedString", r)

	if err != nil {
		return err
	}

	return err

}

// TestRevertedString is a free data retrieval call binding the contract method 0x7237a3bd.
//
// Solidity: function testRevertedString(bool r) pure returns()
func (_Contracts *ContractsSession) TestRevertedString(r bool) error {
	return _Contracts.Contract.TestRevertedString(&_Contracts.CallOpts, r)
}

// TestRevertedString is a free data retrieval call binding the contract method 0x7237a3bd.
//
// Solidity: function testRevertedString(bool r) pure returns()
func (_Contracts *ContractsCallerSession) TestRevertedString(r bool) error {
	return _Contracts.Contract.TestRevertedString(&_Contracts.CallOpts, r)
}

// TestFunc1 is a paid mutator transaction binding the contract method 0x88655d98.
//
// Solidity: function testFunc1(string arg1, uint256 arg2, bytes arg3) returns()
func (_Contracts *ContractsTransactor) TestFunc1(opts *bind.TransactOpts, arg1 string, arg2 *big.Int, arg3 []byte) (*types.Transaction, error) {
	return _Contracts.contract.Transact(opts, "testFunc1", arg1, arg2, arg3)
}

// TestFunc1 is a paid mutator transaction binding the contract method 0x88655d98.
//
// Solidity: function testFunc1(string arg1, uint256 arg2, bytes arg3) returns()
func (_Contracts *ContractsSession) TestFunc1(arg1 string, arg2 *big.Int, arg3 []byte) (*types.Transaction, error) {
	return _Contracts.Contract.TestFunc1(&_Contracts.TransactOpts, arg1, arg2, arg3)
}

// TestFunc1 is a paid mutator transaction binding the contract method 0x88655d98.
//
// Solidity: function testFunc1(string arg1, uint256 arg2, bytes arg3) returns()
func (_Contracts *ContractsTransactorSession) TestFunc1(arg1 string, arg2 *big.Int, arg3 []byte) (*types.Transaction, error) {
	return _Contracts.Contract.TestFunc1(&_Contracts.TransactOpts, arg1, arg2, arg3)
}

// TestSequence is a paid mutator transaction binding the contract method 0x52df6768.
//
// Solidity: function testSequence(uint256 nonce) returns()
func (_Contracts *ContractsTransactor) TestSequence(opts *bind.TransactOpts, nonce *big.Int) (*types.Transaction, error) {
	return _Contracts.contract.Transact(opts, "testSequence", nonce)
}

// TestSequence is a paid mutator transaction binding the contract method 0x52df6768.
//
// Solidity: function testSequence(uint256 nonce) returns()
func (_Contracts *ContractsSession) TestSequence(nonce *big.Int) (*types.Transaction, error) {
	return _Contracts.Contract.TestSequence(&_Contracts.TransactOpts, nonce)
}

// TestSequence is a paid mutator transaction binding the contract method 0x52df6768.
//
// Solidity: function testSequence(uint256 nonce) returns()
func (_Contracts *ContractsTransactorSession) TestSequence(nonce *big.Int) (*types.Transaction, error) {
	return _Contracts.Contract.TestSequence(&_Contracts.TransactOpts, nonce)
}

// ContractsCounterUpdatedIterator is returned from FilterCounterUpdated and is used to iterate over the raw logs and unpacked data for CounterUpdated events raised by the Contracts contract.
type ContractsCounterUpdatedIterator struct {
	Event *ContractsCounterUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractsCounterUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractsCounterUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractsCounterUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractsCounterUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractsCounterUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractsCounterUpdated represents a CounterUpdated event raised by the Contracts contract.
type ContractsCounterUpdated struct {
	Counter *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterCounterUpdated is a free log retrieval operation binding the contract event 0x4785d80d2593e2cb7a3331d31eb5106408bdde2aab0db9e9b616b036a1b6039d.
//
// Solidity: event CounterUpdated(uint256 counter)
func (_Contracts *ContractsFilterer) FilterCounterUpdated(opts *bind.FilterOpts) (*ContractsCounterUpdatedIterator, error) {

	logs, sub, err := _Contracts.contract.FilterLogs(opts, "CounterUpdated")
	if err != nil {
		return nil, err
	}
	return &ContractsCounterUpdatedIterator{contract: _Contracts.contract, event: "CounterUpdated", logs: logs, sub: sub}, nil
}

// WatchCounterUpdated is a free log subscription operation binding the contract event 0x4785d80d2593e2cb7a3331d31eb5106408bdde2aab0db9e9b616b036a1b6039d.
//
// Solidity: event CounterUpdated(uint256 counter)
func (_Contracts *ContractsFilterer) WatchCounterUpdated(opts *bind.WatchOpts, sink chan<- *ContractsCounterUpdated) (event.Subscription, error) {

	logs, sub, err := _Contracts.contract.WatchLogs(opts, "CounterUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractsCounterUpdated)
				if err := _Contracts.contract.UnpackLog(event, "CounterUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseCounterUpdated is a log parse operation binding the contract event 0x4785d80d2593e2cb7a3331d31eb5106408bdde2aab0db9e9b616b036a1b6039d.
//
// Solidity: event CounterUpdated(uint256 counter)
func (_Contracts *ContractsFilterer) ParseCounterUpdated(log types.Log) (*ContractsCounterUpdated, error) {
	event := new(ContractsCounterUpdated)
	if err := _Contracts.contract.UnpackLog(event, "CounterUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractsExecutionIterator is returned from FilterExecution and is used to iterate over the raw logs and unpacked data for Execution events raised by the Contracts contract.
type ContractsExecutionIterator struct {
	Event *ContractsExecution // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractsExecutionIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractsExecution)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractsExecution)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractsExecutionIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractsExecutionIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractsExecution represents a Execution event raised by the Contracts contract.
type ContractsExecution struct {
	Nonce *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterExecution is a free log retrieval operation binding the contract event 0x33e13ecb54c3076d8e8bb8c2881800a4d972b792045ffae98fdf46df365fed75.
//
// Solidity: event Execution(uint256 nonce)
func (_Contracts *ContractsFilterer) FilterExecution(opts *bind.FilterOpts) (*ContractsExecutionIterator, error) {

	logs, sub, err := _Contracts.contract.FilterLogs(opts, "Execution")
	if err != nil {
		return nil, err
	}
	return &ContractsExecutionIterator{contract: _Contracts.contract, event: "Execution", logs: logs, sub: sub}, nil
}

// WatchExecution is a free log subscription operation binding the contract event 0x33e13ecb54c3076d8e8bb8c2881800a4d972b792045ffae98fdf46df365fed75.
//
// Solidity: event Execution(uint256 nonce)
func (_Contracts *ContractsFilterer) WatchExecution(opts *bind.WatchOpts, sink chan<- *ContractsExecution) (event.Subscription, error) {

	logs, sub, err := _Contracts.contract.WatchLogs(opts, "Execution")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractsExecution)
				if err := _Contracts.contract.UnpackLog(event, "Execution", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseExecution is a log parse operation binding the contract event 0x33e13ecb54c3076d8e8bb8c2881800a4d972b792045ffae98fdf46df365fed75.
//
// Solidity: event Execution(uint256 nonce)
func (_Contracts *ContractsFilterer) ParseExecution(log types.Log) (*ContractsExecution, error) {
	event := new(ContractsExecution)
	if err := _Contracts.contract.UnpackLog(event, "Execution", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractsFuncEvent1Iterator is returned from FilterFuncEvent1 and is used to iterate over the raw logs and unpacked data for FuncEvent1 events raised by the Contracts contract.
type ContractsFuncEvent1Iterator struct {
	Event *ContractsFuncEvent1 // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractsFuncEvent1Iterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractsFuncEvent1)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractsFuncEvent1)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractsFuncEvent1Iterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractsFuncEvent1Iterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractsFuncEvent1 represents a FuncEvent1 event raised by the Contracts contract.
type ContractsFuncEvent1 struct {
	Arg1 string
	Arg2 *big.Int
	Arg3 []byte
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterFuncEvent1 is a free log retrieval operation binding the contract event 0xee7ebd5ac9177b3cfe282c440d0220335dc60bc4472338132f06af7b4b9432fc.
//
// Solidity: event FuncEvent1(string arg1, uint256 arg2, bytes arg3)
func (_Contracts *ContractsFilterer) FilterFuncEvent1(opts *bind.FilterOpts) (*ContractsFuncEvent1Iterator, error) {

	logs, sub, err := _Contracts.contract.FilterLogs(opts, "FuncEvent1")
	if err != nil {
		return nil, err
	}
	return &ContractsFuncEvent1Iterator{contract: _Contracts.contract, event: "FuncEvent1", logs: logs, sub: sub}, nil
}

// WatchFuncEvent1 is a free log subscription operation binding the contract event 0xee7ebd5ac9177b3cfe282c440d0220335dc60bc4472338132f06af7b4b9432fc.
//
// Solidity: event FuncEvent1(string arg1, uint256 arg2, bytes arg3)
func (_Contracts *ContractsFilterer) WatchFuncEvent1(opts *bind.WatchOpts, sink chan<- *ContractsFuncEvent1) (event.Subscription, error) {

	logs, sub, err := _Contracts.contract.WatchLogs(opts, "FuncEvent1")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractsFuncEvent1)
				if err := _Contracts.contract.UnpackLog(event, "FuncEvent1", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseFuncEvent1 is a log parse operation binding the contract event 0xee7ebd5ac9177b3cfe282c440d0220335dc60bc4472338132f06af7b4b9432fc.
//
// Solidity: event FuncEvent1(string arg1, uint256 arg2, bytes arg3)
func (_Contracts *ContractsFilterer) ParseFuncEvent1(log types.Log) (*ContractsFuncEvent1, error) {
	event := new(ContractsFuncEvent1)
	if err := _Contracts.contract.UnpackLog(event, "FuncEvent1", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
