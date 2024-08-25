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
const ContractsABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"counter\",\"type\":\"uint256\"}],\"name\":\"CounterUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"arg1\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"arg2\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"arg3\",\"type\":\"bytes\"}],\"name\":\"FuncEvent1\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"counter\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"arg1\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"arg2\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"arg3\",\"type\":\"bytes\"}],\"name\":\"testFunc1\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"testRandomlyReverted\",\"outputs\":[],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"r\",\"type\":\"bool\"}],\"name\":\"testReverted\",\"outputs\":[],\"stateMutability\":\"pure\",\"type\":\"function\"}]"

// ContractsBin is the compiled bytecode used for deploying new contracts.
var ContractsBin = "0x608060405234801561001057600080fd5b50610755806100206000396000f3fe608060405234801561001057600080fd5b506004361061004c5760003560e01c806325e25eed146100515780634293dafd1461006d57806361bc221a1461007757806388655d9814610095575b600080fd5b61006b60048036038101906100669190610219565b6100b1565b005b6100756100ec565b005b61007f610135565b60405161008c919061025f565b60405180910390f35b6100af60048036038101906100aa919061048d565b61013b565b005b80156100e9576040517fe39eae2f00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b50565b60006004436100fb9190610547565b1415610133576040517fe39eae2f00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b565b60005481565b7fee7ebd5ac9177b3cfe282c440d0220335dc60bc4472338132f06af7b4b9432fc83838360405161016e93929190610655565b60405180910390a1600160008082825461018891906106c9565b925050819055507f4785d80d2593e2cb7a3331d31eb5106408bdde2aab0db9e9b616b036a1b6039d6000546040516101c0919061025f565b60405180910390a1505050565b6000604051905090565b600080fd5b600080fd5b60008115159050919050565b6101f6816101e1565b811461020157600080fd5b50565b600081359050610213816101ed565b92915050565b60006020828403121561022f5761022e6101d7565b5b600061023d84828501610204565b91505092915050565b6000819050919050565b61025981610246565b82525050565b60006020820190506102746000830184610250565b92915050565b600080fd5b600080fd5b6000601f19601f8301169050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b6102cd82610284565b810181811067ffffffffffffffff821117156102ec576102eb610295565b5b80604052505050565b60006102ff6101cd565b905061030b82826102c4565b919050565b600067ffffffffffffffff82111561032b5761032a610295565b5b61033482610284565b9050602081019050919050565b82818337600083830152505050565b600061036361035e84610310565b6102f5565b90508281526020810184848401111561037f5761037e61027f565b5b61038a848285610341565b509392505050565b600082601f8301126103a7576103a661027a565b5b81356103b7848260208601610350565b91505092915050565b6103c981610246565b81146103d457600080fd5b50565b6000813590506103e6816103c0565b92915050565b600067ffffffffffffffff82111561040757610406610295565b5b61041082610284565b9050602081019050919050565b600061043061042b846103ec565b6102f5565b90508281526020810184848401111561044c5761044b61027f565b5b610457848285610341565b509392505050565b600082601f8301126104745761047361027a565b5b813561048484826020860161041d565b91505092915050565b6000806000606084860312156104a6576104a56101d7565b5b600084013567ffffffffffffffff8111156104c4576104c36101dc565b5b6104d086828701610392565b93505060206104e1868287016103d7565b925050604084013567ffffffffffffffff811115610502576105016101dc565b5b61050e8682870161045f565b9150509250925092565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b600061055282610246565b915061055d83610246565b92508261056d5761056c610518565b5b828206905092915050565b600081519050919050565b600082825260208201905092915050565b60005b838110156105b2578082015181840152602081019050610597565b838111156105c1576000848401525b50505050565b60006105d282610578565b6105dc8185610583565b93506105ec818560208601610594565b6105f581610284565b840191505092915050565b600081519050919050565b600082825260208201905092915050565b600061062782610600565b610631818561060b565b9350610641818560208601610594565b61064a81610284565b840191505092915050565b6000606082019050818103600083015261066f81866105c7565b905061067e6020830185610250565b8181036040830152610690818461061c565b9050949350505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b60006106d482610246565b91506106df83610246565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff038211156107145761071361069a565b5b82820190509291505056fea26469706673582212207cfff23b894ab5a15b949ebc96ef61a0aa092e4d2c7bdd3a0e04cec190d116cc64736f6c634300080c0033"

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
