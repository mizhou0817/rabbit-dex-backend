package gameassets

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

type BlastAssetsLoaded struct {
	ProfileID uint `db:"profile_id" json:"profile_id"`
	BatchID   uint `db:"batch_id"   json:"batch_id"`

	TradingPoints float64 `db:"trading_points"  json:"trading_points"`
	StakingPoints float64 `db:"staking_points"  json:"staking_points"`
	BonusPoints   float64 `db:"bonus_points"    json:"bonus_points"`
	TotalPoints   float64 `db:"total_points"    json:"total_points"`
	TradingGold   float64 `db:"trading_gold"    json:"trading_gold"`
	StakingGold   float64 `db:"staking_gold"    json:"staking_gold"`
	BonusGold     float64 `db:"bonus_gold"      json:"bonus_gold"`
	TotalGold     float64 `db:"total_gold"      json:"total_gold"`
	VIPExtraBoost float64 `db:"vip_extra_boost" json:"vip_extra_boost"`
}

type BfxAssetsLoaded struct {
	ProfileID uint `db:"profile_id" json:"profile_id"`
	BatchID   uint `db:"batch_id"   json:"batch_id"`

	TradingPoints    float64 `db:"trading_points"       json:"trading_points"`
	StakingPoints    float64 `db:"staking_points"       json:"staking_points"`
	BonusPoints      float64 `db:"bonus_points"         json:"bonus_points"`
	ReferralPoints   float64 `db:"referral_points"      json:"referral_points"`
	TotalPoints      float64 `db:"total_points"         json:"total_points"`
	VIPExtraBoost    float64 `db:"vip_extra_boost"      json:"vip_extra_boost"`
	Wallet           string  `db:"wallet"               json:"wallet"`
	Liquidations     float64 `db:"liquidations"         json:"liquidations"`
	ReferralBoost    float64 `db:"referral_boost"       json:"referral_boost"`
	TradingLevel     string  `db:"trading_level"        json:"trading_level"`
	TradingBoost     float64 `db:"trading_boost"        json:"trading_boost"`
	CumulativeVolume float64 `db:"cumulative_volume"    json:"cumulative_volume"`
	Timestamp        uint64  `db:"timestamp"            json:"timestamp"`

	AveragePositions     AveragePositions `db:"-"                 json:"average_positions"`
	AveragePositionsJSON []byte           `db:"average_positions" json:"-"`
}

type AveragePositions struct {
	Positions map[string]float64 `json:"positions"`
}

type BlastAssets struct {
	BlastAssetsLoaded

	StakedNotionalNet    float64 `json:"staked_notional_net"`
	ReferralBoost        float64 `json:"referral_boost"`
	LevelBoost           float64 `json:"level_boost"`
	TotalBoost           float64 `json:"total_boost"`
	TradingNotional      float64 `json:"trading_notional"`
	LevelTier            uint8   `json:"level_tier"`
	ReferrerProfileID    *uint64 `json:"referrer_profile_id"`
	ReferrerWallet       *string `json:"referrer_wallet"`
	Deposits             float64 `json:"deposits"`
	Withdraws            float64 `json:"withdraws"`
	TotalReferredDeposit float64 `json:"total_referred_deposit"`
	TotalReferredStakes  float64 `json:"total_referred_stakes"`
}

type BfxPoints struct {
	ProfileID      uint    `db:"profile_id"       json:"profile_id"`
	BatchID        uint    `db:"batch_id"         json:"batch_id"`
	ExchangeID     string  `db:"exchange_id"      json:"exchange_id"`
	BonusPoints    float64 `db:"bonus_points"     json:"bonus_points"`
	BfxPointsTotal float64 `db:"bfx_points_total" json:"bfx_points_total"`
	Timestamp      uint64  `db:"timestamp"        json:"timestamp"`

	BfxPoints24H float64 `json:"bfx_points_24h"`
}

type BfxGetPointsRequest struct {
	ProfileID uint `json:"profile_id"`
	BatchID   uint `json:"batch_id"`
}

type BlastLoadBatchResult struct {
	MaxBatchId          uint `json:"max_batch_id"`
	ReplacedRecordCount int  `json:"replaced_record_count"`
}

const BLAST_LOAD_ASSETS_MAX_BATCH_LEN = 1000

// BlastLoadAssetsBatch loads a batch of blast assets records into the database,
// overwriting data for the profile IDs from previous batches.
// Returns the LATEST batch id for Blast from the database and the number of records replaced.
func BlastLoadAssetsBatch(ctx context.Context, db *pgxpool.Pool, batch []BlastAssetsLoaded) (*BlastLoadBatchResult, error) {
	if len(batch) > BLAST_LOAD_ASSETS_MAX_BATCH_LEN {
		return nil, fmt.Errorf("batch size exceeds 1000 records")
	}

	result := &BlastLoadBatchResult{
		MaxBatchId:          0,
		ReplacedRecordCount: 0,
	}
	b := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	if len(batch) > 0 {
		insert := b.
			Insert("app_game_assets").
			Columns(
				"blockchain",
				"profile_id",
				"batch_id",
				"trading_points",
				"staking_points",
				"bonus_points",
				"total_points",
				"trading_gold",
				"staking_gold",
				"bonus_gold",
				"total_gold",
				"vip_extra_boost",
			)
		for _, r := range batch {
			insert = insert.Values(
				model.GAMEASSETS_BLAST,
				r.ProfileID,
				r.BatchID,
				r.TradingPoints,
				r.StakingPoints,
				r.BonusPoints,
				r.TotalPoints,
				r.TradingGold,
				r.StakingGold,
				r.BonusGold,
				r.TotalGold,
				r.VIPExtraBoost,
			)
		}
		// https://github.com/Masterminds/squirrel/issues/372
		onConflict := `
			ON CONFLICT (blockchain, profile_id) DO UPDATE SET
				batch_id                     = EXCLUDED.batch_id
				, trading_points             = EXCLUDED.trading_points
				, staking_points             = EXCLUDED.staking_points
				, bonus_points               = EXCLUDED.bonus_points
				, total_points               = EXCLUDED.total_points
				, trading_gold               = EXCLUDED.trading_gold
				, staking_gold               = EXCLUDED.staking_gold
				, bonus_gold                 = EXCLUDED.bonus_gold
				, total_gold                 = EXCLUDED.total_gold
			WHERE app_game_assets.batch_id <= EXCLUDED.batch_id
		RETURNING 1`
		insertSQL, args, err := insert.ToSql()
		if err != nil {
			return nil, fmt.Errorf("failed to build insert query: %w", err)
		}
		sql := fmt.Sprintf(`
			WITH inserted as NOT MATERIALIZED (
				%s
				%s
			)
			SELECT count(*) FROM inserted`,
			insertSQL, onConflict,
		)
		err = db.QueryRow(ctx, sql, args...).Scan(&result.ReplacedRecordCount)
		if err != nil {
			return nil, fmt.Errorf("failed to insert blast assets batch: %w", err)
		}
	}
	q, args, err := b.
		Select("batch_id").
		From("app_game_assets").
		Where(sq.Eq{"blockchain": model.GAMEASSETS_BLAST}).
		OrderBy("batch_id DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}
	err = db.QueryRow(ctx, q, args...).Scan(&result.MaxBatchId)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to get latest batch id: %w", err)
	}

	return result, nil
}

type BfxLoadBatchResult struct {
	MaxBatchId          uint `json:"max_batch_id"`
	InsertedRecordCount int  `json:"inserted_record_count"`
}

const BFX_LOAD_ASSETS_MAX_BATCH_LEN = 1000

// BfxLoadAssetsBatch loads a batch of bfx assets records into the database,
// Returns the number of records inserted.
func BfxLoadAssetsBatch(ctx context.Context, db *pgxpool.Pool, batch []BfxAssetsLoaded) (*BfxLoadBatchResult, error) {
	if len(batch) > BFX_LOAD_ASSETS_MAX_BATCH_LEN {
		return nil, fmt.Errorf("bfx batch size exceeds 1000 records")
	}

	// Call the function to aggregate and insert app_bfx_points
	if err := aggregateAndInsertBfxPoints(ctx, db, batch); err != nil {
		return nil, err
	}

	result := &BfxLoadBatchResult{
		MaxBatchId:          0,
		InsertedRecordCount: 0,
	}

	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	if len(batch) > 0 {
		insert := builder.
			Insert("app_bfx_game_assets").
			Columns(
				"blockchain",
				"profile_id",
				"batch_id",
				"trading_points",
				"staking_points",
				"bonus_points",
				"referral_points",
				"total_points",
				"vip_extra_boost",
				"wallet",
				"liquidations",
				"referral_boost",
				"trading_level",
				"trading_boost",
				"cumulative_volume",
				"timestamp",
				"average_positions",
			)

		for _, bfxAsset := range batch {
			// Serialize AveragePositions and assign the result to AveragePositionsJSON
			averagePositionsJSON, err := json.Marshal(bfxAsset.AveragePositions)
			if err != nil {
				return nil, fmt.Errorf("failed to serialize BfxAssetsLoaded.AveragePositions: %w", err)
			}

			insert = insert.Values(
				model.GAMEASSETS_BFX,
				bfxAsset.ProfileID,
				bfxAsset.BatchID,
				bfxAsset.TradingPoints,
				bfxAsset.StakingPoints,
				bfxAsset.BonusPoints,
				bfxAsset.ReferralPoints,
				bfxAsset.TotalPoints,
				bfxAsset.VIPExtraBoost,
				bfxAsset.Wallet,
				bfxAsset.Liquidations,
				bfxAsset.ReferralBoost,
				bfxAsset.TradingLevel,
				bfxAsset.TradingBoost,
				bfxAsset.CumulativeVolume,
				bfxAsset.Timestamp,
				averagePositionsJSON,
			)
		}

		insertSQL, args, err := insert.ToSql()
		if err != nil {
			return nil, fmt.Errorf("failed to build bfx insert query: %w", err)
		}
		sql := fmt.Sprintf(`
			WITH inserted as NOT MATERIALIZED (
				%s
				RETURNING 1
			)
			SELECT count(*) FROM inserted`,
			insertSQL,
		)
		err = db.QueryRow(ctx, sql, args...).Scan(&result.InsertedRecordCount)
		if err != nil {
			return nil, fmt.Errorf("failed to insert bfx assets batch: %w", err)
		}
	}
	querySQL, args, err := builder.
		Select("batch_id").
		From("app_bfx_game_assets").
		Where(sq.Eq{"blockchain": model.GAMEASSETS_BFX}).
		OrderBy("batch_id DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build bfx select query: %w", err)
	}
	err = db.QueryRow(ctx, querySQL, args...).Scan(&result.MaxBatchId)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to get latest bfx batch id: %w", err)
	}

	return result, nil
}

// Retrieves the mapping of ProfileID to ExchangeID from the app_profile table.
func getProfileToExchangeIDMapping(ctx context.Context, db *pgxpool.Pool, profileIDs []uint) (map[uint]string, error) {
	profileToExchangeID := make(map[uint]string)

	query := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("id", "exchange_id").
		From("app_profile").
		Where(sq.Eq{"id": profileIDs})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query for profile to exchange ID mapping: %w", err)
	}

	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query for profile to exchange ID mapping: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var profileID uint
		var exchangeID string
		if err := rows.Scan(&profileID, &exchangeID); err != nil {
			return nil, fmt.Errorf("failed to scan row for profile to exchange ID mapping: %w", err)
		}
		profileToExchangeID[profileID] = exchangeID
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over profile to exchange ID mapping rows: %w", err)
	}

	return profileToExchangeID, nil
}

// Aggregates BonusPoints and TotalPoints by ProfileID and inserts them into app_bfx_points table.
func aggregateAndInsertBfxPoints(ctx context.Context, db *pgxpool.Pool, batch []BfxAssetsLoaded) error {

	profileIDs := make([]uint, 0, len(batch))
	for _, bfxAsset := range batch {
		profileIDs = append(profileIDs, bfxAsset.ProfileID)
	}

	profileToExchangeID, err := getProfileToExchangeIDMapping(ctx, db, profileIDs)
	if err != nil {
		return err
	}

	aggregatedData := make(map[uint]BfxPoints)

	for _, bfxAsset := range batch {
		if bfxPoints, exist := aggregatedData[bfxAsset.ProfileID]; exist {
			bfxPoints.BonusPoints += bfxAsset.BonusPoints
			bfxPoints.BfxPointsTotal += bfxAsset.TotalPoints
			if bfxAsset.Timestamp > bfxPoints.Timestamp {
				bfxPoints.Timestamp = bfxAsset.Timestamp
			}
			aggregatedData[bfxAsset.ProfileID] = bfxPoints
		} else {
			aggregatedData[bfxAsset.ProfileID] = BfxPoints{
				ProfileID:      bfxAsset.ProfileID,
				BatchID:        bfxAsset.BatchID,
				ExchangeID:     profileToExchangeID[bfxAsset.ProfileID],
				BonusPoints:    bfxAsset.BonusPoints,
				BfxPointsTotal: bfxAsset.TotalPoints,
				Timestamp:      bfxAsset.Timestamp,
			}
		}
	}

	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	if len(aggregatedData) > 0 {
		insert := builder.
			Insert("app_bfx_points").
			Columns(
				"profile_id",
				"batch_id",
				"exchange_id",
				"bonus_points",
				"bfx_points_total",
				"timestamp",
			)

		for _, bfxPoints := range aggregatedData {
			insert = insert.Values(
				bfxPoints.ProfileID,
				bfxPoints.BatchID,
				bfxPoints.ExchangeID,
				bfxPoints.BonusPoints,
				bfxPoints.BfxPointsTotal,
				bfxPoints.Timestamp,
			)
		}

		insertSQL, args, err := insert.Suffix(`
			ON CONFLICT (profile_id, batch_id) 
			DO UPDATE SET 
				exchange_id 	 = EXCLUDED.exchange_id, 
				bonus_points 	 = EXCLUDED.bonus_points, 
				bfx_points_total = EXCLUDED.bfx_points_total, 
				timestamp 		 = EXCLUDED.timestamp
		`).ToSql()
		if err != nil {
			return fmt.Errorf("failed to build app_bfx_points insert query: %w", err)
		}

		_, err = db.Exec(ctx, insertSQL, args...)
		if err != nil {
			return fmt.Errorf("failed to insert into app_bfx_points: %w", err)
		}
	}

	return nil
}

const REFERRAL_BOOST_LEVEL = 0.16

func BlastGetAssets(ctx context.Context, db *pgxpool.Pool, profileID uint) (BlastAssets, error) {
	assetsFromDB := BlastAssetsLoaded{ProfileID: profileID}
	{
		a, err := getBlastAssetsFromDB(ctx, db, profileID)
		if err != nil {
			return BlastAssets{}, fmt.Errorf("failed to read blast assetsFromDB: %w", err)
		}
		if a != nil {
			assetsFromDB = *a
		}
	}

	referrerData, err := getReferrerData(ctx, db, profileID)
	if err != nil {
		return BlastAssets{}, fmt.Errorf("failed to get referrer profile id: %w", err)
	}
	tradingVolume, err := getTradingVolume(ctx, db, profileID)
	if err != nil {
		return BlastAssets{}, fmt.Errorf("failed to get trading volume: %w", err)
	}
	balOps, err := getBalanceOps(ctx, db, profileID)
	if err != nil {
		return BlastAssets{}, fmt.Errorf("failed to get balance operations: %w", err)
	}
	refBalOps, err := getReferredBalanceOps(ctx, db, profileID)
	if err != nil {
		return BlastAssets{}, fmt.Errorf("failed to get referred balance operations: %w", err)
	}

	volumeBoost := calculateVolumeBoost(tradingVolume)
	tradingNotional := tradingVolume.InexactFloat64()

	var (
		referralBoost     float64
		referrerProfileID *uint64
		referrerWallet    *string
	)
	if referrerData != nil {
		referralBoost = REFERRAL_BOOST_LEVEL
		referrerProfileID = &referrerData.ProfileID
		referrerWallet = &referrerData.Wallet
	}

	totalBoost := referralBoost + volumeBoost.Level + assetsFromDB.VIPExtraBoost
	stakeNotionalNet := balOps.totalStakes - balOps.totalUnstakes

	result := BlastAssets{
		BlastAssetsLoaded: assetsFromDB,

		LevelBoost:           volumeBoost.Level,
		LevelTier:            volumeBoost.Tier,
		TotalBoost:           totalBoost,
		TradingNotional:      tradingNotional,
		ReferralBoost:        referralBoost,
		ReferrerProfileID:    referrerProfileID,
		ReferrerWallet:       referrerWallet,
		Deposits:             balOps.totalDeposits,
		Withdraws:            balOps.totalWithdrawals,
		StakedNotionalNet:    stakeNotionalNet,
		TotalReferredDeposit: refBalOps.totalDeposits,
		TotalReferredStakes:  refBalOps.totalStakes,
	}

	return result, nil
}

// BfxGetPoints retrieves BfxPoints from the app_bfx_points table and calculates BfxPoints24H.
func BfxGetPoints(ctx context.Context, db *pgxpool.Pool, request BfxGetPointsRequest) (BfxPoints, error) {
	// Get BfxPoints from the database
	bfxPoints, err := getBfxPointsFromDB(ctx, db, request.ProfileID, request.BatchID)
	if err != nil {
		return BfxPoints{}, err
	}

	// Calculate BfxPoints24H
	bfxPoints24H, err := calculateBfxPoints24H(ctx, db, request.ProfileID)
	if err != nil {
		return BfxPoints{}, err
	}

	bfxPoints.BfxPoints24H = bfxPoints24H

	return bfxPoints, nil
}

type BlastLeaderboardRecord struct {
	Wallet          string  `db:"wallet"            json:"wallet"`
	InvitedByWallet string  `db:"invited_by_wallet" json:"invited_by_wallet"`
	TotalPoints     float64 `db:"total_points"      json:"total_points"`
	TotalGold       float64 `db:"total_gold"        json:"total_gold"`
	TradingNotional float64 `db:"trading_notional"  json:"trading_notional"`
	Deposits        float64 `db:"deposits"          json:"deposits"`
}

// BlastGetLeaderboard returns the top 100 users by Blast gold.
func BlastGetLeaderboard(ctx context.Context, db *pgxpool.Pool) ([]BlastLeaderboardRecord, error) {
	q, args, err := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select(
			"app_profile.wallet",
			"coalesce(inviter_profile.wallet, '') as invited_by_wallet",
			"app_game_assets.total_points",
			"app_game_assets.total_gold",
			"coalesce(sum(app_fill.price * app_fill.size), 0) as trading_notional",
			"coalesce(sum(app_balance_operation.amount), 0) as deposits",
		).
		From("app_game_assets").
		Join("app_profile ON app_game_assets.profile_id = app_profile.id").
		LeftJoin("app_fill USING (profile_id)").
		LeftJoin(
			`(
SELECT profile_id, amount
FROM
	app_balance_operation
WHERE
      app_balance_operation.ops_type = 'deposit'
  AND app_balance_operation.status = 'success'
) app_balance_operation USING(profile_id)`,
		).
		LeftJoin("app_referral_link ON app_game_assets.profile_id = app_referral_link.invited_id").
		LeftJoin("app_profile inviter_profile ON app_referral_link.profile_id = inviter_profile.id").
		Where(
			sq.And{
				sq.Eq{"app_game_assets.blockchain": model.GAMEASSETS_BLAST},
				sq.Gt{"app_game_assets.total_gold": 0},
			},
		).
		GroupBy(
			"app_profile.wallet",
			"invited_by_wallet",
			"app_game_assets.total_points",
			"app_game_assets.total_gold",
		).
		OrderBy("app_game_assets.total_gold DESC").
		Limit(model.GAMEASSETS_BLAST_LEADERBOARD_ROW_LIMIT).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}
	rows, err := db.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("db query error: %w", err)
	}
	records, err := pgx.CollectRows(rows, pgx.RowToStructByName[BlastLeaderboardRecord])
	if err != nil {
		return nil, fmt.Errorf("db row scan error: %w", err)
	}
	return records, nil
}

func getBlastAssetsFromDB(ctx context.Context, db *pgxpool.Pool, profileID uint) (*BlastAssetsLoaded, error) {
	b := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	q, args, err := b.
		Select(
			"batch_id",
			"profile_id",
			"trading_points",
			"staking_points",
			"bonus_points",
			"total_points",
			"trading_gold",
			"staking_gold",
			"bonus_gold",
			"total_gold",
			"vip_extra_boost",
		).
		From("app_game_assets").
		Where(sq.Eq{"blockchain": model.GAMEASSETS_BLAST}).
		Where(sq.Eq{"profile_id": profileID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}
	rows, err := db.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to read blast assets: %w", err)
	}
	records, err := pgx.CollectRows(rows, pgx.RowToStructByName[BlastAssetsLoaded])
	if err != nil {
		return nil, fmt.Errorf("failed to read blast assets: %w", err)
	}

	if len(records) == 0 {
		return nil, nil
	}
	return &records[0], nil
}

type referrerData struct {
	ProfileID uint64
	Wallet    string
}

func getReferrerData(ctx context.Context, db *pgxpool.Pool, inviteeProfileID uint) (*referrerData, error) {
	b := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	q, args, err := b.
		Select(
			"app_profile.id",
			"app_profile.wallet",
		).
		From("app_referral_link").
		Where(sq.Eq{"invited_id": inviteeProfileID}).
		Join("app_profile ON app_referral_link.profile_id = app_profile.id").
		ToSql()
	if err != nil {
		return nil, err
	}
	var result referrerData
	err = db.QueryRow(ctx, q, args...).Scan(&result.ProfileID, &result.Wallet)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &result, nil
}

type balanceOps struct {
	totalDeposits    float64
	totalWithdrawals float64
	totalStakes      float64
	totalUnstakes    float64
}

func getBalanceOps(ctx context.Context, db *pgxpool.Pool, profileID uint) (balanceOps, error) {
	b := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	q, args, err := b.
		Select("ops_type, sum(amount)").
		From("app_balance_operation").
		Where(sq.Eq{"profile_id": profileID}).
		Where(sq.Eq{"status": "success"}).
		Where(sq.Eq{"ops_type": []string{
			model.BALANCE_OPS_TYPE_DEPOSIT,
			model.BALANCE_OPS_TYPE_WITHDRAW,
			model.BALANCE_OPS_TYPE_STAKE,
			model.BALANCE_OPS_TYPE_UNSTAKE,
		}}).
		GroupBy("ops_type").
		ToSql()
	if err != nil {
		return balanceOps{}, err
	}
	rows, err := db.Query(ctx, q, args...)
	if err != nil {
		return balanceOps{}, err
	}
	defer rows.Close()
	var ops balanceOps
	for rows.Next() {
		var opsType string
		var amount float64
		err = rows.Scan(&opsType, &amount)
		if err != nil {
			return balanceOps{}, err
		}
		switch opsType {
		case model.BALANCE_OPS_TYPE_DEPOSIT:
			ops.totalDeposits = amount
		case model.BALANCE_OPS_TYPE_WITHDRAW:
			ops.totalWithdrawals = amount
		case model.BALANCE_OPS_TYPE_STAKE:
			ops.totalStakes = amount
		case model.BALANCE_OPS_TYPE_UNSTAKE:
			ops.totalUnstakes = amount
		default:
			return ops, fmt.Errorf("bad query: unexpected ops type: %s", opsType)
		}
	}

	return ops, nil
}

type referredBalanceOps struct {
	totalDeposits float64
	totalStakes   float64
}

func getReferredBalanceOps(ctx context.Context, db *pgxpool.Pool, referrerProfileID uint) (referredBalanceOps, error) {
	var ops referredBalanceOps
	b := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	q, args, err := b.
		Select(
			"ops_type",
			"sum(amount)").
		From("app_balance_operation").
		InnerJoin("app_referral_link ON app_balance_operation.profile_id = app_referral_link.invited_id").
		Where(sq.Eq{"app_referral_link.profile_id": referrerProfileID}).
		Where(sq.Eq{"app_balance_operation.status": "success"}).
		Where(sq.Eq{"app_balance_operation.ops_type": []string{
			model.BALANCE_OPS_TYPE_DEPOSIT,
			model.BALANCE_OPS_TYPE_STAKE,
		}}).
		GroupBy("ops_type").
		ToSql()
	if err != nil {
		return referredBalanceOps{}, err
	}
	rows, err := db.Query(ctx, q, args...)
	if err != nil {
		return referredBalanceOps{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var opsType string
		var amount float64
		err = rows.Scan(&opsType, &amount)
		if err != nil {
			return referredBalanceOps{}, err
		}
		switch opsType {
		case model.BALANCE_OPS_TYPE_DEPOSIT:
			ops.totalDeposits = amount
		case model.BALANCE_OPS_TYPE_STAKE:
			ops.totalStakes = amount
		default:
			return ops, fmt.Errorf("bad query: unexpected ops type: %s", opsType)
		}
	}
	return ops, nil
}

func getTradingVolume(ctx context.Context, db *pgxpool.Pool, profileID uint) (decimal.Decimal, error) {
	b := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	q, args, err := b.
		Select("coalesce(sum(price*size), 0)").
		From("app_fill").
		Where(sq.Eq{"profile_id": profileID}).
		ToSql()
	if err != nil {
		return decimal.Zero, err
	}
	var tradingVolume decimal.Decimal
	err = db.QueryRow(ctx, q, args...).Scan(&tradingVolume)
	if err == pgx.ErrNoRows {
		return decimal.Zero, nil
	}
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get trading volume: %w", err)
	}

	return tradingVolume, nil
}

type VolumeBoost struct {
	Tier  uint8
	Level float64
}

func calculateVolumeBoost(tradeVolumeInUSD decimal.Decimal) VolumeBoost {
	vol := tradeVolumeInUSD.IntPart()
	if vol < 100000 {
		return VolumeBoost{Tier: 1, Level: 0.0}
	} else if vol < 400000 {
		return VolumeBoost{Tier: 2, Level: 0.04}
	} else if vol < 1000000 {
		return VolumeBoost{Tier: 3, Level: 0.08}
	} else if vol < 8750000 {
		return VolumeBoost{Tier: 4, Level: 0.12}
	} else if vol < 64000000 {
		return VolumeBoost{Tier: 5, Level: 0.24}
	} else if vol < 460000000 {
		return VolumeBoost{Tier: 6, Level: 0.32}
	} else {
		return VolumeBoost{Tier: 7, Level: 0.40}
	}
}

// Retrieves BfxPoints from the app_bfx_points table based on ProfileID and BatchID.
func getBfxPointsFromDB(ctx context.Context, db *pgxpool.Pool, profileID uint, batchID uint) (BfxPoints, error) {
	var bfxPoints BfxPoints

	query := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("profile_id", "batch_id", "exchange_id", "bonus_points", "bfx_points_total", "timestamp").
		From("app_bfx_points").
		Where(sq.Eq{"profile_id": profileID, "batch_id": batchID})

	sql, args, err := query.ToSql()
	if err != nil {
		return bfxPoints, fmt.Errorf("failed to build query for app_bfx_points: %w", err)
	}

	row := db.QueryRow(ctx, sql, args...)
	err = row.Scan(&bfxPoints.ProfileID, &bfxPoints.BatchID, &bfxPoints.ExchangeID, &bfxPoints.BonusPoints, &bfxPoints.BfxPointsTotal, &bfxPoints.Timestamp)
	if err != nil {
		if err == pgx.ErrNoRows {
			return bfxPoints, fmt.Errorf("no app_bfx_points found for profile_id %d and batch_id %d", profileID, batchID)
		}
		return bfxPoints, fmt.Errorf("failed to scan app_bfx_points: %w", err)
	}

	return bfxPoints, nil
}

// Calculates BfxPoints24H for the given ProfileID.
func calculateBfxPoints24H(ctx context.Context, db *pgxpool.Pool, profileID uint) (float64, error) {
	var bfxPoints24H float64

	// Calculate the timestamp for 24 hours ago
	oneDayAgo := uint(time.Now().Add(-24 * time.Hour).Unix())

	query := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("COALESCE(SUM(total_points), 0)").
		From("app_bfx_game_assets").
		Where(sq.And{
			sq.Eq{"profile_id": profileID},
			sq.GtOrEq{"timestamp": oneDayAgo},
		})

	sql, args, err := query.ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build query for app_bfx_game_assets 24H: %w", err)
	}

	row := db.QueryRow(ctx, sql, args...)
	err = row.Scan(&bfxPoints24H)
	if err != nil {
		return 0, fmt.Errorf("failed to scan app_bfx_game_assets 24H: %w", err)
	}

	return bfxPoints24H, nil
}
