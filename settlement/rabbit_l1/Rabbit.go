// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package rabbit_l1

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

// RabbitL1MetaData contains all meta data concerning the RabbitL1 contract.
var RabbitL1MetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_signer\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_paymentToken\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"trader\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"fromAddress\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256[]\",\"name\":\"payload\",\"type\":\"uint256[]\"}],\"name\":\"MsgNotFound\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"messageType\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256[]\",\"name\":\"payload\",\"type\":\"uint256[]\"}],\"name\":\"UnknownReceipt\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"trader\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Withdraw\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"WithdrawTo\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"trader\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"WithdrawalReceipt\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"new_signer\",\"type\":\"address\"}],\"name\":\"changeSigner\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"deposit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"external_signer\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"paymentToken\",\"outputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"processedWithdrawals\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_paymentToken\",\"type\":\"address\"}],\"name\":\"setPaymentToken\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"trader\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"withdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"withdrawTokensTo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// RabbitL1ABI is the input ABI used to generate the binding from.
// Deprecated: Use RabbitL1MetaData.ABI instead.
var RabbitL1ABI = RabbitL1MetaData.ABI

// RabbitL1 is an auto generated Go binding around an Ethereum contract.
type RabbitL1 struct {
	RabbitL1Caller     // Read-only binding to the contract
	RabbitL1Transactor // Write-only binding to the contract
	RabbitL1Filterer   // Log filterer for contract events
}

// RabbitL1Caller is an auto generated read-only Go binding around an Ethereum contract.
type RabbitL1Caller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RabbitL1Transactor is an auto generated write-only Go binding around an Ethereum contract.
type RabbitL1Transactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RabbitL1Filterer is an auto generated log filtering Go binding around an Ethereum contract events.
type RabbitL1Filterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RabbitL1Session is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type RabbitL1Session struct {
	Contract     *RabbitL1         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// RabbitL1CallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type RabbitL1CallerSession struct {
	Contract *RabbitL1Caller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// RabbitL1TransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type RabbitL1TransactorSession struct {
	Contract     *RabbitL1Transactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// RabbitL1Raw is an auto generated low-level Go binding around an Ethereum contract.
type RabbitL1Raw struct {
	Contract *RabbitL1 // Generic contract binding to access the raw methods on
}

// RabbitL1CallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type RabbitL1CallerRaw struct {
	Contract *RabbitL1Caller // Generic read-only contract binding to access the raw methods on
}

// RabbitL1TransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type RabbitL1TransactorRaw struct {
	Contract *RabbitL1Transactor // Generic write-only contract binding to access the raw methods on
}

// NewRabbitL1 creates a new instance of RabbitL1, bound to a specific deployed contract.
func NewRabbitL1(address common.Address, backend bind.ContractBackend) (*RabbitL1, error) {
	contract, err := bindRabbitL1(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &RabbitL1{RabbitL1Caller: RabbitL1Caller{contract: contract}, RabbitL1Transactor: RabbitL1Transactor{contract: contract}, RabbitL1Filterer: RabbitL1Filterer{contract: contract}}, nil
}

// NewRabbitL1Caller creates a new read-only instance of RabbitL1, bound to a specific deployed contract.
func NewRabbitL1Caller(address common.Address, caller bind.ContractCaller) (*RabbitL1Caller, error) {
	contract, err := bindRabbitL1(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &RabbitL1Caller{contract: contract}, nil
}

// NewRabbitL1Transactor creates a new write-only instance of RabbitL1, bound to a specific deployed contract.
func NewRabbitL1Transactor(address common.Address, transactor bind.ContractTransactor) (*RabbitL1Transactor, error) {
	contract, err := bindRabbitL1(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &RabbitL1Transactor{contract: contract}, nil
}

// NewRabbitL1Filterer creates a new log filterer instance of RabbitL1, bound to a specific deployed contract.
func NewRabbitL1Filterer(address common.Address, filterer bind.ContractFilterer) (*RabbitL1Filterer, error) {
	contract, err := bindRabbitL1(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &RabbitL1Filterer{contract: contract}, nil
}

// bindRabbitL1 binds a generic wrapper to an already deployed contract.
func bindRabbitL1(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(RabbitL1ABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_RabbitL1 *RabbitL1Raw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _RabbitL1.Contract.RabbitL1Caller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_RabbitL1 *RabbitL1Raw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _RabbitL1.Contract.RabbitL1Transactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_RabbitL1 *RabbitL1Raw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _RabbitL1.Contract.RabbitL1Transactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_RabbitL1 *RabbitL1CallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _RabbitL1.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_RabbitL1 *RabbitL1TransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _RabbitL1.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_RabbitL1 *RabbitL1TransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _RabbitL1.Contract.contract.Transact(opts, method, params...)
}

// ExternalSigner is a free data retrieval call binding the contract method 0x0008cecb.
//
// Solidity: function external_signer() view returns(address)
func (_RabbitL1 *RabbitL1Caller) ExternalSigner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _RabbitL1.contract.Call(opts, &out, "external_signer")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ExternalSigner is a free data retrieval call binding the contract method 0x0008cecb.
//
// Solidity: function external_signer() view returns(address)
func (_RabbitL1 *RabbitL1Session) ExternalSigner() (common.Address, error) {
	return _RabbitL1.Contract.ExternalSigner(&_RabbitL1.CallOpts)
}

// ExternalSigner is a free data retrieval call binding the contract method 0x0008cecb.
//
// Solidity: function external_signer() view returns(address)
func (_RabbitL1 *RabbitL1CallerSession) ExternalSigner() (common.Address, error) {
	return _RabbitL1.Contract.ExternalSigner(&_RabbitL1.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_RabbitL1 *RabbitL1Caller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _RabbitL1.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_RabbitL1 *RabbitL1Session) Owner() (common.Address, error) {
	return _RabbitL1.Contract.Owner(&_RabbitL1.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_RabbitL1 *RabbitL1CallerSession) Owner() (common.Address, error) {
	return _RabbitL1.Contract.Owner(&_RabbitL1.CallOpts)
}

// PaymentToken is a free data retrieval call binding the contract method 0x3013ce29.
//
// Solidity: function paymentToken() view returns(address)
func (_RabbitL1 *RabbitL1Caller) PaymentToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _RabbitL1.contract.Call(opts, &out, "paymentToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PaymentToken is a free data retrieval call binding the contract method 0x3013ce29.
//
// Solidity: function paymentToken() view returns(address)
func (_RabbitL1 *RabbitL1Session) PaymentToken() (common.Address, error) {
	return _RabbitL1.Contract.PaymentToken(&_RabbitL1.CallOpts)
}

// PaymentToken is a free data retrieval call binding the contract method 0x3013ce29.
//
// Solidity: function paymentToken() view returns(address)
func (_RabbitL1 *RabbitL1CallerSession) PaymentToken() (common.Address, error) {
	return _RabbitL1.Contract.PaymentToken(&_RabbitL1.CallOpts)
}

// ProcessedWithdrawals is a free data retrieval call binding the contract method 0xdde4d950.
//
// Solidity: function processedWithdrawals(uint256 ) view returns(bool)
func (_RabbitL1 *RabbitL1Caller) ProcessedWithdrawals(opts *bind.CallOpts, arg0 *big.Int) (bool, error) {
	var out []interface{}
	err := _RabbitL1.contract.Call(opts, &out, "processedWithdrawals", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// ProcessedWithdrawals is a free data retrieval call binding the contract method 0xdde4d950.
//
// Solidity: function processedWithdrawals(uint256 ) view returns(bool)
func (_RabbitL1 *RabbitL1Session) ProcessedWithdrawals(arg0 *big.Int) (bool, error) {
	return _RabbitL1.Contract.ProcessedWithdrawals(&_RabbitL1.CallOpts, arg0)
}

// ProcessedWithdrawals is a free data retrieval call binding the contract method 0xdde4d950.
//
// Solidity: function processedWithdrawals(uint256 ) view returns(bool)
func (_RabbitL1 *RabbitL1CallerSession) ProcessedWithdrawals(arg0 *big.Int) (bool, error) {
	return _RabbitL1.Contract.ProcessedWithdrawals(&_RabbitL1.CallOpts, arg0)
}

// ChangeSigner is a paid mutator transaction binding the contract method 0xaad2b723.
//
// Solidity: function changeSigner(address new_signer) returns()
func (_RabbitL1 *RabbitL1Transactor) ChangeSigner(opts *bind.TransactOpts, new_signer common.Address) (*types.Transaction, error) {
	return _RabbitL1.contract.Transact(opts, "changeSigner", new_signer)
}

// ChangeSigner is a paid mutator transaction binding the contract method 0xaad2b723.
//
// Solidity: function changeSigner(address new_signer) returns()
func (_RabbitL1 *RabbitL1Session) ChangeSigner(new_signer common.Address) (*types.Transaction, error) {
	return _RabbitL1.Contract.ChangeSigner(&_RabbitL1.TransactOpts, new_signer)
}

// ChangeSigner is a paid mutator transaction binding the contract method 0xaad2b723.
//
// Solidity: function changeSigner(address new_signer) returns()
func (_RabbitL1 *RabbitL1TransactorSession) ChangeSigner(new_signer common.Address) (*types.Transaction, error) {
	return _RabbitL1.Contract.ChangeSigner(&_RabbitL1.TransactOpts, new_signer)
}

// Deposit is a paid mutator transaction binding the contract method 0xb6b55f25.
//
// Solidity: function deposit(uint256 amount) returns()
func (_RabbitL1 *RabbitL1Transactor) Deposit(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _RabbitL1.contract.Transact(opts, "deposit", amount)
}

// Deposit is a paid mutator transaction binding the contract method 0xb6b55f25.
//
// Solidity: function deposit(uint256 amount) returns()
func (_RabbitL1 *RabbitL1Session) Deposit(amount *big.Int) (*types.Transaction, error) {
	return _RabbitL1.Contract.Deposit(&_RabbitL1.TransactOpts, amount)
}

// Deposit is a paid mutator transaction binding the contract method 0xb6b55f25.
//
// Solidity: function deposit(uint256 amount) returns()
func (_RabbitL1 *RabbitL1TransactorSession) Deposit(amount *big.Int) (*types.Transaction, error) {
	return _RabbitL1.Contract.Deposit(&_RabbitL1.TransactOpts, amount)
}

// SetPaymentToken is a paid mutator transaction binding the contract method 0x6a326ab1.
//
// Solidity: function setPaymentToken(address _paymentToken) returns()
func (_RabbitL1 *RabbitL1Transactor) SetPaymentToken(opts *bind.TransactOpts, _paymentToken common.Address) (*types.Transaction, error) {
	return _RabbitL1.contract.Transact(opts, "setPaymentToken", _paymentToken)
}

// SetPaymentToken is a paid mutator transaction binding the contract method 0x6a326ab1.
//
// Solidity: function setPaymentToken(address _paymentToken) returns()
func (_RabbitL1 *RabbitL1Session) SetPaymentToken(_paymentToken common.Address) (*types.Transaction, error) {
	return _RabbitL1.Contract.SetPaymentToken(&_RabbitL1.TransactOpts, _paymentToken)
}

// SetPaymentToken is a paid mutator transaction binding the contract method 0x6a326ab1.
//
// Solidity: function setPaymentToken(address _paymentToken) returns()
func (_RabbitL1 *RabbitL1TransactorSession) SetPaymentToken(_paymentToken common.Address) (*types.Transaction, error) {
	return _RabbitL1.Contract.SetPaymentToken(&_RabbitL1.TransactOpts, _paymentToken)
}

// Withdraw is a paid mutator transaction binding the contract method 0x61c8e739.
//
// Solidity: function withdraw(uint256 id, address trader, uint256 amount, uint8 v, bytes32 r, bytes32 s) returns()
func (_RabbitL1 *RabbitL1Transactor) Withdraw(opts *bind.TransactOpts, id *big.Int, trader common.Address, amount *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _RabbitL1.contract.Transact(opts, "withdraw", id, trader, amount, v, r, s)
}

// Withdraw is a paid mutator transaction binding the contract method 0x61c8e739.
//
// Solidity: function withdraw(uint256 id, address trader, uint256 amount, uint8 v, bytes32 r, bytes32 s) returns()
func (_RabbitL1 *RabbitL1Session) Withdraw(id *big.Int, trader common.Address, amount *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _RabbitL1.Contract.Withdraw(&_RabbitL1.TransactOpts, id, trader, amount, v, r, s)
}

// Withdraw is a paid mutator transaction binding the contract method 0x61c8e739.
//
// Solidity: function withdraw(uint256 id, address trader, uint256 amount, uint8 v, bytes32 r, bytes32 s) returns()
func (_RabbitL1 *RabbitL1TransactorSession) Withdraw(id *big.Int, trader common.Address, amount *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _RabbitL1.Contract.Withdraw(&_RabbitL1.TransactOpts, id, trader, amount, v, r, s)
}

// WithdrawTokensTo is a paid mutator transaction binding the contract method 0xb743e722.
//
// Solidity: function withdrawTokensTo(uint256 amount, address to) returns()
func (_RabbitL1 *RabbitL1Transactor) WithdrawTokensTo(opts *bind.TransactOpts, amount *big.Int, to common.Address) (*types.Transaction, error) {
	return _RabbitL1.contract.Transact(opts, "withdrawTokensTo", amount, to)
}

// WithdrawTokensTo is a paid mutator transaction binding the contract method 0xb743e722.
//
// Solidity: function withdrawTokensTo(uint256 amount, address to) returns()
func (_RabbitL1 *RabbitL1Session) WithdrawTokensTo(amount *big.Int, to common.Address) (*types.Transaction, error) {
	return _RabbitL1.Contract.WithdrawTokensTo(&_RabbitL1.TransactOpts, amount, to)
}

// WithdrawTokensTo is a paid mutator transaction binding the contract method 0xb743e722.
//
// Solidity: function withdrawTokensTo(uint256 amount, address to) returns()
func (_RabbitL1 *RabbitL1TransactorSession) WithdrawTokensTo(amount *big.Int, to common.Address) (*types.Transaction, error) {
	return _RabbitL1.Contract.WithdrawTokensTo(&_RabbitL1.TransactOpts, amount, to)
}

// RabbitL1DepositIterator is returned from FilterDeposit and is used to iterate over the raw logs and unpacked data for Deposit events raised by the RabbitL1 contract.
type RabbitL1DepositIterator struct {
	Event *RabbitL1Deposit // Event containing the contract specifics and raw log

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
func (it *RabbitL1DepositIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RabbitL1Deposit)
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
		it.Event = new(RabbitL1Deposit)
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
func (it *RabbitL1DepositIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RabbitL1DepositIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RabbitL1Deposit represents a Deposit event raised by the RabbitL1 contract.
type RabbitL1Deposit struct {
	Id     *big.Int
	Trader common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDeposit is a free log retrieval operation binding the contract event 0xeaa18152488ce5959073c9c79c88ca90b3d96c00de1f118cfaad664c3dab06b9.
//
// Solidity: event Deposit(uint256 indexed id, address indexed trader, uint256 amount)
func (_RabbitL1 *RabbitL1Filterer) FilterDeposit(opts *bind.FilterOpts, id []*big.Int, trader []common.Address) (*RabbitL1DepositIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var traderRule []interface{}
	for _, traderItem := range trader {
		traderRule = append(traderRule, traderItem)
	}

	logs, sub, err := _RabbitL1.contract.FilterLogs(opts, "Deposit", idRule, traderRule)
	if err != nil {
		return nil, err
	}
	return &RabbitL1DepositIterator{contract: _RabbitL1.contract, event: "Deposit", logs: logs, sub: sub}, nil
}

// WatchDeposit is a free log subscription operation binding the contract event 0xeaa18152488ce5959073c9c79c88ca90b3d96c00de1f118cfaad664c3dab06b9.
//
// Solidity: event Deposit(uint256 indexed id, address indexed trader, uint256 amount)
func (_RabbitL1 *RabbitL1Filterer) WatchDeposit(opts *bind.WatchOpts, sink chan<- *RabbitL1Deposit, id []*big.Int, trader []common.Address) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var traderRule []interface{}
	for _, traderItem := range trader {
		traderRule = append(traderRule, traderItem)
	}

	logs, sub, err := _RabbitL1.contract.WatchLogs(opts, "Deposit", idRule, traderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RabbitL1Deposit)
				if err := _RabbitL1.contract.UnpackLog(event, "Deposit", log); err != nil {
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

// ParseDeposit is a log parse operation binding the contract event 0xeaa18152488ce5959073c9c79c88ca90b3d96c00de1f118cfaad664c3dab06b9.
//
// Solidity: event Deposit(uint256 indexed id, address indexed trader, uint256 amount)
func (_RabbitL1 *RabbitL1Filterer) ParseDeposit(log types.Log) (*RabbitL1Deposit, error) {
	event := new(RabbitL1Deposit)
	if err := _RabbitL1.contract.UnpackLog(event, "Deposit", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RabbitL1MsgNotFoundIterator is returned from FilterMsgNotFound and is used to iterate over the raw logs and unpacked data for MsgNotFound events raised by the RabbitL1 contract.
type RabbitL1MsgNotFoundIterator struct {
	Event *RabbitL1MsgNotFound // Event containing the contract specifics and raw log

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
func (it *RabbitL1MsgNotFoundIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RabbitL1MsgNotFound)
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
		it.Event = new(RabbitL1MsgNotFound)
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
func (it *RabbitL1MsgNotFoundIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RabbitL1MsgNotFoundIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RabbitL1MsgNotFound represents a MsgNotFound event raised by the RabbitL1 contract.
type RabbitL1MsgNotFound struct {
	FromAddress *big.Int
	Payload     []*big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterMsgNotFound is a free log retrieval operation binding the contract event 0x7f9b4cb43ff44052de179ab10347ef5e3e1ac39ca47e82074aa4fe9eb8c49db2.
//
// Solidity: event MsgNotFound(uint256 indexed fromAddress, uint256[] payload)
func (_RabbitL1 *RabbitL1Filterer) FilterMsgNotFound(opts *bind.FilterOpts, fromAddress []*big.Int) (*RabbitL1MsgNotFoundIterator, error) {

	var fromAddressRule []interface{}
	for _, fromAddressItem := range fromAddress {
		fromAddressRule = append(fromAddressRule, fromAddressItem)
	}

	logs, sub, err := _RabbitL1.contract.FilterLogs(opts, "MsgNotFound", fromAddressRule)
	if err != nil {
		return nil, err
	}
	return &RabbitL1MsgNotFoundIterator{contract: _RabbitL1.contract, event: "MsgNotFound", logs: logs, sub: sub}, nil
}

// WatchMsgNotFound is a free log subscription operation binding the contract event 0x7f9b4cb43ff44052de179ab10347ef5e3e1ac39ca47e82074aa4fe9eb8c49db2.
//
// Solidity: event MsgNotFound(uint256 indexed fromAddress, uint256[] payload)
func (_RabbitL1 *RabbitL1Filterer) WatchMsgNotFound(opts *bind.WatchOpts, sink chan<- *RabbitL1MsgNotFound, fromAddress []*big.Int) (event.Subscription, error) {

	var fromAddressRule []interface{}
	for _, fromAddressItem := range fromAddress {
		fromAddressRule = append(fromAddressRule, fromAddressItem)
	}

	logs, sub, err := _RabbitL1.contract.WatchLogs(opts, "MsgNotFound", fromAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RabbitL1MsgNotFound)
				if err := _RabbitL1.contract.UnpackLog(event, "MsgNotFound", log); err != nil {
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

// ParseMsgNotFound is a log parse operation binding the contract event 0x7f9b4cb43ff44052de179ab10347ef5e3e1ac39ca47e82074aa4fe9eb8c49db2.
//
// Solidity: event MsgNotFound(uint256 indexed fromAddress, uint256[] payload)
func (_RabbitL1 *RabbitL1Filterer) ParseMsgNotFound(log types.Log) (*RabbitL1MsgNotFound, error) {
	event := new(RabbitL1MsgNotFound)
	if err := _RabbitL1.contract.UnpackLog(event, "MsgNotFound", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RabbitL1UnknownReceiptIterator is returned from FilterUnknownReceipt and is used to iterate over the raw logs and unpacked data for UnknownReceipt events raised by the RabbitL1 contract.
type RabbitL1UnknownReceiptIterator struct {
	Event *RabbitL1UnknownReceipt // Event containing the contract specifics and raw log

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
func (it *RabbitL1UnknownReceiptIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RabbitL1UnknownReceipt)
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
		it.Event = new(RabbitL1UnknownReceipt)
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
func (it *RabbitL1UnknownReceiptIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RabbitL1UnknownReceiptIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RabbitL1UnknownReceipt represents a UnknownReceipt event raised by the RabbitL1 contract.
type RabbitL1UnknownReceipt struct {
	MessageType *big.Int
	Payload     []*big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterUnknownReceipt is a free log retrieval operation binding the contract event 0xf7d5eeb592c98005a5953fe011acb6720929e2d93c3fcd96cb168836d54ffe8c.
//
// Solidity: event UnknownReceipt(uint256 indexed messageType, uint256[] payload)
func (_RabbitL1 *RabbitL1Filterer) FilterUnknownReceipt(opts *bind.FilterOpts, messageType []*big.Int) (*RabbitL1UnknownReceiptIterator, error) {

	var messageTypeRule []interface{}
	for _, messageTypeItem := range messageType {
		messageTypeRule = append(messageTypeRule, messageTypeItem)
	}

	logs, sub, err := _RabbitL1.contract.FilterLogs(opts, "UnknownReceipt", messageTypeRule)
	if err != nil {
		return nil, err
	}
	return &RabbitL1UnknownReceiptIterator{contract: _RabbitL1.contract, event: "UnknownReceipt", logs: logs, sub: sub}, nil
}

// WatchUnknownReceipt is a free log subscription operation binding the contract event 0xf7d5eeb592c98005a5953fe011acb6720929e2d93c3fcd96cb168836d54ffe8c.
//
// Solidity: event UnknownReceipt(uint256 indexed messageType, uint256[] payload)
func (_RabbitL1 *RabbitL1Filterer) WatchUnknownReceipt(opts *bind.WatchOpts, sink chan<- *RabbitL1UnknownReceipt, messageType []*big.Int) (event.Subscription, error) {

	var messageTypeRule []interface{}
	for _, messageTypeItem := range messageType {
		messageTypeRule = append(messageTypeRule, messageTypeItem)
	}

	logs, sub, err := _RabbitL1.contract.WatchLogs(opts, "UnknownReceipt", messageTypeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RabbitL1UnknownReceipt)
				if err := _RabbitL1.contract.UnpackLog(event, "UnknownReceipt", log); err != nil {
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

// ParseUnknownReceipt is a log parse operation binding the contract event 0xf7d5eeb592c98005a5953fe011acb6720929e2d93c3fcd96cb168836d54ffe8c.
//
// Solidity: event UnknownReceipt(uint256 indexed messageType, uint256[] payload)
func (_RabbitL1 *RabbitL1Filterer) ParseUnknownReceipt(log types.Log) (*RabbitL1UnknownReceipt, error) {
	event := new(RabbitL1UnknownReceipt)
	if err := _RabbitL1.contract.UnpackLog(event, "UnknownReceipt", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RabbitL1WithdrawIterator is returned from FilterWithdraw and is used to iterate over the raw logs and unpacked data for Withdraw events raised by the RabbitL1 contract.
type RabbitL1WithdrawIterator struct {
	Event *RabbitL1Withdraw // Event containing the contract specifics and raw log

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
func (it *RabbitL1WithdrawIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RabbitL1Withdraw)
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
		it.Event = new(RabbitL1Withdraw)
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
func (it *RabbitL1WithdrawIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RabbitL1WithdrawIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RabbitL1Withdraw represents a Withdraw event raised by the RabbitL1 contract.
type RabbitL1Withdraw struct {
	Trader common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterWithdraw is a free log retrieval operation binding the contract event 0x884edad9ce6fa2440d8a54cc123490eb96d2768479d49ff9c7366125a9424364.
//
// Solidity: event Withdraw(address indexed trader, uint256 amount)
func (_RabbitL1 *RabbitL1Filterer) FilterWithdraw(opts *bind.FilterOpts, trader []common.Address) (*RabbitL1WithdrawIterator, error) {

	var traderRule []interface{}
	for _, traderItem := range trader {
		traderRule = append(traderRule, traderItem)
	}

	logs, sub, err := _RabbitL1.contract.FilterLogs(opts, "Withdraw", traderRule)
	if err != nil {
		return nil, err
	}
	return &RabbitL1WithdrawIterator{contract: _RabbitL1.contract, event: "Withdraw", logs: logs, sub: sub}, nil
}

// WatchWithdraw is a free log subscription operation binding the contract event 0x884edad9ce6fa2440d8a54cc123490eb96d2768479d49ff9c7366125a9424364.
//
// Solidity: event Withdraw(address indexed trader, uint256 amount)
func (_RabbitL1 *RabbitL1Filterer) WatchWithdraw(opts *bind.WatchOpts, sink chan<- *RabbitL1Withdraw, trader []common.Address) (event.Subscription, error) {

	var traderRule []interface{}
	for _, traderItem := range trader {
		traderRule = append(traderRule, traderItem)
	}

	logs, sub, err := _RabbitL1.contract.WatchLogs(opts, "Withdraw", traderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RabbitL1Withdraw)
				if err := _RabbitL1.contract.UnpackLog(event, "Withdraw", log); err != nil {
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

// ParseWithdraw is a log parse operation binding the contract event 0x884edad9ce6fa2440d8a54cc123490eb96d2768479d49ff9c7366125a9424364.
//
// Solidity: event Withdraw(address indexed trader, uint256 amount)
func (_RabbitL1 *RabbitL1Filterer) ParseWithdraw(log types.Log) (*RabbitL1Withdraw, error) {
	event := new(RabbitL1Withdraw)
	if err := _RabbitL1.contract.UnpackLog(event, "Withdraw", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RabbitL1WithdrawToIterator is returned from FilterWithdrawTo and is used to iterate over the raw logs and unpacked data for WithdrawTo events raised by the RabbitL1 contract.
type RabbitL1WithdrawToIterator struct {
	Event *RabbitL1WithdrawTo // Event containing the contract specifics and raw log

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
func (it *RabbitL1WithdrawToIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RabbitL1WithdrawTo)
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
		it.Event = new(RabbitL1WithdrawTo)
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
func (it *RabbitL1WithdrawToIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RabbitL1WithdrawToIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RabbitL1WithdrawTo represents a WithdrawTo event raised by the RabbitL1 contract.
type RabbitL1WithdrawTo struct {
	To     common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterWithdrawTo is a free log retrieval operation binding the contract event 0x47096d7b247e809edf18e9bccfcb92f2af426ce8e6b40c923e65cb1b8394cef7.
//
// Solidity: event WithdrawTo(address indexed to, uint256 amount)
func (_RabbitL1 *RabbitL1Filterer) FilterWithdrawTo(opts *bind.FilterOpts, to []common.Address) (*RabbitL1WithdrawToIterator, error) {

	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _RabbitL1.contract.FilterLogs(opts, "WithdrawTo", toRule)
	if err != nil {
		return nil, err
	}
	return &RabbitL1WithdrawToIterator{contract: _RabbitL1.contract, event: "WithdrawTo", logs: logs, sub: sub}, nil
}

// WatchWithdrawTo is a free log subscription operation binding the contract event 0x47096d7b247e809edf18e9bccfcb92f2af426ce8e6b40c923e65cb1b8394cef7.
//
// Solidity: event WithdrawTo(address indexed to, uint256 amount)
func (_RabbitL1 *RabbitL1Filterer) WatchWithdrawTo(opts *bind.WatchOpts, sink chan<- *RabbitL1WithdrawTo, to []common.Address) (event.Subscription, error) {

	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _RabbitL1.contract.WatchLogs(opts, "WithdrawTo", toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RabbitL1WithdrawTo)
				if err := _RabbitL1.contract.UnpackLog(event, "WithdrawTo", log); err != nil {
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
// Solidity: event WithdrawTo(address indexed to, uint256 amount)
func (_RabbitL1 *RabbitL1Filterer) ParseWithdrawTo(log types.Log) (*RabbitL1WithdrawTo, error) {
	event := new(RabbitL1WithdrawTo)
	if err := _RabbitL1.contract.UnpackLog(event, "WithdrawTo", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RabbitL1WithdrawalReceiptIterator is returned from FilterWithdrawalReceipt and is used to iterate over the raw logs and unpacked data for WithdrawalReceipt events raised by the RabbitL1 contract.
type RabbitL1WithdrawalReceiptIterator struct {
	Event *RabbitL1WithdrawalReceipt // Event containing the contract specifics and raw log

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
func (it *RabbitL1WithdrawalReceiptIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RabbitL1WithdrawalReceipt)
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
		it.Event = new(RabbitL1WithdrawalReceipt)
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
func (it *RabbitL1WithdrawalReceiptIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RabbitL1WithdrawalReceiptIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RabbitL1WithdrawalReceipt represents a WithdrawalReceipt event raised by the RabbitL1 contract.
type RabbitL1WithdrawalReceipt struct {
	Id     *big.Int
	Trader common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterWithdrawalReceipt is a free log retrieval operation binding the contract event 0x64ef09c96beca083de3bd312078bd3b09203dfe40a2923e4704ac55dda16c67d.
//
// Solidity: event WithdrawalReceipt(uint256 indexed id, address indexed trader, uint256 amount)
func (_RabbitL1 *RabbitL1Filterer) FilterWithdrawalReceipt(opts *bind.FilterOpts, id []*big.Int, trader []common.Address) (*RabbitL1WithdrawalReceiptIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var traderRule []interface{}
	for _, traderItem := range trader {
		traderRule = append(traderRule, traderItem)
	}

	logs, sub, err := _RabbitL1.contract.FilterLogs(opts, "WithdrawalReceipt", idRule, traderRule)
	if err != nil {
		return nil, err
	}
	return &RabbitL1WithdrawalReceiptIterator{contract: _RabbitL1.contract, event: "WithdrawalReceipt", logs: logs, sub: sub}, nil
}

// WatchWithdrawalReceipt is a free log subscription operation binding the contract event 0x64ef09c96beca083de3bd312078bd3b09203dfe40a2923e4704ac55dda16c67d.
//
// Solidity: event WithdrawalReceipt(uint256 indexed id, address indexed trader, uint256 amount)
func (_RabbitL1 *RabbitL1Filterer) WatchWithdrawalReceipt(opts *bind.WatchOpts, sink chan<- *RabbitL1WithdrawalReceipt, id []*big.Int, trader []common.Address) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var traderRule []interface{}
	for _, traderItem := range trader {
		traderRule = append(traderRule, traderItem)
	}

	logs, sub, err := _RabbitL1.contract.WatchLogs(opts, "WithdrawalReceipt", idRule, traderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RabbitL1WithdrawalReceipt)
				if err := _RabbitL1.contract.UnpackLog(event, "WithdrawalReceipt", log); err != nil {
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

// ParseWithdrawalReceipt is a log parse operation binding the contract event 0x64ef09c96beca083de3bd312078bd3b09203dfe40a2923e4704ac55dda16c67d.
//
// Solidity: event WithdrawalReceipt(uint256 indexed id, address indexed trader, uint256 amount)
func (_RabbitL1 *RabbitL1Filterer) ParseWithdrawalReceipt(log types.Log) (*RabbitL1WithdrawalReceipt, error) {
	event := new(RabbitL1WithdrawalReceipt)
	if err := _RabbitL1.contract.UnpackLog(event, "WithdrawalReceipt", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
