// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package airdrop_l1

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

// AirdropL1MetaData contains all meta data concerning the AirdropL1 contract.
var AirdropL1MetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_rewardToken\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_signer\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"trader\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Claimed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"WithdrawTo\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"trader\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"claim\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"external_signer\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"processedClaims\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"rewardToken\",\"outputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"withdrawTokensTo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// AirdropL1ABI is the input ABI used to generate the binding from.
// Deprecated: Use AirdropL1MetaData.ABI instead.
var AirdropL1ABI = AirdropL1MetaData.ABI

// AirdropL1 is an auto generated Go binding around an Ethereum contract.
type AirdropL1 struct {
	AirdropL1Caller     // Read-only binding to the contract
	AirdropL1Transactor // Write-only binding to the contract
	AirdropL1Filterer   // Log filterer for contract events
}

// AirdropL1Caller is an auto generated read-only Go binding around an Ethereum contract.
type AirdropL1Caller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AirdropL1Transactor is an auto generated write-only Go binding around an Ethereum contract.
type AirdropL1Transactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AirdropL1Filterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AirdropL1Filterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AirdropL1Session is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AirdropL1Session struct {
	Contract     *AirdropL1        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AirdropL1CallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AirdropL1CallerSession struct {
	Contract *AirdropL1Caller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// AirdropL1TransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AirdropL1TransactorSession struct {
	Contract     *AirdropL1Transactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// AirdropL1Raw is an auto generated low-level Go binding around an Ethereum contract.
type AirdropL1Raw struct {
	Contract *AirdropL1 // Generic contract binding to access the raw methods on
}

// AirdropL1CallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AirdropL1CallerRaw struct {
	Contract *AirdropL1Caller // Generic read-only contract binding to access the raw methods on
}

// AirdropL1TransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AirdropL1TransactorRaw struct {
	Contract *AirdropL1Transactor // Generic write-only contract binding to access the raw methods on
}

// NewAirdropL1 creates a new instance of AirdropL1, bound to a specific deployed contract.
func NewAirdropL1(address common.Address, backend bind.ContractBackend) (*AirdropL1, error) {
	contract, err := bindAirdropL1(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AirdropL1{AirdropL1Caller: AirdropL1Caller{contract: contract}, AirdropL1Transactor: AirdropL1Transactor{contract: contract}, AirdropL1Filterer: AirdropL1Filterer{contract: contract}}, nil
}

// NewAirdropL1Caller creates a new read-only instance of AirdropL1, bound to a specific deployed contract.
func NewAirdropL1Caller(address common.Address, caller bind.ContractCaller) (*AirdropL1Caller, error) {
	contract, err := bindAirdropL1(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AirdropL1Caller{contract: contract}, nil
}

// NewAirdropL1Transactor creates a new write-only instance of AirdropL1, bound to a specific deployed contract.
func NewAirdropL1Transactor(address common.Address, transactor bind.ContractTransactor) (*AirdropL1Transactor, error) {
	contract, err := bindAirdropL1(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AirdropL1Transactor{contract: contract}, nil
}

// NewAirdropL1Filterer creates a new log filterer instance of AirdropL1, bound to a specific deployed contract.
func NewAirdropL1Filterer(address common.Address, filterer bind.ContractFilterer) (*AirdropL1Filterer, error) {
	contract, err := bindAirdropL1(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AirdropL1Filterer{contract: contract}, nil
}

// bindAirdropL1 binds a generic wrapper to an already deployed contract.
func bindAirdropL1(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := AirdropL1MetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AirdropL1 *AirdropL1Raw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AirdropL1.Contract.AirdropL1Caller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AirdropL1 *AirdropL1Raw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AirdropL1.Contract.AirdropL1Transactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AirdropL1 *AirdropL1Raw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AirdropL1.Contract.AirdropL1Transactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AirdropL1 *AirdropL1CallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AirdropL1.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AirdropL1 *AirdropL1TransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AirdropL1.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AirdropL1 *AirdropL1TransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AirdropL1.Contract.contract.Transact(opts, method, params...)
}

// ExternalSigner is a free data retrieval call binding the contract method 0x0008cecb.
//
// Solidity: function external_signer() view returns(address)
func (_AirdropL1 *AirdropL1Caller) ExternalSigner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AirdropL1.contract.Call(opts, &out, "external_signer")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ExternalSigner is a free data retrieval call binding the contract method 0x0008cecb.
//
// Solidity: function external_signer() view returns(address)
func (_AirdropL1 *AirdropL1Session) ExternalSigner() (common.Address, error) {
	return _AirdropL1.Contract.ExternalSigner(&_AirdropL1.CallOpts)
}

// ExternalSigner is a free data retrieval call binding the contract method 0x0008cecb.
//
// Solidity: function external_signer() view returns(address)
func (_AirdropL1 *AirdropL1CallerSession) ExternalSigner() (common.Address, error) {
	return _AirdropL1.Contract.ExternalSigner(&_AirdropL1.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_AirdropL1 *AirdropL1Caller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AirdropL1.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_AirdropL1 *AirdropL1Session) Owner() (common.Address, error) {
	return _AirdropL1.Contract.Owner(&_AirdropL1.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_AirdropL1 *AirdropL1CallerSession) Owner() (common.Address, error) {
	return _AirdropL1.Contract.Owner(&_AirdropL1.CallOpts)
}

// ProcessedClaims is a free data retrieval call binding the contract method 0x653f0323.
//
// Solidity: function processedClaims(uint256 ) view returns(bool)
func (_AirdropL1 *AirdropL1Caller) ProcessedClaims(opts *bind.CallOpts, arg0 *big.Int) (bool, error) {
	var out []interface{}
	err := _AirdropL1.contract.Call(opts, &out, "processedClaims", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// ProcessedClaims is a free data retrieval call binding the contract method 0x653f0323.
//
// Solidity: function processedClaims(uint256 ) view returns(bool)
func (_AirdropL1 *AirdropL1Session) ProcessedClaims(arg0 *big.Int) (bool, error) {
	return _AirdropL1.Contract.ProcessedClaims(&_AirdropL1.CallOpts, arg0)
}

// ProcessedClaims is a free data retrieval call binding the contract method 0x653f0323.
//
// Solidity: function processedClaims(uint256 ) view returns(bool)
func (_AirdropL1 *AirdropL1CallerSession) ProcessedClaims(arg0 *big.Int) (bool, error) {
	return _AirdropL1.Contract.ProcessedClaims(&_AirdropL1.CallOpts, arg0)
}

// RewardToken is a free data retrieval call binding the contract method 0xf7c618c1.
//
// Solidity: function rewardToken() view returns(address)
func (_AirdropL1 *AirdropL1Caller) RewardToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AirdropL1.contract.Call(opts, &out, "rewardToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// RewardToken is a free data retrieval call binding the contract method 0xf7c618c1.
//
// Solidity: function rewardToken() view returns(address)
func (_AirdropL1 *AirdropL1Session) RewardToken() (common.Address, error) {
	return _AirdropL1.Contract.RewardToken(&_AirdropL1.CallOpts)
}

// RewardToken is a free data retrieval call binding the contract method 0xf7c618c1.
//
// Solidity: function rewardToken() view returns(address)
func (_AirdropL1 *AirdropL1CallerSession) RewardToken() (common.Address, error) {
	return _AirdropL1.Contract.RewardToken(&_AirdropL1.CallOpts)
}

// Claim is a paid mutator transaction binding the contract method 0xd2b9c645.
//
// Solidity: function claim(uint256 id, address trader, uint256 amount, uint8 v, bytes32 r, bytes32 s) returns()
func (_AirdropL1 *AirdropL1Transactor) Claim(opts *bind.TransactOpts, id *big.Int, trader common.Address, amount *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _AirdropL1.contract.Transact(opts, "claim", id, trader, amount, v, r, s)
}

// Claim is a paid mutator transaction binding the contract method 0xd2b9c645.
//
// Solidity: function claim(uint256 id, address trader, uint256 amount, uint8 v, bytes32 r, bytes32 s) returns()
func (_AirdropL1 *AirdropL1Session) Claim(id *big.Int, trader common.Address, amount *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _AirdropL1.Contract.Claim(&_AirdropL1.TransactOpts, id, trader, amount, v, r, s)
}

// Claim is a paid mutator transaction binding the contract method 0xd2b9c645.
//
// Solidity: function claim(uint256 id, address trader, uint256 amount, uint8 v, bytes32 r, bytes32 s) returns()
func (_AirdropL1 *AirdropL1TransactorSession) Claim(id *big.Int, trader common.Address, amount *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _AirdropL1.Contract.Claim(&_AirdropL1.TransactOpts, id, trader, amount, v, r, s)
}

// WithdrawTokensTo is a paid mutator transaction binding the contract method 0xb743e722.
//
// Solidity: function withdrawTokensTo(uint256 amount, address to) returns()
func (_AirdropL1 *AirdropL1Transactor) WithdrawTokensTo(opts *bind.TransactOpts, amount *big.Int, to common.Address) (*types.Transaction, error) {
	return _AirdropL1.contract.Transact(opts, "withdrawTokensTo", amount, to)
}

// WithdrawTokensTo is a paid mutator transaction binding the contract method 0xb743e722.
//
// Solidity: function withdrawTokensTo(uint256 amount, address to) returns()
func (_AirdropL1 *AirdropL1Session) WithdrawTokensTo(amount *big.Int, to common.Address) (*types.Transaction, error) {
	return _AirdropL1.Contract.WithdrawTokensTo(&_AirdropL1.TransactOpts, amount, to)
}

// WithdrawTokensTo is a paid mutator transaction binding the contract method 0xb743e722.
//
// Solidity: function withdrawTokensTo(uint256 amount, address to) returns()
func (_AirdropL1 *AirdropL1TransactorSession) WithdrawTokensTo(amount *big.Int, to common.Address) (*types.Transaction, error) {
	return _AirdropL1.Contract.WithdrawTokensTo(&_AirdropL1.TransactOpts, amount, to)
}

// AirdropL1ClaimedIterator is returned from FilterClaimed and is used to iterate over the raw logs and unpacked data for Claimed events raised by the AirdropL1 contract.
type AirdropL1ClaimedIterator struct {
	Event *AirdropL1Claimed // Event containing the contract specifics and raw log

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
func (it *AirdropL1ClaimedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AirdropL1Claimed)
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
		it.Event = new(AirdropL1Claimed)
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
func (it *AirdropL1ClaimedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AirdropL1ClaimedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AirdropL1Claimed represents a Claimed event raised by the AirdropL1 contract.
type AirdropL1Claimed struct {
	Trader common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterClaimed is a free log retrieval operation binding the contract event 0xd8138f8a3f377c5259ca548e70e4c2de94f129f5a11036a15b69513cba2b426a.
//
// Solidity: event Claimed(address indexed trader, uint256 indexed amount)
func (_AirdropL1 *AirdropL1Filterer) FilterClaimed(opts *bind.FilterOpts, trader []common.Address, amount []*big.Int) (*AirdropL1ClaimedIterator, error) {

	var traderRule []interface{}
	for _, traderItem := range trader {
		traderRule = append(traderRule, traderItem)
	}
	var amountRule []interface{}
	for _, amountItem := range amount {
		amountRule = append(amountRule, amountItem)
	}

	logs, sub, err := _AirdropL1.contract.FilterLogs(opts, "Claimed", traderRule, amountRule)
	if err != nil {
		return nil, err
	}
	return &AirdropL1ClaimedIterator{contract: _AirdropL1.contract, event: "Claimed", logs: logs, sub: sub}, nil
}

// WatchClaimed is a free log subscription operation binding the contract event 0xd8138f8a3f377c5259ca548e70e4c2de94f129f5a11036a15b69513cba2b426a.
//
// Solidity: event Claimed(address indexed trader, uint256 indexed amount)
func (_AirdropL1 *AirdropL1Filterer) WatchClaimed(opts *bind.WatchOpts, sink chan<- *AirdropL1Claimed, trader []common.Address, amount []*big.Int) (event.Subscription, error) {

	var traderRule []interface{}
	for _, traderItem := range trader {
		traderRule = append(traderRule, traderItem)
	}
	var amountRule []interface{}
	for _, amountItem := range amount {
		amountRule = append(amountRule, amountItem)
	}

	logs, sub, err := _AirdropL1.contract.WatchLogs(opts, "Claimed", traderRule, amountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AirdropL1Claimed)
				if err := _AirdropL1.contract.UnpackLog(event, "Claimed", log); err != nil {
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

// ParseClaimed is a log parse operation binding the contract event 0xd8138f8a3f377c5259ca548e70e4c2de94f129f5a11036a15b69513cba2b426a.
//
// Solidity: event Claimed(address indexed trader, uint256 indexed amount)
func (_AirdropL1 *AirdropL1Filterer) ParseClaimed(log types.Log) (*AirdropL1Claimed, error) {
	event := new(AirdropL1Claimed)
	if err := _AirdropL1.contract.UnpackLog(event, "Claimed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AirdropL1WithdrawToIterator is returned from FilterWithdrawTo and is used to iterate over the raw logs and unpacked data for WithdrawTo events raised by the AirdropL1 contract.
type AirdropL1WithdrawToIterator struct {
	Event *AirdropL1WithdrawTo // Event containing the contract specifics and raw log

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
func (it *AirdropL1WithdrawToIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AirdropL1WithdrawTo)
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
		it.Event = new(AirdropL1WithdrawTo)
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
func (it *AirdropL1WithdrawToIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AirdropL1WithdrawToIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AirdropL1WithdrawTo represents a WithdrawTo event raised by the AirdropL1 contract.
type AirdropL1WithdrawTo struct {
	To     common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterWithdrawTo is a free log retrieval operation binding the contract event 0x47096d7b247e809edf18e9bccfcb92f2af426ce8e6b40c923e65cb1b8394cef7.
//
// Solidity: event WithdrawTo(address indexed to, uint256 indexed amount)
func (_AirdropL1 *AirdropL1Filterer) FilterWithdrawTo(opts *bind.FilterOpts, to []common.Address, amount []*big.Int) (*AirdropL1WithdrawToIterator, error) {

	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var amountRule []interface{}
	for _, amountItem := range amount {
		amountRule = append(amountRule, amountItem)
	}

	logs, sub, err := _AirdropL1.contract.FilterLogs(opts, "WithdrawTo", toRule, amountRule)
	if err != nil {
		return nil, err
	}
	return &AirdropL1WithdrawToIterator{contract: _AirdropL1.contract, event: "WithdrawTo", logs: logs, sub: sub}, nil
}

// WatchWithdrawTo is a free log subscription operation binding the contract event 0x47096d7b247e809edf18e9bccfcb92f2af426ce8e6b40c923e65cb1b8394cef7.
//
// Solidity: event WithdrawTo(address indexed to, uint256 indexed amount)
func (_AirdropL1 *AirdropL1Filterer) WatchWithdrawTo(opts *bind.WatchOpts, sink chan<- *AirdropL1WithdrawTo, to []common.Address, amount []*big.Int) (event.Subscription, error) {

	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var amountRule []interface{}
	for _, amountItem := range amount {
		amountRule = append(amountRule, amountItem)
	}

	logs, sub, err := _AirdropL1.contract.WatchLogs(opts, "WithdrawTo", toRule, amountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AirdropL1WithdrawTo)
				if err := _AirdropL1.contract.UnpackLog(event, "WithdrawTo", log); err != nil {
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

// ParseWithdrawTo is a log parse operation binding the contract event 0x47096d7b247e809edf18e9bccfcb92f2af426ce8e6b40c923e65cb1b8394cef7.
//
// Solidity: event WithdrawTo(address indexed to, uint256 indexed amount)
func (_AirdropL1 *AirdropL1Filterer) ParseWithdrawTo(log types.Log) (*AirdropL1WithdrawTo, error) {
	event := new(AirdropL1WithdrawTo)
	if err := _AirdropL1.contract.UnpackLog(event, "WithdrawTo", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
