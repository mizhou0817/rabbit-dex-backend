package tests

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/FZambia/tarantool"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TestMarketFeeSuite struct {
	suite.Suite

	ctx    context.Context
	cancel context.CancelFunc

	marketConn *tarantool.Connection
}

func (s *TestMarketFeeSuite) SetupTest() {
	s.ctx, s.cancel = context.WithTimeout(context.Background(), time.Minute)

	broker := ClearAll(s.T(), SkipInstances("api-gateway"))
	require.NotEmpty(s.T(), broker)

	s.marketConn = broker.Pool["BTC-USD"]
	require.NotNil(s.T(), s.marketConn)
}

func (s *TestMarketFeeSuite) TearDownTest() {
	cmd := `
	box.space.balance_operations:truncate()
	box.space.exchange_wallets:truncate()
	`
	require.NoError(s.T(), evalScript(s.ctx, s.marketConn, cmd))
	s.cancel()
}

func (s *TestMarketFeeSuite) TestWithdrawFee_OkPositiveBalance() {
	require.NoError(s.T(), evalScript(s.ctx, s.marketConn, `
		local balance = require("decimal").new(100)
		box.space.exchange_wallets:replace({1, balance, 123456789})
	`))

	require.NoError(s.T(), evalScript(s.ctx, s.marketConn, `
		local balance = require("decimal").new(100)

		local b = require("app.balance")
		local max_fee = require("decimal").new(10000)
		local res, err = b.withdraw_fee(max_fee, "666")
		if err ~= nil then
			error(err)
		end
		if res.error ~= nil then
			error(res.error)
		end

		if res.res ~= balance then
			error("BAD AMOUNT VALUE")
		end

		local tuples = box.space.balance_operations:select()
		if #tuples ~= 1 then
			error("BAD balance operations COUNT")
		end

		ops = tuples[1]:tomap()
		if ops.status ~= "success" then
			error("OPS STATUS SHOULD BE success")
		end
		if ops.txhash ~= "666" then
			error("OPS TXHASH SHOULD BE 666")
		end
		if ops.profile_id ~= (2^63-1ULL) then
			error("OPS PROFILE_ID SHOULD BE (2^63-1")
		end
		if ops.wallet ~= "" then
			error("OPS WALLET SHOULD BE EMPTY")
		end
		if ops.amount ~= balance then
			error("BAD OPS BALANCE")
		end

		local wallet = box.space.exchange_wallets:get({1})
		if wallet[2] ~= 0 then
			error("BAD FEE WALLET BALANCE")
		end
	`))
}

func (s *TestMarketFeeSuite) TestWithdrawFee_OkNegativeBalance() {
	require.NoError(s.T(), evalScript(s.ctx, s.marketConn, `
		local balance = require("decimal").new(-100)
		box.space.exchange_wallets:replace({1, balance, 123456789})
	`))

	require.NoError(s.T(), evalScript(s.ctx, s.marketConn, `
		local balance = require("decimal").new(-100)

		local b = require("app.balance")
		local max_fee = require("decimal").new(10000)
		local res, err = b.withdraw_fee(max_fee, "666")
		if err ~= nil then
			error(err)
		end
		if res.error ~= nil then
			error(res.error)
		end

		if res.res ~= balance then
			error("BAD AMOUNT VALUE")
		end

		local tuples = box.space.balance_operations:select()
		if #tuples ~= 1 then
			error("BAD balance operations COUNT")
		end

		ops = tuples[1]:tomap()
		if ops.status ~= "success" then
			error("OPS STATUS SHOULD BE success")
		end
		if ops.txhash ~= "666" then
			error("OPS TXHASH SHOULD BE 666")
		end
		if ops.profile_id ~= (2^63-1ULL) then
			error("OPS PROFILE_ID SHOULD BE (2^63-1")
		end
		if ops.wallet ~= "" then
			error("OPS WALLET SHOULD BE EMPTY")
		end
		if ops.amount ~= balance then
			error("BAD OPS BALANCE")
		end

		local wallet = box.space.exchange_wallets:get({1})
		if wallet[2] ~= 0 then
			error("BAD FEE WALLET BALANCE")
		end
	`))
}

func (s *TestMarketFeeSuite) TestWithdrawFee_OkMaxFeePositiveBalance() {
	require.NoError(s.T(), evalScript(s.ctx, s.marketConn, `
		local balance = require("decimal").new(10000)
		box.space.exchange_wallets:replace({1, balance, 123456789})
	`))

	require.NoError(s.T(), evalScript(s.ctx, s.marketConn, `
		local balance = require("decimal").new(10000)

		local b = require("app.balance")
		local max_fee = require("decimal").new(100)
		local res, err = b.withdraw_fee(max_fee, "666")
		if err ~= nil then
			error(err)
		end
		if res.error ~= nil then
			error(res.error)
		end

		if res.res ~= max_fee then
			error("BAD AMOUNT VALUE")
		end

		local tuples = box.space.balance_operations:select()
		if #tuples ~= 1 then
			error("BAD balance operations COUNT")
		end

		ops = tuples[1]:tomap()
		if ops.status ~= "success" then
			error("OPS STATUS SHOULD BE success")
		end
		if ops.txhash ~= "666" then
			error("OPS TXHASH SHOULD BE 666")
		end
		if ops.profile_id ~= (2^63-1ULL) then
			error("OPS PROFILE_ID SHOULD BE (2^64-1")
		end
		if ops.wallet ~= "" then
			error("OPS WALLET SHOULD BE EMPTY")
		end
		if ops.amount ~= max_fee then
			error("BAD OPS BALANCE")
		end

		local wallet = box.space.exchange_wallets:get({1})
		if wallet[2] ~= (balance - max_fee) then
			error("BAD FEE WALLET BALANCE")
		end
	`))
}

func (s *TestMarketFeeSuite) TestWithdrawFee_OkMaxFeeNegativeBalance() {
	require.NoError(s.T(), evalScript(s.ctx, s.marketConn, `
		local balance = require("decimal").new(-10000)
		box.space.exchange_wallets:replace({1, balance, 123456789})
	`))

	require.NoError(s.T(), evalScript(s.ctx, s.marketConn, `
		local balance = require("decimal").new(-10000)

		local b = require("app.balance")
		local max_fee = require("decimal").new(100)
		local res, err = b.withdraw_fee(max_fee, "666")
		if err ~= nil then
			error(err)
		end
		if res.error ~= nil then
			error(res.error)
		end

		if res.res ~= -max_fee then
			error("BAD AMOUNT VALUE")
		end

		local tuples = box.space.balance_operations:select()
		if #tuples ~= 1 then
			error("BAD balance operations COUNT")
		end

		ops = tuples[1]:tomap()
		if ops.status ~= "success" then
			error("OPS STATUS SHOULD BE success")
		end
		if ops.txhash ~= "666" then
			error("OPS TXHASH SHOULD BE 666")
		end
		if ops.profile_id ~= (2^63-1ULL) then
			error("OPS PROFILE_ID SHOULD BE (2^64-1")
		end
		if ops.wallet ~= "" then
			error("OPS WALLET SHOULD BE EMPTY")
		end
		if ops.amount ~= -max_fee then
			error("BAD OPS BALANCE")
		end

		local wallet = box.space.exchange_wallets:get({1})
		if wallet[2] ~= (balance + max_fee) then
			error("BAD FEE WALLET BALANCE")
		end
	`))
}

func (s *TestMarketFeeSuite) TestWithdrawFee_AlreadyExists() {
	require.NoError(s.T(), evalScript(s.ctx, s.marketConn, `
		local amount = require("decimal").new(100)
		box.space.balance_operations:replace({
			"1",
			"success",
			"reason",
			"tx",
			2^63,
			"0xwallet",
			"withdraw_fee",
			"666",
			amount,
			123456789,
			0,
			"rbx",
			0,
			"",
			--
			"BTC-USD",
			123456,
		})
	`))

	require.NoError(s.T(), evalScript(s.ctx, s.marketConn, `
		local b = require("app.balance")
		local max_fee = require("decimal").new(100)
		local res, err = b.withdraw_fee(max_fee, "666")
		if err ~= nil then
			error(err)
		end
		if res.error ~= nil then
			error(res.error)
		end

		local amount = require("decimal").new(100)
		if res.res ~= amount then
			error("BAD AMOUNT VALUE")
		end

		local tuples = box.space.balance_operations:select()
		if #tuples ~= 1 then
			error("BAD BALANCE OPERATIONS COUNT")
		end
	`))
}

func (s *TestMarketFeeSuite) TestWithdrawFee_NoFeeWallet() {
	require.NoError(s.T(), evalScript(s.ctx, s.marketConn, `
		local b = require("app.balance")
		local max_fee = require("decimal").new(100)
		local res, err = b.withdraw_fee(max_fee, "666")
		if err ~= nil then
			error(err)
		end
		if res.error == "FEE_WALLET_NOT_FOUND" then
			return
		end
		error("FEE WALLET FOUND")
	`))
}

func (s *TestMarketFeeSuite) TestWithdrawFee_UpdateFailure() {
	require.NoError(s.T(), evalScript(s.ctx, s.marketConn, `
		local balance = require("decimal").new(100)
		box.space.exchange_wallets:replace({1, balance, 123456789})
	`))

	require.Error(s.T(), evalScript(s.ctx, s.marketConn, `
		box.begin()

		local b = require("app.balance")
		local res, err = b.withdraw_fee("666")
		if err ~= nil then
			error(err)
		end
		if res.error ~= nil then
			error(res.error)
		end

		local amount = require("decimal").new(100)
		if res.res ~= amount then
			error("BAD AMOUNT VALUE")
		end

		local tuples = box.space.balance_operations:select()
		if #tuples ~= 1 then
			error("BAD balance operations COUNT")
		end

		local wallet = box.space.exchange_wallets:get({1})
		if wallet[2] ~= 0 then
			error("BAD FEE WALLET BALANCE")
		end
	`),
		"Operation is not permitted when there is an active transaction")
}

type TestProfileFeeSuite struct {
	suite.Suite

	ctx    context.Context
	cancel context.CancelFunc

	profileConn *tarantool.Connection
	marketConns []*tarantool.Connection
}

func (s *TestProfileFeeSuite) SetupTest() {
	s.ctx, s.cancel = context.WithTimeout(context.Background(), time.Minute)

	broker := ClearAll(s.T(), SkipInstances("api-gateway"))
	require.NotEmpty(s.T(), broker)

	s.profileConn = broker.Pool["profile"]
	require.NotNil(s.T(), s.profileConn)

	s.marketConns = nil
	for id, c := range broker.Pool {
		if strings.HasSuffix(id, "-USD") {
			require.NotNil(s.T(), c)
			s.marketConns = append(s.marketConns, c)
		}
	}
}

func (s *TestProfileFeeSuite) TearDownTest() {
	pcmd := `
	box.space.withdraw_fee:truncate()
	box.space.balance_operations:truncate()
	box.space.balance_sum:truncate()
	`
	require.NoError(s.T(), evalScript(s.ctx, s.profileConn, pcmd))

	mcmd := `
	box.space.balance_operations:truncate()
	box.space.exchange_wallets:truncate()
	`
	for _, c := range s.marketConns {
		require.NoError(s.T(), evalScript(s.ctx, c, mcmd))
	}

	s.cancel()
}

func (s *TestProfileFeeSuite) TestProfileWithdrawFee_Ok() {
	k, balance := 1, -50
	for _, c := range s.marketConns {
		require.NoError(s.T(), evalScript(s.ctx, c, `
			local balance = require("decimal").new(`+fmt.Sprint(balance)+`)
			box.space.exchange_wallets:replace({1, balance, 123456789})
		`))
		balance += 100 * k
		k += 1
	}

	require.NoError(s.T(), evalScript(s.ctx, s.profileConn, `
		local fee = require("app.profile.fee")
		local max_fee = require("decimal").new(100)
		local total_fee, err = fee.withdraw_fee(555, "0x555", max_fee)
		if err ~= nil then
			error(err)
		end

		local ops = box.space.balance_operations:select()[1]:tomap()
		if ops.status ~= "pending" then
			error("OPS STATUS SHOULD BE pending")
		end
		if ops.profile_id ~= 555 then
			error("OPS PROFILE_ID SHOULD BE 555")
		end
		if ops.wallet ~= "0x555" then
			error("OPS WALLET SHOULD BE 0x555")
		end
		if ops.amount ~= total_fee then
			error("BAD AMOUNT VALUE")
		end

		local sum = box.space.balance_sum:select({555})[1]:tomap()
		if sum.balance ~= 0 then
			error("BAD BALANCE SHOULD BE ZERO")
		end
	`))
}

func (s *TestProfileFeeSuite) TestProfileWithdrawFee_OkOnboardWallet() {
	require.NoError(s.T(), evalScript(s.ctx, s.profileConn, `
		local balance = require("decimal").new(100)
		box.space.balance_sum:replace({555, balance, 123456789})
	`))

	s.TestProfileWithdrawFee_Ok()
}

func (s *TestProfileFeeSuite) TestProfileWithdrawFee_MarketFailure() {
	for _, c := range s.marketConns[1:] {
		require.NoError(s.T(), evalScript(s.ctx, c, `
			local balance = require("decimal").new(100)
			box.space.exchange_wallets:replace({1, balance, 123456789})
		`))
	}

	require.Error(s.T(), evalScript(s.ctx, s.profileConn, `
		local fee = require("app.profile.fee")
		local max_fee = require("decimal").new(100)
		local _, err = fee.withdraw_fee(555, "0x555", max_fee)
		if err ~= nil then
			error(err)
		end
	`), "FEE_WALLET_NOT_FOUND")
}

func (s *TestProfileFeeSuite) TestProfileWithdrawFee_Duplicate() {
	k, balance := 0, -50
	amount, maxFee := 0, 100

	for _, c := range s.marketConns {
		balance += maxFee * k
		amount += minValue(balance, maxFee)
		k += 1
		require.NoError(s.T(), evalScript(s.ctx, c, `
			local balance = require("decimal").new(`+fmt.Sprint(balance)+`)
			box.space.exchange_wallets:replace({1, balance, 123456789})
		`))
	}

	require.NoError(s.T(), evalScript(s.ctx, s.profileConn, `
		local amount = require("decimal").new(`+fmt.Sprint(amount)+`)

		local fee = require("app.profile.fee")
		local max_fee = require("decimal").new(`+fmt.Sprint(maxFee)+`)
		local total_fee, err = fee.withdraw_fee(555, "0x555", max_fee)
		if err ~= nil then
			error(err)
		end

		if total_fee ~= amount then
			error("BAD AMOUNT VALUE: " .. tostring(total_fee))
		end
	`))

	require.NoError(s.T(), evalScript(s.ctx, s.profileConn, `
		local fee = require("app.profile.fee")
		local max_fee = require("decimal").new(100)
		local total_fee, err = fee.withdraw_fee(555, "0x555", max_fee)
		if err ~= nil then
			error(err)
		end

		local amount = require("decimal").new(`+fmt.Sprint(amount)+`)
		if total_fee ~= amount then
			error("DUPLICATE BAD AMOUNT VALUE: " .. tostring(total_fee))
		end
	`))
}

func TestMarketFee(t *testing.T) {
	suite.Run(t, new(TestMarketFeeSuite))
}

func TestProfileFee(t *testing.T) {
	suite.Run(t, new(TestProfileFeeSuite))
}
