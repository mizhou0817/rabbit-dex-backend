package vaultdata

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/strips-finance/rabbit-dex-backend/api/types"
	"github.com/strips-finance/rabbit-dex-backend/model"
	"github.com/strips-finance/rabbit-dex-backend/portfolio"
	"strings"
)

var (
	ZERO = decimal.NewFromInt(0)
)

type Wallet string

type ExchangeId string

type VaultProfile struct {
	ID          int64
	ProfileType string
}

type VaultHistoryRequest struct {
	VaultWallet string `form:"vault_wallet" binding:"required"`
	Range       string `form:"range" binding:"oneof=1h 1d 1w 1m 1y all"`
}

type VaultHoldingsRequest struct {
	VaultWallet string `form:"vault_wallet"`
}

type VaultHoldingsResponse struct {
	StakerProfileId    uint64 `json:"staker_profile_id"`
	VaultProfileId     uint64 `json:"vault_profile_id"`
	Wallet             string `json:"wallet"`
	ExchangeId         string `json:"exchange_id"`
	Status             string `json:"status"`
	VaultName          string `json:"vault_name"`
	ManagerName        string `json:"manager_name"`
	InceptionTimestamp int64  `json:"inception_timestamp"`

	Shares            decimal.Decimal `json:"shares"`
	UserNav           decimal.Decimal `json:"user_nav"`
	NetWithdrawable   decimal.Decimal `json:"net_withdrawable"`
	PerformanceCharge decimal.Decimal `json:"performance_charge"`
	PerformanceFee    decimal.Decimal `json:"performance_fee"`
}

type VaultBalanceOperationsRequest struct {
	VaultWallet string   `form:"vault_wallet"`
	OpsType     []string `form:"ops_type"`
}

type VaultBalanceOperationsResponse struct {
	Id                 string `json:"id"`
	OpsType            string `json:"ops_type"`
	OpsSubType         string `json:"ops_sub_type"`
	StakerProfileId    uint64 `json:"staker_profile_id"`
	VaultProfileId     uint64 `json:"vault_profile_id"`
	Wallet             string `json:"wallet"`
	ExchangeId         string `json:"exchange_id"`
	Status             string `json:"status"`
	VaultName          string `json:"vault_name"`
	ManagerName        string `json:"manager_name"`
	InceptionTimestamp int64  `json:"inception_timestamp"`
	Timestamp          int64  `json:"timestamp"`

	StakeUSDT   decimal.Decimal `json:"stake_usdt"`
	StakeShares decimal.Decimal `json:"stake_shares"`

	UnstakeShares  decimal.Decimal `json:"unstake_shares"`
	UnstakeUSDT    decimal.Decimal `json:"unstake_usdt"`
	UnstakeFeeUSDT decimal.Decimal `json:"unstake_fee_usdt"`
}

type VaultRequest struct {
	VaultWallet string `form:"vault_wallet"`
}

type VaultResponse struct {
	VaultProfileId     uint64          `json:"vault_profile_id"`
	Wallet             string          `json:"wallet"`
	ExchangeId         string          `json:"exchange_id"`
	AccountEquity      decimal.Decimal `json:"account_equity"`
	TotalShares        decimal.Decimal `json:"total_shares"`
	SharePrice         decimal.Decimal `json:"share_price"`
	APY                decimal.Decimal `json:"apy"`
	Status             string          `json:"status"`
	PerformanceFee     decimal.Decimal `json:"performance_fee"`
	ManagerName        string          `json:"manager_name"`
	VaultName          string          `json:"vault_name"`
	InceptionTimestamp int64           `json:"inception_timestamp"`
}

func split(value string) []string {
	result := strings.Split(value, ",")
	if len(result) == 1 && result[0] == "" {
		return nil
	}
	return result
}

func splitSlice(result []string) []string {
	if len(result) == 1 && result[0] == "" {
		return nil
	}
	return result
}

func HandleVaultHoldings(ctx context.Context, db *pgxpool.Pool, request VaultHoldingsRequest, profileId uint, exchangeId string, pagination types.PaginationRequestParams) ([]VaultHoldingsResponse, *types.PaginationResponse, error) {
	filterVaultWalletIds := split(request.VaultWallet)
	filterVaultWalletIds = model.GetWalletsStringInRabbitTntStandardFormat(filterVaultWalletIds)

	qSelect := `SELECT 
		vh.staker_profile_id,
		v.vault_profile_id,
		p.wallet,
		p.exchange_id,
		v.status,
		v.vault_name,
		v.manager_name,
		v.initialised_at,
		vh.shares,
		vh.entry_nav,
		vh.entry_price,
		c.account_equity,
		v.total_shares,
		v.performance_fee
`
	qFrom := `
	FROM app_profile as p
	JOIN app_vault_last as v on v.vault_profile_id = p.id
	JOIN app_profile_cache_last as c on c.id=v.vault_profile_id
	JOIN app_vault_holdings_last vh on vh.vault_profile_id=v.vault_profile_id
	WHERE vh.staker_profile_id = @staker_profile_id 
	AND p.profile_type = 'vault' AND p.exchange_id = @exchange_id
	`

	args := pgx.NamedArgs{
		"staker_profile_id": profileId,
		"exchange_id":       exchangeId,
	}

	if len(filterVaultWalletIds) > 0 {
		qFrom += " AND p.wallet = ANY(@wallets) "
		args["wallets"] = filterVaultWalletIds
	}

	paginationResponse := &types.PaginationResponse{
		Limit: pagination.Limit,
		Page:  pagination.Page,
		Order: pagination.Order,
	}
	totalQuery := `SELECT COUNT(*) ` + qFrom
	err := db.QueryRow(ctx, totalQuery, args).Scan(&paginationResponse.Total)
	if err != nil {
		return nil, nil, errors.Wrap(err, "execute total query")
	}

	selectQuery := qSelect + qFrom +
		" ORDER BY v.vault_profile_id " + pagination.Order +
		" LIMIT @limit OFFSET @offset"

	args["limit"] = pagination.Limit
	args["offset"] = pagination.Limit * pagination.Page

	rows, err := db.Query(ctx, selectQuery, args)
	if err != nil {
		return nil, nil, errors.Wrap(err, "execute query")
	}

	defer rows.Close()

	response := make([]VaultHoldingsResponse, 0)
	for rows.Next() {
		var entryNav decimal.Decimal
		var entryPrice decimal.Decimal
		var accountEquity decimal.Decimal
		var totalShares decimal.Decimal
		var performanceFee decimal.Decimal

		var r VaultHoldingsResponse
		err = rows.Scan(
			&r.StakerProfileId,
			&r.VaultProfileId,
			&r.Wallet,
			&r.ExchangeId,
			&r.Status,
			&r.VaultName,
			&r.ManagerName,
			&r.InceptionTimestamp,
			&r.Shares,
			&entryNav,
			&entryPrice,
			&accountEquity,
			&totalShares,
			&performanceFee,
		)
		if err != nil {
			return nil, nil, errors.Wrap(err, "scan row")
		}

		var performanceCharge decimal.Decimal
		if totalShares.Cmp(ZERO) > 0 {
			r.UserNav = accountEquity.Mul(r.Shares).Div(totalShares)

			currentPrice := accountEquity.Div(totalShares)
			if currentPrice.GreaterThan(entryPrice) && currentPrice.Cmp(ZERO) > 0 && entryPrice.Cmp(ZERO) > 0 {
				performanceCharge =
					performanceFee.Mul(r.UserNav).Mul(currentPrice.Sub(entryPrice)).Div(currentPrice)
			} else {
				performanceCharge = ZERO
			}

			r.NetWithdrawable = r.UserNav.Sub(performanceCharge)
			if r.NetWithdrawable.Cmp(ZERO) < 0 {
				r.NetWithdrawable = ZERO
			}
		} else {
			r.UserNav = accountEquity
			r.NetWithdrawable = r.UserNav
			performanceCharge = ZERO
		}

		r.PerformanceFee = performanceFee
		r.PerformanceCharge = performanceCharge

		response = append(response, r)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, errors.Wrap(err, "rows error")
	}

	return response, paginationResponse, nil
}

func HandleVaultBalanceOperations(ctx context.Context, db *pgxpool.Pool, request VaultBalanceOperationsRequest, profileId uint, exchangeId string, pagination types.PaginationRequestParams) ([]VaultBalanceOperationsResponse, *types.PaginationResponse, error) {
	filterVaultWalletIds := split(request.VaultWallet)
	filterVaultWalletIds = model.GetWalletsStringInRabbitTntStandardFormat(filterVaultWalletIds)
	filterTypes := splitSlice(request.OpsType)
	filterTypes = normalizeTypes(filterTypes)

	qSelect := `SELECT
		op.id,
		op.ops_type,
		op.ops_sub_type,
		op.staker_profile_id,
		op.vault_profile_id,
		op.vault_wallet,
		op.vault_exchange_id,
		op.status as op_status,
		v.vault_name,
		v.manager_name,
		v.initialised_at,
		op.timestamp,
		op.stake_usdt,
		op.stake_shares,
		op.unstake_shares,
		op.unstake_usdt,
		op.unstake_fee_usdt
	`

	qFrom := `
	FROM app_vault_balance_operation_last as op
	JOIN app_profile as p on p.wallet = op.vault_wallet AND p.exchange_id = op.vault_exchange_id
	JOIN app_vault_last as v on v.vault_profile_id = p.id
	WHERE op.staker_profile_id = @staker_profile_id 
	AND p.profile_type = 'vault' AND p.exchange_id = @exchange_id
	`

	args := pgx.NamedArgs{
		"staker_profile_id": profileId,
		"exchange_id":       exchangeId,
	}

	if len(filterVaultWalletIds) > 0 {
		qFrom += " AND op.vault_wallet = ANY(@wallets) "
		args["wallets"] = filterVaultWalletIds
	}
	if len(filterTypes) > 0 {
		qFrom += " AND op.ops_type = ANY(@ops_type) "
		args["ops_type"] = filterTypes
	}

	paginationResponse := &types.PaginationResponse{
		Limit: pagination.Limit,
		Page:  pagination.Page,
		Order: pagination.Order,
	}
	totalQuery := `SELECT COUNT(*) ` + qFrom
	err := db.QueryRow(ctx, totalQuery, args).Scan(&paginationResponse.Total)
	if err != nil {
		return nil, nil, errors.Wrap(err, "execute total query")
	}

	selectQuery := qSelect + qFrom +
		" ORDER BY op.timestamp " + pagination.Order +
		" LIMIT @limit OFFSET @offset"

	args["limit"] = pagination.Limit
	args["offset"] = pagination.Limit * pagination.Page

	rows, err := db.Query(ctx, selectQuery, args)
	if err != nil {
		return nil, nil, errors.Wrap(err, "execute query")
	}

	defer rows.Close()

	response := make([]VaultBalanceOperationsResponse, 0)
	for rows.Next() {
		var r VaultBalanceOperationsResponse
		err = rows.Scan(
			&r.Id,
			&r.OpsType,
			&r.OpsSubType,
			&r.StakerProfileId,
			&r.VaultProfileId,
			&r.Wallet,
			&r.ExchangeId,
			&r.Status,
			&r.VaultName,
			&r.ManagerName,
			&r.InceptionTimestamp,
			&r.Timestamp,
			&r.StakeUSDT,
			&r.StakeShares,
			&r.UnstakeShares,
			&r.UnstakeUSDT,
			&r.UnstakeFeeUSDT,
		)
		if err != nil {
			return nil, nil, errors.Wrap(err, "scan row")
		}

		response = append(response, r)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, errors.Wrap(err, "rows error")
	}

	return response, paginationResponse, nil
}

func normalizeTypes(values []string) []string {
	result := make([]string, len(values))
	for i, w := range values {
		result[i] = normalizeType(w)
	}
	return result
}

func normalizeType(value string) string {
	return strings.ToLower(value)
}

func HandleVault(ctx context.Context, db *pgxpool.Pool, request VaultRequest, exchangeId string, pagination types.PaginationRequestParams) ([]VaultResponse, *types.PaginationResponse, error) {
	filterVaultWalletIds := split(request.VaultWallet)
	filterVaultWalletIds = model.GetWalletsStringInRabbitTntStandardFormat(filterVaultWalletIds)

	qSelect := `SELECT
    	v.vault_profile_id,
    	p.wallet,
    	p.exchange_id,
    	c.account_equity,
    	v.total_shares,
    	COALESCE(a.share_price,1) as share_price,
    	COALESCE(a.apy_total,0) as apy_total,
    	v.status,
    	v.performance_fee,
    	v.manager_name,
		v.initialised_at,
    	v.vault_name
`
	qFrom := `
	FROM app_profile as p
	JOIN app_vault_last as v on v.vault_profile_id = p.id
	JOIN app_profile_cache_last as c on c.id=v.vault_profile_id
	LEFT JOIN app_vault_aggregate_last as a on a.vault_profile_id = v.vault_profile_id
	WHERE p.profile_type = 'vault' AND p.exchange_id = @exchange_id
	`

	args := pgx.NamedArgs{
		"exchange_id": exchangeId,
	}
	if len(filterVaultWalletIds) > 0 {
		qFrom += " AND p.wallet = ANY(@wallets) "
		args["wallets"] = filterVaultWalletIds
	}

	paginationResponse := &types.PaginationResponse{
		Limit: pagination.Limit,
		Page:  pagination.Page,
		Order: pagination.Order,
	}
	totalQuery := `SELECT COUNT(*) ` + qFrom
	err := db.QueryRow(ctx, totalQuery, args).Scan(&paginationResponse.Total)
	if err != nil {
		return nil, nil, errors.Wrap(err, "execute total query")
	}

	selectQuery := qSelect + qFrom +
		" ORDER BY v.vault_profile_id " + pagination.Order +
		" LIMIT @limit OFFSET @offset"

	args["limit"] = pagination.Limit
	args["offset"] = pagination.Limit * pagination.Page

	rows, err := db.Query(ctx, selectQuery, args)
	if err != nil {
		return nil, nil, errors.Wrap(err, "execute query")
	}

	defer rows.Close()

	response := make([]VaultResponse, 0)
	for rows.Next() {
		var r VaultResponse
		err = rows.Scan(
			&r.VaultProfileId,
			&r.Wallet,
			&r.ExchangeId,
			&r.AccountEquity,
			&r.TotalShares,
			&r.SharePrice,
			&r.APY,
			&r.Status,
			&r.PerformanceFee,
			&r.ManagerName,
			&r.InceptionTimestamp,
			&r.VaultName,
		)
		if err != nil {
			return nil, nil, errors.Wrap(err, "scan row")
		}
		response = append(response, r)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, errors.Wrap(err, "rows error")
	}

	return response, paginationResponse, nil
}

func HandleNavHistory(ctx context.Context, db *pgxpool.Pool, request VaultHistoryRequest, exchangeId string) ([]portfolio.PortfolioData, error) {
	vaultWallet := model.GetWalletStringInRabbitTntStandardFormat(request.VaultWallet)

	vaultProfile, err := GetVaultProfile(ctx, db, Wallet(vaultWallet), ExchangeId(exchangeId))
	if err != nil {
		return nil, fmt.Errorf(
			"GET_VAULT_PROFILE_ERROR: wallet=%s, err=%s",
			vaultWallet,
			err.Error(),
		)
	}
	if vaultProfile == nil {
		return nil, fmt.Errorf(
			"VAULT_PROFILE_NOT_FOUND: wallet=%s",
			vaultWallet,
		)
	}
	if vaultProfile.ProfileType != model.PROFILE_TYPE_VAULT {
		return nil, fmt.Errorf("TARGET_IS_NOT_A_VAULT %s", vaultWallet)
	}

	portfolioRequest := portfolio.PortfolioRequest{
		Range: request.Range,
	}
	return portfolio.HandlePortfolioList(ctx, db, portfolioRequest, uint(vaultProfile.ID))
}

func GetVaultProfile(ctx context.Context, db *pgxpool.Pool, vaultWallet Wallet, exchangeId ExchangeId) (*VaultProfile, error) {
	q := fmt.Sprintf(`SELECT
    		p.id,
    		p.profile_type
			FROM app_profile as p
			WHERE p.exchange_id = @exchange_id AND p.wallet = @wallet`)

	args := pgx.NamedArgs{
		"exchange_id": exchangeId,
		"wallet":      vaultWallet,
	}
	rows, err := db.Query(ctx, q, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []VaultProfile
	for rows.Next() {
		var r VaultProfile
		err = rows.Scan(
			&r.ID,
			&r.ProfileType)

		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, nil
	}
	if len(result) > 1 {
		return nil, fmt.Errorf("found more than 1 profile by wallet %v and exchange id %s", vaultWallet, exchangeId)
	}
	return &result[0], nil
}
