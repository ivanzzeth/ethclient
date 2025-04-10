// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package compatibilityFallbackHandler

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

// CompatibilityFallbackHandlerMetaData contains all meta data concerning the CompatibilityFallbackHandler contract.
var CompatibilityFallbackHandlerMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"NAME\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getMessageHash\",\"inputs\":[{\"name\":\"message\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getMessageHashForSafe\",\"inputs\":[{\"name\":\"safe\",\"type\":\"address\",\"internalType\":\"contractGnosisSafe\"},{\"name\":\"message\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getModules\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address[]\",\"internalType\":\"address[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isValidSignature\",\"inputs\":[{\"name\":\"_dataHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"_signature\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isValidSignature\",\"inputs\":[{\"name\":\"_data\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"_signature\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"onERC1155BatchReceived\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"},{\"name\":\"\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"},{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"onERC1155Received\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"onERC721Received\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"simulate\",\"inputs\":[{\"name\":\"targetContract\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"calldataPayload\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"response\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"supportsInterface\",\"inputs\":[{\"name\":\"interfaceId\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"tokensReceived\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"pure\"}]",
	Bin: "0x60808060405234601557610d71908161001a8239f35b5f80fdfe60806040526004361015610011575f80fd5b5f3560e01c806223de29146100e357806301ffc9a7146100de5780630a1028c4146100d9578063150b7a02146100d45780631626ba7e146100cf57806320c13b0b146100ca5780636ac24784146100c5578063a3f4df7e146100c0578063b2494df3146100bb578063bc197c81146100b6578063bd61951d146100b1578063f23a6e61146100ac5763ffa1ad74146100a7575f80fd5b610a9a565b610a27565b61097a565b6108c2565b610775565b6106c7565b61063c565b6104d2565b6103a9565b61034f565b61030b565b6101c0565b610138565b73ffffffffffffffffffffffffffffffffffffffff81160361010657565b5f80fd5b9181601f840112156101065782359167ffffffffffffffff8311610106576020838186019501011161010657565b346101065760c0366003190112610106576101546004356100e8565b61015f6024356100e8565b61016a6044356100e8565b60843567ffffffffffffffff81116101065761018a90369060040161010a565b505060a43567ffffffffffffffff8111610106576101ac90369060040161010a565b005b6001600160e01b031981160361010657565b346101065760203660031901126101065760206001600160e01b03196004356101e8816101ae565b167f4e2312e0000000000000000000000000000000000000000000000000000000008114908115610250575b8115610226575b506040519015158152f35b7f01ffc9a7000000000000000000000000000000000000000000000000000000009150145f61021b565b630a85bd0160e11b81149150610214565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b90601f8019910116810190811067ffffffffffffffff8211176102b057604052565b610261565b81601f820112156101065780359067ffffffffffffffff82116102b057604051926102ea601f8401601f19166020018561028e565b8284526020838301011161010657815f926020809301838601378301015290565b346101065760203660031901126101065760043567ffffffffffffffff81116101065761034761034160209236906004016102b5565b33610bf2565b604051908152f35b346101065760803660031901126101065761036b6004356100e8565b6103766024356100e8565b60643567ffffffffffffffff81116101065761039690369060040161010a565b50506020604051630a85bd0160e11b8152f35b346101065760403660031901126101065760243567ffffffffffffffff8111610106576103dc602091369060040161010a565b9190604051610406816103f86004358683019190602083019252565b03601f19810183528261028e565b61042460405194859384936320c13b0b60e01b855260048501610b0e565b0381335afa9081156104cd576320c13b0b60e01b916001600160e01b0319915f9161049e575b501603610495576104917f1626ba7e000000000000000000000000000000000000000000000000000000005b6040516001600160e01b031990911681529081906020820190565b0390f35b6104915f610476565b6104c0915060203d6020116104c6575b6104b8818361028e565b810190610af9565b5f61044a565b503d6104ae565b610b48565b346101065760403660031901126101065760043567ffffffffffffffff8111610106576105039036906004016102b5565b60243567ffffffffffffffff8111610106576105239036906004016102b5565b9061052e8133610bf2565b908251155f146105cc57506040517f5ae6bd3700000000000000000000000000000000000000000000000000000000815260048101919091529050602081602481335afa80156104cd5761058b915f9161059d575b501515610b8d565b6040516320c13b0b60e01b8152602090f35b6105bf915060203d6020116105c5575b6105b7818361028e565b810190610b7e565b5f610583565b503d6105ad565b333b15610106575f9161060c60405194859384937f934f3a1100000000000000000000000000000000000000000000000000000000855260048501610b53565b0381335afa80156104cd57610622575b5061058b565b806106305f6106369361028e565b80610685565b5f61061c565b3461010657604036600319011261010657600435610659816100e8565b60243567ffffffffffffffff81116101065760209161067f6103479236906004016102b5565b90610bf2565b5f91031261010657565b805180835260209291819084018484015e5f828201840152601f01601f1916010190565b9060206106c492818152019061068f565b90565b34610106575f366003190112610106576104916040516106e860408261028e565b601881527f44656661756c742043616c6c6261636b2048616e646c65720000000000000000602082015260405191829160208352602083019061068f565b60206040818301928281528451809452019201905f5b8181106107495750505090565b825173ffffffffffffffffffffffffffffffffffffffff1684526020938401939092019160010161073c565b34610106575f366003190112610106576040517fcc2f845200000000000000000000000000000000000000000000000000000000815260016004820152600a60248201525f81604481335afa80156104cd575f906107de575b6104919060405191829182610726565b503d805f833e6107ee818361028e565b810160408282031261010657815167ffffffffffffffff81116101065782019080601f830112156101065781519167ffffffffffffffff83116102b0578260051b9060405193610841602084018661028e565b845260208085019282010192831161010657602001905b8282106108775750505061087160206104919301610d2e565b506107ce565b602080918351610886816100e8565b815201910190610858565b9181601f840112156101065782359167ffffffffffffffff8311610106576020808501948460051b01011161010657565b346101065760a0366003190112610106576108de6004356100e8565b6108e96024356100e8565b60443567ffffffffffffffff811161010657610909903690600401610891565b505060643567ffffffffffffffff81116101065761092b903690600401610891565b505060843567ffffffffffffffff81116101065761094d90369060040161010a565b50506040517fbc197c81000000000000000000000000000000000000000000000000000000008152602090f35b3461010657366003190160408112610106576109976004356100e8565b60243567ffffffffffffffff8111610106576020916109bb5f92369060040161010a565b5050604051907fb4faba09000000000000000000000000000000000000000000000000000000008252600480830137369082335af1503d60405190601f1981830101604052601f19016020823e5f5115610a1f5761049190604051918291826106b3565b602081519101fd5b346101065760a036600319011261010657610a436004356100e8565b610a4e6024356100e8565b60843567ffffffffffffffff811161010657610a6e90369060040161010a565b505060206040517ff23a6e61000000000000000000000000000000000000000000000000000000008152f35b34610106575f36600319011261010657610491604051610abb60408261028e565b600581527f312e302e30000000000000000000000000000000000000000000000000000000602082015260405191829160208352602083019061068f565b9081602091031261010657516106c4816101ae565b9192602093610b26829360408652604086019061068f565b9385818603910152818452848401375f828201840152601f01601f1916010190565b6040513d5f823e3d90fd5b91610b70906106c49492845260606020850152606084019061068f565b91604081840391015261068f565b90816020910312610106575190565b15610b9457565b60646040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601160248201527f48617368206e6f7420617070726f7665640000000000000000000000000000006044820152fd5b602073ffffffffffffffffffffffffffffffffffffffff92818151910120604051610c4f816103f885820194859190602060408401937f60b3cbf8b4a223d68d641b3b6ddf9a298e7f33710cf3d3a9d1146b5a6150fbca81520152565b519020916004604051809581937ff698da25000000000000000000000000000000000000000000000000000000008352165afa9182156104cd575f92610d05575b506040517f1900000000000000000000000000000000000000000000000000000000000000602082019081527f0100000000000000000000000000000000000000000000000000000000000000602183015260228201939093526042810191909152610cff81606281016103f8565b51902090565b6103f8919250610d26610cff9160203d6020116105c5576105b7818361028e565b929150610c90565b5190610d39826100e8565b56fea264697066735822122062572258d81ad821c552d8728edeb23cb9ed61c4763fa1554dd4f02b491ccc2864736f6c634300081c0033",
}

// CompatibilityFallbackHandlerABI is the input ABI used to generate the binding from.
// Deprecated: Use CompatibilityFallbackHandlerMetaData.ABI instead.
var CompatibilityFallbackHandlerABI = CompatibilityFallbackHandlerMetaData.ABI

// CompatibilityFallbackHandlerBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use CompatibilityFallbackHandlerMetaData.Bin instead.
var CompatibilityFallbackHandlerBin = CompatibilityFallbackHandlerMetaData.Bin

// DeployCompatibilityFallbackHandler deploys a new Ethereum contract, binding an instance of CompatibilityFallbackHandler to it.
func DeployCompatibilityFallbackHandler(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *CompatibilityFallbackHandler, error) {
	parsed, err := CompatibilityFallbackHandlerMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(CompatibilityFallbackHandlerBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &CompatibilityFallbackHandler{CompatibilityFallbackHandlerCaller: CompatibilityFallbackHandlerCaller{contract: contract}, CompatibilityFallbackHandlerTransactor: CompatibilityFallbackHandlerTransactor{contract: contract}, CompatibilityFallbackHandlerFilterer: CompatibilityFallbackHandlerFilterer{contract: contract}}, nil
}

// CompatibilityFallbackHandler is an auto generated Go binding around an Ethereum contract.
type CompatibilityFallbackHandler struct {
	CompatibilityFallbackHandlerCaller     // Read-only binding to the contract
	CompatibilityFallbackHandlerTransactor // Write-only binding to the contract
	CompatibilityFallbackHandlerFilterer   // Log filterer for contract events
}

// CompatibilityFallbackHandlerCaller is an auto generated read-only Go binding around an Ethereum contract.
type CompatibilityFallbackHandlerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CompatibilityFallbackHandlerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type CompatibilityFallbackHandlerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CompatibilityFallbackHandlerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type CompatibilityFallbackHandlerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CompatibilityFallbackHandlerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type CompatibilityFallbackHandlerSession struct {
	Contract     *CompatibilityFallbackHandler // Generic contract binding to set the session for
	CallOpts     bind.CallOpts                 // Call options to use throughout this session
	TransactOpts bind.TransactOpts             // Transaction auth options to use throughout this session
}

// CompatibilityFallbackHandlerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type CompatibilityFallbackHandlerCallerSession struct {
	Contract *CompatibilityFallbackHandlerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                       // Call options to use throughout this session
}

// CompatibilityFallbackHandlerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type CompatibilityFallbackHandlerTransactorSession struct {
	Contract     *CompatibilityFallbackHandlerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                       // Transaction auth options to use throughout this session
}

// CompatibilityFallbackHandlerRaw is an auto generated low-level Go binding around an Ethereum contract.
type CompatibilityFallbackHandlerRaw struct {
	Contract *CompatibilityFallbackHandler // Generic contract binding to access the raw methods on
}

// CompatibilityFallbackHandlerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type CompatibilityFallbackHandlerCallerRaw struct {
	Contract *CompatibilityFallbackHandlerCaller // Generic read-only contract binding to access the raw methods on
}

// CompatibilityFallbackHandlerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type CompatibilityFallbackHandlerTransactorRaw struct {
	Contract *CompatibilityFallbackHandlerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewCompatibilityFallbackHandler creates a new instance of CompatibilityFallbackHandler, bound to a specific deployed contract.
func NewCompatibilityFallbackHandler(address common.Address, backend bind.ContractBackend) (*CompatibilityFallbackHandler, error) {
	contract, err := bindCompatibilityFallbackHandler(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &CompatibilityFallbackHandler{CompatibilityFallbackHandlerCaller: CompatibilityFallbackHandlerCaller{contract: contract}, CompatibilityFallbackHandlerTransactor: CompatibilityFallbackHandlerTransactor{contract: contract}, CompatibilityFallbackHandlerFilterer: CompatibilityFallbackHandlerFilterer{contract: contract}}, nil
}

// NewCompatibilityFallbackHandlerCaller creates a new read-only instance of CompatibilityFallbackHandler, bound to a specific deployed contract.
func NewCompatibilityFallbackHandlerCaller(address common.Address, caller bind.ContractCaller) (*CompatibilityFallbackHandlerCaller, error) {
	contract, err := bindCompatibilityFallbackHandler(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &CompatibilityFallbackHandlerCaller{contract: contract}, nil
}

// NewCompatibilityFallbackHandlerTransactor creates a new write-only instance of CompatibilityFallbackHandler, bound to a specific deployed contract.
func NewCompatibilityFallbackHandlerTransactor(address common.Address, transactor bind.ContractTransactor) (*CompatibilityFallbackHandlerTransactor, error) {
	contract, err := bindCompatibilityFallbackHandler(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &CompatibilityFallbackHandlerTransactor{contract: contract}, nil
}

// NewCompatibilityFallbackHandlerFilterer creates a new log filterer instance of CompatibilityFallbackHandler, bound to a specific deployed contract.
func NewCompatibilityFallbackHandlerFilterer(address common.Address, filterer bind.ContractFilterer) (*CompatibilityFallbackHandlerFilterer, error) {
	contract, err := bindCompatibilityFallbackHandler(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &CompatibilityFallbackHandlerFilterer{contract: contract}, nil
}

// bindCompatibilityFallbackHandler binds a generic wrapper to an already deployed contract.
func bindCompatibilityFallbackHandler(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := CompatibilityFallbackHandlerMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _CompatibilityFallbackHandler.Contract.CompatibilityFallbackHandlerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CompatibilityFallbackHandler.Contract.CompatibilityFallbackHandlerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _CompatibilityFallbackHandler.Contract.CompatibilityFallbackHandlerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _CompatibilityFallbackHandler.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CompatibilityFallbackHandler.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _CompatibilityFallbackHandler.Contract.contract.Transact(opts, method, params...)
}

// NAME is a free data retrieval call binding the contract method 0xa3f4df7e.
//
// Solidity: function NAME() view returns(string)
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerCaller) NAME(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _CompatibilityFallbackHandler.contract.Call(opts, &out, "NAME")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// NAME is a free data retrieval call binding the contract method 0xa3f4df7e.
//
// Solidity: function NAME() view returns(string)
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerSession) NAME() (string, error) {
	return _CompatibilityFallbackHandler.Contract.NAME(&_CompatibilityFallbackHandler.CallOpts)
}

// NAME is a free data retrieval call binding the contract method 0xa3f4df7e.
//
// Solidity: function NAME() view returns(string)
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerCallerSession) NAME() (string, error) {
	return _CompatibilityFallbackHandler.Contract.NAME(&_CompatibilityFallbackHandler.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(string)
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerCaller) VERSION(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _CompatibilityFallbackHandler.contract.Call(opts, &out, "VERSION")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(string)
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerSession) VERSION() (string, error) {
	return _CompatibilityFallbackHandler.Contract.VERSION(&_CompatibilityFallbackHandler.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(string)
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerCallerSession) VERSION() (string, error) {
	return _CompatibilityFallbackHandler.Contract.VERSION(&_CompatibilityFallbackHandler.CallOpts)
}

// GetMessageHash is a free data retrieval call binding the contract method 0x0a1028c4.
//
// Solidity: function getMessageHash(bytes message) view returns(bytes32)
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerCaller) GetMessageHash(opts *bind.CallOpts, message []byte) ([32]byte, error) {
	var out []interface{}
	err := _CompatibilityFallbackHandler.contract.Call(opts, &out, "getMessageHash", message)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetMessageHash is a free data retrieval call binding the contract method 0x0a1028c4.
//
// Solidity: function getMessageHash(bytes message) view returns(bytes32)
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerSession) GetMessageHash(message []byte) ([32]byte, error) {
	return _CompatibilityFallbackHandler.Contract.GetMessageHash(&_CompatibilityFallbackHandler.CallOpts, message)
}

// GetMessageHash is a free data retrieval call binding the contract method 0x0a1028c4.
//
// Solidity: function getMessageHash(bytes message) view returns(bytes32)
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerCallerSession) GetMessageHash(message []byte) ([32]byte, error) {
	return _CompatibilityFallbackHandler.Contract.GetMessageHash(&_CompatibilityFallbackHandler.CallOpts, message)
}

// GetMessageHashForSafe is a free data retrieval call binding the contract method 0x6ac24784.
//
// Solidity: function getMessageHashForSafe(address safe, bytes message) view returns(bytes32)
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerCaller) GetMessageHashForSafe(opts *bind.CallOpts, safe common.Address, message []byte) ([32]byte, error) {
	var out []interface{}
	err := _CompatibilityFallbackHandler.contract.Call(opts, &out, "getMessageHashForSafe", safe, message)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetMessageHashForSafe is a free data retrieval call binding the contract method 0x6ac24784.
//
// Solidity: function getMessageHashForSafe(address safe, bytes message) view returns(bytes32)
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerSession) GetMessageHashForSafe(safe common.Address, message []byte) ([32]byte, error) {
	return _CompatibilityFallbackHandler.Contract.GetMessageHashForSafe(&_CompatibilityFallbackHandler.CallOpts, safe, message)
}

// GetMessageHashForSafe is a free data retrieval call binding the contract method 0x6ac24784.
//
// Solidity: function getMessageHashForSafe(address safe, bytes message) view returns(bytes32)
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerCallerSession) GetMessageHashForSafe(safe common.Address, message []byte) ([32]byte, error) {
	return _CompatibilityFallbackHandler.Contract.GetMessageHashForSafe(&_CompatibilityFallbackHandler.CallOpts, safe, message)
}

// GetModules is a free data retrieval call binding the contract method 0xb2494df3.
//
// Solidity: function getModules() view returns(address[])
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerCaller) GetModules(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _CompatibilityFallbackHandler.contract.Call(opts, &out, "getModules")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetModules is a free data retrieval call binding the contract method 0xb2494df3.
//
// Solidity: function getModules() view returns(address[])
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerSession) GetModules() ([]common.Address, error) {
	return _CompatibilityFallbackHandler.Contract.GetModules(&_CompatibilityFallbackHandler.CallOpts)
}

// GetModules is a free data retrieval call binding the contract method 0xb2494df3.
//
// Solidity: function getModules() view returns(address[])
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerCallerSession) GetModules() ([]common.Address, error) {
	return _CompatibilityFallbackHandler.Contract.GetModules(&_CompatibilityFallbackHandler.CallOpts)
}

// IsValidSignature is a free data retrieval call binding the contract method 0x1626ba7e.
//
// Solidity: function isValidSignature(bytes32 _dataHash, bytes _signature) view returns(bytes4)
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerCaller) IsValidSignature(opts *bind.CallOpts, _dataHash [32]byte, _signature []byte) ([4]byte, error) {
	var out []interface{}
	err := _CompatibilityFallbackHandler.contract.Call(opts, &out, "isValidSignature", _dataHash, _signature)

	if err != nil {
		return *new([4]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([4]byte)).(*[4]byte)

	return out0, err

}

// IsValidSignature is a free data retrieval call binding the contract method 0x1626ba7e.
//
// Solidity: function isValidSignature(bytes32 _dataHash, bytes _signature) view returns(bytes4)
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerSession) IsValidSignature(_dataHash [32]byte, _signature []byte) ([4]byte, error) {
	return _CompatibilityFallbackHandler.Contract.IsValidSignature(&_CompatibilityFallbackHandler.CallOpts, _dataHash, _signature)
}

// IsValidSignature is a free data retrieval call binding the contract method 0x1626ba7e.
//
// Solidity: function isValidSignature(bytes32 _dataHash, bytes _signature) view returns(bytes4)
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerCallerSession) IsValidSignature(_dataHash [32]byte, _signature []byte) ([4]byte, error) {
	return _CompatibilityFallbackHandler.Contract.IsValidSignature(&_CompatibilityFallbackHandler.CallOpts, _dataHash, _signature)
}

// IsValidSignature0 is a free data retrieval call binding the contract method 0x20c13b0b.
//
// Solidity: function isValidSignature(bytes _data, bytes _signature) view returns(bytes4)
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerCaller) IsValidSignature0(opts *bind.CallOpts, _data []byte, _signature []byte) ([4]byte, error) {
	var out []interface{}
	err := _CompatibilityFallbackHandler.contract.Call(opts, &out, "isValidSignature0", _data, _signature)

	if err != nil {
		return *new([4]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([4]byte)).(*[4]byte)

	return out0, err

}

// IsValidSignature0 is a free data retrieval call binding the contract method 0x20c13b0b.
//
// Solidity: function isValidSignature(bytes _data, bytes _signature) view returns(bytes4)
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerSession) IsValidSignature0(_data []byte, _signature []byte) ([4]byte, error) {
	return _CompatibilityFallbackHandler.Contract.IsValidSignature0(&_CompatibilityFallbackHandler.CallOpts, _data, _signature)
}

// IsValidSignature0 is a free data retrieval call binding the contract method 0x20c13b0b.
//
// Solidity: function isValidSignature(bytes _data, bytes _signature) view returns(bytes4)
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerCallerSession) IsValidSignature0(_data []byte, _signature []byte) ([4]byte, error) {
	return _CompatibilityFallbackHandler.Contract.IsValidSignature0(&_CompatibilityFallbackHandler.CallOpts, _data, _signature)
}

// OnERC1155BatchReceived is a free data retrieval call binding the contract method 0xbc197c81.
//
// Solidity: function onERC1155BatchReceived(address , address , uint256[] , uint256[] , bytes ) pure returns(bytes4)
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerCaller) OnERC1155BatchReceived(opts *bind.CallOpts, arg0 common.Address, arg1 common.Address, arg2 []*big.Int, arg3 []*big.Int, arg4 []byte) ([4]byte, error) {
	var out []interface{}
	err := _CompatibilityFallbackHandler.contract.Call(opts, &out, "onERC1155BatchReceived", arg0, arg1, arg2, arg3, arg4)

	if err != nil {
		return *new([4]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([4]byte)).(*[4]byte)

	return out0, err

}

// OnERC1155BatchReceived is a free data retrieval call binding the contract method 0xbc197c81.
//
// Solidity: function onERC1155BatchReceived(address , address , uint256[] , uint256[] , bytes ) pure returns(bytes4)
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerSession) OnERC1155BatchReceived(arg0 common.Address, arg1 common.Address, arg2 []*big.Int, arg3 []*big.Int, arg4 []byte) ([4]byte, error) {
	return _CompatibilityFallbackHandler.Contract.OnERC1155BatchReceived(&_CompatibilityFallbackHandler.CallOpts, arg0, arg1, arg2, arg3, arg4)
}

// OnERC1155BatchReceived is a free data retrieval call binding the contract method 0xbc197c81.
//
// Solidity: function onERC1155BatchReceived(address , address , uint256[] , uint256[] , bytes ) pure returns(bytes4)
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerCallerSession) OnERC1155BatchReceived(arg0 common.Address, arg1 common.Address, arg2 []*big.Int, arg3 []*big.Int, arg4 []byte) ([4]byte, error) {
	return _CompatibilityFallbackHandler.Contract.OnERC1155BatchReceived(&_CompatibilityFallbackHandler.CallOpts, arg0, arg1, arg2, arg3, arg4)
}

// OnERC1155Received is a free data retrieval call binding the contract method 0xf23a6e61.
//
// Solidity: function onERC1155Received(address , address , uint256 , uint256 , bytes ) pure returns(bytes4)
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerCaller) OnERC1155Received(opts *bind.CallOpts, arg0 common.Address, arg1 common.Address, arg2 *big.Int, arg3 *big.Int, arg4 []byte) ([4]byte, error) {
	var out []interface{}
	err := _CompatibilityFallbackHandler.contract.Call(opts, &out, "onERC1155Received", arg0, arg1, arg2, arg3, arg4)

	if err != nil {
		return *new([4]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([4]byte)).(*[4]byte)

	return out0, err

}

// OnERC1155Received is a free data retrieval call binding the contract method 0xf23a6e61.
//
// Solidity: function onERC1155Received(address , address , uint256 , uint256 , bytes ) pure returns(bytes4)
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerSession) OnERC1155Received(arg0 common.Address, arg1 common.Address, arg2 *big.Int, arg3 *big.Int, arg4 []byte) ([4]byte, error) {
	return _CompatibilityFallbackHandler.Contract.OnERC1155Received(&_CompatibilityFallbackHandler.CallOpts, arg0, arg1, arg2, arg3, arg4)
}

// OnERC1155Received is a free data retrieval call binding the contract method 0xf23a6e61.
//
// Solidity: function onERC1155Received(address , address , uint256 , uint256 , bytes ) pure returns(bytes4)
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerCallerSession) OnERC1155Received(arg0 common.Address, arg1 common.Address, arg2 *big.Int, arg3 *big.Int, arg4 []byte) ([4]byte, error) {
	return _CompatibilityFallbackHandler.Contract.OnERC1155Received(&_CompatibilityFallbackHandler.CallOpts, arg0, arg1, arg2, arg3, arg4)
}

// OnERC721Received is a free data retrieval call binding the contract method 0x150b7a02.
//
// Solidity: function onERC721Received(address , address , uint256 , bytes ) pure returns(bytes4)
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerCaller) OnERC721Received(opts *bind.CallOpts, arg0 common.Address, arg1 common.Address, arg2 *big.Int, arg3 []byte) ([4]byte, error) {
	var out []interface{}
	err := _CompatibilityFallbackHandler.contract.Call(opts, &out, "onERC721Received", arg0, arg1, arg2, arg3)

	if err != nil {
		return *new([4]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([4]byte)).(*[4]byte)

	return out0, err

}

// OnERC721Received is a free data retrieval call binding the contract method 0x150b7a02.
//
// Solidity: function onERC721Received(address , address , uint256 , bytes ) pure returns(bytes4)
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerSession) OnERC721Received(arg0 common.Address, arg1 common.Address, arg2 *big.Int, arg3 []byte) ([4]byte, error) {
	return _CompatibilityFallbackHandler.Contract.OnERC721Received(&_CompatibilityFallbackHandler.CallOpts, arg0, arg1, arg2, arg3)
}

// OnERC721Received is a free data retrieval call binding the contract method 0x150b7a02.
//
// Solidity: function onERC721Received(address , address , uint256 , bytes ) pure returns(bytes4)
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerCallerSession) OnERC721Received(arg0 common.Address, arg1 common.Address, arg2 *big.Int, arg3 []byte) ([4]byte, error) {
	return _CompatibilityFallbackHandler.Contract.OnERC721Received(&_CompatibilityFallbackHandler.CallOpts, arg0, arg1, arg2, arg3)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _CompatibilityFallbackHandler.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _CompatibilityFallbackHandler.Contract.SupportsInterface(&_CompatibilityFallbackHandler.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _CompatibilityFallbackHandler.Contract.SupportsInterface(&_CompatibilityFallbackHandler.CallOpts, interfaceId)
}

// TokensReceived is a free data retrieval call binding the contract method 0x0023de29.
//
// Solidity: function tokensReceived(address , address , address , uint256 , bytes , bytes ) pure returns()
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerCaller) TokensReceived(opts *bind.CallOpts, arg0 common.Address, arg1 common.Address, arg2 common.Address, arg3 *big.Int, arg4 []byte, arg5 []byte) error {
	var out []interface{}
	err := _CompatibilityFallbackHandler.contract.Call(opts, &out, "tokensReceived", arg0, arg1, arg2, arg3, arg4, arg5)

	if err != nil {
		return err
	}

	return err

}

// TokensReceived is a free data retrieval call binding the contract method 0x0023de29.
//
// Solidity: function tokensReceived(address , address , address , uint256 , bytes , bytes ) pure returns()
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerSession) TokensReceived(arg0 common.Address, arg1 common.Address, arg2 common.Address, arg3 *big.Int, arg4 []byte, arg5 []byte) error {
	return _CompatibilityFallbackHandler.Contract.TokensReceived(&_CompatibilityFallbackHandler.CallOpts, arg0, arg1, arg2, arg3, arg4, arg5)
}

// TokensReceived is a free data retrieval call binding the contract method 0x0023de29.
//
// Solidity: function tokensReceived(address , address , address , uint256 , bytes , bytes ) pure returns()
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerCallerSession) TokensReceived(arg0 common.Address, arg1 common.Address, arg2 common.Address, arg3 *big.Int, arg4 []byte, arg5 []byte) error {
	return _CompatibilityFallbackHandler.Contract.TokensReceived(&_CompatibilityFallbackHandler.CallOpts, arg0, arg1, arg2, arg3, arg4, arg5)
}

// Simulate is a paid mutator transaction binding the contract method 0xbd61951d.
//
// Solidity: function simulate(address targetContract, bytes calldataPayload) returns(bytes response)
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerTransactor) Simulate(opts *bind.TransactOpts, targetContract common.Address, calldataPayload []byte) (*types.Transaction, error) {
	return _CompatibilityFallbackHandler.contract.Transact(opts, "simulate", targetContract, calldataPayload)
}

// Simulate is a paid mutator transaction binding the contract method 0xbd61951d.
//
// Solidity: function simulate(address targetContract, bytes calldataPayload) returns(bytes response)
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerSession) Simulate(targetContract common.Address, calldataPayload []byte) (*types.Transaction, error) {
	return _CompatibilityFallbackHandler.Contract.Simulate(&_CompatibilityFallbackHandler.TransactOpts, targetContract, calldataPayload)
}

// Simulate is a paid mutator transaction binding the contract method 0xbd61951d.
//
// Solidity: function simulate(address targetContract, bytes calldataPayload) returns(bytes response)
func (_CompatibilityFallbackHandler *CompatibilityFallbackHandlerTransactorSession) Simulate(targetContract common.Address, calldataPayload []byte) (*types.Transaction, error) {
	return _CompatibilityFallbackHandler.Contract.Simulate(&_CompatibilityFallbackHandler.TransactOpts, targetContract, calldataPayload)
}
