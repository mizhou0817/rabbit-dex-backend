package vaultdata

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/strips-finance/rabbit-dex-backend/api/types"
	"github.com/strips-finance/rabbit-dex-backend/portfolio"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/strips-finance/rabbit-dex-backend/dbtestsuite"
)

const (
	wallet            = "0x1a05a1507c35c763035fdb151af9286d4d90a81b"
	traderWallet      = "0x1a05a1507c35c763035fdb151af9286d4d90a81c"
	vaultWallet2      = "0x1a05a1507c35c763035fdb151af9286d4d90a81d"
	exchangeId        = "rbx"
	treasurerID       = uint64(0)
	vaultID1          = uint64(1)
	traderID          = uint64(2)
	vaultID2          = uint64(3)
	vaultProfileType  = "vault"
	traderProfileType = "trader"
)

var tables = []string{
	"app_profile",
	"app_vault_last",
	"app_profile_cache_last",
	"app_vault_holdings_last",
	"app_vault_aggregate_last",
	"app_vault_balance_operation_last",
	"app_balance_operation",
	"app_vault_aggregate_last",
	"app_vault_aggregate_history",
	"app_profile_cache",
	"app_profile_cache_1m",
	"app_profile_cache_15m",
	"app_profile_cache_30m",
	"app_profile_cache_1h",
	"app_profile_cache_1d",
	"app_profile_cache_1d",
}

type dbTestSuite struct {
	dbtestsuite.DBTestSuite
}

func Test_dbTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timescaledb integration test")
	}
	testSuite := new(dbTestSuite)
	suite.Run(t, testSuite)
}

func (s *dbTestSuite) SetupSuite() {
	s.BaseSetupSuite()
	ApplyTestMigrations(s.T(), s.MigrationConnectionString())
}

func (s *dbTestSuite) TearDownSuite() {
	s.BaseTearDownSuite()
}

func (s *dbTestSuite) SetupTest() {
	s.DeleteTables(tables)
}

func (s *dbTestSuite) TestHandleVaultHoldings() {
	performanceFee := decimal.NewFromFloat(0.1)
	status := "status1"
	vaultName := "vaultName1"
	managerName := "managerName1"
	initialisedAt := time.Now()
	accountEquity := decimal.NewFromFloat(2000)
	accountEquity2 := decimal.NewFromFloat(500)
	shares := decimal.NewFromFloat(10)
	totalShares := decimal.NewFromFloat(100)
	entryPrice := decimal.NewFromFloat(1)
	entryNav := decimal.NewFromFloat(1000)
	entryNav2 := decimal.NewFromFloat(1000)
	userNav := decimal.NewFromFloat(200)
	userNav2 := decimal.NewFromFloat(50)
	performanceCharge := decimal.NewFromFloat(19)
	performanceCharge2 := decimal.NewFromFloat(4)
	netWithdrawable := decimal.NewFromFloat(181)
	netWithdrawable2 := decimal.NewFromFloat(46)

	insertCommon := func() {
		s.insertProfile(vaultID1, vaultProfileType, wallet, exchangeId)
		s.insertVaultLast(vaultID1, performanceFee, totalShares, status, vaultName, managerName, initialisedAt)
		s.insertProfileCacheLast(vaultID1, vaultProfileType, wallet, accountEquity)
		s.insertVaultHoldingsLast(vaultID1, traderID, shares, entryNav, entryPrice)

		s.insertProfile(traderID, traderProfileType, traderWallet, exchangeId)
		s.insertVaultLast(traderID, performanceFee, totalShares, status, vaultName, managerName, initialisedAt)
		s.insertProfileCacheLast(traderID, traderProfileType, traderWallet, accountEquity)

		s.insertProfile(vaultID2, vaultProfileType, vaultWallet2, exchangeId)
		s.insertVaultLast(vaultID2, performanceFee, totalShares, status, vaultName, managerName, initialisedAt)
		s.insertProfileCacheLast(vaultID2, vaultProfileType, vaultWallet2, accountEquity2)
		s.insertVaultHoldingsLast(vaultID2, traderID, shares, entryNav2, entryPrice)
	}

	type args struct {
		request    VaultHoldingsRequest
		profileId  uint64
		exchangeId string
		pagination *types.PaginationRequestParams
	}
	tests := []struct {
		name           string
		args           args
		insertData     func()
		want           []VaultHoldingsResponse
		wantPagination *types.PaginationResponse
		wantErr        bool
	}{
		{
			name: "should return one response for first wallet",
			args: args{
				request: VaultHoldingsRequest{
					VaultWallet: wallet,
				},
				profileId:  traderID,
				exchangeId: exchangeId,
				pagination: &types.PaginationRequestParams{
					Page:  0,
					Limit: 50,
					Order: "ASC",
				},
			},
			insertData: insertCommon,
			want: []VaultHoldingsResponse{
				{
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID1,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					VaultName:          vaultName,
					ManagerName:        managerName,
					InceptionTimestamp: initialisedAt.UnixMicro(),
					Status:             status,
					Shares:             shares,
					UserNav:            userNav,
					NetWithdrawable:    netWithdrawable,
					PerformanceCharge:  performanceCharge,
					PerformanceFee:     performanceFee,
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 1,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
		{
			name: "should return net_withdrawable = 0, when net_withdrawable < 0",
			args: args{
				request: VaultHoldingsRequest{
					VaultWallet: wallet,
				},
				profileId:  traderID,
				exchangeId: exchangeId,
				pagination: &types.PaginationRequestParams{
					Page:  0,
					Limit: 50,
					Order: "ASC",
				},
			},
			insertData: func() {
				s.insertProfile(vaultID1, vaultProfileType, wallet, exchangeId)
				s.insertVaultLast(vaultID1, decimal.NewFromFloat(2.0), totalShares, status, vaultName, managerName, initialisedAt)
				s.insertProfileCacheLast(vaultID1, vaultProfileType, wallet, accountEquity)
				s.insertVaultHoldingsLast(vaultID1, traderID, shares, entryNav, entryPrice)
			},
			want: []VaultHoldingsResponse{
				{
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID1,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					VaultName:          vaultName,
					ManagerName:        managerName,
					InceptionTimestamp: initialisedAt.UnixMicro(),
					Status:             status,
					Shares:             shares,
					UserNav:            userNav,
					NetWithdrawable:    ZERO,
					PerformanceCharge:  decimal.NewFromFloat(380.0),
					PerformanceFee:     decimal.NewFromFloat(2.0),
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 1,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
		{
			name: "should return UserNav=accountEquity, when totalShares = 0",
			args: args{
				request: VaultHoldingsRequest{
					VaultWallet: wallet,
				},
				profileId:  traderID,
				exchangeId: exchangeId,
				pagination: &types.PaginationRequestParams{
					Page:  0,
					Limit: 50,
					Order: "ASC",
				},
			},
			insertData: func() {
				s.insertProfile(vaultID1, vaultProfileType, wallet, exchangeId)
				s.insertVaultLast(vaultID1, performanceFee, ZERO, status, vaultName, managerName, initialisedAt)
				s.insertProfileCacheLast(vaultID1, vaultProfileType, wallet, accountEquity)
				s.insertVaultHoldingsLast(vaultID1, traderID, shares, entryNav, entryPrice)
			},
			want: []VaultHoldingsResponse{
				{
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID1,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					VaultName:          vaultName,
					ManagerName:        managerName,
					InceptionTimestamp: initialisedAt.UnixMicro(),
					Status:             status,
					Shares:             shares,
					UserNav:            accountEquity,
					NetWithdrawable:    accountEquity,
					PerformanceCharge:  ZERO,
					PerformanceFee:     performanceFee,
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 1,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
		{
			name: "should return UserNav=accountEquity, when totalShares = -1",
			args: args{
				request: VaultHoldingsRequest{
					VaultWallet: wallet,
				},
				profileId:  traderID,
				exchangeId: exchangeId,
				pagination: &types.PaginationRequestParams{
					Page:  0,
					Limit: 50,
					Order: "ASC",
				},
			},
			insertData: func() {
				s.insertProfile(vaultID1, vaultProfileType, wallet, exchangeId)
				s.insertVaultLast(vaultID1, performanceFee, decimal.NewFromFloat(-1.0), status, vaultName, managerName, initialisedAt)
				s.insertProfileCacheLast(vaultID1, vaultProfileType, wallet, accountEquity)
				s.insertVaultHoldingsLast(vaultID1, traderID, shares, entryNav, entryPrice)
			},
			want: []VaultHoldingsResponse{
				{
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID1,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					VaultName:          vaultName,
					ManagerName:        managerName,
					InceptionTimestamp: initialisedAt.UnixMicro(),
					Status:             status,
					Shares:             shares,
					UserNav:            accountEquity,
					NetWithdrawable:    accountEquity,
					PerformanceCharge:  ZERO,
					PerformanceFee:     performanceFee,
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 1,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
		{
			name: "should return PerformanceCharge=0, when currentPrice < entryPrice",
			args: args{
				request: VaultHoldingsRequest{
					VaultWallet: wallet,
				},
				profileId:  traderID,
				exchangeId: exchangeId,
				pagination: &types.PaginationRequestParams{
					Page:  0,
					Limit: 50,
					Order: "ASC",
				},
			},
			insertData: func() {
				s.insertProfile(vaultID1, vaultProfileType, wallet, exchangeId)
				s.insertVaultLast(vaultID1, performanceFee, totalShares, status, vaultName, managerName, initialisedAt)
				s.insertProfileCacheLast(vaultID1, vaultProfileType, wallet, accountEquity)
				s.insertVaultHoldingsLast(vaultID1, traderID, shares, entryNav, decimal.NewFromFloat(1000.0))
			},
			want: []VaultHoldingsResponse{
				{
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID1,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					VaultName:          vaultName,
					ManagerName:        managerName,
					InceptionTimestamp: initialisedAt.UnixMicro(),
					Status:             status,
					Shares:             shares,
					UserNav:            userNav,
					NetWithdrawable:    userNav,
					PerformanceCharge:  ZERO,
					PerformanceFee:     performanceFee,
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 1,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
		{
			name: "should return PerformanceCharge=0, when currentPrice = entryPrice",
			args: args{
				request: VaultHoldingsRequest{
					VaultWallet: wallet,
				},
				profileId:  traderID,
				exchangeId: exchangeId,
				pagination: &types.PaginationRequestParams{
					Page:  0,
					Limit: 50,
					Order: "ASC",
				},
			},
			insertData: func() {
				s.insertProfile(vaultID1, vaultProfileType, wallet, exchangeId)
				s.insertVaultLast(vaultID1, performanceFee, totalShares, status, vaultName, managerName, initialisedAt)
				s.insertProfileCacheLast(vaultID1, vaultProfileType, wallet, accountEquity)
				s.insertVaultHoldingsLast(vaultID1, traderID, shares, entryNav, decimal.NewFromFloat(20.0))
			},
			want: []VaultHoldingsResponse{
				{
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID1,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					VaultName:          vaultName,
					ManagerName:        managerName,
					InceptionTimestamp: initialisedAt.UnixMicro(),
					Status:             status,
					Shares:             shares,
					UserNav:            userNav,
					NetWithdrawable:    userNav,
					PerformanceCharge:  ZERO,
					PerformanceFee:     performanceFee,
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 1,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
		{
			name: "should return PerformanceCharge=0, when currentPrice = 0 and entryPrice = -1",
			args: args{
				request: VaultHoldingsRequest{
					VaultWallet: wallet,
				},
				profileId:  traderID,
				exchangeId: exchangeId,
				pagination: &types.PaginationRequestParams{
					Page:  0,
					Limit: 50,
					Order: "ASC",
				},
			},
			insertData: func() {
				s.insertProfile(vaultID1, vaultProfileType, wallet, exchangeId)
				s.insertVaultLast(vaultID1, performanceFee, totalShares, status, vaultName, managerName, initialisedAt)
				s.insertProfileCacheLast(vaultID1, vaultProfileType, wallet, ZERO)
				s.insertVaultHoldingsLast(vaultID1, traderID, shares, entryNav, decimal.NewFromFloat(-1.0))
			},
			want: []VaultHoldingsResponse{
				{
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID1,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					VaultName:          vaultName,
					ManagerName:        managerName,
					InceptionTimestamp: initialisedAt.UnixMicro(),
					Status:             status,
					Shares:             shares,
					UserNav:            ZERO,
					NetWithdrawable:    ZERO,
					PerformanceCharge:  ZERO,
					PerformanceFee:     performanceFee,
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 1,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
		{
			name: "should return PerformanceCharge=0, when accountEquity = -2000, currentPrice = -20 and entryPrice = -21",
			args: args{
				request: VaultHoldingsRequest{
					VaultWallet: wallet,
				},
				profileId:  traderID,
				exchangeId: exchangeId,
				pagination: &types.PaginationRequestParams{
					Page:  0,
					Limit: 50,
					Order: "ASC",
				},
			},
			insertData: func() {
				s.insertProfile(vaultID1, vaultProfileType, wallet, exchangeId)
				s.insertVaultLast(vaultID1, performanceFee, totalShares, status, vaultName, managerName, initialisedAt)
				s.insertProfileCacheLast(vaultID1, vaultProfileType, wallet, accountEquity.Neg())
				s.insertVaultHoldingsLast(vaultID1, traderID, shares, entryNav, decimal.NewFromFloat(-21.0))
			},
			want: []VaultHoldingsResponse{
				{
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID1,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					VaultName:          vaultName,
					ManagerName:        managerName,
					InceptionTimestamp: initialisedAt.UnixMicro(),
					Status:             status,
					Shares:             shares,
					UserNav:            decimal.NewFromFloat(-200.0),
					NetWithdrawable:    ZERO,
					PerformanceCharge:  ZERO,
					PerformanceFee:     performanceFee,
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 1,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
		{
			name: "should return one response for other wallet",
			args: args{
				request: VaultHoldingsRequest{
					VaultWallet: vaultWallet2,
				},
				profileId:  traderID,
				exchangeId: exchangeId,
				pagination: &types.PaginationRequestParams{
					Page:  0,
					Limit: 50,
					Order: "ASC",
				},
			},
			insertData: insertCommon,
			want: []VaultHoldingsResponse{
				{
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID2,
					Wallet:             vaultWallet2,
					ExchangeId:         exchangeId,
					VaultName:          vaultName,
					ManagerName:        managerName,
					InceptionTimestamp: initialisedAt.UnixMicro(),
					Status:             status,
					Shares:             shares,
					UserNav:            userNav2,
					NetWithdrawable:    netWithdrawable2,
					PerformanceCharge:  performanceCharge2,
					PerformanceFee:     performanceFee,
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 1,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
		{
			name: "should return all two response when VaultWallet is empty",
			args: args{
				request:    VaultHoldingsRequest{},
				profileId:  traderID,
				exchangeId: exchangeId,
				pagination: &types.PaginationRequestParams{
					Page:  0,
					Limit: 50,
					Order: "ASC",
				},
			},
			insertData: insertCommon,
			want: []VaultHoldingsResponse{
				{
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID1,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					VaultName:          vaultName,
					ManagerName:        managerName,
					InceptionTimestamp: initialisedAt.UnixMicro(),
					Status:             status,
					Shares:             shares,
					UserNav:            userNav,
					NetWithdrawable:    netWithdrawable,
					PerformanceCharge:  performanceCharge,
					PerformanceFee:     performanceFee,
				},
				{
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID2,
					Wallet:             vaultWallet2,
					ExchangeId:         exchangeId,
					VaultName:          vaultName,
					ManagerName:        managerName,
					InceptionTimestamp: initialisedAt.UnixMicro(),
					Status:             status,
					Shares:             shares,
					UserNav:            userNav2,
					NetWithdrawable:    netWithdrawable2,
					PerformanceCharge:  performanceCharge2,
					PerformanceFee:     performanceFee,
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 2,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
		{
			name: "should return first when page 0 limit 1",
			args: args{
				request:    VaultHoldingsRequest{},
				profileId:  traderID,
				exchangeId: exchangeId,
				pagination: &types.PaginationRequestParams{
					Page:  0,
					Limit: 1,
					Order: "ASC",
				},
			},
			insertData: insertCommon,
			want: []VaultHoldingsResponse{
				{
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID1,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					VaultName:          vaultName,
					ManagerName:        managerName,
					InceptionTimestamp: initialisedAt.UnixMicro(),
					Status:             status,
					Shares:             shares,
					UserNav:            userNav,
					NetWithdrawable:    netWithdrawable,
					PerformanceCharge:  performanceCharge,
					PerformanceFee:     performanceFee,
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 2,
				Page:  0,
				Limit: 1,
				Order: "ASC",
			},
		},
		{
			name: "should return second when page 1 limit 1",
			args: args{
				request:    VaultHoldingsRequest{},
				profileId:  traderID,
				exchangeId: exchangeId,
				pagination: &types.PaginationRequestParams{
					Page:  1,
					Limit: 1,
					Order: "ASC",
				},
			},
			insertData: insertCommon,
			want: []VaultHoldingsResponse{
				{
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID2,
					Wallet:             vaultWallet2,
					ExchangeId:         exchangeId,
					VaultName:          vaultName,
					ManagerName:        managerName,
					InceptionTimestamp: initialisedAt.UnixMicro(),
					Status:             status,
					Shares:             shares,
					UserNav:            userNav2,
					NetWithdrawable:    netWithdrawable2,
					PerformanceCharge:  performanceCharge2,
					PerformanceFee:     performanceFee,
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 2,
				Page:  1,
				Limit: 1,
				Order: "ASC",
			},
		},
		{
			name: "should return second when page 0 limit 1 order by DESC",
			args: args{
				request:    VaultHoldingsRequest{},
				profileId:  traderID,
				exchangeId: exchangeId,
				pagination: &types.PaginationRequestParams{
					Page:  0,
					Limit: 1,
					Order: "DESC",
				},
			},
			insertData: insertCommon,
			want: []VaultHoldingsResponse{
				{
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID2,
					Wallet:             vaultWallet2,
					ExchangeId:         exchangeId,
					VaultName:          vaultName,
					ManagerName:        managerName,
					InceptionTimestamp: initialisedAt.UnixMicro(),
					Status:             status,
					Shares:             shares,
					UserNav:            userNav2,
					NetWithdrawable:    netWithdrawable2,
					PerformanceCharge:  performanceCharge2,
					PerformanceFee:     performanceFee,
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 2,
				Page:  0,
				Limit: 1,
				Order: "DESC",
			},
		},
		{
			name: "should return zero responses when profileId not found",
			args: args{
				request: VaultHoldingsRequest{
					VaultWallet: wallet,
				},
				profileId:  uint64(0),
				exchangeId: exchangeId,
				pagination: &types.PaginationRequestParams{
					Page:  0,
					Limit: 50,
					Order: "ASC",
				},
			},
			insertData: insertCommon,
			want:       []VaultHoldingsResponse{},
			wantPagination: &types.PaginationResponse{
				Total: 0,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
		{
			name: "should return zero responses when wallet not found",
			args: args{
				request: VaultHoldingsRequest{
					VaultWallet: "WALLET_NOT_FOUND",
				},
				profileId:  traderID,
				exchangeId: exchangeId,
				pagination: &types.PaginationRequestParams{
					Page:  0,
					Limit: 50,
					Order: "ASC",
				},
			},
			insertData: insertCommon,
			want:       []VaultHoldingsResponse{},
			wantPagination: &types.PaginationResponse{
				Total: 0,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
		{
			name: "should return zero responses when exchangeId not found",
			args: args{
				request: VaultHoldingsRequest{
					VaultWallet: wallet,
				},
				profileId:  traderID,
				exchangeId: "EXCHANGE_ID_NOT_FOUND",
				pagination: &types.PaginationRequestParams{
					Page:  0,
					Limit: 50,
					Order: "ASC",
				},
			},
			insertData: insertCommon,
			want:       []VaultHoldingsResponse{},
			wantPagination: &types.PaginationResponse{
				Total: 0,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
		{
			name: "should return zero responses when profile type is not vault",
			args: args{
				request: VaultHoldingsRequest{
					VaultWallet: traderWallet,
				},
				profileId:  traderID,
				exchangeId: exchangeId,
				pagination: &types.PaginationRequestParams{
					Page:  0,
					Limit: 50,
					Order: "ASC",
				},
			},
			insertData: insertCommon,
			want:       []VaultHoldingsResponse{},
			wantPagination: &types.PaginationResponse{
				Total: 0,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.DeleteTables(tables)
			tt.insertData()

			pagination := types.PaginationRequestParams{
				Page:  0,
				Limit: 50,
				Order: "ASC",
			}
			if tt.args.pagination != nil {
				pagination = *tt.args.pagination
			}

			// when
			got, paginationResponse, err := HandleVaultHoldings(context.Background(), s.GetDB(), tt.args.request, uint(tt.args.profileId), tt.args.exchangeId, pagination)

			// then
			if tt.wantErr {
				s.NotNil(err)
				s.Nil(got)
				s.Nil(paginationResponse)
			} else {
				s.NoError(err)
				s.assertJSON(tt.want, got)
				s.Equal(tt.wantPagination, paginationResponse)
			}
		})
	}
}

func (s *dbTestSuite) TestHandleVault() {
	performanceFee := decimal.NewFromFloat(10.01)
	totalShares := decimal.NewFromFloat(10.25)
	status := "status1"
	vaultName := "vaultName1"
	managerName := "managerName1"
	initialisedAt := time.Now()
	accountEquity := decimal.NewFromFloat(102.50)
	sharePrice := decimal.NewFromFloat(10.02)
	apyTotal := decimal.NewFromFloat(10.03)

	s.insertProfile(vaultID1, vaultProfileType, wallet, exchangeId)
	s.insertVaultLast(vaultID1, performanceFee, totalShares, status, vaultName, managerName, initialisedAt)
	s.insertProfileCacheLast(vaultID1, vaultProfileType, wallet, accountEquity)
	s.insertVaultAggregateLast(vaultID1, sharePrice, apyTotal)

	s.insertProfile(traderID, traderProfileType, traderWallet, exchangeId)
	s.insertVaultLast(traderID, performanceFee, totalShares, status, vaultName, managerName, initialisedAt)
	s.insertProfileCacheLast(traderID, traderProfileType, traderWallet, accountEquity)

	s.insertProfile(vaultID2, vaultProfileType, vaultWallet2, exchangeId)
	s.insertVaultLast(vaultID2, performanceFee, totalShares, status, vaultName, managerName, initialisedAt)
	s.insertProfileCacheLast(vaultID2, vaultProfileType, vaultWallet2, accountEquity)

	type args struct {
		request    VaultRequest
		exchangeId string
		pagination *types.PaginationRequestParams
	}
	tests := []struct {
		name           string
		args           args
		want           []VaultResponse
		wantPagination *types.PaginationResponse
		wantErr        bool
	}{
		{
			name: "should return one response for first wallet",
			args: args{
				request: VaultRequest{
					VaultWallet: wallet,
				},
				exchangeId: exchangeId,
				pagination: &types.PaginationRequestParams{
					Page:  0,
					Limit: 50,
					Order: "ASC",
				},
			},
			want: []VaultResponse{
				{
					VaultProfileId:     vaultID1,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					AccountEquity:      accountEquity,
					TotalShares:        totalShares,
					SharePrice:         sharePrice,
					APY:                apyTotal,
					Status:             status,
					PerformanceFee:     performanceFee,
					ManagerName:        managerName,
					VaultName:          vaultName,
					InceptionTimestamp: initialisedAt.UnixMicro(),
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 1,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
		{
			name: "should return one response with zero share price and apy for other wallet when no record in app_vault_aggregate_last",
			args: args{
				request: VaultRequest{
					VaultWallet: vaultWallet2,
				},
				exchangeId: exchangeId,
				pagination: &types.PaginationRequestParams{
					Page:  0,
					Limit: 50,
					Order: "ASC",
				},
			},
			want: []VaultResponse{
				{
					VaultProfileId:     vaultID2,
					Wallet:             vaultWallet2,
					ExchangeId:         exchangeId,
					AccountEquity:      accountEquity,
					TotalShares:        totalShares,
					SharePrice:         decimal.NewFromFloat(1),
					APY:                decimal.NewFromFloat(0),
					Status:             status,
					PerformanceFee:     performanceFee,
					ManagerName:        managerName,
					VaultName:          vaultName,
					InceptionTimestamp: initialisedAt.UnixMicro(),
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 1,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
		{
			name: "should return all two response when VaultWallet is empty",
			args: args{
				request:    VaultRequest{},
				exchangeId: exchangeId,
				pagination: &types.PaginationRequestParams{
					Page:  0,
					Limit: 50,
					Order: "ASC",
				},
			},
			want: []VaultResponse{
				{
					VaultProfileId:     vaultID1,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					AccountEquity:      accountEquity,
					TotalShares:        totalShares,
					SharePrice:         sharePrice,
					APY:                apyTotal,
					Status:             status,
					PerformanceFee:     performanceFee,
					ManagerName:        managerName,
					VaultName:          vaultName,
					InceptionTimestamp: initialisedAt.UnixMicro(),
				},
				{
					VaultProfileId:     vaultID2,
					Wallet:             vaultWallet2,
					ExchangeId:         exchangeId,
					AccountEquity:      accountEquity,
					TotalShares:        totalShares,
					SharePrice:         decimal.NewFromFloat(1),
					APY:                decimal.NewFromFloat(0),
					Status:             status,
					PerformanceFee:     performanceFee,
					ManagerName:        managerName,
					VaultName:          vaultName,
					InceptionTimestamp: initialisedAt.UnixMicro(),
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 2,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
		{
			name: "should return first response when page 0 limit 1",
			args: args{
				request:    VaultRequest{},
				exchangeId: exchangeId,
				pagination: &types.PaginationRequestParams{
					Page:  0,
					Limit: 1,
					Order: "ASC",
				},
			},
			want: []VaultResponse{
				{
					VaultProfileId:     vaultID1,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					AccountEquity:      accountEquity,
					TotalShares:        totalShares,
					SharePrice:         sharePrice,
					APY:                apyTotal,
					Status:             status,
					PerformanceFee:     performanceFee,
					ManagerName:        managerName,
					VaultName:          vaultName,
					InceptionTimestamp: initialisedAt.UnixMicro(),
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 2,
				Page:  0,
				Limit: 1,
				Order: "ASC",
			},
		},
		{
			name: "should return second response when page 1 limit 1",
			args: args{
				request:    VaultRequest{},
				exchangeId: exchangeId,
				pagination: &types.PaginationRequestParams{
					Page:  1,
					Limit: 1,
					Order: "ASC",
				},
			},
			want: []VaultResponse{
				{
					VaultProfileId:     vaultID2,
					Wallet:             vaultWallet2,
					ExchangeId:         exchangeId,
					AccountEquity:      accountEquity,
					TotalShares:        totalShares,
					SharePrice:         decimal.NewFromFloat(1),
					APY:                decimal.NewFromFloat(0),
					Status:             status,
					PerformanceFee:     performanceFee,
					ManagerName:        managerName,
					VaultName:          vaultName,
					InceptionTimestamp: initialisedAt.UnixMicro(),
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 2,
				Page:  1,
				Limit: 1,
				Order: "ASC",
			},
		},
		{
			name: "should return second response when page 0 limit 1 order by DESC",
			args: args{
				request:    VaultRequest{},
				exchangeId: exchangeId,
				pagination: &types.PaginationRequestParams{
					Page:  0,
					Limit: 1,
					Order: "DESC",
				},
			},
			want: []VaultResponse{
				{
					VaultProfileId:     vaultID2,
					Wallet:             vaultWallet2,
					ExchangeId:         exchangeId,
					AccountEquity:      accountEquity,
					TotalShares:        totalShares,
					SharePrice:         decimal.NewFromFloat(1),
					APY:                decimal.NewFromFloat(0),
					Status:             status,
					PerformanceFee:     performanceFee,
					ManagerName:        managerName,
					VaultName:          vaultName,
					InceptionTimestamp: initialisedAt.UnixMicro(),
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 2,
				Page:  0,
				Limit: 1,
				Order: "DESC",
			},
		},
		{
			name: "should return zero responses when wallet not found",
			args: args{
				request: VaultRequest{
					VaultWallet: "WALLET_NOT_FOUND",
				},
				exchangeId: exchangeId,
				pagination: &types.PaginationRequestParams{
					Page:  0,
					Limit: 50,
					Order: "ASC",
				},
			},
			want: []VaultResponse{},
			wantPagination: &types.PaginationResponse{
				Total: 0,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
		{
			name: "should return zero responses when exchangeId not found",
			args: args{
				request: VaultRequest{
					VaultWallet: wallet,
				},
				exchangeId: "EXCHANGE_ID_NOT_FOUND",
				pagination: &types.PaginationRequestParams{
					Page:  0,
					Limit: 50,
					Order: "ASC",
				},
			},
			want: []VaultResponse{},
			wantPagination: &types.PaginationResponse{
				Total: 0,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
		{
			name: "should return zero responses when profile type is not vault",
			args: args{
				request: VaultRequest{
					VaultWallet: traderWallet,
				},
				exchangeId: exchangeId,
				pagination: &types.PaginationRequestParams{
					Page:  0,
					Limit: 50,
					Order: "ASC",
				},
			},
			want: []VaultResponse{},
			wantPagination: &types.PaginationResponse{
				Total: 0,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			pagination := types.PaginationRequestParams{
				Page:  0,
				Limit: 50,
				Order: "ASC",
			}
			if tt.args.pagination != nil {
				pagination = *tt.args.pagination
			}

			// when
			got, paginationResponse, err := HandleVault(context.Background(), s.GetDB(), tt.args.request, tt.args.exchangeId, pagination)

			// then
			if tt.wantErr {
				s.NotNil(err)
				s.Nil(got)
				s.Nil(paginationResponse)
			} else {
				s.NoError(err)
				s.assertJSON(tt.want, got)
				s.Equal(tt.wantPagination, paginationResponse)
			}
		})
	}
}

func (s *dbTestSuite) TestHandleVaultBalanceOperations() {
	status := "pending"
	performanceFee := decimal.NewFromFloat(2)
	vaultName := "vaultName1"
	managerName := "managerName1"
	initialisedAt := time.Now()
	accountEquity := decimal.NewFromFloat(2000)
	accountEquity2 := decimal.NewFromFloat(500)
	shares := decimal.NewFromFloat(10)
	totalShares := decimal.NewFromFloat(100)
	entryNav := decimal.NewFromFloat(1000)
	entryNav2 := decimal.NewFromFloat(1000)
	entryPrice := decimal.NewFromFloat(1)

	stakeUSDT := decimal.NewFromFloat(1.01)
	stakeShares := decimal.NewFromFloat(1.02)
	stakeVaultUSDT := decimal.NewFromFloat(1.03)

	unstakeShares := decimal.NewFromFloat(1.04)
	unstakeVaultShares := decimal.NewFromFloat(1.05)
	unstakeUSDT := decimal.NewFromFloat(1.06)
	unstakeFeeUSDT := decimal.NewFromFloat(1.07)
	unstakeVaultUSDT := decimal.NewFromFloat(1.08)

	insertCommon := func() {
		s.insertProfile(vaultID1, vaultProfileType, wallet, exchangeId)
		s.insertVaultLast(vaultID1, performanceFee, totalShares, status, vaultName, managerName, initialisedAt)
		s.insertProfileCacheLast(vaultID1, vaultProfileType, wallet, accountEquity)
		s.insertVaultHoldingsLast(vaultID1, traderID, shares, entryNav, entryPrice)

		s.insertProfile(traderID, traderProfileType, traderWallet, exchangeId)
		s.insertVaultLast(traderID, performanceFee, totalShares, status, vaultName, managerName, initialisedAt)
		s.insertProfileCacheLast(traderID, traderProfileType, traderWallet, accountEquity)

		s.insertProfile(vaultID2, vaultProfileType, vaultWallet2, exchangeId)
		s.insertVaultLast(vaultID2, performanceFee, totalShares, status, vaultName, managerName, initialisedAt)
		s.insertProfileCacheLast(vaultID2, vaultProfileType, vaultWallet2, accountEquity2)
		s.insertVaultHoldingsLast(vaultID2, traderID, shares, entryNav2, entryPrice)
	}

	timestampStake := int64(1)
	timestampStake2 := int64(2)
	timestampStakeFromBalance := int64(3)
	timestampUnstake := int64(4)

	insertStake := func() {
		s.insertBalanceOperation("d624ad3f-bedd-4396-b2b3-048d51459525", "s_132", "stake", traderID, wallet, exchangeId, stakeUSDT, timestampStake, status)
		s.insertBalanceOperation("ss_d624ad3f-bedd-4396-b2b3-048d51459525", "ss_s_132", "stake_shares", traderID, wallet, exchangeId, stakeShares, timestampStake, "")
		s.insertBalanceOperation("vs_s_d624ad3f-bedd-4396-b2b3-048d51459525", "vs_s_132", "vault_stake", vaultID1, wallet, exchangeId, stakeVaultUSDT, timestampStake, "")
	}

	insertStake2 := func() {
		s.insertBalanceOperation("d624ad3f-bedd-4396-b2b3-048d51459526", "s_133", "stake", traderID, vaultWallet2, exchangeId, stakeUSDT, timestampStake2, status)
		s.insertBalanceOperation("ss_d624ad3f-bedd-4396-b2b3-048d51459526", "ss_s_133", "stake_shares", traderID, vaultWallet2, exchangeId, stakeShares, timestampStake2, "")
		s.insertBalanceOperation("vs_s_d624ad3f-bedd-4396-b2b3-048d51459526", "vs_s_133", "vault_stake", vaultID2, vaultWallet2, exchangeId, stakeVaultUSDT, timestampStake2, "")
	}

	insertStakeFromBalance := func() {
		s.insertBalanceOperation("772b0381-125d-48e9-8cdd-a6e383a0ac2f", "772b0381-125d-48e9-8cdd-a6e383a0ac2f", "stake_from_balance", traderID, wallet, exchangeId, stakeUSDT, timestampStakeFromBalance, status)
		s.insertBalanceOperation("ss_772b0381-125d-48e9-8cdd-a6e383a0ac2f", "ss_772b0381-125d-48e9-8cdd-a6e383a0ac2f", "stake_shares", traderID, wallet, exchangeId, stakeShares, timestampStakeFromBalance, "")
		s.insertBalanceOperation("vs_s_772b0381-125d-48e9-8cdd-a6e383a0ac2f", "vs_772b0381-125d-48e9-8cdd-a6e383a0ac2f", "vault_stake", vaultID1, wallet, exchangeId, stakeVaultUSDT, timestampStakeFromBalance, "")
	}

	insertUnstake := func() {
		s.insertBalanceOperation("u_345", "u_345", "unstake_shares", traderID, wallet, exchangeId, unstakeShares, timestampUnstake, status)
		s.insertBalanceOperation("vus_u_345", "vus_u_345", "vault_unstake_shares", vaultID1, wallet, exchangeId, unstakeVaultShares, timestampUnstake, "")
		s.insertBalanceOperation("uv_u_345", "uv_u_345", "unstake_value", traderID, wallet, exchangeId, unstakeUSDT, timestampUnstake, "")
		s.insertBalanceOperation("uf_u_345", "uf_u_345", "unstake_fee", treasurerID, wallet, exchangeId, unstakeFeeUSDT, timestampUnstake, "")
		s.insertBalanceOperation("vuv_u_345", "vuv_u_345", "vault_unstake_value", vaultID1, wallet, exchangeId, unstakeVaultUSDT, timestampUnstake, "")
	}

	type args struct {
		request    VaultBalanceOperationsRequest
		profileId  uint64
		exchangeId string
		pagination *types.PaginationRequestParams
	}
	tests := []struct {
		name           string
		args           args
		insertData     func()
		want           []VaultBalanceOperationsResponse
		wantPagination *types.PaginationResponse
		wantErr        bool
	}{
		{
			name: "should map stake operation",
			args: args{
				request:    VaultBalanceOperationsRequest{},
				profileId:  traderID,
				exchangeId: exchangeId,
			},
			insertData: func() {
				insertCommon()
				s.insertBalanceOperation("d624ad3f-bedd-4396-b2b3-048d51459525", "s_132", "stake", traderID, wallet, exchangeId, stakeUSDT, timestampStake, status)
			},
			want: []VaultBalanceOperationsResponse{
				{
					Id:                 "d624ad3f-bedd-4396-b2b3-048d51459525",
					OpsType:            "stake",
					OpsSubType:         "stake",
					StakerProfileId:    traderID,
					VaultProfileId:     0,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					Status:             status,
					VaultName:          vaultName,
					ManagerName:        managerName,
					Timestamp:          timestampStake,
					StakeUSDT:          stakeUSDT,
					StakeShares:        decimal.NewFromFloat(0),
					InceptionTimestamp: initialisedAt.UnixMicro(),
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 1,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
		{
			name: "should map stake_shares operation",
			args: args{
				request:    VaultBalanceOperationsRequest{},
				profileId:  traderID,
				exchangeId: exchangeId,
			},
			insertData: func() {
				insertCommon()
				s.insertBalanceOperation("ss_d624ad3f-bedd-4396-b2b3-048d51459525", "ss_s_132", "stake_shares", traderID, wallet, exchangeId, stakeShares, timestampStake, status)
			},
			want: []VaultBalanceOperationsResponse{
				{
					Id:                 "d624ad3f-bedd-4396-b2b3-048d51459525",
					OpsType:            "",
					OpsSubType:         "",
					StakerProfileId:    traderID,
					VaultProfileId:     0,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					Status:             "",
					VaultName:          vaultName,
					ManagerName:        managerName,
					Timestamp:          0,
					StakeUSDT:          decimal.NewFromFloat(0),
					StakeShares:        stakeShares,
					InceptionTimestamp: initialisedAt.UnixMicro(),
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 1,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
		{
			name: "should return operation type stake sub type stake",
			args: args{
				request: VaultBalanceOperationsRequest{
					VaultWallet: wallet,
				},
				profileId:  traderID,
				exchangeId: exchangeId,
			},
			insertData: func() {
				insertCommon()
				insertStake()
			},
			want: []VaultBalanceOperationsResponse{
				{
					Id:                 "d624ad3f-bedd-4396-b2b3-048d51459525",
					OpsType:            "stake",
					OpsSubType:         "stake",
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID1,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					Status:             status,
					VaultName:          vaultName,
					ManagerName:        managerName,
					Timestamp:          timestampStake,
					StakeUSDT:          stakeVaultUSDT,
					StakeShares:        stakeShares,
					InceptionTimestamp: initialisedAt.UnixMicro(),
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 1,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
		{
			name: "should return operation type stake sub type stake_from_balance",
			args: args{
				request: VaultBalanceOperationsRequest{
					VaultWallet: wallet,
				},
				profileId:  traderID,
				exchangeId: exchangeId,
			},
			insertData: func() {
				insertCommon()
				insertStakeFromBalance()
			},
			want: []VaultBalanceOperationsResponse{
				{
					Id:                 "772b0381-125d-48e9-8cdd-a6e383a0ac2f",
					OpsType:            "stake",
					OpsSubType:         "stake_from_balance",
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID1,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					Status:             status,
					VaultName:          vaultName,
					ManagerName:        managerName,
					Timestamp:          timestampStakeFromBalance,
					StakeUSDT:          stakeVaultUSDT,
					StakeShares:        stakeShares,
					InceptionTimestamp: initialisedAt.UnixMicro(),
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 1,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
		{
			name: "should map operation type unstake_shares",
			args: args{
				request:    VaultBalanceOperationsRequest{},
				profileId:  traderID,
				exchangeId: exchangeId,
			},
			insertData: func() {
				insertCommon()
				s.insertBalanceOperation("u_345", "u_345", "unstake_shares", traderID, wallet, exchangeId, unstakeShares, timestampUnstake, status)
			},
			want: []VaultBalanceOperationsResponse{
				{
					Id:                 "u_345",
					OpsType:            "unstake",
					OpsSubType:         "unstake",
					StakerProfileId:    traderID,
					VaultProfileId:     0,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					Status:             status,
					VaultName:          vaultName,
					ManagerName:        managerName,
					Timestamp:          timestampUnstake,
					UnstakeShares:      unstakeShares,
					UnstakeUSDT:        decimal.NewFromFloat(0),
					UnstakeFeeUSDT:     decimal.NewFromFloat(0),
					InceptionTimestamp: initialisedAt.UnixMicro(),
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 1,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
		{
			name: "should map operation type unstake_value",
			args: args{
				request:    VaultBalanceOperationsRequest{},
				profileId:  traderID,
				exchangeId: exchangeId,
			},
			insertData: func() {
				insertCommon()
				s.insertBalanceOperation("uv_u_345", "uv_u_345", "unstake_value", traderID, wallet, exchangeId, unstakeUSDT, timestampUnstake, status)
			},
			want: []VaultBalanceOperationsResponse{
				{
					Id:                 "u_345",
					OpsType:            "",
					OpsSubType:         "",
					StakerProfileId:    traderID,
					VaultProfileId:     0,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					Status:             "",
					VaultName:          vaultName,
					ManagerName:        managerName,
					Timestamp:          0,
					UnstakeShares:      decimal.NewFromFloat(0),
					UnstakeUSDT:        unstakeUSDT,
					UnstakeFeeUSDT:     decimal.NewFromFloat(0),
					InceptionTimestamp: initialisedAt.UnixMicro(),
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 1,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
		{
			name: "should return operation type unstake",
			args: args{
				request: VaultBalanceOperationsRequest{
					VaultWallet: wallet,
				},
				profileId:  traderID,
				exchangeId: exchangeId,
			},
			insertData: func() {
				insertCommon()
				insertUnstake()
			},
			want: []VaultBalanceOperationsResponse{
				{
					Id:                 "u_345",
					OpsType:            "unstake",
					OpsSubType:         "unstake",
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID1,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					Status:             status,
					VaultName:          vaultName,
					ManagerName:        managerName,
					Timestamp:          timestampUnstake,
					UnstakeShares:      unstakeVaultShares,
					UnstakeUSDT:        unstakeUSDT,
					UnstakeFeeUSDT:     unstakeFeeUSDT,
					InceptionTimestamp: initialisedAt.UnixMicro(),
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 1,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
		{
			name: "should return all operation types for 1 wallet",
			args: args{
				request: VaultBalanceOperationsRequest{
					VaultWallet: wallet,
				},
				profileId:  traderID,
				exchangeId: exchangeId,
			},
			insertData: func() {
				insertCommon()
				insertStake()
				insertStakeFromBalance()
				insertUnstake()
			},
			want: []VaultBalanceOperationsResponse{
				{
					Id:                 "d624ad3f-bedd-4396-b2b3-048d51459525",
					OpsType:            "stake",
					OpsSubType:         "stake",
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID1,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					Status:             status,
					VaultName:          vaultName,
					ManagerName:        managerName,
					Timestamp:          timestampStake,
					StakeUSDT:          stakeVaultUSDT,
					StakeShares:        stakeShares,
					InceptionTimestamp: initialisedAt.UnixMicro(),
				},
				{
					Id:                 "772b0381-125d-48e9-8cdd-a6e383a0ac2f",
					OpsType:            "stake",
					OpsSubType:         "stake_from_balance",
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID1,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					Status:             status,
					VaultName:          vaultName,
					ManagerName:        managerName,
					Timestamp:          timestampStakeFromBalance,
					StakeUSDT:          stakeVaultUSDT,
					StakeShares:        stakeShares,
					InceptionTimestamp: initialisedAt.UnixMicro(),
				},
				{
					Id:                 "u_345",
					OpsType:            "unstake",
					OpsSubType:         "unstake",
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID1,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					Status:             status,
					VaultName:          vaultName,
					ManagerName:        managerName,
					Timestamp:          timestampUnstake,
					UnstakeShares:      unstakeVaultShares,
					UnstakeUSDT:        unstakeUSDT,
					UnstakeFeeUSDT:     unstakeFeeUSDT,
					InceptionTimestamp: initialisedAt.UnixMicro(),
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 3,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
		{
			name: "should return all operation types for 1 wallet - page 0",
			args: args{
				request: VaultBalanceOperationsRequest{
					VaultWallet: wallet,
				},
				profileId:  traderID,
				exchangeId: exchangeId,
				pagination: &types.PaginationRequestParams{
					Page:  0,
					Limit: 1,
					Order: "ASC",
				},
			},
			insertData: func() {
				insertCommon()
				insertStake()
				insertStakeFromBalance()
				insertUnstake()
			},
			want: []VaultBalanceOperationsResponse{
				{
					Id:                 "d624ad3f-bedd-4396-b2b3-048d51459525",
					OpsType:            "stake",
					OpsSubType:         "stake",
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID1,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					Status:             status,
					VaultName:          vaultName,
					ManagerName:        managerName,
					Timestamp:          timestampStake,
					StakeUSDT:          stakeVaultUSDT,
					StakeShares:        stakeShares,
					InceptionTimestamp: initialisedAt.UnixMicro(),
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 3,
				Page:  0,
				Limit: 1,
				Order: "ASC",
			},
		},
		{
			name: "should return all operation types for 1 wallet - page 1",
			args: args{
				request: VaultBalanceOperationsRequest{
					VaultWallet: wallet,
				},
				profileId:  traderID,
				exchangeId: exchangeId,
				pagination: &types.PaginationRequestParams{
					Page:  1,
					Limit: 1,
					Order: "ASC",
				},
			},
			insertData: func() {
				insertCommon()
				insertStake()
				insertStakeFromBalance()
				insertUnstake()
			},
			want: []VaultBalanceOperationsResponse{
				{
					Id:                 "772b0381-125d-48e9-8cdd-a6e383a0ac2f",
					OpsType:            "stake",
					OpsSubType:         "stake_from_balance",
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID1,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					Status:             status,
					VaultName:          vaultName,
					ManagerName:        managerName,
					Timestamp:          timestampStakeFromBalance,
					StakeUSDT:          stakeVaultUSDT,
					StakeShares:        stakeShares,
					InceptionTimestamp: initialisedAt.UnixMicro(),
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 3,
				Page:  1,
				Limit: 1,
				Order: "ASC",
			},
		},
		{
			name: "should return all operation types for 1 wallet - page 2",
			args: args{
				request: VaultBalanceOperationsRequest{
					VaultWallet: wallet,
				},
				profileId:  traderID,
				exchangeId: exchangeId,
				pagination: &types.PaginationRequestParams{
					Page:  2,
					Limit: 1,
					Order: "ASC",
				},
			},
			insertData: func() {
				insertCommon()
				insertStake()
				insertStakeFromBalance()
				insertUnstake()
			},
			want: []VaultBalanceOperationsResponse{
				{
					Id:                 "u_345",
					OpsType:            "unstake",
					OpsSubType:         "unstake",
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID1,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					Status:             status,
					VaultName:          vaultName,
					ManagerName:        managerName,
					Timestamp:          timestampUnstake,
					UnstakeShares:      unstakeVaultShares,
					UnstakeUSDT:        unstakeUSDT,
					UnstakeFeeUSDT:     unstakeFeeUSDT,
					InceptionTimestamp: initialisedAt.UnixMicro(),
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 3,
				Page:  2,
				Limit: 1,
				Order: "ASC",
			},
		},
		{
			name: "should return all operation types for 1 wallet - page 0 sort DESC",
			args: args{
				request: VaultBalanceOperationsRequest{
					VaultWallet: wallet,
				},
				profileId:  traderID,
				exchangeId: exchangeId,
				pagination: &types.PaginationRequestParams{
					Page:  0,
					Limit: 1,
					Order: "DESC",
				},
			},
			insertData: func() {
				insertCommon()
				insertStake()
				insertStakeFromBalance()
				insertUnstake()
			},
			want: []VaultBalanceOperationsResponse{
				{
					Id:                 "u_345",
					OpsType:            "unstake",
					OpsSubType:         "unstake",
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID1,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					Status:             status,
					VaultName:          vaultName,
					ManagerName:        managerName,
					Timestamp:          timestampUnstake,
					UnstakeShares:      unstakeVaultShares,
					UnstakeUSDT:        unstakeUSDT,
					UnstakeFeeUSDT:     unstakeFeeUSDT,
					InceptionTimestamp: initialisedAt.UnixMicro(),
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 3,
				Page:  0,
				Limit: 1,
				Order: "DESC",
			},
		},
		{
			name: "should return all operation types for 2 wallets",
			args: args{
				request: VaultBalanceOperationsRequest{
					VaultWallet: wallet + "," + vaultWallet2,
					OpsType:     []string{"stake", "unstake"},
				},
				profileId:  traderID,
				exchangeId: exchangeId,
			},
			insertData: func() {
				insertCommon()
				insertStake()
				insertStakeFromBalance()
				insertUnstake()
				insertStake2()
			},
			want: []VaultBalanceOperationsResponse{
				{
					Id:                 "d624ad3f-bedd-4396-b2b3-048d51459525",
					OpsType:            "stake",
					OpsSubType:         "stake",
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID1,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					Status:             status,
					VaultName:          vaultName,
					ManagerName:        managerName,
					Timestamp:          timestampStake,
					StakeUSDT:          stakeVaultUSDT,
					StakeShares:        stakeShares,
					InceptionTimestamp: initialisedAt.UnixMicro(),
				},
				{
					Id:                 "d624ad3f-bedd-4396-b2b3-048d51459526",
					OpsType:            "stake",
					OpsSubType:         "stake",
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID2,
					Wallet:             vaultWallet2,
					ExchangeId:         exchangeId,
					Status:             status,
					VaultName:          vaultName,
					ManagerName:        managerName,
					Timestamp:          timestampStake2,
					StakeUSDT:          stakeVaultUSDT,
					StakeShares:        stakeShares,
					InceptionTimestamp: initialisedAt.UnixMicro(),
				},
				{
					Id:                 "772b0381-125d-48e9-8cdd-a6e383a0ac2f",
					OpsType:            "stake",
					OpsSubType:         "stake_from_balance",
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID1,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					Status:             status,
					VaultName:          vaultName,
					ManagerName:        managerName,
					Timestamp:          timestampStakeFromBalance,
					StakeUSDT:          stakeVaultUSDT,
					StakeShares:        stakeShares,
					InceptionTimestamp: initialisedAt.UnixMicro(),
				},
				{
					Id:                 "u_345",
					OpsType:            "unstake",
					OpsSubType:         "unstake",
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID1,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					Status:             status,
					VaultName:          vaultName,
					ManagerName:        managerName,
					Timestamp:          timestampUnstake,
					UnstakeShares:      unstakeVaultShares,
					UnstakeUSDT:        unstakeUSDT,
					UnstakeFeeUSDT:     unstakeFeeUSDT,
					InceptionTimestamp: initialisedAt.UnixMicro(),
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 4,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
		{
			name: "should return all operation types for all wallets",
			args: args{
				request:    VaultBalanceOperationsRequest{},
				profileId:  traderID,
				exchangeId: exchangeId,
			},
			insertData: func() {
				insertCommon()
				insertStake()
				insertStakeFromBalance()
				insertUnstake()
				insertStake2()
			},
			want: []VaultBalanceOperationsResponse{
				{
					Id:                 "d624ad3f-bedd-4396-b2b3-048d51459525",
					OpsType:            "stake",
					OpsSubType:         "stake",
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID1,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					Status:             status,
					VaultName:          vaultName,
					ManagerName:        managerName,
					Timestamp:          timestampStake,
					StakeUSDT:          stakeVaultUSDT,
					StakeShares:        stakeShares,
					InceptionTimestamp: initialisedAt.UnixMicro(),
				},
				{
					Id:                 "d624ad3f-bedd-4396-b2b3-048d51459526",
					OpsType:            "stake",
					OpsSubType:         "stake",
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID2,
					Wallet:             vaultWallet2,
					ExchangeId:         exchangeId,
					Status:             status,
					VaultName:          vaultName,
					ManagerName:        managerName,
					Timestamp:          timestampStake2,
					StakeUSDT:          stakeVaultUSDT,
					StakeShares:        stakeShares,
					InceptionTimestamp: initialisedAt.UnixMicro(),
				},
				{
					Id:                 "772b0381-125d-48e9-8cdd-a6e383a0ac2f",
					OpsType:            "stake",
					OpsSubType:         "stake_from_balance",
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID1,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					Status:             status,
					VaultName:          vaultName,
					ManagerName:        managerName,
					Timestamp:          timestampStakeFromBalance,
					StakeUSDT:          stakeVaultUSDT,
					StakeShares:        stakeShares,
					InceptionTimestamp: initialisedAt.UnixMicro(),
				},
				{
					Id:                 "u_345",
					OpsType:            "unstake",
					OpsSubType:         "unstake",
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID1,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					Status:             status,
					VaultName:          vaultName,
					ManagerName:        managerName,
					Timestamp:          timestampUnstake,
					UnstakeShares:      unstakeVaultShares,
					UnstakeUSDT:        unstakeUSDT,
					UnstakeFeeUSDT:     unstakeFeeUSDT,
					InceptionTimestamp: initialisedAt.UnixMicro(),
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 4,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
		{
			name: "should return stake operations",
			args: args{
				request: VaultBalanceOperationsRequest{
					VaultWallet: wallet + "," + vaultWallet2,
					OpsType:     []string{"stake"},
				},
				profileId:  traderID,
				exchangeId: exchangeId,
			},
			insertData: func() {
				insertCommon()
				insertStake()
				insertStakeFromBalance()
				insertUnstake()
				insertStake2()
			},
			want: []VaultBalanceOperationsResponse{
				{
					Id:                 "d624ad3f-bedd-4396-b2b3-048d51459525",
					OpsType:            "stake",
					OpsSubType:         "stake",
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID1,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					Status:             status,
					VaultName:          vaultName,
					ManagerName:        managerName,
					Timestamp:          timestampStake,
					StakeUSDT:          stakeVaultUSDT,
					StakeShares:        stakeShares,
					InceptionTimestamp: initialisedAt.UnixMicro(),
				},
				{
					Id:                 "d624ad3f-bedd-4396-b2b3-048d51459526",
					OpsType:            "stake",
					OpsSubType:         "stake",
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID2,
					Wallet:             vaultWallet2,
					ExchangeId:         exchangeId,
					Status:             status,
					VaultName:          vaultName,
					ManagerName:        managerName,
					Timestamp:          timestampStake2,
					StakeUSDT:          stakeVaultUSDT,
					StakeShares:        stakeShares,
					InceptionTimestamp: initialisedAt.UnixMicro(),
				},
				{
					Id:                 "772b0381-125d-48e9-8cdd-a6e383a0ac2f",
					OpsType:            "stake",
					OpsSubType:         "stake_from_balance",
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID1,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					Status:             status,
					VaultName:          vaultName,
					ManagerName:        managerName,
					Timestamp:          timestampStakeFromBalance,
					StakeUSDT:          stakeVaultUSDT,
					StakeShares:        stakeShares,
					InceptionTimestamp: initialisedAt.UnixMicro(),
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 3,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
		{
			name: "should return unstake operations",
			args: args{
				request: VaultBalanceOperationsRequest{
					VaultWallet: wallet,
					OpsType:     []string{"unstake"},
				},
				profileId:  traderID,
				exchangeId: exchangeId,
			},
			insertData: func() {
				insertCommon()
				insertStake()
				insertStakeFromBalance()
				insertUnstake()
				insertStake2()
			},
			want: []VaultBalanceOperationsResponse{
				{
					Id:                 "u_345",
					OpsType:            "unstake",
					OpsSubType:         "unstake",
					StakerProfileId:    traderID,
					VaultProfileId:     vaultID1,
					Wallet:             wallet,
					ExchangeId:         exchangeId,
					Status:             status,
					VaultName:          vaultName,
					ManagerName:        managerName,
					Timestamp:          timestampUnstake,
					UnstakeShares:      unstakeVaultShares,
					UnstakeUSDT:        unstakeUSDT,
					UnstakeFeeUSDT:     unstakeFeeUSDT,
					InceptionTimestamp: initialisedAt.UnixMicro(),
				},
			},
			wantPagination: &types.PaginationResponse{
				Total: 1,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
		{
			name: "should return zero responses when profileId not found",
			args: args{
				request: VaultBalanceOperationsRequest{
					VaultWallet: wallet,
				},
				profileId:  uint64(0),
				exchangeId: exchangeId,
			},
			insertData: func() {
				insertCommon()
				insertStake()
				insertStakeFromBalance()
				insertUnstake()
				insertStake2()
			},
			want: []VaultBalanceOperationsResponse{},
			wantPagination: &types.PaginationResponse{
				Total: 0,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
		{
			name: "should return zero responses when wallet not found",
			args: args{
				request: VaultBalanceOperationsRequest{
					VaultWallet: "WALLET_NOT_FOUND",
				},
				profileId:  traderID,
				exchangeId: exchangeId,
			},
			insertData: func() {
				insertCommon()
				insertStake()
				insertStakeFromBalance()
				insertUnstake()
				insertStake2()
			},
			want: []VaultBalanceOperationsResponse{},
			wantPagination: &types.PaginationResponse{
				Total: 0,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
		{
			name: "should return zero responses when exchangeId not found",
			args: args{
				request: VaultBalanceOperationsRequest{
					VaultWallet: wallet,
				},
				profileId:  traderID,
				exchangeId: "EXCHANGE_ID_NOT_FOUND",
			},
			insertData: func() {
				insertCommon()
				insertStake()
				insertStakeFromBalance()
				insertUnstake()
				insertStake2()
			},
			want: []VaultBalanceOperationsResponse{},
			wantPagination: &types.PaginationResponse{
				Total: 0,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
		{
			name: "should return zero responses when type not found",
			args: args{
				request: VaultBalanceOperationsRequest{
					VaultWallet: wallet,
					OpsType:     []string{"OPERATION_TYPE_NOT_FOUND"},
				},
				profileId:  traderID,
				exchangeId: exchangeId,
			},
			insertData: func() {
				insertCommon()
				insertStake()
				insertStakeFromBalance()
				insertUnstake()
				insertStake2()
			},
			want: []VaultBalanceOperationsResponse{},
			wantPagination: &types.PaginationResponse{
				Total: 0,
				Page:  0,
				Limit: 50,
				Order: "ASC",
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.DeleteTables(tables)
			tt.insertData()

			pagination := types.PaginationRequestParams{
				Page:  0,
				Limit: 50,
				Order: "ASC",
			}
			if tt.args.pagination != nil {
				pagination = *tt.args.pagination
			}

			// when
			got, paginationResponse, err := HandleVaultBalanceOperations(context.Background(), s.GetDB(), tt.args.request, uint(tt.args.profileId), tt.args.exchangeId, pagination)

			// then
			if tt.wantErr {
				s.NotNil(err)
				s.Nil(got)
				s.Nil(paginationResponse)
			} else {
				s.NoError(err)
				s.assertJSON(tt.want, got)
				s.Equal(tt.wantPagination, paginationResponse)
			}
		})
	}
}

func (s *dbTestSuite) TestRefreshAppVaultAggregateLast() {
	now := time.Now().Add(-time.Second)

	initialSharePrice := appVaultAggregate{
		VaultProfileId:   vaultID1,
		SharePrice:       decimal.NewFromFloat(1),
		APYTotal:         decimal.NewFromFloat(0),
		APYUSDT:          decimal.NewFromFloat(0),
		APYRBX:           decimal.NewFromFloat(0),
		ArchiveTimestamp: now.UnixMicro(),
	}
	existing := appVaultAggregate{
		VaultProfileId:   vaultID1,
		SharePrice:       decimal.NewFromFloat(100),
		APYTotal:         decimal.NewFromFloat(200),
		APYUSDT:          decimal.NewFromFloat(300),
		APYRBX:           decimal.NewFromFloat(400),
		ArchiveTimestamp: now.Add(-time.Hour).UnixMicro(),
	}
	tests := []struct {
		name        string
		insertData  func()
		want        []appVaultAggregate
		wantHistory []appVaultAggregate
	}{
		{
			name: "should start with empty tables",
		},
		{
			name: "should calculate when profile cache archive_timestamp is equal to vault archive_timestamp",
			insertData: func() {
				initialisedAt := now
				vaultArchiveTimestamp := now
				cacheArchiveTimestamp := now
				accountEquity := decimal.NewFromFloat(55)
				totalShares := decimal.NewFromFloat(5)

				s.insertVaultLast2(vaultID1, totalShares, initialisedAt, vaultArchiveTimestamp)
				s.insertProfileCacheLast2(vaultID1, vaultProfileType, wallet, accountEquity, cacheArchiveTimestamp)
			},
			want: []appVaultAggregate{
				{
					VaultProfileId:   vaultID1,
					SharePrice:       decimal.NewFromFloat(11),
					APYTotal:         decimal.NewFromFloat(3650),
					APYUSDT:          decimal.NewFromFloat(3650),
					APYRBX:           decimal.NewFromFloat(0),
					ArchiveTimestamp: now.UnixMicro(),
				},
			},
			wantHistory: []appVaultAggregate{
				{
					VaultProfileId:   vaultID1,
					SharePrice:       decimal.NewFromFloat(11),
					APYTotal:         decimal.NewFromFloat(3650),
					APYUSDT:          decimal.NewFromFloat(3650),
					APYRBX:           decimal.NewFromFloat(0),
					ArchiveTimestamp: now.UnixMicro(),
				},
			},
		},
		{
			name: "should calculate when profile cache archive_timestamp is after vault archive_timestamp",
			insertData: func() {
				initialisedAt := now
				vaultArchiveTimestamp := now
				cacheArchiveTimestamp := now.Add(time.Hour)
				accountEquity := decimal.NewFromFloat(55)
				totalShares := decimal.NewFromFloat(5)

				s.insertVaultLast2(vaultID1, totalShares, initialisedAt, vaultArchiveTimestamp)
				s.insertProfileCacheLast2(vaultID1, vaultProfileType, wallet, accountEquity, cacheArchiveTimestamp)
			},
			want: []appVaultAggregate{
				{
					VaultProfileId:   vaultID1,
					SharePrice:       decimal.NewFromFloat(11),
					APYTotal:         decimal.NewFromFloat(3650),
					APYUSDT:          decimal.NewFromFloat(3650),
					APYRBX:           decimal.NewFromFloat(0),
					ArchiveTimestamp: now.Add(time.Hour).UnixMicro(),
				},
			},
			wantHistory: []appVaultAggregate{
				{
					VaultProfileId:   vaultID1,
					SharePrice:       decimal.NewFromFloat(11),
					APYTotal:         decimal.NewFromFloat(3650),
					APYUSDT:          decimal.NewFromFloat(3650),
					APYRBX:           decimal.NewFromFloat(0),
					ArchiveTimestamp: now.Add(time.Hour).UnixMicro(),
				},
			},
		},
		{
			name: "should not update apy_total, but should insert share price 1 when profile cache archive_timestamp is before vault archive_timestamp",
			insertData: func() {
				initialisedAt := now
				vaultArchiveTimestamp := now
				cacheArchiveTimestamp := now.Add(-time.Hour)
				accountEquity := decimal.NewFromFloat(55)
				totalShares := decimal.NewFromFloat(5)

				s.insertVaultLast2(vaultID1, totalShares, initialisedAt, vaultArchiveTimestamp)
				s.insertProfileCacheLast2(vaultID1, vaultProfileType, wallet, accountEquity, cacheArchiveTimestamp)
			},
			want: []appVaultAggregate{
				{
					VaultProfileId:   vaultID1,
					SharePrice:       decimal.NewFromFloat(1),
					APYTotal:         decimal.NewFromFloat(0),
					APYUSDT:          decimal.NewFromFloat(0),
					APYRBX:           decimal.NewFromFloat(0),
					ArchiveTimestamp: now.Add(-time.Hour).UnixMicro(),
				},
			},
			wantHistory: []appVaultAggregate{
				{
					VaultProfileId:   vaultID1,
					SharePrice:       decimal.NewFromFloat(1),
					APYTotal:         decimal.NewFromFloat(0),
					APYUSDT:          decimal.NewFromFloat(0),
					APYRBX:           decimal.NewFromFloat(0),
					ArchiveTimestamp: now.Add(-time.Hour).UnixMicro(),
				},
			},
		},
		{
			name: "should calculate APY roughly 2 times lower on next day since inception",
			insertData: func() {
				initialisedAt := now.Add(-24 * time.Hour)
				vaultArchiveTimestamp := now
				cacheArchiveTimestamp := now
				accountEquity := decimal.NewFromFloat(55)
				totalShares := decimal.NewFromFloat(5)

				s.insertVaultLast2(vaultID1, totalShares, initialisedAt, vaultArchiveTimestamp)
				s.insertProfileCacheLast2(vaultID1, vaultProfileType, wallet, accountEquity, cacheArchiveTimestamp)
			},
			want: []appVaultAggregate{
				{
					VaultProfileId:   vaultID1,
					SharePrice:       decimal.NewFromFloat(11),
					APYTotal:         decimal.NewFromFloat(1820),
					APYUSDT:          decimal.NewFromFloat(1820),
					APYRBX:           decimal.NewFromFloat(0),
					ArchiveTimestamp: now.UnixMicro(),
				},
			},
			wantHistory: []appVaultAggregate{
				{
					VaultProfileId:   vaultID1,
					SharePrice:       decimal.NewFromFloat(11),
					APYTotal:         decimal.NewFromFloat(1820),
					APYUSDT:          decimal.NewFromFloat(1820),
					APYRBX:           decimal.NewFromFloat(0),
					ArchiveTimestamp: now.UnixMicro(),
				},
			},
		},
		{
			name: "should not calculate when profile cache is missing",
			insertData: func() {
				initialisedAt := now
				vaultArchiveTimestamp := now
				totalShares := decimal.NewFromFloat(5)

				s.insertVaultLast2(vaultID1, totalShares, initialisedAt, vaultArchiveTimestamp)
			},
		},
		{
			name: "should not calculate when vault is missing",
			insertData: func() {
				cacheArchiveTimestamp := now.Add(-time.Hour)
				accountEquity := decimal.NewFromFloat(55)

				s.insertProfileCacheLast2(vaultID1, vaultProfileType, wallet, accountEquity, cacheArchiveTimestamp)
			},
		},
		{
			name: "should calculate share price 1 and apy 0 when total_shares is 0",
			insertData: func() {
				initialisedAt := now
				vaultArchiveTimestamp := now
				cacheArchiveTimestamp := now
				accountEquity := decimal.NewFromFloat(55)
				totalShares := decimal.NewFromFloat(0)

				s.insertVaultLast2(vaultID1, totalShares, initialisedAt, vaultArchiveTimestamp)
				s.insertProfileCacheLast2(vaultID1, vaultProfileType, wallet, accountEquity, cacheArchiveTimestamp)
			},
			want:        []appVaultAggregate{initialSharePrice},
			wantHistory: []appVaultAggregate{initialSharePrice},
		},
		{
			name: "should calculate share price 1 and apy 0 when total_shares < 0",
			insertData: func() {
				initialisedAt := now
				vaultArchiveTimestamp := now
				cacheArchiveTimestamp := now
				accountEquity := decimal.NewFromFloat(55)
				totalShares := decimal.NewFromFloat(-1)

				s.insertVaultLast2(vaultID1, totalShares, initialisedAt, vaultArchiveTimestamp)
				s.insertProfileCacheLast2(vaultID1, vaultProfileType, wallet, accountEquity, cacheArchiveTimestamp)
			},
			want:        []appVaultAggregate{initialSharePrice},
			wantHistory: []appVaultAggregate{initialSharePrice},
		},
		{
			name: "should calculate share price 1 and apy 0 when account_equity is 0",
			insertData: func() {
				initialisedAt := now
				vaultArchiveTimestamp := now
				cacheArchiveTimestamp := now
				accountEquity := decimal.NewFromFloat(0)
				totalShares := decimal.NewFromFloat(5)

				s.insertVaultLast2(vaultID1, totalShares, initialisedAt, vaultArchiveTimestamp)
				s.insertProfileCacheLast2(vaultID1, vaultProfileType, wallet, accountEquity, cacheArchiveTimestamp)
			},
			want:        []appVaultAggregate{initialSharePrice},
			wantHistory: []appVaultAggregate{initialSharePrice},
		},
		{
			name: "should calculate share price 1 and apy 0 when account_equity < 0",
			insertData: func() {
				initialisedAt := now
				vaultArchiveTimestamp := now
				cacheArchiveTimestamp := now
				accountEquity := decimal.NewFromFloat(-1)
				totalShares := decimal.NewFromFloat(5)

				s.insertVaultLast2(vaultID1, totalShares, initialisedAt, vaultArchiveTimestamp)
				s.insertProfileCacheLast2(vaultID1, vaultProfileType, wallet, accountEquity, cacheArchiveTimestamp)
			},
			want:        []appVaultAggregate{initialSharePrice},
			wantHistory: []appVaultAggregate{initialSharePrice},
		},
		{
			name: "should calculate share price 1 and apy 0 when initialised_at is in the future",
			insertData: func() {
				initialisedAt := now.Add(time.Hour)
				vaultArchiveTimestamp := now
				cacheArchiveTimestamp := now
				accountEquity := decimal.NewFromFloat(55)
				totalShares := decimal.NewFromFloat(5)

				s.insertVaultLast2(vaultID1, totalShares, initialisedAt, vaultArchiveTimestamp)
				s.insertProfileCacheLast2(vaultID1, vaultProfileType, wallet, accountEquity, cacheArchiveTimestamp)
			},
			want:        []appVaultAggregate{initialSharePrice},
			wantHistory: []appVaultAggregate{initialSharePrice},
		},
		{
			name: "should create history on insert",
			insertData: func() {
				s.insertVaultAggregateLast2(existing)
			},
			want: []appVaultAggregate{
				existing,
			},
			wantHistory: []appVaultAggregate{
				existing,
			},
		},
		{
			name: "should update",
			insertData: func() {
				initialisedAt := now
				vaultArchiveTimestamp := now
				cacheArchiveTimestamp := now
				accountEquity := decimal.NewFromFloat(55)
				totalShares := decimal.NewFromFloat(5)

				s.insertVaultAggregateLast2(existing)
				s.insertVaultLast2(vaultID1, totalShares, initialisedAt, vaultArchiveTimestamp)
				s.insertProfileCacheLast2(vaultID1, vaultProfileType, wallet, accountEquity, cacheArchiveTimestamp)
			},
			want: []appVaultAggregate{
				{
					VaultProfileId:   vaultID1,
					SharePrice:       decimal.NewFromFloat(11),
					APYTotal:         decimal.NewFromFloat(3650),
					APYUSDT:          decimal.NewFromFloat(3650),
					APYRBX:           decimal.NewFromFloat(0),
					ArchiveTimestamp: now.UnixMicro(),
				},
			},
			wantHistory: []appVaultAggregate{
				existing,
				{
					VaultProfileId:   vaultID1,
					SharePrice:       decimal.NewFromFloat(11),
					APYTotal:         decimal.NewFromFloat(3650),
					APYUSDT:          decimal.NewFromFloat(3650),
					APYRBX:           decimal.NewFromFloat(0),
					ArchiveTimestamp: now.UnixMicro(),
				},
			},
		},
		{
			name: "should not update",
			insertData: func() {
				initialisedAt := now
				vaultArchiveTimestamp := now
				cacheArchiveTimestamp := now.Add(-time.Hour)
				accountEquity := decimal.NewFromFloat(55)
				totalShares := decimal.NewFromFloat(5)

				s.insertVaultAggregateLast2(existing)
				s.insertVaultLast2(vaultID1, totalShares, initialisedAt, vaultArchiveTimestamp)
				s.insertProfileCacheLast2(vaultID1, vaultProfileType, wallet, accountEquity, cacheArchiveTimestamp)

			},
			want: []appVaultAggregate{
				existing,
			},
			wantHistory: []appVaultAggregate{
				existing,
			},
		},
		{
			name: "should not update when account_equity < 0",
			insertData: func() {
				initialisedAt := now.Add(time.Hour)
				vaultArchiveTimestamp := now.Add(time.Hour)
				cacheArchiveTimestamp := now.Add(time.Hour)
				accountEquity := decimal.NewFromFloat(-1)
				totalShares := decimal.NewFromFloat(10)

				s.insertVaultAggregateLast2(existing)
				s.insertVaultLast2(vaultID1, totalShares, initialisedAt, vaultArchiveTimestamp)
				s.insertProfileCacheLast2(vaultID1, vaultProfileType, wallet, accountEquity, cacheArchiveTimestamp)
			},
			want: []appVaultAggregate{
				existing,
			},
			wantHistory: []appVaultAggregate{
				existing,
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			// clean
			s.DeleteTables(tables)

			// given
			if tt.insertData != nil {
				tt.insertData()
			}

			// when
			s.Execute("SELECT refresh_app_vault_aggregate_last(1, '{}'::jsonb)")

			// then
			got := s.getAppVaultAggregateLast()
			s.assertJSON(tt.want, got)

			gotHistory := s.getAppVaultAggregateHistory()
			s.assertJSON(tt.wantHistory, gotHistory)
		})
	}
}

func (s *dbTestSuite) TestGetVaultProfile() {
	type args struct {
		wallet     string
		exchangeId string
	}

	tests := []struct {
		name       string
		insertData func()
		args       args
		want       *VaultProfile
		wantErr    bool
	}{
		{
			name: "should start with empty tables",
			args: args{
				wallet:     wallet,
				exchangeId: exchangeId,
			},
			want: nil,
		},
		{
			name: "should return vault profile",
			insertData: func() {
				s.insertProfile(vaultID1, vaultProfileType, wallet, exchangeId)
			},
			args: args{
				wallet:     wallet,
				exchangeId: exchangeId,
			},
			want: &VaultProfile{
				ID:          int64(vaultID1),
				ProfileType: vaultProfileType,
			},
		},
		{
			name: "should return error when found 2 wallets",
			insertData: func() {
				s.insertProfile(vaultID1, vaultProfileType, wallet, exchangeId)
				s.insertProfile(vaultID2, vaultProfileType, wallet, exchangeId)
			},
			args: args{
				wallet:     wallet,
				exchangeId: exchangeId,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			// clean
			s.DeleteTables(tables)

			// given
			if tt.insertData != nil {
				tt.insertData()
			}

			// when
			got, err := GetVaultProfile(context.Background(), s.GetDB(), Wallet(tt.args.wallet), ExchangeId(tt.args.exchangeId))

			// then
			r := require.New(s.T())
			if tt.wantErr {
				r.Error(err)
				r.Nil(got)
			} else {
				r.NoError(err)
				r.Equal(tt.want, got)
			}
		})
	}
}

func (s *dbTestSuite) TestHandleNavHistory() {
	now := time.Now()
	ts := now.UnixMicro()
	item := portfolio.PortfolioData{
		Time:  ts,
		Value: decimal.NewFromFloat(10.25),
	}

	type args struct {
		request    VaultHistoryRequest
		exchangeId string
	}

	tests := []struct {
		name       string
		insertData func()
		args       args
		want       []portfolio.PortfolioData
		wantErr    bool
		expectErr  string
	}{
		{
			name: "should return empty list",
			args: args{
				request: VaultHistoryRequest{
					VaultWallet: wallet,
					Range:       "1h",
				},
				exchangeId: exchangeId,
			},
			insertData: func() {
				s.insertProfile(vaultID1, vaultProfileType, wallet, exchangeId)
			},
			want: []portfolio.PortfolioData{},
		},
		{
			name: "should return 1 item",
			args: args{
				request: VaultHistoryRequest{
					VaultWallet: wallet,
					Range:       "1h",
				},
				exchangeId: exchangeId,
			},
			insertData: func() {
				s.insertProfile(vaultID1, vaultProfileType, wallet, exchangeId)
				s.insertProfileCachePeriod("1m", vaultID1, item)
			},
			want: []portfolio.PortfolioData{item},
		},
		{
			name: "should return error when found 2 wallets",
			args: args{
				request: VaultHistoryRequest{
					VaultWallet: wallet,
					Range:       "1h",
				},
				exchangeId: exchangeId,
			},
			insertData: func() {
				s.insertProfile(vaultID1, vaultProfileType, wallet, exchangeId)
				s.insertProfile(vaultID2, vaultProfileType, wallet, exchangeId)
			},
			wantErr:   true,
			expectErr: "GET_VAULT_PROFILE_ERROR: wallet=0x1a05a1507c35c763035fdb151af9286d4d90a81b, err=found more than 1 profile by wallet 0x1a05a1507c35c763035fdb151af9286d4d90a81b and exchange id rbx",
		},
		{
			name: "should return error when vault not found",
			args: args{
				request: VaultHistoryRequest{
					VaultWallet: wallet,
					Range:       "1h",
				},
				exchangeId: exchangeId,
			},
			wantErr:   true,
			expectErr: "VAULT_PROFILE_NOT_FOUND: wallet=0x1a05a1507c35c763035fdb151af9286d4d90a81b",
		},
		{
			name: "should return error when profile type is not found",
			args: args{
				request: VaultHistoryRequest{
					VaultWallet: wallet,
					Range:       "1h",
				},
				exchangeId: exchangeId,
			},
			insertData: func() {
				s.insertProfile(vaultID1, "PROFILE_TYPE_IS_NOT_VAULT", wallet, exchangeId)
			},
			wantErr:   true,
			expectErr: "TARGET_IS_NOT_A_VAULT 0x1a05a1507c35c763035fdb151af9286d4d90a81b",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			// clean
			s.DeleteTables(tables)

			// given
			if tt.insertData != nil {
				tt.insertData()
			}

			// when
			got, err := HandleNavHistory(context.Background(), s.GetDB(), tt.args.request, tt.args.exchangeId)

			// then
			if tt.wantErr {
				s.Error(err)
				s.Equal(tt.expectErr, err.Error())
				s.Nil(got)
			} else {
				s.NoError(err)
				s.Equal(tt.want, got)
			}
		})
	}
}

type appVaultAggregate struct {
	VaultProfileId   uint64          `json:"vault_profile_id"`
	SharePrice       decimal.Decimal `json:"share_price"`
	APYTotal         decimal.Decimal `json:"apy_total"`
	APYUSDT          decimal.Decimal `json:"apy_usdt"`
	APYRBX           decimal.Decimal `json:"apy_rbx"`
	ArchiveTimestamp int64           `json:"archive_timestamp"`
}

func (s *dbTestSuite) getAppVaultAggregateLast() []appVaultAggregate {
	return s.getAppVaultAggregateCommon("app_vault_aggregate_last", "vault_profile_id")
}

func (s *dbTestSuite) getAppVaultAggregateHistory() []appVaultAggregate {
	return s.getAppVaultAggregateCommon("app_vault_aggregate_history", "archive_timestamp, vault_profile_id")
}

func (s *dbTestSuite) getAppVaultAggregateCommon(tableName, orderBy string) []appVaultAggregate {
	r := require.New(s.T())

	q := fmt.Sprintf(`SELECT
    	vault_profile_id,
    	share_price,
    	apy_total,
		apy_usdt,
		apy_rbx,
		archive_timestamp
	FROM %s
	ORDER BY %s
	`, tableName, orderBy)

	rows, err := s.GetDB().Query(context.Background(), q)
	r.NoError(err)

	defer rows.Close()

	var response []appVaultAggregate
	for rows.Next() {
		var row appVaultAggregate
		err = rows.Scan(
			&row.VaultProfileId,
			&row.SharePrice,
			&row.APYTotal,
			&row.APYUSDT,
			&row.APYRBX,
			&row.ArchiveTimestamp,
		)
		r.NoError(err)
		response = append(response, row)
	}

	r.NoError(rows.Err())

	return response
}

func (s *dbTestSuite) assertJSON(expected, got any) {
	expectedB, err := json.Marshal(expected)
	s.NoError(err)
	gotB, err := json.Marshal(got)
	s.NoError(err)
	s.Equal(string(expectedB), string(gotB))

}

func (s *dbTestSuite) insertProfile(vaultID uint64, profileType, wallet, exchangeId string) {
	q := `INSERT INTO app_profile
	(id, profile_type, status, wallet, created_at, shard_id, archive_id, archive_timestamp, exchange_id)
	VALUES(@id, @profileType, 'active',@wallet, @now, 'profile', (select COALESCE(max(archive_id),0) + 1 from app_profile), @now, @exchangeId)
	`
	now := time.Now().UnixMicro()
	args := pgx.NamedArgs{
		"id":          vaultID,
		"profileType": profileType,
		"wallet":      wallet,
		"exchangeId":  exchangeId,
		"now":         now,
	}
	s.Execute(q, args)
}

func (s *dbTestSuite) insertVaultLast(vaultID uint64, performanceFee, totalShares decimal.Decimal, status, vaultName, managerName string, initialisedAt time.Time) {
	q := `INSERT INTO app_vault_last
	(vault_profile_id, manager_profile_id, treasurer_profile_id, performance_fee, status, total_shares, vault_name, manager_name, initialised_at, shard_id, archive_id, archive_timestamp)
	VALUES(@id, 0, 0, @performanceFee, @status, @totalShares, @vaultName, @managerName, @initialisedAt, 'profile', (select COALESCE(max(archive_id),0) + 1 from app_profile), @now)
	`
	now := time.Now().UnixMicro()
	args := pgx.NamedArgs{
		"id":             vaultID,
		"performanceFee": performanceFee,
		"totalShares":    totalShares,
		"now":            now,
		"status":         status,
		"vaultName":      vaultName,
		"managerName":    managerName,
		"initialisedAt":  initialisedAt.UnixMicro(),
	}
	s.Execute(q, args)
}

func (s *dbTestSuite) insertVaultLast2(vaultID uint64, totalShares decimal.Decimal, initialisedAt, archiveTimestamp time.Time) {
	q := `INSERT INTO app_vault_last
	(vault_profile_id, manager_profile_id, treasurer_profile_id, performance_fee, status, total_shares, vault_name, manager_name, initialised_at, shard_id, archive_id, archive_timestamp)
	VALUES(@id, 0, 0, 0, '', @totalShares, '', '', @initialisedAt, 'profile', (select COALESCE(max(archive_id),0) + 1 from app_profile), @archiveTimestamp)
	`
	now := time.Now().UnixMicro()
	args := pgx.NamedArgs{
		"id":               vaultID,
		"totalShares":      totalShares,
		"now":              now,
		"initialisedAt":    initialisedAt.UnixMicro(),
		"archiveTimestamp": archiveTimestamp.UnixMicro(),
	}
	s.Execute(q, args)
}

func (s *dbTestSuite) insertProfileCache(vaultID uint64, profileType, wallet string, accountEquity decimal.Decimal, archiveTimestamp int64) {
	q := `INSERT INTO app_profile_cache
	(id, profile_type, status, wallet, last_update, balance, account_equity, total_position_margin, total_order_margin, total_notional, account_margin, withdrawable_balance, cum_unrealized_pnl, health, account_leverage, cum_trading_volume, leverage, last_liq_check, shard_id, archive_id, archive_timestamp)
	VALUES(@id, @profileType, 'active',@wallet, @now, 0, @accountEquity, 0, 0, 0, 0, 0, 0, 0, 0, 0, '{}', 0, 'profile', (select COALESCE(max(archive_id),0) + 1 from app_profile), @archive_timestamp)
	`
	now := time.Now().UnixMicro()
	args := pgx.NamedArgs{
		"id":                vaultID,
		"profileType":       profileType,
		"wallet":            wallet,
		"now":               now,
		"accountEquity":     accountEquity,
		"archive_timestamp": archiveTimestamp,
	}
	s.Execute(q, args)
}

func (s *dbTestSuite) insertProfileCacheLast(vaultID uint64, profileType, wallet string, accountEquity decimal.Decimal) {
	q := `INSERT INTO app_profile_cache_last
	(id, profile_type, status, wallet, last_update, balance, account_equity, total_position_margin, total_order_margin, total_notional, account_margin, withdrawable_balance, cum_unrealized_pnl, health, account_leverage, cum_trading_volume, leverage, last_liq_check, shard_id, archive_id, archive_timestamp)
	VALUES(@id, @profileType, 'active',@wallet, @now, 0, @accountEquity, 0, 0, 0, 0, 0, 0, 0, 0, 0, '{}', 0, 'profile', (select COALESCE(max(archive_id),0) + 1 from app_profile), @now)
	`
	now := time.Now().UnixMicro()
	args := pgx.NamedArgs{
		"id":            vaultID,
		"profileType":   profileType,
		"wallet":        wallet,
		"now":           now,
		"accountEquity": accountEquity,
	}
	s.Execute(q, args)
}

func (s *dbTestSuite) insertProfileCacheLast2(vaultID uint64, profileType, wallet string, accountEquity decimal.Decimal, archiveTimestamp time.Time) {
	q := `INSERT INTO app_profile_cache_last
	(id, profile_type, status, wallet, last_update, balance, account_equity, total_position_margin, total_order_margin, total_notional, account_margin, withdrawable_balance, cum_unrealized_pnl, health, account_leverage, cum_trading_volume, leverage, last_liq_check, shard_id, archive_id, archive_timestamp)
	VALUES(@id, @profileType, 'active',@wallet, @now, 0, @accountEquity, 0, 0, 0, 0, 0, 0, 0, 0, 0, '{}', 0, 'profile', (select COALESCE(max(archive_id),0) + 1 from app_profile), @archiveTimestamp)
	`
	now := time.Now().UnixMicro()
	args := pgx.NamedArgs{
		"id":               vaultID,
		"profileType":      profileType,
		"wallet":           wallet,
		"now":              now,
		"accountEquity":    accountEquity,
		"archiveTimestamp": archiveTimestamp.UnixMicro(),
	}
	s.Execute(q, args)
}

func (s *dbTestSuite) insertProfileCachePeriod(period string, id uint64, item portfolio.PortfolioData) {
	q := fmt.Sprintf(`INSERT INTO app_profile_cache_%s
	(id, archive_timestamp, account_equity)
	VALUES(@id, @archive_timestamp, @account_equity)
	`, period)
	args := pgx.NamedArgs{
		"id":                id,
		"archive_timestamp": item.Time,
		"account_equity":    item.Value,
	}
	s.Execute(q, args)
}

func (s *dbTestSuite) insertVaultHoldingsLast(vaultID, staker uint64, shares, entryNav, entryPrice decimal.Decimal) {
	q := `INSERT INTO app_vault_holdings_last
	(vault_profile_id, staker_profile_id, shares, entry_nav, shard_id, archive_id, archive_timestamp, entry_price)
	VALUES(@id, @staker, @shares, @entryNav, 'profile', (select COALESCE(max(archive_id),0) + 1 from app_vault_holdings_last), @now, @entryPrice)
	`
	now := time.Now().UnixMicro()
	args := pgx.NamedArgs{
		"id":         vaultID,
		"staker":     staker,
		"shares":     shares,
		"entryNav":   entryNav,
		"now":        now,
		"entryPrice": entryPrice,
	}
	s.Execute(q, args)
}

func (s *dbTestSuite) insertVaultAggregateLast(vaultID uint64, sharePrice, apyTotal decimal.Decimal) {
	q := `INSERT INTO app_vault_aggregate_last
	(vault_profile_id, share_price, apy_total, apy_usdt, apy_rbx, archive_timestamp)
	VALUES(@id, @sharePrice, @apyTotal, 0, 0, @now)
	`
	now := time.Now().UnixMicro()
	args := pgx.NamedArgs{
		"id":         vaultID,
		"sharePrice": sharePrice,
		"apyTotal":   apyTotal,
		"now":        now,
	}
	s.Execute(q, args)
}

func (s *dbTestSuite) insertVaultAggregateLast2(item appVaultAggregate) {
	q := `INSERT INTO app_vault_aggregate_last
    (vault_profile_id, share_price, apy_total, apy_usdt, apy_rbx, archive_timestamp)
    VALUES(@vault_profile_id, @share_price, @apy_total, @apy_usdt, @apy_rbx, @archive_timestamp)
	`
	args := pgx.NamedArgs{
		"vault_profile_id":  item.VaultProfileId,
		"share_price":       item.SharePrice,
		"apy_total":         item.APYTotal,
		"apy_usdt":          item.APYUSDT,
		"apy_rbx":           item.APYRBX,
		"archive_timestamp": item.ArchiveTimestamp,
	}
	s.Execute(q, args)
}

func (s *dbTestSuite) insertBalanceOperation(id, id2, opsType string, staker uint64, wallet, exchangeId string, amount decimal.Decimal, timestamp int64, status string) {
	q := `INSERT INTO app_balance_operation
	(id, status, reason, txhash, profile_id, wallet, ops_type, ops_id2, amount, timestamp, shard_id, archive_id, archive_timestamp, due_block, exchange_id, chain_id, contract_address)
	VALUES(@id, @status, '', '', @staker, @wallet, @opsType, @id2, @amount,  @timestamp, 'profile', (select COALESCE(max(archive_id),0) + 1 from app_balance_operation), @now, 0, @exchangeId, 0, '')
	`
	now := time.Now().UnixMicro()
	args := pgx.NamedArgs{
		"id":         id,
		"staker":     staker,
		"wallet":     wallet,
		"exchangeId": exchangeId,
		"opsType":    opsType,
		"id2":        id2,
		"amount":     amount,
		"now":        now,
		"timestamp":  timestamp,
		"status":     status,
	}
	s.Execute(q, args)
}
