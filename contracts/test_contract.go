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
	ABI: "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"a\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"b\",\"type\":\"uint256\"}],\"name\":\"TestRevert\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"counter\",\"type\":\"uint256\"}],\"name\":\"CounterUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"arg1\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"arg2\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"arg3\",\"type\":\"bytes\"}],\"name\":\"FuncEvent1\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"counter\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"arg1\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"arg2\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"arg3\",\"type\":\"bytes\"}],\"name\":\"testFunc1\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"testRandomlyReverted\",\"outputs\":[],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"r\",\"type\":\"bool\"}],\"name\":\"testReverted\",\"outputs\":[],\"stateMutability\":\"pure\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561001057600080fd5b5061081c806100206000396000f3fe608060405234801561001057600080fd5b506004361061004c5760003560e01c806325e25eed146100515780634293dafd1461006d57806361bc221a1461007757806388655d9814610095575b600080fd5b61006b60048036038101906100669190610237565b6100b1565b005b6100756100fb565b005b61007f610153565b60405161008c919061027d565b60405180910390f35b6100af60048036038101906100aa91906104ab565b610159565b005b80156100f857600160026040517fbff0f7d70000000000000000000000000000000000000000000000000000000081526004016100ef9291906105b6565b60405180910390fd5b50565b600060044361010a919061060e565b141561015157600160026040517fbff0f7d70000000000000000000000000000000000000000000000000000000081526004016101489291906105b6565b60405180910390fd5b565b60005481565b7fee7ebd5ac9177b3cfe282c440d0220335dc60bc4472338132f06af7b4b9432fc83838360405161018c9392919061071c565b60405180910390a160016000808282546101a69190610790565b925050819055507f4785d80d2593e2cb7a3331d31eb5106408bdde2aab0db9e9b616b036a1b6039d6000546040516101de919061027d565b60405180910390a1505050565b6000604051905090565b600080fd5b600080fd5b60008115159050919050565b610214816101ff565b811461021f57600080fd5b50565b6000813590506102318161020b565b92915050565b60006020828403121561024d5761024c6101f5565b5b600061025b84828501610222565b91505092915050565b6000819050919050565b61027781610264565b82525050565b6000602082019050610292600083018461026e565b92915050565b600080fd5b600080fd5b6000601f19601f8301169050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b6102eb826102a2565b810181811067ffffffffffffffff8211171561030a576103096102b3565b5b80604052505050565b600061031d6101eb565b905061032982826102e2565b919050565b600067ffffffffffffffff821115610349576103486102b3565b5b610352826102a2565b9050602081019050919050565b82818337600083830152505050565b600061038161037c8461032e565b610313565b90508281526020810184848401111561039d5761039c61029d565b5b6103a884828561035f565b509392505050565b600082601f8301126103c5576103c4610298565b5b81356103d584826020860161036e565b91505092915050565b6103e781610264565b81146103f257600080fd5b50565b600081359050610404816103de565b92915050565b600067ffffffffffffffff821115610425576104246102b3565b5b61042e826102a2565b9050602081019050919050565b600061044e6104498461040a565b610313565b90508281526020810184848401111561046a5761046961029d565b5b61047584828561035f565b509392505050565b600082601f83011261049257610491610298565b5b81356104a284826020860161043b565b91505092915050565b6000806000606084860312156104c4576104c36101f5565b5b600084013567ffffffffffffffff8111156104e2576104e16101fa565b5b6104ee868287016103b0565b93505060206104ff868287016103f5565b925050604084013567ffffffffffffffff8111156105205761051f6101fa565b5b61052c8682870161047d565b9150509250925092565b6000819050919050565b6000819050919050565b600061056561056061055b84610536565b610540565b610264565b9050919050565b6105758161054a565b82525050565b6000819050919050565b60006105a061059b6105968461057b565b610540565b610264565b9050919050565b6105b081610585565b82525050565b60006040820190506105cb600083018561056c565b6105d860208301846105a7565b9392505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b600061061982610264565b915061062483610264565b925082610634576106336105df565b5b828206905092915050565b600081519050919050565b600082825260208201905092915050565b60005b8381101561067957808201518184015260208101905061065e565b83811115610688576000848401525b50505050565b60006106998261063f565b6106a3818561064a565b93506106b381856020860161065b565b6106bc816102a2565b840191505092915050565b600081519050919050565b600082825260208201905092915050565b60006106ee826106c7565b6106f881856106d2565b935061070881856020860161065b565b610711816102a2565b840191505092915050565b60006060820190508181036000830152610736818661068e565b9050610745602083018561026e565b818103604083015261075781846106e3565b9050949350505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b600061079b82610264565b91506107a683610264565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff038211156107db576107da610761565b5b82820190509291505056fea2646970667358221220393e4eb737c7bececb165a0fb9ec218ed9d66b273630b65fe4c73f354120f48064736f6c634300080c0033",
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
