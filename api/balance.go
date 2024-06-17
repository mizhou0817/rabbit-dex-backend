package api

import (
	"errors"
	"fmt"
	"math"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/strips-finance/rabbit-dex-backend/model"
	"github.com/strips-finance/rabbit-dex-backend/tdecimal"
	"github.com/strips-finance/rabbit-dex-backend/tick"
)

type DepositRequest struct {
	TxHash string  `json:"txhash" binding:"required"`
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

type BalanceOpsIdRequest struct {
	Id string `json:"id" binding:"required"`
}

type WithdrawalRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

type ProcessingWithdrawalRequest struct {
	Id     string `json:"id" binding:"required"`
	TxHash string `json:"txhash" binding:"required"`
}

type InitVaultRequest struct {
	VaultWallet     string  `json:"vault_wallet" binding:"required"`
	ManagerWallet   string  `json:"manager_wallet" binding:"required"`
	TreasurerWallet string  `json:"treasurer_wallet" binding:"required"`
	PerformanceFee  float64 `json:"performance_fee" binding:"required"`
}

type ReactivateVaultRequest struct {
	VaultWallet string `json:"vault_wallet" binding:"required"`
}

type StakeRequest struct {
	VaultWallet string  `json:"vault_wallet" binding:"required"`
	TxHash      string  `json:"txhash" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
}

type UnstakeRequest struct {
	VaultWallet string  `json:"vault_wallet" binding:"required"`
	Shares      float64 `json:"shares" binding:"required,gt=0"`
}

type ProcessUnstakesRequest struct {
	VaultWallet string `json:"vault_wallet" binding:"required"`
	FromId      uint   `json:"from_id" binding:"omitempty,min=0"`
	ToId        uint   `json:"to_id" binding:"omitempty,min=0"`
}

var (
	ZERO = decimal.NewFromInt(0)
	ONE  = decimal.NewFromInt(1)
)

func HandleDeposit(c *gin.Context) {
	var request DepositRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	//CHECKPROFILE
	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	//TODO: make it more general somewhere in router
	if ctx.Profile != nil && ctx.Profile.Type == model.PROFILE_TYPE_VAULT {
		ErrorResponse(c, fmt.Errorf("NOT_ALLOWED_FOR_VAULT"))
		return
	}

	rounded_amount := tick.RoundDownToUsdtTick(request.Amount)

	amount := tdecimal.NewDecimal(decimal.NewFromFloat(rounded_amount))
	res, err := apiModel.CreateDeposit(c.Request.Context(),
		ctx.Profile.ProfileId,
		ctx.Profile.Wallet,
		amount,
		request.TxHash,
		ctx.ExchangeId,
		ctx.ExchangeCfg.ChainId,
	)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	//3. InvalidateCache
	_, err = apiModel.InvalidateCacheAndNotify(c.Request.Context(), ctx.Profile.ProfileId)
	if err != nil {
		logrus.Error(err)
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, res)
}

func HandleStake(c *gin.Context) {
	var request StakeRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)
	stakerProfile := ctx.Profile
	apiModel := model.NewApiModel(ctx.Broker)

	var err error

	_, err = apiModel.AcquireStakeLock(c.Request.Context(), ctx.Profile.ProfileId)
	if err != nil {
		ErrorResponse(c, errors.New("STAKE_UNAVAILABLE_LOCK_ACQUIRE_ERROR"))
		return
	}
	defer func() {
		_, err := apiModel.ReleaseStakeLock(c.Request.Context(), ctx.Profile.ProfileId)
		if err != nil {
			logrus.Error(err)
		}
	}()

	// check the staker's profile type
	if stakerProfile.Type == model.PROFILE_TYPE_VAULT {
		err = fmt.Errorf("NOT_ALLOWED_FOR_VAULT")
		logrus.Error(err)
		ErrorResponse(c, err)
		return
	}

	// check the vaultWallet profile
	vaultWallet := model.GetWalletStringInRabbitTntStandardFormat(request.VaultWallet)
	vaultProfile, err := apiModel.GetProfileByWalletForExchangeId(c.Request.Context(), vaultWallet, ctx.ExchangeId)
	if err != nil {
		logrus.Error(err)
		ErrorResponse(c, fmt.Errorf("GET_VAULT_PROFILE_ERROR %s", err.Error()))
		return
	}
	if vaultProfile == nil || vaultProfile.Type != model.PROFILE_TYPE_VAULT {
		err = fmt.Errorf("STAKE_TARGET_IS_NOT_A_VAULT %s", vaultWallet)
		logrus.Error(err)
		ErrorResponse(c, err)
		return
	}
	bop, err := apiModel.CreateStake(
		c.Request.Context(),
		stakerProfile,
		vaultWallet,
		request.Amount,
		request.TxHash,
		ctx.ExchangeId,
		ctx.ExchangeCfg.ChainId,
	)

	if err != nil {
		logrus.Error(err)
		ErrorResponse(c, errors.New("CREATE_STAKE_ERROR"))
		return
	}

	_, err = apiModel.InvalidateCacheAndNotify(c.Request.Context(), ctx.Profile.ProfileId)
	if err != nil {
		logrus.Error(err)
		ErrorResponse(c, errors.New("STAKE_CACHE_ERROR"))
		return
	}

	SuccessResponse(c, bop)
}

func HandleStakeFromBalance(c *gin.Context) {
	var request StakeRequest
	var err error
	if err = c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)
	stakerProfile := ctx.Profile
	apiModel := model.NewApiModel(ctx.Broker)

	_, err = apiModel.AcquireWithdrawLock(c.Request.Context(), ctx.Profile.ProfileId)
	if err != nil {
		ErrorResponse(c, errors.New("STAKE_UNAVAILABLE_LOCK_ACQUIRE_ERROR"))
		return
	}
	defer func() {
		_, err := apiModel.ReleaseWithdrawLock(c.Request.Context(), ctx.Profile.ProfileId)
		if err != nil {
			logrus.Error(err)
		}
	}()

	amount := decimal.NewFromFloat(request.Amount)
	stakerCache, err := apiModel.InvalidateCache(c.Request.Context(), stakerProfile.ProfileId)
	if err != nil {
		ErrorResponse(c, err)
		return
	}
	if stakerCache.WithdrawableBalance.LessThan(amount) {
		ErrorResponse(c, errors.New("NOT_ENOUGH_FUNDS"))
	}
	vaultWallet := model.GetWalletStringInRabbitTntStandardFormat(request.VaultWallet)
	var vaultProfile *model.Profile
	vaultProfile, err = apiModel.GetProfileByWalletForExchangeId(c.Request.Context(), vaultWallet, ctx.ExchangeId)
	if err != nil {
		ErrorResponse(c, err)
		return
	}
	if vaultProfile == nil || vaultProfile.Type != model.PROFILE_TYPE_VAULT {
		ErrorResponse(c, fmt.Errorf("STAKE_TARGET_IS_NOT_A_VAULT, wallet=%s", vaultWallet))
	}

	vaultCache, err := apiModel.InvalidateCache(c.Request.Context(), vaultProfile.ProfileId)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	stake := model.Stake{
		Id:             "",
		VaultProfileId: vaultProfile.ProfileId,
		VaultWallet:    vaultWallet,
		Amount:         tdecimal.NewDecimal(amount),
		CurrentNav:     vaultCache.AccountEquity,
		Tx:             "",
	}

	bop, err := apiModel.ProcessStake(c.Request.Context(), stakerProfile.ProfileId, stake, true, ctx.ExchangeId)

	if err != nil {
		logrus.Error(err)
		ErrorResponse(c, errors.New("STAKE_FROM_BALANCE_ERROR"))
		return
	}

	_, err = apiModel.InvalidateCacheAndNotify(c.Request.Context(), ctx.Profile.ProfileId)
	if err != nil {
		logrus.Error(err)
		ErrorResponse(c, errors.New("STAKE_CACHE_ERROR"))
		return
	}

	SuccessResponse(c, bop)
}

func BeginUnstake(c *gin.Context) {
	var request UnstakeRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)
	unstakerProfile := ctx.Profile
	apiModel := model.NewApiModel(ctx.Broker)

	shares := tdecimal.NewDecimal(decimal.NewFromFloat(request.Shares))
	goCtx := c.Request.Context()

	_, err := apiModel.AcquireUnstakeLock(c.Request.Context(), ctx.Profile.ProfileId)
	if err != nil {
		ErrorResponse(c, errors.New("UNSTAKE_UNAVAILABLE_LOCK_ACQUIRE_ERROR"))
		return
	}
	defer func() {
		_, err := apiModel.ReleaseUnstakeLock(c.Request.Context(), ctx.Profile.ProfileId)
		if err != nil {
			logrus.Error(err)
		}
	}()

	wallet := model.GetWalletStringInRabbitTntStandardFormat(request.VaultWallet)
	vaultProfile, err := apiModel.GetProfileByWalletForExchangeId(c.Request.Context(), wallet, ctx.ExchangeId)
	if err != nil {
		logrus.Error(err)
		ErrorResponse(c, fmt.Errorf(
			"GET_VAULT_PROFILE_ERROR %s, wallet=%s, exchange_id=%s",
			err.Error(),
			wallet,
			ctx.ExchangeId,
		))
		return
	}
	if vaultProfile == nil || vaultProfile.Type != model.PROFILE_TYPE_VAULT {
		err = fmt.Errorf("UNSTAKE_TARGET_IS_NOT_A_VAULT %s", wallet)
		logrus.Error(err)
		ErrorResponse(c, err)
		return
	}

	res, err := apiModel.CreateUnstake(
		goCtx,
		unstakerProfile.ProfileId,
		vaultProfile.ProfileId,
		wallet,
		shares,
		ctx.ExchangeId,
		ctx.ExchangeCfg.ChainId)

	if err == nil {
		SuccessResponse(c, res)
	} else {
		ErrorResponse(c, err)
	}

}

func InitVault(c *gin.Context) {
	var request InitVaultRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	vaultWallet := model.GetWalletStringInRabbitTntStandardFormat(request.VaultWallet)
	vaultProfile, err := apiModel.GetProfileByWalletForExchangeId(c.Request.Context(), vaultWallet, ctx.ExchangeId)
	if err != nil {
		logrus.Error(err)
		ErrorResponse(c, fmt.Errorf("GET_VAULT_PROFILE_ERROR %s", err.Error()))
		return
	}
	if vaultProfile == nil || vaultProfile.Type != model.PROFILE_TYPE_VAULT {
		err = fmt.Errorf("INIT_TARGET_IS_NOT_A_VAULT %s", vaultWallet)
		logrus.Error(err)
		ErrorResponse(c, err)
		return
	}

	if vaultProfile.ProfileId != ctx.Profile.ProfileId {
		ErrorResponse(c, fmt.Errorf(
			"NOT_AUTHORISED_TO_INIT_VAULT profile_id=%d, vault profile_id=%d",
			ctx.Profile.ProfileId,
			vaultProfile.ProfileId,
		))
		return
	}

	managerWallet := model.GetWalletStringInRabbitTntStandardFormat(request.ManagerWallet)
	managerProfile, err := apiModel.GetProfileByWalletForExchangeId(c.Request.Context(), managerWallet, ctx.ExchangeId)
	if err != nil {
		logrus.Error(err)
		ErrorResponse(c, fmt.Errorf("GET_VAULT_MANAGER_PROFILE_ERROR %s", err.Error()))
		return
	}
	if managerProfile == nil {
		err = fmt.Errorf("VAULT_MANAGER_PROFILE_NOT_FOUND %s", managerWallet)
		logrus.Error(err)
		ErrorResponse(c, err)
		return
	}

	treasurerWallet := model.GetWalletStringInRabbitTntStandardFormat(request.TreasurerWallet)
	treasurerProfile, err := apiModel.GetProfileByWalletForExchangeId(c.Request.Context(), treasurerWallet, ctx.ExchangeId)
	if err != nil {
		logrus.Error(err)
		ErrorResponse(c, fmt.Errorf("GET_VAULT_TREASURER_PROFILE_ERROR %s", err.Error()))
		return
	}
	if treasurerProfile == nil {
		err = fmt.Errorf("VAULT_TREASURER_PROFILE_NOT_FOUND %s", treasurerWallet)
		logrus.Error(err)
		ErrorResponse(c, err)
		return
	}

	err = apiModel.InitVault(c.Request.Context(), ctx.Profile.ProfileId, managerProfile.ProfileId, treasurerProfile.ProfileId, request.PerformanceFee)
	if err != nil {
		logrus.Error(fmt.Errorf("INIT_VAULT_FAILED: %s", err.Error()))
		ErrorResponse(c, fmt.Errorf("INIT_VAULT_FAILED: %s", err.Error()))
		return
	}
	SuccessResponse(c, "")
}

func ReactivateVault(c *gin.Context) {
	var request ReactivateVaultRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	vaultWallet := model.GetWalletStringInRabbitTntStandardFormat(request.VaultWallet)
	vaultProfile, err := apiModel.GetProfileByWalletForExchangeId(c.Request.Context(), vaultWallet, ctx.ExchangeId)
	if err != nil {
		logrus.Error(err)
		ErrorResponse(c, fmt.Errorf("GET_VAULT_PROFILE_ERROR %s", err.Error()))
		return
	}
	if vaultProfile == nil || vaultProfile.Type != model.PROFILE_TYPE_VAULT {
		err = fmt.Errorf("REACTIVATE_TARGET_IS_NOT_A_VAULT %s", vaultWallet)
		logrus.Error(err)
		ErrorResponse(c, err)
		return
	}
	vaultProfileId := vaultProfile.ProfileId

	// vault_manager_profile_id, err := apiModel.GetVaultManagerProfileId(
	// 	c.Request.Context(),
	// 	vaultProfile.ProfileId,
	// )
	// if err != nil {
	// 	ErrorResponse(
	// 		c,
	// 		fmt.Errorf("GET_VAULT_MANAGER_ERROR %s", err.Error()),
	// 	)
	// 	return
	// }
	// if vault_manager_profile_id != ctx.Profile.ProfileId {
	// 	ErrorResponse(
	// 		c,
	// 		fmt.Errorf("NOT_VAULT_MANAGER profile_id=%d", ctx.Profile.ProfileId),
	// 	)
	// 	return
	// }

	if vaultProfileId != ctx.Profile.ProfileId {
		ErrorResponse(
			c,
			fmt.Errorf("NOT_VAULT_MANAGER profile_id=%d", ctx.Profile.ProfileId),
		)
		return
	}

	vaultInfo, err := apiModel.GetVaultInfo(c.Request.Context(), vaultProfile.ProfileId)
	if err != nil {
		logrus.Error(err)
		ErrorResponse(
			c,
			fmt.Errorf("GET_VAULT_INFO_ERROR %s", err.Error()),
		)
		return
	}

	if vaultInfo == nil {
		logrus.Error("VAULT_PROFILE_NOT_FOUND")
		ErrorResponse(
			c,
			errors.New("VAULT_PROFILE_NOT_FOUND"),
		)
		return
	}

	if vaultInfo.Status == model.VAULT_STATUS_ACTIVE {
		logrus.Error("VAULT_ALREADY_ACTIVE")
		ErrorResponse(
			c,
			fmt.Errorf("VAULT_ALREADY_ACTIVE %d", vaultProfile.ProfileId),
		)
		return
	}

	err = apiModel.ReactivateVault(c.Request.Context(), ctx.Profile.ProfileId)
	if err != nil {
		logrus.Error(fmt.Errorf("REACTIVATE_VAULT_FAILED: %s", err.Error()))
		ErrorResponse(c, fmt.Errorf("REACTIVATE_VAULT_FAILED: %s", err.Error()))
	} else {
		SuccessResponse(c, "")
	}
}

func ProcessUnstakes(c *gin.Context) {
	var request ProcessUnstakesRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	vaultWallet := model.GetWalletStringInRabbitTntStandardFormat(request.VaultWallet)
	vaultProfile, err := apiModel.GetProfileByWalletForExchangeId(c.Request.Context(), vaultWallet, ctx.ExchangeId)
	if err != nil {
		logrus.Error(err)
		ErrorResponse(c, fmt.Errorf("GET_VAULT_PROFILE_ERROR %s", err.Error()))
		return
	}
	if vaultProfile == nil || vaultProfile.Type != model.PROFILE_TYPE_VAULT {
		err = fmt.Errorf("PROCESS_UNSTAKES_TARGET_IS_NOT_A_VAULT %s", vaultWallet)
		logrus.Error(err)
		ErrorResponse(c, err)
		return
	}
	vaultProfileId := vaultProfile.ProfileId

	// vault_manager_profile_id, err := apiModel.GetVaultManagerProfileId(
	// 	c.Request.Context(),
	// 	vaultProfile.ProfileId,
	// )
	// if err != nil {
	// 	ErrorResponse(
	// 		c,
	// 		fmt.Errorf("PROCESS_UNSTAKES_VAULT_MANAGER_ERROR %s", err.Error()),
	// 	)
	// 	return
	// }
	// if vault_manager_profile_id != ctx.Profile.ProfileId {
	if vaultProfileId != ctx.Profile.ProfileId {
		ErrorResponse(
			c,
			fmt.Errorf("NOT_VAULT_MANAGER profile_id=%d", ctx.Profile.ProfileId),
		)
		return
	}

	_, err = apiModel.AcquireWithdrawLock(c.Request.Context(), vaultProfile.ProfileId)
	if err != nil {
		ErrorResponse(
			c,
			errors.New("PROCESS_UNSTAKES_UNAVAILABLE_LOCK_ACQUIRE_ERROR"),
		)
		return
	}
	defer func() {
		_, err := apiModel.ReleaseWithdrawLock(c.Request.Context(), vaultProfile.ProfileId)
		if err != nil {
			logrus.Error(err)
		}
	}()

	vaultInfo, err := apiModel.GetVaultInfo(c.Request.Context(), vaultProfile.ProfileId)
	if err != nil {
		logrus.Error(err)
		ErrorResponse(
			c,
			fmt.Errorf("GET_VAULT_INFO_ERROR %s", err.Error()),
		)
		return
	}

	if vaultInfo == nil {
		logrus.Error("VAULT_PROFILE_NOT_FOUND")
		ErrorResponse(
			c,
			errors.New("VAULT_PROFILE_NOT_FOUND"),
		)
		return
	}

	if vaultInfo.Status != model.VAULT_STATUS_ACTIVE {
		logrus.Error("VAULT_NOT_ACTIVE")
		ErrorResponse(
			c,
			fmt.Errorf("VAULT_NOT_ACTIVE %d", vaultProfile.ProfileId),
		)
		return
	}

	if vaultInfo.PerformanceFee.Cmp(ONE) >= 0 ||
		vaultInfo.PerformanceFee.Cmp(ZERO) < 0 {
		err = fmt.Errorf(
			"WRONG_PERFORMANCE_FEE %s for vault %d",
			vaultInfo.PerformanceFee.String(),
			vaultProfile.ProfileId,
		)
		logrus.Error(err)
		ErrorResponse(c, err)
		return
	}

	vaultCache, err := apiModel.InvalidateCache(c.Request.Context(), vaultProfile.ProfileId)
	if err != nil {
		logrus.Error(fmt.Errorf("PROCESS_UNSTAKES_CACHE_ERROR: %s", err.Error()))
		ErrorResponse(c, fmt.Errorf("PROCESS_UNSTAKES_CACHE_ERROR: %s", err.Error()))
		return
	}

	ids, err := apiModel.ProcessUnstakes(c.Request.Context(), vaultProfile.ProfileId, request.FromId, request.ToId, vaultCache.AccountEquity, vaultCache.WithdrawableBalance, &vaultInfo.PerformanceFee, vaultInfo.TreasurerProfileId, &vaultInfo.TotalShares, ctx.ExchangeId)
	if err != nil {
		logrus.Error(fmt.Errorf("PROCESS_UNSTAKES_FAILED: %s", err.Error()))
		ErrorResponse(c, fmt.Errorf("PROCESS_UNSTAKES_FAILED: %s", err.Error()))
	} else {
		apiModel.InvalidateCacheAndNotify(c.Request.Context(), vaultProfile.ProfileId)

		for _, pid := range ids {
			apiModel.InvalidateCacheAndNotify(c.Request.Context(), pid)
		}
		SuccessResponse(c, "")
	}
}

func CancelUnstake(c *gin.Context) {
	var request BalanceOpsIdRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	_, err := apiModel.AcquireWithdrawLock(c.Request.Context(), ctx.Profile.ProfileId)
	if err != nil {
		ErrorResponse(
			c,
			errors.New("CANCEL_UNSTAKE_UNAVAILABLE_LOCK_ACQUIRE_ERROR"),
		)
		return
	}
	defer func() {
		_, err := apiModel.ReleaseWithdrawLock(c.Request.Context(), ctx.Profile.ProfileId)
		if err != nil {
			logrus.Error(err)
		}
	}()

	_, err = apiModel.CancelUnstake(c.Request.Context(), ctx.Profile.ProfileId, request.Id)
	if err != nil {
		logrus.Error(fmt.Errorf("CANCEL_UNSTAKE_FAILED: %s", err.Error()))
		ErrorResponse(c, fmt.Errorf("CANCEL_UNSTAKE_FAILED: %s", err.Error()))
		return
	}

	_, err = apiModel.InvalidateCacheAndNotify(c.Request.Context(), ctx.Profile.ProfileId)
	if err != nil {
		logrus.Error(err)
		ErrorResponse(c, errors.New("UNSTAKE_CACHE_ERROR"))
		return
	}

	SuccessResponse(c, "")
}

func HandleRequestedUnstakesList(c *gin.Context) {
	ctx := GetRabbitContext(c)
	if ctx.Profile == nil || ctx.Profile.Type != model.PROFILE_TYPE_VAULT {
		ErrorResponse(c, fmt.Errorf("ONLY_ALLOWED_FOR_VAULT"))
		return
	}
	apiModel := model.NewApiModel(ctx.Broker)

	res, err := apiModel.GetRequestedUnstakes(c.Request.Context(), ctx.Profile.ProfileId)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, res...)
}

/*
Withdraw for the user. Critical function
1. Need to acquire withdraw lock: even if we have several api instances, lock is managed by 1 tarantool profile instance
2. check that withdraw is allowed for profile_id: If any pending withdraws, then it's not allowed
3. Invalidate profile cache
4. Check conditions: inv3, withdrawble balance
5. make withdraw
6. release lock
*/
func HandleWithdrawal(c *gin.Context) {
	var request WithdrawalRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	if ctx.Profile.Status != model.PROFILE_STATUS_ACTIVE {
		ErrorResponse(c, errors.New("PROFILE_NOT_ACTIVE"))
		return
	}

	//1. Acquire withdraw Lock
	_, err := apiModel.AcquireWithdrawLock(c.Request.Context(), ctx.Profile.ProfileId)
	if err != nil {
		ErrorResponse(c, errors.New("WITHDRAW_UNAVAILABLE_LOCK_ACQUIRE_ERROR"))
		return
	}
	defer func() {
		_, err := apiModel.ReleaseWithdrawLock(c.Request.Context(), ctx.Profile.ProfileId)
		if err != nil {
			logrus.Error(err)
		}
	}()

	//2. Check that no pending withdraw exist
	//CHECK withdraw allowed
	is_allowed := apiModel.CheckWithdrawAllowed(c.Request.Context(), ctx.Profile.ProfileId)
	if !is_allowed {
		ErrorResponse(c, errors.New("WITHDRAW_UNAVAILABLE_PENDING_EXIST"))
		return
	}

	//3. Receive cache
	cache, err := apiModel.InvalidateCache(c.Request.Context(), ctx.Profile.ProfileId)
	if err != nil {
		logrus.Error(err)
		ErrorResponse(c, err)
		return
	} else if cache == nil {
		ErrorResponse(c, errors.New("WITHDRAW_UNAVAILABLE_CACHE_NOT_EXIST"))
		return
	}

	valid, err := apiModel.CachedIsInv3Valid(c.Request.Context(), 0)
	if err != nil {
		ErrorResponse(c, err)
		return
	}
	if !valid {
		ErrorResponse(c, errors.New("INV3_BROKEN"))
		return
	}

	if request.Amount <= 0 {
		ErrorResponse(c, errors.New("WRONG_AMOUNT"))
		return
	}

	if cache.WithdrawableBalance.LessThanOrEqual(decimal.NewFromInt(0)) {
		ErrorResponse(c, errors.New("NOT_ENOUGH_FUNDS"))
		return
	}

	available := math.Min(cache.WithdrawableBalance.InexactFloat64(), request.Amount)
	available = tick.RoundDownToUsdtTick(available)

	if cache.WithdrawableBalance.LessThan(decimal.NewFromFloat(available)) {
		ErrorResponse(c, errors.New("NOT_ENOUGH_FUNDS"))
		return
	}

	// profile_id uint, market_id, order_type, side string, price, size float64
	amount := tdecimal.NewDecimal(decimal.NewFromFloat(available))
	res, err := apiModel.CreateWithdrawal(c.Request.Context(),
		ctx.Profile.ProfileId,
		ctx.Profile.Wallet,
		amount,
		ctx.ExchangeId,
	)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	_, err = apiModel.InvalidateCacheAndNotify(c.Request.Context(), ctx.Profile.ProfileId)
	if err != nil {
		logrus.Error(err)
		ErrorResponse(c, errors.New("WITHDRAW_CACHE_ERROR"))
		return
	}

	SuccessResponse(c, res)
}

func ProcessingWithdrawal(c *gin.Context) {
	var request ProcessingWithdrawalRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	//TODO: make it more general somewhere in router
	if ctx.Profile != nil && ctx.Profile.Type == model.PROFILE_TYPE_VAULT {
		ErrorResponse(c, fmt.Errorf("NOT_ALLOWED_FOR_VAULT"))
		return
	}

	err := apiModel.ProcessingWithdrawal(c.Request.Context(), ctx.Profile.ProfileId, request.TxHash, request.Id)
	if err != nil {
		logrus.Error(fmt.Errorf("PROCESSING_WITHDRAWAL_FAILED: %s", err.Error()))
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, "")
}

func CancelWithdrawal(c *gin.Context) {
	var request BalanceOpsIdRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	_, err := apiModel.AcquireWithdrawLock(c.Request.Context(), ctx.Profile.ProfileId)
	if err != nil {
		ErrorResponse(c, errors.New("CANCEL_WITHDRAWAL_UNAVAILABLE_LOCK_ACQUIRE_ERROR"))
		return
	}
	defer func() {
		_, err := apiModel.ReleaseWithdrawLock(c.Request.Context(), ctx.Profile.ProfileId)
		if err != nil {
			logrus.Error(err)
		}
	}()

	_, err = apiModel.CancelWithdrawal(c.Request.Context(), ctx.Profile.ProfileId, request.Id)
	if err != nil {
		logrus.Error(fmt.Errorf("CANCEL_WITHDRAWAL_FAILED: %s", err.Error()))
		ErrorResponse(c, err)
		return
	}

	_, err = apiModel.InvalidateCache(c.Request.Context(), ctx.Profile.ProfileId)
	if err != nil {
		logrus.Error(err)
		ErrorResponse(c, errors.New("CANCEL_WITHDRAWAL_UNAVAILABLE_CACHE_ERROR"))
		return
	}

	valid, err := apiModel.CachedIsInv3Valid(c.Request.Context(), 0)
	if err != nil {
		ErrorResponse(c, err)
		return
	}
	if !valid {
		ErrorResponse(c, errors.New("INV3_BROKEN"))
		return
	}

	_, err = apiModel.InvalidateCacheAndNotify(c.Request.Context(), ctx.Profile.ProfileId)
	if err != nil {
		logrus.Error(err)
		ErrorResponse(c, errors.New("CANCEL_WITHDRAWAL_UNAVAILABLE_CACHE_ERROR"))
		return
	}

	SuccessResponse(c, "")
}

func ClaimWithdrawal(c *gin.Context) {
	var request BalanceOpsIdRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	_, err := apiModel.AcquireWithdrawLock(c.Request.Context(), ctx.Profile.ProfileId)
	if err != nil {
		ErrorResponse(c, errors.New("CLAIM_WITHDRAWAL_UNAVAILABLE_LOCK_ACQUIRE_ERROR"))
		return
	}
	defer func() {
		_, err := apiModel.ReleaseWithdrawLock(c.Request.Context(), ctx.Profile.ProfileId)
		if err != nil {
			logrus.Error(err)
		}
	}()

	resp, err := apiModel.GetClaimWithdrawalResponse(ctx.Profile.ProfileId, ctx.Signer, request.Id, ctx.ExchangeId)
	if err != nil {
		ErrorResponse(c, err)
	} else {
		SuccessResponse(c, resp)
	}
}

func stripPrefix(input string, charsToRemove int) string {
	asRunes := []rune(input)
	return string(asRunes[charsToRemove:])
}
