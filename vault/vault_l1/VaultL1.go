// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package vault_l1

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

// VaultL1MetaData contains all meta data concerning the VaultL1 contract.
var VaultL1MetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"trader\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Stake\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"trader\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"shares\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Unstake\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"signer\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"role\",\"type\":\"uint256\"}],\"name\":\"isValidSigner\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// VaultL1ABI is the input ABI used to generate the binding from.
// Deprecated: Use VaultL1MetaData.ABI instead.
var VaultL1ABI = VaultL1MetaData.ABI

// VaultL1 is an auto generated Go binding around an Ethereum contract.
type VaultL1 struct {
	VaultL1Caller     // Read-only binding to the contract
	VaultL1Transactor // Write-only binding to the contract
	VaultL1Filterer   // Log filterer for contract events
}

// VaultL1Caller is an auto generated read-only Go binding around an Ethereum contract.
type VaultL1Caller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// VaultL1Transactor is an auto generated write-only Go binding around an Ethereum contract.
type VaultL1Transactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// VaultL1Filterer is an auto generated log filtering Go binding around an Ethereum contract events.
type VaultL1Filterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// VaultL1Session is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type VaultL1Session struct {
	Contract     *VaultL1          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// VaultL1CallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type VaultL1CallerSession struct {
	Contract *VaultL1Caller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// VaultL1TransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type VaultL1TransactorSession struct {
	Contract     *VaultL1Transactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// VaultL1Raw is an auto generated low-level Go binding around an Ethereum contract.
type VaultL1Raw struct {
	Contract *VaultL1 // Generic contract binding to access the raw methods on
}

// VaultL1CallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type VaultL1CallerRaw struct {
	Contract *VaultL1Caller // Generic read-only contract binding to access the raw methods on
}

// VaultL1TransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type VaultL1TransactorRaw struct {
	Contract *VaultL1Transactor // Generic write-only contract binding to access the raw methods on
}

// NewVaultL1 creates a new instance of VaultL1, bound to a specific deployed contract.
func NewVaultL1(address common.Address, backend bind.ContractBackend) (*VaultL1, error) {
	contract, err := bindVaultL1(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &VaultL1{VaultL1Caller: VaultL1Caller{contract: contract}, VaultL1Transactor: VaultL1Transactor{contract: contract}, VaultL1Filterer: VaultL1Filterer{contract: contract}}, nil
}

// NewVaultL1Caller creates a new read-only instance of VaultL1, bound to a specific deployed contract.
func NewVaultL1Caller(address common.Address, caller bind.ContractCaller) (*VaultL1Caller, error) {
	contract, err := bindVaultL1(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &VaultL1Caller{contract: contract}, nil
}

// NewVaultL1Transactor creates a new write-only instance of VaultL1, bound to a specific deployed contract.
func NewVaultL1Transactor(address common.Address, transactor bind.ContractTransactor) (*VaultL1Transactor, error) {
	contract, err := bindVaultL1(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &VaultL1Transactor{contract: contract}, nil
}

// NewVaultL1Filterer creates a new log filterer instance of VaultL1, bound to a specific deployed contract.
func NewVaultL1Filterer(address common.Address, filterer bind.ContractFilterer) (*VaultL1Filterer, error) {
	contract, err := bindVaultL1(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &VaultL1Filterer{contract: contract}, nil
}

// bindVaultL1 binds a generic wrapper to an already deployed contract.
func bindVaultL1(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(VaultL1ABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_VaultL1 *VaultL1Raw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _VaultL1.Contract.VaultL1Caller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_VaultL1 *VaultL1Raw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _VaultL1.Contract.VaultL1Transactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_VaultL1 *VaultL1Raw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _VaultL1.Contract.VaultL1Transactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_VaultL1 *VaultL1CallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _VaultL1.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_VaultL1 *VaultL1TransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _VaultL1.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_VaultL1 *VaultL1TransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _VaultL1.Contract.contract.Transact(opts, method, params...)
}

// IsValidSigner is a free data retrieval call binding the contract method 0x53635d76.
//
// Solidity: function isValidSigner(address signer, uint256 role) view returns(bool)
func (_VaultL1 *VaultL1Caller) IsValidSigner(opts *bind.CallOpts, signer common.Address, role *big.Int) (bool, error) {
	var out []interface{}
	err := _VaultL1.contract.Call(opts, &out, "isValidSigner", signer, role)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsValidSigner is a free data retrieval call binding the contract method 0x53635d76.
//
// Solidity: function isValidSigner(address signer, uint256 role) view returns(bool)
func (_VaultL1 *VaultL1Session) IsValidSigner(signer common.Address, role *big.Int) (bool, error) {
	return _VaultL1.Contract.IsValidSigner(&_VaultL1.CallOpts, signer, role)
}

// IsValidSigner is a free data retrieval call binding the contract method 0x53635d76.
//
// Solidity: function isValidSigner(address signer, uint256 role) view returns(bool)
func (_VaultL1 *VaultL1CallerSession) IsValidSigner(signer common.Address, role *big.Int) (bool, error) {
	return _VaultL1.Contract.IsValidSigner(&_VaultL1.CallOpts, signer, role)
}

// VaultL1StakeIterator is returned from FilterStake and is used to iterate over the raw logs and unpacked data for Stake events raised by the VaultL1 contract.
type VaultL1StakeIterator struct {
	Event *VaultL1Stake // Event containing the contract specifics and raw log

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
func (it *VaultL1StakeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(VaultL1Stake)
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
		it.Event = new(VaultL1Stake)
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
func (it *VaultL1StakeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *VaultL1StakeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// VaultL1Stake represents a Stake event raised by the VaultL1 contract.
type VaultL1Stake struct {
	Id     *big.Int
	Trader common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterStake is a free log retrieval operation binding the contract event 0x02567b2553aeb44e4ddd5d68462774dc3de158cb0f2c2da1740e729b22086aff.
//
// Solidity: event Stake(uint256 indexed id, address indexed trader, uint256 amount)
func (_VaultL1 *VaultL1Filterer) FilterStake(opts *bind.FilterOpts, id []*big.Int, trader []common.Address) (*VaultL1StakeIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var traderRule []interface{}
	for _, traderItem := range trader {
		traderRule = append(traderRule, traderItem)
	}

	logs, sub, err := _VaultL1.contract.FilterLogs(opts, "Stake", idRule, traderRule)
	if err != nil {
		return nil, err
	}
	return &VaultL1StakeIterator{contract: _VaultL1.contract, event: "Stake", logs: logs, sub: sub}, nil
}

// WatchStake is a free log subscription operation binding the contract event 0x02567b2553aeb44e4ddd5d68462774dc3de158cb0f2c2da1740e729b22086aff.
//
// Solidity: event Stake(uint256 indexed id, address indexed trader, uint256 amount)
func (_VaultL1 *VaultL1Filterer) WatchStake(opts *bind.WatchOpts, sink chan<- *VaultL1Stake, id []*big.Int, trader []common.Address) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var traderRule []interface{}
	for _, traderItem := range trader {
		traderRule = append(traderRule, traderItem)
	}

	logs, sub, err := _VaultL1.contract.WatchLogs(opts, "Stake", idRule, traderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(VaultL1Stake)
				if err := _VaultL1.contract.UnpackLog(event, "Stake", log); err != nil {
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

// ParseStake is a log parse operation binding the contract event 0x02567b2553aeb44e4ddd5d68462774dc3de158cb0f2c2da1740e729b22086aff.
//
// Solidity: event Stake(uint256 indexed id, address indexed trader, uint256 amount)
func (_VaultL1 *VaultL1Filterer) ParseStake(log types.Log) (*VaultL1Stake, error) {
	event := new(VaultL1Stake)
	if err := _VaultL1.contract.UnpackLog(event, "Stake", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// VaultL1UnstakeIterator is returned from FilterUnstake and is used to iterate over the raw logs and unpacked data for Unstake events raised by the VaultL1 contract.
type VaultL1UnstakeIterator struct {
	Event *VaultL1Unstake // Event containing the contract specifics and raw log

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
func (it *VaultL1UnstakeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(VaultL1Unstake)
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
		it.Event = new(VaultL1Unstake)
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
func (it *VaultL1UnstakeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *VaultL1UnstakeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// VaultL1Unstake represents a Unstake event raised by the VaultL1 contract.
type VaultL1Unstake struct {
	Id     *big.Int
	Trader common.Address
	Shares *big.Int
	Value  *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterUnstake is a free log retrieval operation binding the contract event 0xc1e00202ee2c06861d326fc6374026b751863ff64218ccbaa38c3e603a8e72c2.
//
// Solidity: event Unstake(uint256 indexed id, address indexed trader, uint256 shares, uint256 value)
func (_VaultL1 *VaultL1Filterer) FilterUnstake(opts *bind.FilterOpts, id []*big.Int, trader []common.Address) (*VaultL1UnstakeIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var traderRule []interface{}
	for _, traderItem := range trader {
		traderRule = append(traderRule, traderItem)
	}

	logs, sub, err := _VaultL1.contract.FilterLogs(opts, "Unstake", idRule, traderRule)
	if err != nil {
		return nil, err
	}
	return &VaultL1UnstakeIterator{contract: _VaultL1.contract, event: "Unstake", logs: logs, sub: sub}, nil
}

// WatchUnstake is a free log subscription operation binding the contract event 0xc1e00202ee2c06861d326fc6374026b751863ff64218ccbaa38c3e603a8e72c2.
//
// Solidity: event Unstake(uint256 indexed id, address indexed trader, uint256 shares, uint256 value)
func (_VaultL1 *VaultL1Filterer) WatchUnstake(opts *bind.WatchOpts, sink chan<- *VaultL1Unstake, id []*big.Int, trader []common.Address) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var traderRule []interface{}
	for _, traderItem := range trader {
		traderRule = append(traderRule, traderItem)
	}

	logs, sub, err := _VaultL1.contract.WatchLogs(opts, "Unstake", idRule, traderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(VaultL1Unstake)
				if err := _VaultL1.contract.UnpackLog(event, "Unstake", log); err != nil {
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

// ParseUnstake is a log parse operation binding the contract event 0xc1e00202ee2c06861d326fc6374026b751863ff64218ccbaa38c3e603a8e72c2.
//
// Solidity: event Unstake(uint256 indexed id, address indexed trader, uint256 shares, uint256 value)
func (_VaultL1 *VaultL1Filterer) ParseUnstake(log types.Log) (*VaultL1Unstake, error) {
	event := new(VaultL1Unstake)
	if err := _VaultL1.contract.UnpackLog(event, "Unstake", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
