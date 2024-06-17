package model

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum/common"
	"github.com/strips-finance/rabbit-dex-backend/signer"
	"github.com/strips-finance/rabbit-dex-backend/tdecimal"
	"github.com/strips-finance/rabbit-dex-backend/tick"
)

const (
	ACQUIRE_STAKE_LOCK     = "balance.acquire_stake_lock"
	RELEASE_STAKE_LOCK     = "balance.release_stake_lock"
	ACQUIRE_UNSTAKE_LOCK   = "balance.acquire_unstake_lock"
	RELEASE_UNSTAKE_LOCK   = "balance.release_unstake_lock"
	ACQUIRE_WITHDRAW_LOCK  = "balance.acquire_withdraw_lock"
	RELEASE_WITHDRAW_LOCK  = "balance.release_withdraw_lock"
	CHECK_WITHDRAW_ALLOWED = "balance.check_withdraw_allowed"

	GET_LAST_PROCESSED_BLOCK_NUMBER = "balance.get_last_processed_block_number"
	SET_LAST_PROCESSED_BLOCK_NUMBER = "balance.set_last_processed_block_number"

	GET_SETTLEMENT_STATE = "balance.get_settlement_state"
	GET_PROCESSING_OPS   = "balance.get_processing_ops"
	MERGE_UNKNOWN_OPS    = "profile.merge_unknown_ops"
	DELETE_UNKNOWN_OPS   = "balance.delete_unknown_ops"
	MAX_WITHDRAW_AMOUNT  = "balance.max_withdraw_amount"
	ROLLING_24_WDS       = "profile.wds_per_24h"

	ADD_CONTRACT_MAP             = "balance.add_to_contract_map"
	INIT_VAULT                   = "balance.init_vault"
	REACTIVATE_VAULT             = "balance.reactivate_vault"
	CREATE_DEPOSIT               = "balance.create_deposit"
	PROCESS_DEPOSIT              = "balance.process_deposit"
	PROCESS_DEPOSIT_UNKNOWN      = "balance.process_deposit_unknown"
	PROCESS_YIELD                = "cache.process_yield_and_invalidate"
	PROCESS_STAKE                = "balance.process_stake"
	CREATE_WITHDRAWAL            = "balance.create_withdrawal"
	CANCEL_WITHDRAWAL            = "balance.cancel_withdrawal"
	CLAIM_WITHDRAWAL             = "balance.claim_withdrawal"
	CREATE_STAKE                 = "balance.create_stake"
	CREATE_UNSTAKE               = "balance.create_unstake"
	CANCEL_UNSTAKE               = "balance.cancel_unstake"
	PROCESS_UNSTAKES             = "balance.process_unstakes"
	GET_PENDING_DEPOSITS         = "balance.get_pending_deposits"
	GET_PENDING_STAKES           = "balance.get_pending_stakes"
	GET_PENDING_WITHDRAWALS      = "balance.get_pending_withdrawals"
	GET_ALL_PENDING_WITHDRAWALS  = "balance.get_all_pending_withdrawals"
	GET_WITHDRAWABLE_UNSTAKES    = "balance.get_withdrawable_unstakes"
	GET_VAULT_MANAGER_PROFILE_ID = "balance.get_vault_manager_profile_id"
	GET_VAULT_INFO               = "balance.get_vault_info"
	GET_HOLDING_INFO             = "balance.get_holding_info"
	PENDING_DEPOSIT_CANCELED     = "balance.pending_deposit_canceled"
	UPDATE_PENDING_WITHDRAWALS   = "balance.update_pending_withdrawals"
	PROCESSING_WITHDRAWAL        = "balance.processing_withdrawal"
	COMPLETED_WITHDRAWALS        = "balance.completed_withdrawals"
	GET_WITHDRAWALS_SUSPENDED    = "balance.get_withdrawals_suspended"
	SUSPEND_WITHDRAWALS          = "balance.suspend_withdrawals"
)

var (
	decimalZero                 = decimal.NewFromInt(0)
	_decimals6  decimal.Decimal = decimal.NewFromInt32(10).Pow(decimal.NewFromInt32(6))
	_decimals18 decimal.Decimal = decimal.NewFromInt32(10).Pow(decimal.NewFromInt32(18))
)

func validateDecimals(_decimals decimal.Decimal, exchangeId string) error {
	if _decimals != _decimals6 && _decimals != _decimals18 {
		return fmt.Errorf("DECIMALS_NOT_ALLOWED _decimals=%s", _decimals.String())
	}

	if exchangeId == EXCHANGE_RBX && _decimals != _decimals6 {
		return fmt.Errorf("RBX WRONG DECIMALS _decimals=%s required=%s", _decimals.String(), _decimals6.String())
	}

	if exchangeId == EXCHANGE_BFX && _decimals != _decimals18 {
		return fmt.Errorf("BFX WRONG DECIMALS _decimals=%s required=%s", _decimals.String(), _decimals18.String())
	}

	return nil
}

//TODO: create separate token config and move this part there.
//all calcs with decimals should USE one token config later.

var eidToDecimals = map[string]decimal.Decimal{
	EXCHANGE_RBX: _decimals6,
	EXCHANGE_BFX: _decimals18,
}

type WithdrawalResponse struct {
	Withdrawal *BalanceOps `json:"withdrawal"`
	BnAmount   string      `json:"bn_amount"`
	R          string      `json:"r"`
	S          string      `json:"s"`
	V          uint        `json:"v"`
}

type StakeResponse struct {
	Id       string `json:"id"`
	Staker   string `json:"staker"`
	BnAmount string `json:"bn_amount"`
	BnNav    string `json:"bn_nav"`
	Block    string `json:"block"`
	R        string `json:"r"`
	S        string `json:"s"`
	V        uint   `json:"v"`
}

type UnstakeResponse struct {
	Id       string `json:"id"`
	Staker   string `json:"staker"`
	BnShares string `json:"bn_shares"`
	BnNav    string `json:"bn_nav"`
	Block    string `json:"block"`
	R        string `json:"r"`
	S        string `json:"s"`
	V        uint   `json:"v"`
}

type Deposit struct {
	Id              string            `msgpack:"id"`
	Wallet          string            `msgpack:"wallet"`
	Amount          *tdecimal.Decimal `msgpack:"amount"`
	Tx              string            `msgpack:"tx"`
	ExchangeId      string            `msgpack:"exchange_id"`
	ChainId         uint              `msgpack:"chain_id"`
	ExchangeAddress string            `msgpack:"exchange_address"`
}

type Yield struct {
	Id              string            `msgpack:"id"`
	Amount          *tdecimal.Decimal `msgpack:"amount"`
	Tx              string            `msgpack:"tx"`
	ExchangeId      string            `msgpack:"exchange_id"`
	ChainId         uint              `msgpack:"chain_id"`
	ExchangeAddress string            `msgpack:"exchange_address"`
}

type Stake struct {
	Id             string            `msgpack:"id"`
	VaultProfileId uint              `msgpack:"vault_profile_id"`
	VaultWallet    string            `msgpack:"vault_wallet"`
	Amount         *tdecimal.Decimal `msgpack:"amount"`
	CurrentNav     *tdecimal.Decimal `msgpack:"current_nav"`
	Tx             string            `msgpack:"tx"`
}

type StakeLock struct {
	ProfileId uint  `msgpack:"id"`
	Locked    bool  `msgpack:"locked"`
	Timestamp int64 `msgpack:"timestamp"`
}

type WithdrawLock struct {
	ProfileId uint  `msgpack:"id"`
	Locked    bool  `msgpack:"locked"`
	Timestamp int64 `msgpack:"timestamp"`
}

type WithdrawData struct {
	Id                 uint   `msgpack:"id" json:"id"`
	LastProcessedBlock string `msgpack:"last_processed_block" json:"last_processed_block"`
	Suspended          bool   `msgpack:"withdrawals_suspended" json:"withdrawals_suspended"`
}

type DepositData struct {
	Id                 uint   `msgpack:"id" json:"id"`
	LastProcessedBlock string `msgpack:"last_processed_block" json:"last_processed_block"`
}

type SettlementState struct {
	WithdrawState *WithdrawData `msgpack:"withdraw_state" json:"withdraw_state"`
	DepositState  *DepositData  `msgpack:"deposit_state" json:"deposit_state"`
}

type WithdrawalTxInfo struct {
	Id     string `msgpack:"id"`
	TxHash string `msgpack:"txhash"`
}

type ContractMap struct {
	ContractAddress string `msgpack:"contract_address" json:"contract_address"`
	ChainId         uint   `msgpack:"chain_id" json:"chain_id"`
	ExchangeId      string `msgpack:"exchange_id" json:"exchange_id"`
}

func (api *ApiModel) AcquireStakeLock(ctx context.Context, profileId uint) (*StakeLock, error) {
	ops, err := DataResponse[*StakeLock]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		ACQUIRE_STAKE_LOCK,
		[]interface{}{
			profileId,
		},
	)

	return ops, err
}

func (api *ApiModel) ReleaseStakeLock(ctx context.Context, profileId uint) (*StakeLock, error) {
	ops, err := DataResponse[*StakeLock]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		RELEASE_STAKE_LOCK,
		[]interface{}{
			profileId,
		},
	)

	return ops, err
}

func (api *ApiModel) AcquireUnstakeLock(ctx context.Context, profileId uint) (*StakeLock, error) {
	ops, err := DataResponse[*StakeLock]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		ACQUIRE_UNSTAKE_LOCK,
		[]interface{}{
			profileId,
		},
	)

	return ops, err
}

func (api *ApiModel) ReleaseUnstakeLock(ctx context.Context, profileId uint) (*StakeLock, error) {
	ops, err := DataResponse[*StakeLock]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		RELEASE_UNSTAKE_LOCK,
		[]interface{}{
			profileId,
		},
	)

	return ops, err
}

func (api *ApiModel) AcquireWithdrawLock(ctx context.Context, profileId uint) (*WithdrawLock, error) {
	ops, err := DataResponse[*WithdrawLock]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		ACQUIRE_WITHDRAW_LOCK,
		[]interface{}{
			profileId,
		},
	)

	return ops, err
}

func (api *ApiModel) ReleaseWithdrawLock(ctx context.Context, profileId uint) (*WithdrawLock, error) {
	ops, err := DataResponse[*WithdrawLock]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		RELEASE_WITHDRAW_LOCK,
		[]interface{}{
			profileId,
		},
	)

	return ops, err
}

func (api *ApiModel) CheckWithdrawAllowed(ctx context.Context, profileId uint) bool {
	_, err := DataResponse[interface{}]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		CHECK_WITHDRAW_ALLOWED,
		[]interface{}{
			profileId,
		},
	)
	return err == nil
}

func (api *ApiModel) CancelWithdrawal(ctx context.Context, profileId uint, bopsId string) (bool, error) {
	res, err := DataResponse[bool]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		CANCEL_WITHDRAWAL,
		[]interface{}{
			profileId,
			bopsId,
		},
	)
	return res, err
}

func GetWalletsStringInRabbitTntStandardFormat(initialWallet []string) []string {
	result := make([]string, len(initialWallet))
	for i, w := range initialWallet {
		result[i] = GetWalletStringInRabbitTntStandardFormat(w)
	}
	return result
}

func GetWalletStringInRabbitTntStandardFormat(initialWallet string) string {
	wallet := strings.ToLower(initialWallet)
	if !strings.HasPrefix(wallet, "0x") {
		return "0x" + wallet
	}
	return wallet
}

// Returns the input with its first charsToRemove characters removed.
// Deposit and withdrawal ids are strings "d_123", "w_1112" etc. in
// Tarantool, but are integers 123, 1112 on StarkNet and Ethereum.
// Also used to remove the leading '0x' from a hex number string.
func StripPrefix(input string, charsToRemove int) string {
	asRunes := []rune(input)
	return string(asRunes[charsToRemove:])
}

func getVaultAddress(vault string) common.Address {
	if strings.HasPrefix(vault, "0x") {
		vault = StripPrefix(vault, 2)
	}
	return common.HexToAddress(vault)
}

func (api *ApiModel) CreateStake(ctx context.Context, stakerProfile *Profile, vaultWallet string, requestedAmount float64, txhash string, exchangeId string, chainId uint) (*BalanceOps, error) {

	rounded_amount := tick.RoundDownToUsdtTick(requestedAmount)

	amount := tdecimal.NewDecimal(decimal.NewFromFloat(rounded_amount))

	ops, err := DataResponse[*BalanceOps]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		CREATE_STAKE,
		[]interface{}{
			stakerProfile.ProfileId,
			vaultWallet,
			amount,
			txhash,
			exchangeId,
			chainId,
		},
	)

	return ops, err
}

func (api *ApiModel) GetClaimWithdrawalResponse(profileId uint, signer *signer.WithdrawalSigner, bopsId string, exchangeId string) (*WithdrawalResponse, error) {
	withdrawal, err := api.ClaimWithdrawal(context.Background(), profileId, bopsId)
	if err != nil || withdrawal == nil {
		logrus.Errorf("CLAIM_WITHDRAWAL_FAILED: %s", err.Error())
		return nil, err
	}

	opsIdIntStr := StripPrefix(withdrawal.OpsId, 2)
	bigIntId, success := new(big.Int).SetString(opsIdIntStr, 10)
	if !success {
		err = fmt.Errorf("CLAIM_WITHDRAWAL_FAILED_BAD_ID: %s", opsIdIntStr)
		logrus.Errorf(err.Error())
		return nil, err
	}

	_decimals, ok := eidToDecimals[exchangeId]
	if !ok {
		return nil, fmt.Errorf("NO_DECIMALS in the map for exchangeId=%s", exchangeId)
	}

	if err := validateDecimals(_decimals, exchangeId); err != nil {
		return nil, err
	}

	bigIntAmount := withdrawal.Amount.Mul(_decimals).Round(0).BigInt()
	r, s, v, err := signer.Sign(bigIntId, withdrawal.Wallet, bigIntAmount)
	if err != nil {
		logrus.Errorf("SIGN_CLAIM_WITHDRAWAL_FAILED: %s", err.Error())
		return nil, err
	}

	R := fmt.Sprintf("0x%x", r)
	S := fmt.Sprintf("0x%x", s)

	resp := &WithdrawalResponse{
		Withdrawal: withdrawal,
		BnAmount:   bigIntAmount.String(),
		R:          R,
		S:          S,
		V:          v,
	}
	return resp, nil
}

func (api *ApiModel) ClaimWithdrawal(ctx context.Context, profileId uint, bopsId string) (*BalanceOps, error) {
	res, err := DataResponse[*BalanceOps]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		CLAIM_WITHDRAWAL,
		[]interface{}{
			profileId,
			bopsId,
		},
	)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (api *ApiModel) CreateDeposit(ctx context.Context, profileId uint, wallet string, amount *tdecimal.Decimal, txhash string, exchangeId string, chainId uint) (*BalanceOps, error) {
	exchangeId = strings.ToLower(exchangeId)

	ops, err := DataResponse[*BalanceOps]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		CREATE_DEPOSIT,
		[]interface{}{
			profileId,
			wallet,
			amount,
			txhash,
			exchangeId,
			chainId,
		},
	)

	return ops, err
}

func (api *ApiModel) ProcessDeposit(ctx context.Context, profileId uint, deposit Deposit, isPoolDeposit bool) error {

	_, err := DataResponse[*BalanceOps]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		PROCESS_DEPOSIT,
		[]interface{}{
			profileId,
			deposit.Wallet,
			deposit.Id,
			deposit.Amount,
			deposit.Tx,
			isPoolDeposit,
			deposit.ExchangeId,
			deposit.ChainId,
			deposit.ExchangeAddress,
		},
	)

	return err
}

func (api *ApiModel) ProcessDepositUnknown(ctx context.Context, deposit Deposit) error {

	_, err := DataResponse[*BalanceOps]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		PROCESS_DEPOSIT_UNKNOWN,
		[]interface{}{
			deposit.Wallet,
			deposit.Id,
			deposit.Amount,
			deposit.Tx,
			deposit.ExchangeId,
			deposit.ChainId,
			deposit.ExchangeAddress,
		},
	)

	return err
}

func (api *ApiModel) ProcessYield(ctx context.Context, yield Yield) error {

	_, err := DataResponse[*BalanceOps]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		PROCESS_YIELD,
		[]interface{}{
			yield.Id,
			yield.Amount,
			yield.Tx,
			yield.ExchangeId,
			yield.ChainId,
			yield.ExchangeAddress,
		},
	)

	return err
}

func (api *ApiModel) InitVault(ctx context.Context, vaultProfileId uint, managerProfileId uint, treasurerProfileId uint, performanceFee float64) error {

	d_performanceFee := tdecimal.NewDecimal(decimal.NewFromFloat(performanceFee))

	_, err := DataResponse[*VaultInfo]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		INIT_VAULT,
		[]interface{}{
			vaultProfileId,
			managerProfileId,
			treasurerProfileId,
			d_performanceFee,
		},
	)

	return err
}

func (api *ApiModel) ReactivateVault(ctx context.Context, vaultProfileId uint) error {
	_, err := DataResponse[*VaultInfo]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		REACTIVATE_VAULT,
		[]interface{}{
			vaultProfileId,
		},
	)

	return err
}

func (api *ApiModel) ProcessStake(ctx context.Context, stakerProfileId uint, stake Stake, fromBalance bool, exchange_id string) (*BalanceOps, error) {

	ops, err := DataResponse[*BalanceOps]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		PROCESS_STAKE,
		[]interface{}{
			stakerProfileId,
			stake.VaultProfileId,
			stake.VaultWallet,
			stake.Id,
			stake.Amount,
			stake.CurrentNav,
			stake.Tx,
			fromBalance,
			exchange_id,
		},
	)

	return ops, err
}

func (api *ApiModel) CreateWithdrawal(ctx context.Context, profileId uint, wallet string, amount *tdecimal.Decimal, exchangeId string) (*BalanceOps, error) {

	ops, err := DataResponse[*BalanceOps]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		CREATE_WITHDRAWAL,
		[]interface{}{
			profileId,
			wallet,
			amount,
			exchangeId,
		},
	)

	return ops, err
}

func (api *ApiModel) CreateUnstake(ctx context.Context, stakerProfileId uint, vaultProfileId uint, vaultWallet string, shares *tdecimal.Decimal, exchangeId string, chainId uint) (*BalanceOps, error) {
	ops, err := DataResponse[*BalanceOps]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		CREATE_UNSTAKE,
		[]interface{}{
			stakerProfileId,
			vaultProfileId,
			vaultWallet,
			shares,
			exchangeId,
			chainId,
		},
	)

	return ops, err
}

func (api *ApiModel) GetVaultInfo(ctx context.Context, vaultProfileId uint) (*VaultInfo, error) {
	vlt, err := DataResponse[*VaultInfo]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		GET_VAULT_INFO,
		[]interface{}{
			vaultProfileId,
		},
	)

	return vlt, err
}

func (api *ApiModel) GetVaultHoldingInfo(ctx context.Context, vaultProfileId uint, stakerProfileId uint) (*VaultHoldingInfo, error) {
	holding, err := DataResponse[*VaultHoldingInfo]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		GET_HOLDING_INFO,
		[]interface{}{
			vaultProfileId,
			stakerProfileId,
		},
	)

	return holding, err
}

func (api *ApiModel) CancelUnstake(ctx context.Context, profileId uint, bopsId string) (bool, error) {
	res, err := DataResponse[bool]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		CANCEL_UNSTAKE,
		[]interface{}{
			profileId,
			bopsId,
		},
	)
	return res, err
}

func (api *ApiModel) ProcessUnstakes(ctx context.Context, vaultProfileId uint, fromId uint, toId uint, currentNav *tdecimal.Decimal, withdrawableBalance *tdecimal.Decimal, performanceFee *tdecimal.Decimal, treasurerProfileId uint, totalShares *tdecimal.Decimal, exchangeId string) ([]uint, error) {
	res, err := DataResponse[[]uint]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		PROCESS_UNSTAKES,
		[]interface{}{
			vaultProfileId,
			fromId,
			toId,
			currentNav,
			withdrawableBalance,
			performanceFee,
			treasurerProfileId,
			totalShares,
			exchangeId,
		},
	)
	return res, err
}

func (api *ApiModel) GetVaultManagerProfileId(ctx context.Context, vaultProfileId uint) (uint, error) {
	res, err := DataResponse[uint]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		GET_VAULT_MANAGER_PROFILE_ID,
		[]interface{}{
			vaultProfileId,
		},
	)
	return res, err
}

func (api *ApiModel) PendingDepositCanceled(ctx context.Context, opsId string) (bool, error) {
	res, err := DataResponse[bool]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		PENDING_DEPOSIT_CANCELED,
		[]interface{}{
			opsId,
		},
	)
	return res, err
}

func (api *ApiModel) GetPendingDeposits(ctx context.Context, exchangeId string, chainId uint) ([]*BalanceOps, error) {
	exchangeId = strings.ToLower(exchangeId)

	res, err := DataResponse[[]*BalanceOps]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		GET_PENDING_DEPOSITS,
		[]interface{}{exchangeId, chainId},
	)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (api *ApiModel) GetPendingStakes(ctx context.Context, exchangeId string, chainId uint) ([]*BalanceOps, error) {
	exchangeId = strings.ToLower(exchangeId)

	res, err := DataResponse[[]*BalanceOps]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		GET_PENDING_STAKES,
		[]interface{}{exchangeId, chainId},
	)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (api *ApiModel) GetPendingWithdrawals(ctx context.Context, exchangeId string, chainId uint) ([]*BalanceOps, error) {
	exchangeId = strings.ToLower(exchangeId)

	res, err := DataResponse[[]*BalanceOps]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		GET_PENDING_WITHDRAWALS,
		[]interface{}{exchangeId, chainId},
	)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (api *ApiModel) GetAllPendingWithdrawals(ctx context.Context) ([]*BalanceOps, error) {
	res, err := DataResponse[[]*BalanceOps]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		GET_ALL_PENDING_WITHDRAWALS,
		[]interface{}{},
	)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (api *ApiModel) CompletedWithdrawals(ctx context.Context, ids []*WithdrawalTxInfo) error {
	_, err := DataResponse[[]*WithdrawalTxInfo]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		COMPLETED_WITHDRAWALS,
		[]interface{}{ids},
	)
	return err
}

func (api *ApiModel) GetWithdrawalsSuspended(ctx context.Context) (bool, error) {
	res, err := DataResponse[bool]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		GET_WITHDRAWALS_SUSPENDED,
		[]interface{}{},
	)
	return res, err
}

func (api *ApiModel) SuspendWithdrawals(ctx context.Context) error {
	_, err := DataResponse[[]interface{}]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		SUSPEND_WITHDRAWALS,
		[]interface{}{},
	)
	return err
}

func (api *ApiModel) UpdatePendingWithdrawals(ctx context.Context, currentBlock *big.Int, future_block *big.Int, for_contract string) error {
	for_contract = strings.ToLower(for_contract)
	_, err := DataResponse[string]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		UPDATE_PENDING_WITHDRAWALS,
		[]interface{}{currentBlock.Uint64(), future_block.Uint64(), for_contract},
	)
	return err
}

func (api *ApiModel) ProcessingWithdrawal(ctx context.Context, profileId uint, txhash string, bopsId string) error {
	_, err := DataResponse[string]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		PROCESSING_WITHDRAWAL,
		[]interface{}{profileId, txhash, bopsId},
	)
	return err
}

func (api *ApiModel) GetLastProcessedBlockNumber(ctx context.Context, forContract string, chainId uint, eventType string) (*big.Int, error) {
	forContract = strings.ToLower(forContract)
	eventType = strings.ToLower(eventType)

	res, err := DataResponse[string]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		GET_LAST_PROCESSED_BLOCK_NUMBER,
		[]interface{}{forContract, chainId, eventType},
	)
	if err != nil {
		return nil, err
	}
	blockNum, success := new(big.Int).SetString(res, 10)
	if success {
		return blockNum, err
	} else {
		return nil, fmt.Errorf("eventType=%s forContract=%s chainId=%d unable to parse last processed block number: res=%s",
			eventType, forContract, chainId, res)
	}
}

func (api *ApiModel) SetLastProcessedBlockNumber(ctx context.Context, lastProcessed *big.Int, forContract string, chainId uint, eventType string) error {
	forContract = strings.ToLower(forContract)
	eventType = strings.ToLower(eventType)

	_, err := DataResponse[[]interface{}]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		SET_LAST_PROCESSED_BLOCK_NUMBER,
		[]interface{}{lastProcessed.String(), forContract, chainId, eventType},
	)
	return err
}

// Only for debug purpose: never use in prod

func (api *ApiModel) GetSettlementState(ctx context.Context) (bool, error) {
	res, err := DataResponse[bool]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		GET_SETTLEMENT_STATE,
		[]interface{}{},
	)

	if err != nil {
		return false, err
	}

	return res, nil
}

func (api *ApiModel) GetProcessingOps(ctx context.Context) ([]*BalanceOps, error) {
	res, err := DataResponse[[]*BalanceOps]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		GET_PROCESSING_OPS,
		[]interface{}{},
	)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (api *ApiModel) MergeUnknown(ctx context.Context) (uint, error) {
	total, err := DataResponse[uint]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		MERGE_UNKNOWN_OPS,
		[]interface{}{},
	)

	return total, err
}

func (api *ApiModel) DeleteUnknown(ctx context.Context) (uint, error) {
	total, err := DataResponse[uint]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		DELETE_UNKNOWN_OPS,
		[]interface{}{},
	)

	return total, err
}

func (api *ApiModel) AddContractMap(ctx context.Context, contract_address string, chain_id uint, exchange_id string) (*ContractMap, error) {
	contract_address = strings.ToLower(contract_address)
	exchange_id = strings.ToLower(exchange_id)

	res, err := DataResponse[*ContractMap]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		ADD_CONTRACT_MAP,
		[]interface{}{contract_address, chain_id, exchange_id},
	)

	return res, err
}

func (api *ApiModel) MaxWithdrawAmount(ctx context.Context) (uint, error) {

	amount, err := DataResponse[uint]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		MAX_WITHDRAW_AMOUNT,
		[]interface{}{},
	)

	return amount, err
}

func (api *ApiModel) Rolling24hWds(ctx context.Context) (*tdecimal.Decimal, error) {

	amount, err := DataResponse[*tdecimal.Decimal]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		ROLLING_24_WDS,
		[]interface{}{},
	)

	return amount, err
}
