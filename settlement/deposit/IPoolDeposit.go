// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package deposit

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
)

// Contribution is an auto generated low-level Go binding around an user-defined struct.
type Contribution struct {
	Contributor common.Address
	Amount      *big.Int
}

// DepositMetaData contains all meta data concerning the Deposit contract.
var DepositMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"trader\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"poolId\",\"type\":\"uint256\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"PooledDeposit\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"contributor\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"individualDeposit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"contributor\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structContribution[]\",\"name\":\"contributions\",\"type\":\"tuple[]\"}],\"name\":\"pooledDeposit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// DepositABI is the input ABI used to generate the binding from.
// Deprecated: Use DepositMetaData.ABI instead.
var DepositABI = DepositMetaData.ABI

// Deposit is an auto generated Go binding around an Ethereum contract.
type Deposit struct {
	DepositCaller     // Read-only binding to the contract
	DepositTransactor // Write-only binding to the contract
	DepositFilterer   // Log filterer for contract events
}

// DepositCaller is an auto generated read-only Go binding around an Ethereum contract.
type DepositCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DepositTransactor is an auto generated write-only Go binding around an Ethereum contract.
type DepositTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DepositFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type DepositFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DepositSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type DepositSession struct {
	Contract     *Deposit          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// DepositCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type DepositCallerSession struct {
	Contract *DepositCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// DepositTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type DepositTransactorSession struct {
	Contract     *DepositTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// DepositRaw is an auto generated low-level Go binding around an Ethereum contract.
type DepositRaw struct {
	Contract *Deposit // Generic contract binding to access the raw methods on
}

// DepositCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type DepositCallerRaw struct {
	Contract *DepositCaller // Generic read-only contract binding to access the raw methods on
}

// DepositTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type DepositTransactorRaw struct {
	Contract *DepositTransactor // Generic write-only contract binding to access the raw methods on
}

// NewDeposit creates a new instance of Deposit, bound to a specific deployed contract.
func NewDeposit(address common.Address, backend bind.ContractBackend) (*Deposit, error) {
	contract, err := bindDeposit(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Deposit{DepositCaller: DepositCaller{contract: contract}, DepositTransactor: DepositTransactor{contract: contract}, DepositFilterer: DepositFilterer{contract: contract}}, nil
}

// NewDepositCaller creates a new read-only instance of Deposit, bound to a specific deployed contract.
func NewDepositCaller(address common.Address, caller bind.ContractCaller) (*DepositCaller, error) {
	contract, err := bindDeposit(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &DepositCaller{contract: contract}, nil
}

// NewDepositTransactor creates a new write-only instance of Deposit, bound to a specific deployed contract.
func NewDepositTransactor(address common.Address, transactor bind.ContractTransactor) (*DepositTransactor, error) {
	contract, err := bindDeposit(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &DepositTransactor{contract: contract}, nil
}

// NewDepositFilterer creates a new log filterer instance of Deposit, bound to a specific deployed contract.
func NewDepositFilterer(address common.Address, filterer bind.ContractFilterer) (*DepositFilterer, error) {
	contract, err := bindDeposit(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &DepositFilterer{contract: contract}, nil
}

// bindDeposit binds a generic wrapper to an already deployed contract.
func bindDeposit(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(DepositABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Deposit *DepositRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Deposit.Contract.DepositCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Deposit *DepositRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Deposit.Contract.DepositTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Deposit *DepositRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Deposit.Contract.DepositTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Deposit *DepositCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Deposit.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Deposit *DepositTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Deposit.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Deposit *DepositTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Deposit.Contract.contract.Transact(opts, method, params...)
}

// IndividualDeposit is a paid mutator transaction binding the contract method 0xee74567c.
//
// Solidity: function individualDeposit(address contributor, uint256 amount) returns()
func (_Deposit *DepositTransactor) IndividualDeposit(opts *bind.TransactOpts, contributor common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Deposit.contract.Transact(opts, "individualDeposit", contributor, amount)
}

// IndividualDeposit is a paid mutator transaction binding the contract method 0xee74567c.
//
// Solidity: function individualDeposit(address contributor, uint256 amount) returns()
func (_Deposit *DepositSession) IndividualDeposit(contributor common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Deposit.Contract.IndividualDeposit(&_Deposit.TransactOpts, contributor, amount)
}

// IndividualDeposit is a paid mutator transaction binding the contract method 0xee74567c.
//
// Solidity: function individualDeposit(address contributor, uint256 amount) returns()
func (_Deposit *DepositTransactorSession) IndividualDeposit(contributor common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Deposit.Contract.IndividualDeposit(&_Deposit.TransactOpts, contributor, amount)
}

// PooledDeposit is a paid mutator transaction binding the contract method 0x57a7ce82.
//
// Solidity: function pooledDeposit((address,uint256)[] contributions) returns()
func (_Deposit *DepositTransactor) PooledDeposit(opts *bind.TransactOpts, contributions []Contribution) (*types.Transaction, error) {
	return _Deposit.contract.Transact(opts, "pooledDeposit", contributions)
}

// PooledDeposit is a paid mutator transaction binding the contract method 0x57a7ce82.
//
// Solidity: function pooledDeposit((address,uint256)[] contributions) returns()
func (_Deposit *DepositSession) PooledDeposit(contributions []Contribution) (*types.Transaction, error) {
	return _Deposit.Contract.PooledDeposit(&_Deposit.TransactOpts, contributions)
}

// PooledDeposit is a paid mutator transaction binding the contract method 0x57a7ce82.
//
// Solidity: function pooledDeposit((address,uint256)[] contributions) returns()
func (_Deposit *DepositTransactorSession) PooledDeposit(contributions []Contribution) (*types.Transaction, error) {
	return _Deposit.Contract.PooledDeposit(&_Deposit.TransactOpts, contributions)
}

// DepositDepositIterator is returned from FilterDeposit and is used to iterate over the raw logs and unpacked data for Deposit events raised by the Deposit contract.
type DepositDepositIterator struct {
	Event *DepositDeposit // Event containing the contract specifics and raw log

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
func (it *DepositDepositIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DepositDeposit)
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
		it.Event = new(DepositDeposit)
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
func (it *DepositDepositIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DepositDepositIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DepositDeposit represents a Deposit event raised by the Deposit contract.
type DepositDeposit struct {
	Id     *big.Int
	Trader common.Address
	Amount *big.Int
	PoolId *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDeposit is a free log retrieval operation binding the contract event 0xd36a2f67d06d285786f61a32b052b9ace6b0b7abef5177b54358abdc83a0b69b.
//
// Solidity: event Deposit(uint256 indexed id, address indexed trader, uint256 amount, uint256 indexed poolId)
func (_Deposit *DepositFilterer) FilterDeposit(opts *bind.FilterOpts, id []*big.Int, trader []common.Address, poolId []*big.Int) (*DepositDepositIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var traderRule []interface{}
	for _, traderItem := range trader {
		traderRule = append(traderRule, traderItem)
	}

	var poolIdRule []interface{}
	for _, poolIdItem := range poolId {
		poolIdRule = append(poolIdRule, poolIdItem)
	}

	logs, sub, err := _Deposit.contract.FilterLogs(opts, "Deposit", idRule, traderRule, poolIdRule)
	if err != nil {
		return nil, err
	}
	return &DepositDepositIterator{contract: _Deposit.contract, event: "Deposit", logs: logs, sub: sub}, nil
}

// WatchDeposit is a free log subscription operation binding the contract event 0xd36a2f67d06d285786f61a32b052b9ace6b0b7abef5177b54358abdc83a0b69b.
//
// Solidity: event Deposit(uint256 indexed id, address indexed trader, uint256 amount, uint256 indexed poolId)
func (_Deposit *DepositFilterer) WatchDeposit(opts *bind.WatchOpts, sink chan<- *DepositDeposit, id []*big.Int, trader []common.Address, poolId []*big.Int) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var traderRule []interface{}
	for _, traderItem := range trader {
		traderRule = append(traderRule, traderItem)
	}

	var poolIdRule []interface{}
	for _, poolIdItem := range poolId {
		poolIdRule = append(poolIdRule, poolIdItem)
	}

	logs, sub, err := _Deposit.contract.WatchLogs(opts, "Deposit", idRule, traderRule, poolIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DepositDeposit)
				if err := _Deposit.contract.UnpackLog(event, "Deposit", log); err != nil {
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

// ParseDeposit is a log parse operation binding the contract event 0xd36a2f67d06d285786f61a32b052b9ace6b0b7abef5177b54358abdc83a0b69b.
//
// Solidity: event Deposit(uint256 indexed id, address indexed trader, uint256 amount, uint256 indexed poolId)
func (_Deposit *DepositFilterer) ParseDeposit(log types.Log) (*DepositDeposit, error) {
	event := new(DepositDeposit)
	if err := _Deposit.contract.UnpackLog(event, "Deposit", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DepositPooledDepositIterator is returned from FilterPooledDeposit and is used to iterate over the raw logs and unpacked data for PooledDeposit events raised by the Deposit contract.
type DepositPooledDepositIterator struct {
	Event *DepositPooledDeposit // Event containing the contract specifics and raw log

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
func (it *DepositPooledDepositIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DepositPooledDeposit)
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
		it.Event = new(DepositPooledDeposit)
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
func (it *DepositPooledDepositIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DepositPooledDepositIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DepositPooledDeposit represents a PooledDeposit event raised by the Deposit contract.
type DepositPooledDeposit struct {
	Id     *big.Int
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterPooledDeposit is a free log retrieval operation binding the contract event 0x667ed75d7a503a5dc36e11fbc72336a5a2a2b95577a97ecd2aac380f1ac1b640.
//
// Solidity: event PooledDeposit(uint256 indexed id, uint256 amount)
func (_Deposit *DepositFilterer) FilterPooledDeposit(opts *bind.FilterOpts, id []*big.Int) (*DepositPooledDepositIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}

	logs, sub, err := _Deposit.contract.FilterLogs(opts, "PooledDeposit", idRule)
	if err != nil {
		return nil, err
	}
	return &DepositPooledDepositIterator{contract: _Deposit.contract, event: "PooledDeposit", logs: logs, sub: sub}, nil
}

// WatchPooledDeposit is a free log subscription operation binding the contract event 0x667ed75d7a503a5dc36e11fbc72336a5a2a2b95577a97ecd2aac380f1ac1b640.
//
// Solidity: event PooledDeposit(uint256 indexed id, uint256 amount)
func (_Deposit *DepositFilterer) WatchPooledDeposit(opts *bind.WatchOpts, sink chan<- *DepositPooledDeposit, id []*big.Int) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}

	logs, sub, err := _Deposit.contract.WatchLogs(opts, "PooledDeposit", idRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DepositPooledDeposit)
				if err := _Deposit.contract.UnpackLog(event, "PooledDeposit", log); err != nil {
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

// ParsePooledDeposit is a log parse operation binding the contract event 0x667ed75d7a503a5dc36e11fbc72336a5a2a2b95577a97ecd2aac380f1ac1b640.
//
// Solidity: event PooledDeposit(uint256 indexed id, uint256 amount)
func (_Deposit *DepositFilterer) ParsePooledDeposit(log types.Log) (*DepositPooledDeposit, error) {
	event := new(DepositPooledDeposit)
	if err := _Deposit.contract.UnpackLog(event, "PooledDeposit", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
