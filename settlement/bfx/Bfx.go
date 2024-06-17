// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bfx

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

// BfxMetaData contains all meta data concerning the Bfx contract.
var BfxMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_signer\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_claimer\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_paymentToken\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"ClaimedYield\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"trader\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"signer\",\"type\":\"address\"}],\"name\":\"SetSigner\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"SetToken\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"WithdrawTo\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"trader\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"WithdrawalReceipt\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"new_signer\",\"type\":\"address\"}],\"name\":\"changeSigner\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"claimYield\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"claimer\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"deposit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"external_signer\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"paymentToken\",\"outputs\":[{\"internalType\":\"contractIERC20Rebasing\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"processedWithdrawals\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_paymentToken\",\"type\":\"address\"}],\"name\":\"setPaymentToken\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"trader\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"withdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"withdrawTokensTo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// BfxABI is the input ABI used to generate the binding from.
// Deprecated: Use BfxMetaData.ABI instead.
var BfxABI = BfxMetaData.ABI

// Bfx is an auto generated Go binding around an Ethereum contract.
type Bfx struct {
	BfxCaller     // Read-only binding to the contract
	BfxTransactor // Write-only binding to the contract
	BfxFilterer   // Log filterer for contract events
}

// BfxCaller is an auto generated read-only Go binding around an Ethereum contract.
type BfxCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BfxTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BfxTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BfxFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BfxFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BfxSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BfxSession struct {
	Contract     *Bfx              // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BfxCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BfxCallerSession struct {
	Contract *BfxCaller    // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// BfxTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BfxTransactorSession struct {
	Contract     *BfxTransactor    // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BfxRaw is an auto generated low-level Go binding around an Ethereum contract.
type BfxRaw struct {
	Contract *Bfx // Generic contract binding to access the raw methods on
}

// BfxCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BfxCallerRaw struct {
	Contract *BfxCaller // Generic read-only contract binding to access the raw methods on
}

// BfxTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BfxTransactorRaw struct {
	Contract *BfxTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBfx creates a new instance of Bfx, bound to a specific deployed contract.
func NewBfx(address common.Address, backend bind.ContractBackend) (*Bfx, error) {
	contract, err := bindBfx(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Bfx{BfxCaller: BfxCaller{contract: contract}, BfxTransactor: BfxTransactor{contract: contract}, BfxFilterer: BfxFilterer{contract: contract}}, nil
}

// NewBfxCaller creates a new read-only instance of Bfx, bound to a specific deployed contract.
func NewBfxCaller(address common.Address, caller bind.ContractCaller) (*BfxCaller, error) {
	contract, err := bindBfx(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BfxCaller{contract: contract}, nil
}

// NewBfxTransactor creates a new write-only instance of Bfx, bound to a specific deployed contract.
func NewBfxTransactor(address common.Address, transactor bind.ContractTransactor) (*BfxTransactor, error) {
	contract, err := bindBfx(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BfxTransactor{contract: contract}, nil
}

// NewBfxFilterer creates a new log filterer instance of Bfx, bound to a specific deployed contract.
func NewBfxFilterer(address common.Address, filterer bind.ContractFilterer) (*BfxFilterer, error) {
	contract, err := bindBfx(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BfxFilterer{contract: contract}, nil
}

// bindBfx binds a generic wrapper to an already deployed contract.
func bindBfx(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(BfxABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Bfx *BfxRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Bfx.Contract.BfxCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Bfx *BfxRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bfx.Contract.BfxTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Bfx *BfxRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Bfx.Contract.BfxTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Bfx *BfxCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Bfx.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Bfx *BfxTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bfx.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Bfx *BfxTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Bfx.Contract.contract.Transact(opts, method, params...)
}

// Claimer is a free data retrieval call binding the contract method 0xd379be23.
//
// Solidity: function claimer() view returns(address)
func (_Bfx *BfxCaller) Claimer(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Bfx.contract.Call(opts, &out, "claimer")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Claimer is a free data retrieval call binding the contract method 0xd379be23.
//
// Solidity: function claimer() view returns(address)
func (_Bfx *BfxSession) Claimer() (common.Address, error) {
	return _Bfx.Contract.Claimer(&_Bfx.CallOpts)
}

// Claimer is a free data retrieval call binding the contract method 0xd379be23.
//
// Solidity: function claimer() view returns(address)
func (_Bfx *BfxCallerSession) Claimer() (common.Address, error) {
	return _Bfx.Contract.Claimer(&_Bfx.CallOpts)
}

// ExternalSigner is a free data retrieval call binding the contract method 0x0008cecb.
//
// Solidity: function external_signer() view returns(address)
func (_Bfx *BfxCaller) ExternalSigner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Bfx.contract.Call(opts, &out, "external_signer")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ExternalSigner is a free data retrieval call binding the contract method 0x0008cecb.
//
// Solidity: function external_signer() view returns(address)
func (_Bfx *BfxSession) ExternalSigner() (common.Address, error) {
	return _Bfx.Contract.ExternalSigner(&_Bfx.CallOpts)
}

// ExternalSigner is a free data retrieval call binding the contract method 0x0008cecb.
//
// Solidity: function external_signer() view returns(address)
func (_Bfx *BfxCallerSession) ExternalSigner() (common.Address, error) {
	return _Bfx.Contract.ExternalSigner(&_Bfx.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Bfx *BfxCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Bfx.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Bfx *BfxSession) Owner() (common.Address, error) {
	return _Bfx.Contract.Owner(&_Bfx.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Bfx *BfxCallerSession) Owner() (common.Address, error) {
	return _Bfx.Contract.Owner(&_Bfx.CallOpts)
}

// PaymentToken is a free data retrieval call binding the contract method 0x3013ce29.
//
// Solidity: function paymentToken() view returns(address)
func (_Bfx *BfxCaller) PaymentToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Bfx.contract.Call(opts, &out, "paymentToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PaymentToken is a free data retrieval call binding the contract method 0x3013ce29.
//
// Solidity: function paymentToken() view returns(address)
func (_Bfx *BfxSession) PaymentToken() (common.Address, error) {
	return _Bfx.Contract.PaymentToken(&_Bfx.CallOpts)
}

// PaymentToken is a free data retrieval call binding the contract method 0x3013ce29.
//
// Solidity: function paymentToken() view returns(address)
func (_Bfx *BfxCallerSession) PaymentToken() (common.Address, error) {
	return _Bfx.Contract.PaymentToken(&_Bfx.CallOpts)
}

// ProcessedWithdrawals is a free data retrieval call binding the contract method 0xdde4d950.
//
// Solidity: function processedWithdrawals(uint256 ) view returns(bool)
func (_Bfx *BfxCaller) ProcessedWithdrawals(opts *bind.CallOpts, arg0 *big.Int) (bool, error) {
	var out []interface{}
	err := _Bfx.contract.Call(opts, &out, "processedWithdrawals", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// ProcessedWithdrawals is a free data retrieval call binding the contract method 0xdde4d950.
//
// Solidity: function processedWithdrawals(uint256 ) view returns(bool)
func (_Bfx *BfxSession) ProcessedWithdrawals(arg0 *big.Int) (bool, error) {
	return _Bfx.Contract.ProcessedWithdrawals(&_Bfx.CallOpts, arg0)
}

// ProcessedWithdrawals is a free data retrieval call binding the contract method 0xdde4d950.
//
// Solidity: function processedWithdrawals(uint256 ) view returns(bool)
func (_Bfx *BfxCallerSession) ProcessedWithdrawals(arg0 *big.Int) (bool, error) {
	return _Bfx.Contract.ProcessedWithdrawals(&_Bfx.CallOpts, arg0)
}

// ChangeSigner is a paid mutator transaction binding the contract method 0xaad2b723.
//
// Solidity: function changeSigner(address new_signer) returns()
func (_Bfx *BfxTransactor) ChangeSigner(opts *bind.TransactOpts, new_signer common.Address) (*types.Transaction, error) {
	return _Bfx.contract.Transact(opts, "changeSigner", new_signer)
}

// ChangeSigner is a paid mutator transaction binding the contract method 0xaad2b723.
//
// Solidity: function changeSigner(address new_signer) returns()
func (_Bfx *BfxSession) ChangeSigner(new_signer common.Address) (*types.Transaction, error) {
	return _Bfx.Contract.ChangeSigner(&_Bfx.TransactOpts, new_signer)
}

// ChangeSigner is a paid mutator transaction binding the contract method 0xaad2b723.
//
// Solidity: function changeSigner(address new_signer) returns()
func (_Bfx *BfxTransactorSession) ChangeSigner(new_signer common.Address) (*types.Transaction, error) {
	return _Bfx.Contract.ChangeSigner(&_Bfx.TransactOpts, new_signer)
}

// ClaimYield is a paid mutator transaction binding the contract method 0x406cf229.
//
// Solidity: function claimYield() returns()
func (_Bfx *BfxTransactor) ClaimYield(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bfx.contract.Transact(opts, "claimYield")
}

// ClaimYield is a paid mutator transaction binding the contract method 0x406cf229.
//
// Solidity: function claimYield() returns()
func (_Bfx *BfxSession) ClaimYield() (*types.Transaction, error) {
	return _Bfx.Contract.ClaimYield(&_Bfx.TransactOpts)
}

// ClaimYield is a paid mutator transaction binding the contract method 0x406cf229.
//
// Solidity: function claimYield() returns()
func (_Bfx *BfxTransactorSession) ClaimYield() (*types.Transaction, error) {
	return _Bfx.Contract.ClaimYield(&_Bfx.TransactOpts)
}

// Deposit is a paid mutator transaction binding the contract method 0xb6b55f25.
//
// Solidity: function deposit(uint256 amount) returns()
func (_Bfx *BfxTransactor) Deposit(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _Bfx.contract.Transact(opts, "deposit", amount)
}

// Deposit is a paid mutator transaction binding the contract method 0xb6b55f25.
//
// Solidity: function deposit(uint256 amount) returns()
func (_Bfx *BfxSession) Deposit(amount *big.Int) (*types.Transaction, error) {
	return _Bfx.Contract.Deposit(&_Bfx.TransactOpts, amount)
}

// Deposit is a paid mutator transaction binding the contract method 0xb6b55f25.
//
// Solidity: function deposit(uint256 amount) returns()
func (_Bfx *BfxTransactorSession) Deposit(amount *big.Int) (*types.Transaction, error) {
	return _Bfx.Contract.Deposit(&_Bfx.TransactOpts, amount)
}

// SetPaymentToken is a paid mutator transaction binding the contract method 0x6a326ab1.
//
// Solidity: function setPaymentToken(address _paymentToken) returns()
func (_Bfx *BfxTransactor) SetPaymentToken(opts *bind.TransactOpts, _paymentToken common.Address) (*types.Transaction, error) {
	return _Bfx.contract.Transact(opts, "setPaymentToken", _paymentToken)
}

// SetPaymentToken is a paid mutator transaction binding the contract method 0x6a326ab1.
//
// Solidity: function setPaymentToken(address _paymentToken) returns()
func (_Bfx *BfxSession) SetPaymentToken(_paymentToken common.Address) (*types.Transaction, error) {
	return _Bfx.Contract.SetPaymentToken(&_Bfx.TransactOpts, _paymentToken)
}

// SetPaymentToken is a paid mutator transaction binding the contract method 0x6a326ab1.
//
// Solidity: function setPaymentToken(address _paymentToken) returns()
func (_Bfx *BfxTransactorSession) SetPaymentToken(_paymentToken common.Address) (*types.Transaction, error) {
	return _Bfx.Contract.SetPaymentToken(&_Bfx.TransactOpts, _paymentToken)
}

// Withdraw is a paid mutator transaction binding the contract method 0x61c8e739.
//
// Solidity: function withdraw(uint256 id, address trader, uint256 amount, uint8 v, bytes32 r, bytes32 s) returns()
func (_Bfx *BfxTransactor) Withdraw(opts *bind.TransactOpts, id *big.Int, trader common.Address, amount *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _Bfx.contract.Transact(opts, "withdraw", id, trader, amount, v, r, s)
}

// Withdraw is a paid mutator transaction binding the contract method 0x61c8e739.
//
// Solidity: function withdraw(uint256 id, address trader, uint256 amount, uint8 v, bytes32 r, bytes32 s) returns()
func (_Bfx *BfxSession) Withdraw(id *big.Int, trader common.Address, amount *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _Bfx.Contract.Withdraw(&_Bfx.TransactOpts, id, trader, amount, v, r, s)
}

// Withdraw is a paid mutator transaction binding the contract method 0x61c8e739.
//
// Solidity: function withdraw(uint256 id, address trader, uint256 amount, uint8 v, bytes32 r, bytes32 s) returns()
func (_Bfx *BfxTransactorSession) Withdraw(id *big.Int, trader common.Address, amount *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _Bfx.Contract.Withdraw(&_Bfx.TransactOpts, id, trader, amount, v, r, s)
}

// WithdrawTokensTo is a paid mutator transaction binding the contract method 0xb743e722.
//
// Solidity: function withdrawTokensTo(uint256 amount, address to) returns()
func (_Bfx *BfxTransactor) WithdrawTokensTo(opts *bind.TransactOpts, amount *big.Int, to common.Address) (*types.Transaction, error) {
	return _Bfx.contract.Transact(opts, "withdrawTokensTo", amount, to)
}

// WithdrawTokensTo is a paid mutator transaction binding the contract method 0xb743e722.
//
// Solidity: function withdrawTokensTo(uint256 amount, address to) returns()
func (_Bfx *BfxSession) WithdrawTokensTo(amount *big.Int, to common.Address) (*types.Transaction, error) {
	return _Bfx.Contract.WithdrawTokensTo(&_Bfx.TransactOpts, amount, to)
}

// WithdrawTokensTo is a paid mutator transaction binding the contract method 0xb743e722.
//
// Solidity: function withdrawTokensTo(uint256 amount, address to) returns()
func (_Bfx *BfxTransactorSession) WithdrawTokensTo(amount *big.Int, to common.Address) (*types.Transaction, error) {
	return _Bfx.Contract.WithdrawTokensTo(&_Bfx.TransactOpts, amount, to)
}

// BfxClaimedYieldIterator is returned from FilterClaimedYield and is used to iterate over the raw logs and unpacked data for ClaimedYield events raised by the Bfx contract.
type BfxClaimedYieldIterator struct {
	Event *BfxClaimedYield // Event containing the contract specifics and raw log

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
func (it *BfxClaimedYieldIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BfxClaimedYield)
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
		it.Event = new(BfxClaimedYield)
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
func (it *BfxClaimedYieldIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BfxClaimedYieldIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BfxClaimedYield represents a ClaimedYield event raised by the Bfx contract.
type BfxClaimedYield struct {
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterClaimedYield is a free log retrieval operation binding the contract event 0x32919048af00a655979f73cd0a43b0e566a60e04a81a751b9329c2658441e05b.
//
// Solidity: event ClaimedYield(uint256 amount)
func (_Bfx *BfxFilterer) FilterClaimedYield(opts *bind.FilterOpts) (*BfxClaimedYieldIterator, error) {

	logs, sub, err := _Bfx.contract.FilterLogs(opts, "ClaimedYield")
	if err != nil {
		return nil, err
	}
	return &BfxClaimedYieldIterator{contract: _Bfx.contract, event: "ClaimedYield", logs: logs, sub: sub}, nil
}

// WatchClaimedYield is a free log subscription operation binding the contract event 0x32919048af00a655979f73cd0a43b0e566a60e04a81a751b9329c2658441e05b.
//
// Solidity: event ClaimedYield(uint256 amount)
func (_Bfx *BfxFilterer) WatchClaimedYield(opts *bind.WatchOpts, sink chan<- *BfxClaimedYield) (event.Subscription, error) {

	logs, sub, err := _Bfx.contract.WatchLogs(opts, "ClaimedYield")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BfxClaimedYield)
				if err := _Bfx.contract.UnpackLog(event, "ClaimedYield", log); err != nil {
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

// ParseClaimedYield is a log parse operation binding the contract event 0x32919048af00a655979f73cd0a43b0e566a60e04a81a751b9329c2658441e05b.
//
// Solidity: event ClaimedYield(uint256 amount)
func (_Bfx *BfxFilterer) ParseClaimedYield(log types.Log) (*BfxClaimedYield, error) {
	event := new(BfxClaimedYield)
	if err := _Bfx.contract.UnpackLog(event, "ClaimedYield", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BfxDepositIterator is returned from FilterDeposit and is used to iterate over the raw logs and unpacked data for Deposit events raised by the Bfx contract.
type BfxDepositIterator struct {
	Event *BfxDeposit // Event containing the contract specifics and raw log

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
func (it *BfxDepositIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BfxDeposit)
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
		it.Event = new(BfxDeposit)
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
func (it *BfxDepositIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BfxDepositIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BfxDeposit represents a Deposit event raised by the Bfx contract.
type BfxDeposit struct {
	Id     *big.Int
	Trader common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDeposit is a free log retrieval operation binding the contract event 0xeaa18152488ce5959073c9c79c88ca90b3d96c00de1f118cfaad664c3dab06b9.
//
// Solidity: event Deposit(uint256 indexed id, address indexed trader, uint256 amount)
func (_Bfx *BfxFilterer) FilterDeposit(opts *bind.FilterOpts, id []*big.Int, trader []common.Address) (*BfxDepositIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var traderRule []interface{}
	for _, traderItem := range trader {
		traderRule = append(traderRule, traderItem)
	}

	logs, sub, err := _Bfx.contract.FilterLogs(opts, "Deposit", idRule, traderRule)
	if err != nil {
		return nil, err
	}
	return &BfxDepositIterator{contract: _Bfx.contract, event: "Deposit", logs: logs, sub: sub}, nil
}

// WatchDeposit is a free log subscription operation binding the contract event 0xeaa18152488ce5959073c9c79c88ca90b3d96c00de1f118cfaad664c3dab06b9.
//
// Solidity: event Deposit(uint256 indexed id, address indexed trader, uint256 amount)
func (_Bfx *BfxFilterer) WatchDeposit(opts *bind.WatchOpts, sink chan<- *BfxDeposit, id []*big.Int, trader []common.Address) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var traderRule []interface{}
	for _, traderItem := range trader {
		traderRule = append(traderRule, traderItem)
	}

	logs, sub, err := _Bfx.contract.WatchLogs(opts, "Deposit", idRule, traderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BfxDeposit)
				if err := _Bfx.contract.UnpackLog(event, "Deposit", log); err != nil {
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
func (_Bfx *BfxFilterer) ParseDeposit(log types.Log) (*BfxDeposit, error) {
	event := new(BfxDeposit)
	if err := _Bfx.contract.UnpackLog(event, "Deposit", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BfxSetSignerIterator is returned from FilterSetSigner and is used to iterate over the raw logs and unpacked data for SetSigner events raised by the Bfx contract.
type BfxSetSignerIterator struct {
	Event *BfxSetSigner // Event containing the contract specifics and raw log

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
func (it *BfxSetSignerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BfxSetSigner)
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
		it.Event = new(BfxSetSigner)
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
func (it *BfxSetSignerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BfxSetSignerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BfxSetSigner represents a SetSigner event raised by the Bfx contract.
type BfxSetSigner struct {
	Signer common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterSetSigner is a free log retrieval operation binding the contract event 0xbb10aee7ef5a307b8097c6a7f2892b909ff1736fd24a6a5260640c185f7153b6.
//
// Solidity: event SetSigner(address indexed signer)
func (_Bfx *BfxFilterer) FilterSetSigner(opts *bind.FilterOpts, signer []common.Address) (*BfxSetSignerIterator, error) {

	var signerRule []interface{}
	for _, signerItem := range signer {
		signerRule = append(signerRule, signerItem)
	}

	logs, sub, err := _Bfx.contract.FilterLogs(opts, "SetSigner", signerRule)
	if err != nil {
		return nil, err
	}
	return &BfxSetSignerIterator{contract: _Bfx.contract, event: "SetSigner", logs: logs, sub: sub}, nil
}

// WatchSetSigner is a free log subscription operation binding the contract event 0xbb10aee7ef5a307b8097c6a7f2892b909ff1736fd24a6a5260640c185f7153b6.
//
// Solidity: event SetSigner(address indexed signer)
func (_Bfx *BfxFilterer) WatchSetSigner(opts *bind.WatchOpts, sink chan<- *BfxSetSigner, signer []common.Address) (event.Subscription, error) {

	var signerRule []interface{}
	for _, signerItem := range signer {
		signerRule = append(signerRule, signerItem)
	}

	logs, sub, err := _Bfx.contract.WatchLogs(opts, "SetSigner", signerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BfxSetSigner)
				if err := _Bfx.contract.UnpackLog(event, "SetSigner", log); err != nil {
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

// ParseSetSigner is a log parse operation binding the contract event 0xbb10aee7ef5a307b8097c6a7f2892b909ff1736fd24a6a5260640c185f7153b6.
//
// Solidity: event SetSigner(address indexed signer)
func (_Bfx *BfxFilterer) ParseSetSigner(log types.Log) (*BfxSetSigner, error) {
	event := new(BfxSetSigner)
	if err := _Bfx.contract.UnpackLog(event, "SetSigner", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BfxSetTokenIterator is returned from FilterSetToken and is used to iterate over the raw logs and unpacked data for SetToken events raised by the Bfx contract.
type BfxSetTokenIterator struct {
	Event *BfxSetToken // Event containing the contract specifics and raw log

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
func (it *BfxSetTokenIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BfxSetToken)
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
		it.Event = new(BfxSetToken)
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
func (it *BfxSetTokenIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BfxSetTokenIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BfxSetToken represents a SetToken event raised by the Bfx contract.
type BfxSetToken struct {
	Token common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterSetToken is a free log retrieval operation binding the contract event 0xefc1fd16ea80a922086ee4e995739d59b025c1bcea6d1f67855747480c83214b.
//
// Solidity: event SetToken(address indexed token)
func (_Bfx *BfxFilterer) FilterSetToken(opts *bind.FilterOpts, token []common.Address) (*BfxSetTokenIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _Bfx.contract.FilterLogs(opts, "SetToken", tokenRule)
	if err != nil {
		return nil, err
	}
	return &BfxSetTokenIterator{contract: _Bfx.contract, event: "SetToken", logs: logs, sub: sub}, nil
}

// WatchSetToken is a free log subscription operation binding the contract event 0xefc1fd16ea80a922086ee4e995739d59b025c1bcea6d1f67855747480c83214b.
//
// Solidity: event SetToken(address indexed token)
func (_Bfx *BfxFilterer) WatchSetToken(opts *bind.WatchOpts, sink chan<- *BfxSetToken, token []common.Address) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _Bfx.contract.WatchLogs(opts, "SetToken", tokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BfxSetToken)
				if err := _Bfx.contract.UnpackLog(event, "SetToken", log); err != nil {
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

// ParseSetToken is a log parse operation binding the contract event 0xefc1fd16ea80a922086ee4e995739d59b025c1bcea6d1f67855747480c83214b.
//
// Solidity: event SetToken(address indexed token)
func (_Bfx *BfxFilterer) ParseSetToken(log types.Log) (*BfxSetToken, error) {
	event := new(BfxSetToken)
	if err := _Bfx.contract.UnpackLog(event, "SetToken", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BfxWithdrawToIterator is returned from FilterWithdrawTo and is used to iterate over the raw logs and unpacked data for WithdrawTo events raised by the Bfx contract.
type BfxWithdrawToIterator struct {
	Event *BfxWithdrawTo // Event containing the contract specifics and raw log

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
func (it *BfxWithdrawToIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BfxWithdrawTo)
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
		it.Event = new(BfxWithdrawTo)
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
func (it *BfxWithdrawToIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BfxWithdrawToIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BfxWithdrawTo represents a WithdrawTo event raised by the Bfx contract.
type BfxWithdrawTo struct {
	To     common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterWithdrawTo is a free log retrieval operation binding the contract event 0x47096d7b247e809edf18e9bccfcb92f2af426ce8e6b40c923e65cb1b8394cef7.
//
// Solidity: event WithdrawTo(address indexed to, uint256 amount)
func (_Bfx *BfxFilterer) FilterWithdrawTo(opts *bind.FilterOpts, to []common.Address) (*BfxWithdrawToIterator, error) {

	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _Bfx.contract.FilterLogs(opts, "WithdrawTo", toRule)
	if err != nil {
		return nil, err
	}
	return &BfxWithdrawToIterator{contract: _Bfx.contract, event: "WithdrawTo", logs: logs, sub: sub}, nil
}

// WatchWithdrawTo is a free log subscription operation binding the contract event 0x47096d7b247e809edf18e9bccfcb92f2af426ce8e6b40c923e65cb1b8394cef7.
//
// Solidity: event WithdrawTo(address indexed to, uint256 amount)
func (_Bfx *BfxFilterer) WatchWithdrawTo(opts *bind.WatchOpts, sink chan<- *BfxWithdrawTo, to []common.Address) (event.Subscription, error) {

	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _Bfx.contract.WatchLogs(opts, "WithdrawTo", toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BfxWithdrawTo)
				if err := _Bfx.contract.UnpackLog(event, "WithdrawTo", log); err != nil {
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
func (_Bfx *BfxFilterer) ParseWithdrawTo(log types.Log) (*BfxWithdrawTo, error) {
	event := new(BfxWithdrawTo)
	if err := _Bfx.contract.UnpackLog(event, "WithdrawTo", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BfxWithdrawalReceiptIterator is returned from FilterWithdrawalReceipt and is used to iterate over the raw logs and unpacked data for WithdrawalReceipt events raised by the Bfx contract.
type BfxWithdrawalReceiptIterator struct {
	Event *BfxWithdrawalReceipt // Event containing the contract specifics and raw log

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
func (it *BfxWithdrawalReceiptIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BfxWithdrawalReceipt)
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
		it.Event = new(BfxWithdrawalReceipt)
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
func (it *BfxWithdrawalReceiptIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BfxWithdrawalReceiptIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BfxWithdrawalReceipt represents a WithdrawalReceipt event raised by the Bfx contract.
type BfxWithdrawalReceipt struct {
	Id     *big.Int
	Trader common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterWithdrawalReceipt is a free log retrieval operation binding the contract event 0x64ef09c96beca083de3bd312078bd3b09203dfe40a2923e4704ac55dda16c67d.
//
// Solidity: event WithdrawalReceipt(uint256 indexed id, address indexed trader, uint256 amount)
func (_Bfx *BfxFilterer) FilterWithdrawalReceipt(opts *bind.FilterOpts, id []*big.Int, trader []common.Address) (*BfxWithdrawalReceiptIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var traderRule []interface{}
	for _, traderItem := range trader {
		traderRule = append(traderRule, traderItem)
	}

	logs, sub, err := _Bfx.contract.FilterLogs(opts, "WithdrawalReceipt", idRule, traderRule)
	if err != nil {
		return nil, err
	}
	return &BfxWithdrawalReceiptIterator{contract: _Bfx.contract, event: "WithdrawalReceipt", logs: logs, sub: sub}, nil
}

// WatchWithdrawalReceipt is a free log subscription operation binding the contract event 0x64ef09c96beca083de3bd312078bd3b09203dfe40a2923e4704ac55dda16c67d.
//
// Solidity: event WithdrawalReceipt(uint256 indexed id, address indexed trader, uint256 amount)
func (_Bfx *BfxFilterer) WatchWithdrawalReceipt(opts *bind.WatchOpts, sink chan<- *BfxWithdrawalReceipt, id []*big.Int, trader []common.Address) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var traderRule []interface{}
	for _, traderItem := range trader {
		traderRule = append(traderRule, traderItem)
	}

	logs, sub, err := _Bfx.contract.WatchLogs(opts, "WithdrawalReceipt", idRule, traderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BfxWithdrawalReceipt)
				if err := _Bfx.contract.UnpackLog(event, "WithdrawalReceipt", log); err != nil {
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
func (_Bfx *BfxFilterer) ParseWithdrawalReceipt(log types.Log) (*BfxWithdrawalReceipt, error) {
	event := new(BfxWithdrawalReceipt)
	if err := _Bfx.contract.UnpackLog(event, "WithdrawalReceipt", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
