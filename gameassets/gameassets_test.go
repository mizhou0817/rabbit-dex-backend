package gameassets_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/strips-finance/rabbit-dex-backend/dbtestsuite"
	"github.com/strips-finance/rabbit-dex-backend/gameassets"
	"github.com/strips-finance/rabbit-dex-backend/migrations"
)

type gameAssetsTestSuite struct {
	dbtestsuite.DBTestSuite
}

func (s *gameAssetsTestSuite) SetupSuite() {
	r := require.New(s.T())
	s.BaseSetupSuite()
	migrateDir := func(dirname string) {
		r.NoError(migrations.ApplyMigrations(
			s.MigrationConnectionString(),
			dirname,
			dirname+"_db_version",
		))
	}
	migrateDir("analytics")
	migrateDir("archiver")
	migrateDir("referrals")
}

func (s *gameAssetsTestSuite) TearDownSuite() {
	s.BaseTearDownSuite()
}

func (s *gameAssetsTestSuite) SetupTest() {
	s.truncateAllTables()
}

func TestGameAssets(t *testing.T) {
	suite.Run(t, new(gameAssetsTestSuite))
}

func (s *gameAssetsTestSuite) TestBlastAssets() {
	require := s.Require()
	assert := s.Assert()
	ctx := context.Background()

	{
		// get from empty table
		record, err := gameassets.BlastGetAssets(ctx, s.GetDB(), 1)
		require.NoError(err)
		require.Equal(
			gameassets.BlastAssets{
				BlastAssetsLoaded: gameassets.BlastAssetsLoaded{ProfileID: 1},
				LevelTier:         1,
			},
			record,
		)
	}

	{ // load empty batch
		res, err := gameassets.BlastLoadAssetsBatch(ctx, s.GetDB(), []gameassets.BlastAssetsLoaded{})
		require.NoError(err)
		require.NotNil(res)
		require.Equal(
			&gameassets.BlastLoadBatchResult{MaxBatchId: 0, ReplacedRecordCount: 0},
			res)
	}

	var loadedRecords []gameassets.BlastAssetsLoaded

	{ // load non-empty batch
		records := []gameassets.BlastAssetsLoaded{
			{
				BatchID:   1,
				ProfileID: 1,

				TradingPoints: 101.1,
				StakingPoints: 102.2,
				BonusPoints:   103.3,
				TotalPoints:   104.4,
				TradingGold:   105.5,
				StakingGold:   106.6,
				BonusGold:     107.7,
				TotalGold:     108.8,
				VIPExtraBoost: 0,
			},
			{
				BatchID:   1,
				ProfileID: 2,

				TradingPoints: 201.1,
				StakingPoints: 202.2,
				BonusPoints:   203.3,
				TotalPoints:   204.4,
				TradingGold:   205.5,
				StakingGold:   206.6,
				BonusGold:     207.7,
				TotalGold:     208.8,
				VIPExtraBoost: 0,
			},
		}
		for i := uint(0); i < 998; i++ {
			records = append(records,
				gameassets.BlastAssetsLoaded{
					BatchID:   1,
					ProfileID: 1000000 + i,

					TradingPoints: 1.1,
					StakingPoints: 2.2,
					BonusPoints:   3.3,
					TotalPoints:   4.4,
					TradingGold:   5.5,
					StakingGold:   6.6,
					BonusGold:     7.7,
					TotalGold:     8.8,
					VIPExtraBoost: 0.001 * float64(i),
				},
			)
		}
		require.Equal(1000, len(records))
		res, err := gameassets.BlastLoadAssetsBatch(ctx, s.GetDB(), records)
		require.NoError(err)
		require.NotNil(res)
		require.Equal(
			&gameassets.BlastLoadBatchResult{MaxBatchId: 1, ReplacedRecordCount: 1000},
			res)
		loadedRecords = records
	}

	{ // get from non-empty table
		gotAssets0, err := gameassets.BlastGetAssets(ctx, s.GetDB(), 1)
		require.NoError(err)
		require.NotNil(gotAssets0)

		gotAssets1, err := gameassets.BlastGetAssets(ctx, s.GetDB(), 2)
		require.NoError(err)
		require.NotNil(gotAssets1)

		wantAssets0 := gameassets.BlastAssets{BlastAssetsLoaded: loadedRecords[0], LevelTier: 1}
		wantAssets1 := gameassets.BlastAssets{BlastAssetsLoaded: loadedRecords[1], LevelTier: 1}

		require.Equal(wantAssets0, gotAssets0)
		require.Equal(wantAssets1, gotAssets1)
	}

	{ // subsequent load of the same data
		res, err := gameassets.BlastLoadAssetsBatch(ctx, s.GetDB(), loadedRecords)
		require.NoError(err)
		require.Equal(
			&gameassets.BlastLoadBatchResult{MaxBatchId: 1, ReplacedRecordCount: 1000},
			res)
	}

	{ // load batch with id less than max
		records := []gameassets.BlastAssetsLoaded{
			{
				BatchID:   0,
				ProfileID: 1000005,
			},
		}
		res, err := gameassets.BlastLoadAssetsBatch(ctx, s.GetDB(), records)
		require.NoError(err)
		require.Equal(
			&gameassets.BlastLoadBatchResult{MaxBatchId: 1, ReplacedRecordCount: 0},
			res)
	}

	{ // load batch with id more than max
		records := []gameassets.BlastAssetsLoaded{
			{
				BatchID:   2,
				ProfileID: 1000006,
			},
			{
				BatchID:   3,
				ProfileID: 1000007,
			},
		}
		res, err := gameassets.BlastLoadAssetsBatch(ctx, s.GetDB(), records)
		require.NoError(err)
		require.Equal(
			&gameassets.BlastLoadBatchResult{MaxBatchId: 3, ReplacedRecordCount: 2},
			res)
	}

	{ // check calculated fields
		{
			invitee := 1000001
			inviter := 1
			s.addReferralLink(inviter, invitee)
			s.addFill(map[string]any{"profile_id": invitee, "price": 64000000, "size": 1})
			s.addFill(map[string]any{"profile_id": invitee, "price": 123, "size": 20})
			s.addBalanceOp("deposit", invitee, 1234567890)
			s.addBalanceOp("withdraw", invitee, 12345)
			s.addBalanceOp("stake", invitee, 5000)
			s.addBalanceOp("unstake", invitee, 10)
			s.addProfile(invitee, "0x12345", "exchangeid")
			s.addProfile(inviter, "0x1", "exchangeid")
		}
		{
			inviter := 1000001
			invitee1 := 1000991
			invitee2 := 1000992
			s.addReferralLink(inviter, invitee1)
			s.addReferralLink(inviter, invitee2)
			s.addBalanceOp("deposit", invitee1, 101)
			s.addBalanceOp("stake", invitee1, 102)
			s.addBalanceOp("deposit", invitee2, 201)
			s.addBalanceOp("stake", invitee2, 202)
		}

		record, err := gameassets.BlastGetAssets(ctx, s.GetDB(), 1000001)
		require.NoError(err)
		require.NotNil(record)
		assert.EqualValues(0.16, record.ReferralBoost)
		assert.EqualValues(0.32, record.LevelBoost)
		assert.EqualValues(0.481, record.TotalBoost) // referral + level + vip
		assert.EqualValues(6, record.LevelTier)
		assert.EqualValues(64000000+20*123, record.TradingNotional)
		require.NotNil(record.ReferrerProfileID)
		require.NotNil(record.ReferrerWallet)
		assert.EqualValues(1, *record.ReferrerProfileID)
		assert.EqualValues("0x1", *record.ReferrerWallet)
		assert.EqualValues(1234567890, record.Deposits)
		assert.EqualValues(12345, record.Withdraws)
		assert.EqualValues(4990, record.StakedNotionalNet)
		assert.EqualValues(302, record.TotalReferredDeposit)
		assert.EqualValues(304, record.TotalReferredStakes)
		assert.EqualValues(0.001, record.VIPExtraBoost)
	}
}

func (s *gameAssetsTestSuite) TestBlastLeaderboard() {
	require, assert := s.Require(), s.Assert()
	ctx := context.Background()

	{ // setup
		s.addProfile(0, "wallet0", "exchangeid")
		s.addProfile(1, "wallet1", "exchangeid")
		s.addProfile(2, "wallet2", "exchangeid")
		inviter, invitee := 1, 2
		s.addReferralLink(inviter, invitee)
		s.addFill(map[string]any{"profile_id": 2, "price": 2222, "size": 1})
		s.addFill(map[string]any{"profile_id": 1, "price": 1111, "size": 1})
		s.addBalanceOp("deposit", 2, 22222)
		s.addBalanceOp("deposit", 1, 11111)

		batch := []gameassets.BlastAssetsLoaded{
			{
				ProfileID:   0,
				TotalGold:   0,
				TotalPoints: 0,
			},
			{
				ProfileID:   1,
				TotalGold:   1.1,
				TotalPoints: 11.1,
			},
			{
				ProfileID:   2,
				TotalGold:   2.2,
				TotalPoints: 22.2,
			},
		}

		loadResult, err := gameassets.BlastLoadAssetsBatch(ctx, s.GetDB(), batch)
		require.NoError(err)
		require.Equal(&gameassets.BlastLoadBatchResult{MaxBatchId: 0, ReplacedRecordCount: 3}, loadResult)
	}

	wantLeaderboard := []gameassets.BlastLeaderboardRecord{
		{
			Wallet:          "wallet2",
			InvitedByWallet: "wallet1",
			TotalGold:       2.2,
			TotalPoints:     22.2,
			TradingNotional: 2222,
			Deposits:        22222,
		},
		{
			Wallet:          "wallet1",
			InvitedByWallet: "",
			TotalGold:       1.1,
			TotalPoints:     11.1,
			TradingNotional: 1111,
			Deposits:        11111,
		},
	}
	gotLeaderboard, err := gameassets.BlastGetLeaderboard(ctx, s.GetDB())
	require.NoError(err)
	assert.Equal(wantLeaderboard, gotLeaderboard)
}

func (s *gameAssetsTestSuite) truncateAllTables() {
	var tables []string
	// postgres
	rows, err := s.GetDB().Query(context.Background(), `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = current_schema()
	`)
	s.Require().NoError(err)
	defer rows.Close()
	for rows.Next() {
		var table string
		err = rows.Scan(&table)
		s.Require().NoError(err)
		tables = append(tables, table)
	}
	for _, table := range tables {
		_, err = s.GetDB().Exec(context.Background(), "truncate table "+table)
		s.Require().NoError(err)
	}
}

var ids = 999999

func nextId() int {
	ids++
	return ids
}

func (s *gameAssetsTestSuite) insert(tableName string, row map[string]any) {
	sql, args := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Insert(tableName).
		SetMap(row).
		MustSql()
	_, err := s.GetDB().Exec(context.Background(), sql, args...)
	s.Require().NoError(err)
}

func (s *gameAssetsTestSuite) addReferralLink(inviter int, invitee int) {
	row := map[string]any{
		"profile_id": inviter,
		"invited_id": invitee,
	}
	s.insert("app_referral_link", row)
}

func (s *gameAssetsTestSuite) addProfile(profileId int, wallet string, exchangeId string) {
	row := map[string]any{
		"id":                profileId,
		"profile_type":      "profile_type",
		"status":            "status",
		"wallet":            wallet,
		"created_at":        1712664891 + nextId(),
		"exchange_id":       exchangeId,
		"shard_id":          "1",
		"archive_id":        nextId(),
		"archive_timestamp": 1712664891 + nextId(),
	}
	s.insert("app_profile", row)
}

func (s *gameAssetsTestSuite) addBalanceOp(opType string, profileId int, amount int) {
	row := map[string]any{
		"id":                "id" + strconv.Itoa(nextId()),
		"status":            "success",
		"reason":            "reason",
		"txhash":            "txhash" + strconv.Itoa(nextId()),
		"profile_id":        profileId,
		"wallet":            "wallet" + strconv.Itoa(profileId),
		"ops_type":          opType,
		"ops_id2":           "ops_id2",
		"amount":            amount,
		"timestamp":         1712664891 + nextId(),
		"exchange_id":       "exchange_id",
		"chain_id":          1,
		"contract_address":  "contract_address",
		"shard_id":          "1",
		"archive_id":        nextId(),
		"archive_timestamp": 1712664891 + nextId(),
		"due_block":         1,
	}
	s.insert("app_balance_operation", row)
}

func (s *gameAssetsTestSuite) addFill(colVals map[string]any) {
	getOr := func(key string, fallback any) any {
		if val, ok := colVals[key]; ok {
			return val
		}
		return fallback
	}
	row := map[string]any{
		"id":                getOr("id", "id"+strconv.Itoa(nextId())),
		"profile_id":        getOr("profile_id", nextId()),
		"market_id":         "ETH-USDT",
		"order_id":          getOr("order_id", "order_id"+strconv.Itoa(nextId())),
		"timestamp":         1712664891 + nextId(),
		"trade_id":          getOr("trade_id", "trade_id"+strconv.Itoa(nextId())),
		"price":             getOr("price", 1000*nextId()),
		"size":              getOr("size", nextId()),
		"side":              getOr("side", "buy"),
		"is_maker":          getOr("is_maker", false),
		"fee":               getOr("fee", 0.1),
		"liquidation":       false,
		"shard_id":          "1",
		"archive_id":        getOr("archive_id", nextId()),
		"archive_timestamp": 1712664891 + nextId(),
		"client_order_id":   getOr("client_order_id", "client_order_id"+strconv.Itoa(nextId())),
	}
	s.insert("app_fill", row)
}

func (s *gameAssetsTestSuite) TestDBError() {
	require := s.Require()
	ctx := context.Background()

	_, err := s.GetDB().Exec(ctx, "drop table app_game_assets")
	require.NoError(err)

	res, err := gameassets.BlastLoadAssetsBatch(ctx, s.GetDB(), []gameassets.BlastAssetsLoaded{
		{
			BatchID:   2,
			ProfileID: 1000006,
		},
	})
	require.Error(err)
	require.Nil(res)

	record, err := gameassets.BlastGetAssets(ctx, s.GetDB(), 1)
	require.Error(err)
	require.Equal(gameassets.BlastAssets{}, record)
}

func (s *gameAssetsTestSuite) TestBatchTooLarge() {
	require := s.Require()
	ctx := context.Background()

	var records []gameassets.BlastAssetsLoaded
	for i := 0; i < 1001; i++ {
		records = append(records, gameassets.BlastAssetsLoaded{
			BatchID:   1,
			ProfileID: 1,
		})
	}

	res, err := gameassets.BlastLoadAssetsBatch(ctx, s.GetDB(), records)
	require.ErrorContains(err, "batch size exceeds 1000 records")
	require.Nil(res)
}

func (s *gameAssetsTestSuite) TestBfxLoadAssetsBatch() {
	require := s.Require()
	ctx := context.Background()

	// Create a batch with the same batch_id and multiple records for the same profile_id
	batchID := uint(1)
	profileID := uint(123)
	batch := []gameassets.BfxAssetsLoaded{
		{
			ProfileID:        profileID,
			BatchID:          batchID,
			TradingPoints:    10.0,
			StakingPoints:    20.0,
			BonusPoints:      5.0,
			ReferralPoints:   2.0,
			TotalPoints:      37.0,
			VIPExtraBoost:    1.0,
			Wallet:           "wallet1",
			Liquidations:     0.0,
			ReferralBoost:    0.0,
			TradingLevel:     "level1",
			TradingBoost:     0.0,
			CumulativeVolume: 100.0,
			Timestamp:        uint64(time.Now().Unix()),
			AveragePositions: gameassets.AveragePositions{
				Positions: map[string]float64{"BTC": 0.5},
			},
		},
		{
			ProfileID:        profileID,
			BatchID:          batchID,
			TradingPoints:    15.0,
			StakingPoints:    25.0,
			BonusPoints:      10.0,
			ReferralPoints:   5.0,
			TotalPoints:      55.0,
			VIPExtraBoost:    2.0,
			Wallet:           "wallet2",
			Liquidations:     0.0,
			ReferralBoost:    0.0,
			TradingLevel:     "level2",
			TradingBoost:     0.0,
			CumulativeVolume: 200.0,
			Timestamp:        uint64(time.Now().Unix()),
			AveragePositions: gameassets.AveragePositions{
				Positions: map[string]float64{"ETH": 1.0},
			},
		},
	}

	s.addProfile(int(profileID), "wallet1", "exchangeid123")

	// Test basic functionality
	result, err := gameassets.BfxLoadAssetsBatch(ctx, s.GetDB(), batch)
	require.NoError(err)
	require.NotNil(result)
	require.Equal(2, result.InsertedRecordCount)
	// Verify aggregated data in app_bfx_points table
	row := s.GetDB().QueryRow(ctx, `
		SELECT exchange_id, bonus_points, bfx_points_total
		FROM app_bfx_points
		WHERE profile_id = $1 AND batch_id = $2
	`, profileID, batchID)
	var exchangeId string
	var bonusPoints, bfxPointsTotal float64
	err = row.Scan(&exchangeId, &bonusPoints, &bfxPointsTotal)
	require.NoError(err)
	require.Equal("exchangeid123", exchangeId)
	require.Equal(15.0, bonusPoints)
	require.Equal(92.0, bfxPointsTotal)
	// Test batch size limit
	largeBatch := make([]gameassets.BfxAssetsLoaded, gameassets.BFX_LOAD_ASSETS_MAX_BATCH_LEN+1)
	for i := range largeBatch {
		largeBatch[i] = batch[0]
	}
	_, err = gameassets.BfxLoadAssetsBatch(ctx, s.GetDB(), largeBatch)
	require.Error(err)
	require.Contains(err.Error(), "bfx batch size exceeds 1000 records")
	// Test performance with a large batch
	performanceBatch := make([]gameassets.BfxAssetsLoaded, gameassets.BFX_LOAD_ASSETS_MAX_BATCH_LEN)
	for i := range performanceBatch {
		performanceBatch[i] = batch[0]
	}
	startTime := time.Now()
	_, err = gameassets.BfxLoadAssetsBatch(ctx, s.GetDB(), performanceBatch)
	require.NoError(err)
	duration := time.Since(startTime)
	require.Less(duration.Seconds(), 5.0, "Performance test should complete within 5 seconds")
	// Test concurrent processing
	concurrency := 10
	ch := make(chan error, concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			_, err := gameassets.BfxLoadAssetsBatch(ctx, s.GetDB(), batch)
			ch <- err
		}()
	}
	for i := 0; i < concurrency; i++ {
		err := <-ch
		require.NoError(err)
	}
}

func (s *gameAssetsTestSuite) TestBfxGetPoints() {
	require := s.Require()
	ctx := context.Background()
	// Step 1: Insert test data using BfxLoadAssetsBatch
	profileID := uint(123)
	batchID := uint(1)
	now := time.Now()
	batch := []gameassets.BfxAssetsLoaded{
		{
			ProfileID:        profileID,
			BatchID:          batchID,
			TradingPoints:    10.0,
			StakingPoints:    20.0,
			BonusPoints:      5.0,
			ReferralPoints:   2.0,
			TotalPoints:      37.0,
			VIPExtraBoost:    1.0,
			Wallet:           "wallet1",
			Liquidations:     0.0,
			ReferralBoost:    0.0,
			TradingLevel:     "level1",
			TradingBoost:     0.0,
			CumulativeVolume: 100.0,
			Timestamp:        uint64(now.Unix()),
			AveragePositions: gameassets.AveragePositions{
				Positions: map[string]float64{"BTC": 0.5},
			},
		},
		{
			ProfileID:        profileID,
			BatchID:          batchID,
			TradingPoints:    15.0,
			StakingPoints:    25.0,
			BonusPoints:      10.0,
			ReferralPoints:   5.0,
			TotalPoints:      55.0,
			VIPExtraBoost:    2.0,
			Wallet:           "wallet2",
			Liquidations:     0.0,
			ReferralBoost:    0.0,
			TradingLevel:     "level2",
			TradingBoost:     0.0,
			CumulativeVolume: 200.0,
			Timestamp:        uint64(now.Add(-23 * time.Hour).Unix()), // within 24 hours
			AveragePositions: gameassets.AveragePositions{
				Positions: map[string]float64{"ETH": 1.0},
			},
		},
		{
			ProfileID:        profileID,
			BatchID:          batchID,
			TradingPoints:    5.0,
			StakingPoints:    10.0,
			BonusPoints:      2.0,
			ReferralPoints:   1.0,
			TotalPoints:      18.0,
			VIPExtraBoost:    0.5,
			Wallet:           "wallet3",
			Liquidations:     0.0,
			ReferralBoost:    0.0,
			TradingLevel:     "level3",
			TradingBoost:     0.0,
			CumulativeVolume: 50.0,
			Timestamp:        uint64(now.Add(-25 * time.Hour).Unix()), // outside 24 hours
			AveragePositions: gameassets.AveragePositions{
				Positions: map[string]float64{"LTC": 0.3},
			},
		},
	}
	// Insert profile for the given profileID
	s.addProfile(int(profileID), "wallet1", "exchangeid123")
	// Load the batch into the database
	result, err := gameassets.BfxLoadAssetsBatch(ctx, s.GetDB(), batch)
	require.NoError(err)
	require.NotNil(result)
	require.Equal(3, result.InsertedRecordCount)
	// Step 2: Call BfxGetPoints to retrieve the calculated points
	request := gameassets.BfxGetPointsRequest{
		ProfileID: profileID,
		BatchID:   batchID,
	}
	points, err := gameassets.BfxGetPoints(ctx, s.GetDB(), request)
	require.NoError(err)

	// Step 3: Validate the total points calculated in the last 24 hours from all records for the profile id
	expectedTotalPoints24H := 37.0 + 55.0
	require.Equal(expectedTotalPoints24H, points.BfxPoints24H)

	// Step 4: Validate other fields
	expectedTotalPoints := 37.0 + 55.0 + 18.0 // Total points for the batch id + profile id
	expectedBonusPoints := 5.0 + 10.0 + 2.0   // Total bonus points for the batch id + profile id
	expectedTimestamp := uint64(now.Unix())
	require.Equal(profileID, points.ProfileID)
	require.Equal(batchID, points.BatchID)
	require.Equal("exchangeid123", points.ExchangeID)
	require.Equal(expectedTotalPoints, points.BfxPointsTotal)
	require.Equal(expectedBonusPoints, points.BonusPoints)
	require.Equal(expectedTimestamp, points.Timestamp)
}
