// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contracts

import (
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
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// ContractsABI is the input ABI used to generate the binding from.
const ContractsABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"counter\",\"type\":\"uint256\"}],\"name\":\"CounterUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"arg1\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"arg2\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"arg3\",\"type\":\"bytes\"}],\"name\":\"FuncEvent1\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"counter\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"testControlledReverted\",\"outputs\":[],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"arg1\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"arg2\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"arg3\",\"type\":\"bytes\"}],\"name\":\"testFunc1\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"r\",\"type\":\"bool\"}],\"name\":\"testReverted\",\"outputs\":[],\"stateMutability\":\"pure\",\"type\":\"function\"}]"

// ContractsBin is the compiled bytecode used for deploying new contracts.
var ContractsBin = "0x608060405234801561001057600080fd5b506107d3806100206000396000f3fe608060405234801561001057600080fd5b506004361061004c5760003560e01c806325e25eed1461005157806325e9af121461006d57806361bc221a1461007757806388655d9814610095575b600080fd5b61006b6004803603810190610066919061022b565b6100b1565b005b6100756100f5565b005b61007f610147565b60405161008c9190610271565b60405180910390f35b6100af60048036038101906100aa919061049f565b61014d565b005b80156100f2576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016100e990610587565b60405180910390fd5b50565b600060044361010491906105d6565b1415610145576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161013c90610587565b60405180910390fd5b565b60005481565b7fee7ebd5ac9177b3cfe282c440d0220335dc60bc4472338132f06af7b4b9432fc838383604051610180939291906106d3565b60405180910390a1600160008082825461019a9190610747565b925050819055507f4785d80d2593e2cb7a3331d31eb5106408bdde2aab0db9e9b616b036a1b6039d6000546040516101d29190610271565b60405180910390a1505050565b6000604051905090565b600080fd5b600080fd5b60008115159050919050565b610208816101f3565b811461021357600080fd5b50565b600081359050610225816101ff565b92915050565b600060208284031215610241576102406101e9565b5b600061024f84828501610216565b91505092915050565b6000819050919050565b61026b81610258565b82525050565b60006020820190506102866000830184610262565b92915050565b600080fd5b600080fd5b6000601f19601f8301169050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b6102df82610296565b810181811067ffffffffffffffff821117156102fe576102fd6102a7565b5b80604052505050565b60006103116101df565b905061031d82826102d6565b919050565b600067ffffffffffffffff82111561033d5761033c6102a7565b5b61034682610296565b9050602081019050919050565b82818337600083830152505050565b600061037561037084610322565b610307565b90508281526020810184848401111561039157610390610291565b5b61039c848285610353565b509392505050565b600082601f8301126103b9576103b861028c565b5b81356103c9848260208601610362565b91505092915050565b6103db81610258565b81146103e657600080fd5b50565b6000813590506103f8816103d2565b92915050565b600067ffffffffffffffff821115610419576104186102a7565b5b61042282610296565b9050602081019050919050565b600061044261043d846103fe565b610307565b90508281526020810184848401111561045e5761045d610291565b5b610469848285610353565b509392505050565b600082601f8301126104865761048561028c565b5b813561049684826020860161042f565b91505092915050565b6000806000606084860312156104b8576104b76101e9565b5b600084013567ffffffffffffffff8111156104d6576104d56101ee565b5b6104e2868287016103a4565b93505060206104f3868287016103e9565b925050604084013567ffffffffffffffff811115610514576105136101ee565b5b61052086828701610471565b9150509250925092565b600082825260208201905092915050565b7f7465737420726576657274656400000000000000000000000000000000000000600082015250565b6000610571600d8361052a565b915061057c8261053b565b602082019050919050565b600060208201905081810360008301526105a081610564565b9050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b60006105e182610258565b91506105ec83610258565b9250826105fc576105fb6105a7565b5b828206905092915050565b600081519050919050565b60005b83811015610630578082015181840152602081019050610615565b8381111561063f576000848401525b50505050565b600061065082610607565b61065a818561052a565b935061066a818560208601610612565b61067381610296565b840191505092915050565b600081519050919050565b600082825260208201905092915050565b60006106a58261067e565b6106af8185610689565b93506106bf818560208601610612565b6106c881610296565b840191505092915050565b600060608201905081810360008301526106ed8186610645565b90506106fc6020830185610262565b818103604083015261070e818461069a565b9050949350505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b600061075282610258565b915061075d83610258565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0382111561079257610791610718565b5b82820190509291505056fea26469706673582212204f10559557cf5697d5077c65f6f3fd9f6919995de2b1b9fbd7aaa3915099edf564736f6c634300080c0033"

// DeployContracts deploys a new Ethereum contract, binding an instance of Contracts to it.
func DeployContracts(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Contracts, error) {
	parsed, err := abi.JSON(strings.NewReader(ContractsABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(ContractsBin), backend)
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
	parsed, err := abi.JSON(strings.NewReader(ContractsABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
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

// TestControlledReverted is a free data retrieval call binding the contract method 0x25e9af12.
//
// Solidity: function testControlledReverted() view returns()
func (_Contracts *ContractsCaller) TestControlledReverted(opts *bind.CallOpts) error {
	var out []interface{}
	err := _Contracts.contract.Call(opts, &out, "testControlledReverted")

	if err != nil {
		return err
	}

	return err

}

// TestControlledReverted is a free data retrieval call binding the contract method 0x25e9af12.
//
// Solidity: function testControlledReverted() view returns()
func (_Contracts *ContractsSession) TestControlledReverted() error {
	return _Contracts.Contract.TestControlledReverted(&_Contracts.CallOpts)
}

// TestControlledReverted is a free data retrieval call binding the contract method 0x25e9af12.
//
// Solidity: function testControlledReverted() view returns()
func (_Contracts *ContractsCallerSession) TestControlledReverted() error {
	return _Contracts.Contract.TestControlledReverted(&_Contracts.CallOpts)
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
