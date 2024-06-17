package api

import (
	"context"
	"errors"
	sq "github.com/Masterminds/squirrel"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/strips-finance/rabbit-dex-backend/referrals"
	"time"
)

type CreateReferralRequest struct {
	Model string `json:"model" binding:"oneof=percentage"`
}

type EditReferralRequest struct {
	NewReferralCode string `json:"new_referral_code" binding:"required,min=3,max=15"`
}

type LeaderBoardReferralRequest struct {
	Range string `form:"range" binding:"oneof=1w 1m all"`
}

type LeaderBoardReferralResponse struct {
	ProfileId      uint64          `json:"profile_id"`
	ExchangeId     string          `json:"exchange_id"`
	Volume         decimal.Decimal `json:"volume"`
	PreviousRank   uint64          `json:"previous_rank"`
	CurrentRank    uint64          `json:"current_rank"`
	Change         uint64          `json:"change"`
	InvitedCounter uint64          `json:"invited_counter"`
	Wallet         string          `json:"wallet"`
}

type ReferralResponse struct {
	ProfileId                  uint64                        `json:"profile_id"`
	ShortCode                  string                        `json:"short_code"`
	Model                      string                        `json:"model"`
	ModelFeePercent            *decimal.Decimal              `json:"model_fee_percent"`
	AmendCounter               uint64                        `json:"amend_counter"`
	Timestamp                  time.Time                     `json:"timestamp"`
	InvitedCounter             uint64                        `json:"invited_counter"`
	EarningsAllTime            decimal.Decimal               `json:"earnings_all_time"`
	Payout24h                  decimal.Decimal               `json:"payout_24h"`
	ReferralLevelStatus        referrals.ReferralLevelStatus `json:"referral_level_status"`
	LeaderBoardReStatusWeekly  LeaderBoardReferralResponse   `json:"leader_board_status_weekly"`
	LeaderBoardReStatusMonthly LeaderBoardReferralResponse   `json:"leader_board_status_monthly"`
	LeaderBoardReStatusAll     LeaderBoardReferralResponse   `json:"leader_board_status_all"`
}

func getReferralCode(db *pgxpool.Pool, profileId uint64) (*ReferralResponse, error) {
	sqlBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	sql, args, err := sqlBuilder.
		Select("profile_id", "short_code", "model", "model_fee_percent", "amend_counter", "timestamp").
		From("app_referral_code").
		Where(sq.Eq{"profile_id": profileId}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var resp ReferralResponse
	err = db.QueryRow(context.Background(), sql, args...).Scan(
		&resp.ProfileId,
		&resp.ShortCode,
		&resp.Model,
		&resp.ModelFeePercent,
		&resp.AmendCounter,
		&resp.Timestamp,
	)

	if err != nil {
		return nil, err
	}

	sqlCounter, args, err := sqlBuilder.
		Select("COALESCE(SUM(invited_counter), 0)").
		From("app_referral_counter").
		Where(sq.Eq{"profile_id": profileId}).
		ToSql()
	if err != nil {
		return nil, err
	}

	err = db.QueryRow(context.Background(), sqlCounter, args...).Scan(&resp.InvitedCounter)
	if err != nil {
		return nil, err
	}

	sqlAllTimeEarning, args, err := sqlBuilder.
		Select("COALESCE(SUM(amount), 0)").
		From("referral_payout").
		Where(sq.Eq{"profile_id": profileId}).
		ToSql()
	if err != nil {
		return nil, err
	}

	err = db.QueryRow(context.Background(), sqlAllTimeEarning, args...).Scan(&resp.EarningsAllTime)
	if err != nil {
		return nil, err
	}

	sqlPayout24h, args, err := sqlBuilder.
		Select("COALESCE(SUM(amount), 0)").
		From("referral_payout").
		Where("profile_id = ? AND timestamp >= (unix_now() - interval_to_micros('24h'))", profileId).
		ToSql()
	if err != nil {
		return nil, err
	}

	err = db.QueryRow(context.Background(), sqlPayout24h, args...).Scan(&resp.Payout24h)
	if err != nil {
		return nil, err
	}

	// total lifetime volume
	var volume decimal.Decimal
	sqlVolume, args, err := sqlBuilder.
		Select("COALESCE(volume, 0)").
		From("referral_volumes").
		Where(sq.Eq{"profile_id": profileId}).
		ToSql()
	if err != nil {
		return nil, err
	}

	err = db.QueryRow(context.Background(), sqlVolume, args...).Scan(&volume)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	sqlRankWeekly, args, err := sqlBuilder.
		Select("v.profile_id",
			"v.exchange_id",
			"COALESCE(v.current_volume, 0)",
			"COALESCE(v.previous_rank, 0)",
			"COALESCE(v.current_rank, 0)",
			"COALESCE(v.previous_rank, 0) - COALESCE(v.current_rank, 0)",
			"COALESCE(c.invited_counter, 0)",
			"p.wallet").
		From("referral_leaderboard_weekly_rank v").
		LeftJoin("app_referral_counter c ON c.profile_id = v.profile_id").
		LeftJoin("app_profile p ON p.id = v.profile_id").
		Where(sq.Eq{"v.profile_id": profileId}).
		ToSql()
	if err != nil {
		return nil, err
	}

	err = db.QueryRow(context.Background(), sqlRankWeekly, args...).Scan(
		&resp.LeaderBoardReStatusWeekly.ProfileId,
		&resp.LeaderBoardReStatusWeekly.ExchangeId,
		&resp.LeaderBoardReStatusWeekly.Volume,
		&resp.LeaderBoardReStatusWeekly.PreviousRank,
		&resp.LeaderBoardReStatusWeekly.CurrentRank,
		&resp.LeaderBoardReStatusWeekly.Change,
		&resp.LeaderBoardReStatusWeekly.InvitedCounter,
		&resp.LeaderBoardReStatusWeekly.Wallet,
	)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	sqlRankMonthly, args, err := sqlBuilder.
		Select("v.profile_id",
			"v.exchange_id",
			"COALESCE(v.current_volume, 0)",
			"COALESCE(v.previous_rank, 0)",
			"COALESCE(v.current_rank, 0)",
			"COALESCE(v.previous_rank, 0) - COALESCE(v.current_rank, 0)",
			"COALESCE(c.invited_counter, 0)",
			"p.wallet").
		From("referral_leaderboard_monthly_rank v").
		LeftJoin("app_referral_counter c ON c.profile_id = v.profile_id").
		LeftJoin("app_profile p ON p.id = v.profile_id").
		Where(sq.Eq{"v.profile_id": profileId}).
		ToSql()
	if err != nil {
		return nil, err
	}

	err = db.QueryRow(context.Background(), sqlRankMonthly, args...).Scan(
		&resp.LeaderBoardReStatusMonthly.ProfileId,
		&resp.LeaderBoardReStatusMonthly.ExchangeId,
		&resp.LeaderBoardReStatusMonthly.Volume,
		&resp.LeaderBoardReStatusMonthly.PreviousRank,
		&resp.LeaderBoardReStatusMonthly.CurrentRank,
		&resp.LeaderBoardReStatusMonthly.Change,
		&resp.LeaderBoardReStatusMonthly.InvitedCounter,
		&resp.LeaderBoardReStatusMonthly.Wallet,
	)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	sqlRankAll, args, err := sqlBuilder.
		Select("v.profile_id",
			"v.exchange_id",
			"COALESCE(v.current_volume, 0)",
			"COALESCE(v.previous_rank, 0)",
			"COALESCE(v.current_rank, 0)",
			"COALESCE(v.previous_rank, 0) - COALESCE(v.current_rank, 0)",
			"COALESCE(c.invited_counter, 0)",
			"p.wallet").
		From("referral_leaderboard_lifetime_rank v").
		LeftJoin("app_referral_counter c ON c.profile_id = v.profile_id").
		LeftJoin("app_profile p ON p.id = v.profile_id").
		Where(sq.Eq{"v.profile_id": profileId}).
		ToSql()
	if err != nil {
		return nil, err
	}

	err = db.QueryRow(context.Background(), sqlRankAll, args...).Scan(
		&resp.LeaderBoardReStatusAll.ProfileId,
		&resp.LeaderBoardReStatusAll.ExchangeId,
		&resp.LeaderBoardReStatusAll.Volume,
		&resp.LeaderBoardReStatusAll.PreviousRank,
		&resp.LeaderBoardReStatusAll.CurrentRank,
		&resp.LeaderBoardReStatusAll.Change,
		&resp.LeaderBoardReStatusAll.InvitedCounter,
		&resp.LeaderBoardReStatusAll.Wallet,
	)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	resp.ReferralLevelStatus = referrals.GetLevel(volume)

	return &resp, nil
}

func HandleReferralCreate(c *gin.Context) {
	var r CreateReferralRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		ErrorResponse(c, err)
		return
	}

	shortCode, err := referrals.GenShortCode(10)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)
	profileId := ctx.Profile.ProfileId
	sqlBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	insertBuilder := sqlBuilder.
		Insert("app_referral_code").
		Columns("profile_id", "short_code", "model").
		Values(profileId, shortCode, r.Model)

	sql, args, err := insertBuilder.ToSql()
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	db := ctx.TimeScaleDB
	_, err = db.Exec(context.Background(), sql, args...)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	resp, err := getReferralCode(db, uint64(profileId))
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, resp)
}

func HandleReferralGet(c *gin.Context) {
	ctx := GetRabbitContext(c)
	db := ctx.TimeScaleDB

	resp, err := getReferralCode(db, uint64(ctx.Profile.ProfileId))
	if errors.Is(err, pgx.ErrNoRows) {
		SuccessResponse(c, false)
		return
	}

	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, resp)
}

func HandleReferralEdit(c *gin.Context) {
	var r EditReferralRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)
	db := ctx.TimeScaleDB
	sqlBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	profileId := ctx.Profile.ProfileId
	sql, args, err := sqlBuilder.
		Update("app_referral_code").
		Set("short_code", r.NewReferralCode).
		Set("amend_counter", sq.Expr("amend_counter + 1")).
		Where(sq.Eq{"profile_id": profileId}).
		ToSql()
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	_, err = db.Exec(context.Background(), sql, args...)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	resp, err := getReferralCode(db, uint64(profileId))
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, resp)
}

func HandleGetLeaderBoard(c *gin.Context) {
	var r LeaderBoardReferralRequest
	if err := c.ShouldBindQuery(&r); err != nil {
		ErrorResponse(c, err)
		return
	}

	var view string
	switch r.Range {
	case "1w":
		view = "referral_leaderboard_weekly_rank AS v"
	case "1m":
		view = "referral_leaderboard_monthly_rank AS v"
	case "all":
		view = "referral_leaderboard_lifetime_rank AS v"
	}

	ctx := GetRabbitContext(c)

	sqlBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	selectBuilder := sqlBuilder.
		Select("v.profile_id",
			"v.exchange_id",
			"v.current_volume",
			"COALESCE(v.previous_rank, 0)",
			"COALESCE(v.current_rank, 0)",
			"COALESCE(v.previous_rank, 0) - COALESCE(v.current_rank, 0)",
			"COALESCE(c.invited_counter, 0)",
			"p.wallet").
		From(view).
		LeftJoin("app_referral_counter c ON c.profile_id = v.profile_id").
		LeftJoin("app_profile p ON p.id = v.profile_id").
		Where("v.exchange_id = ?", ctx.ExchangeId).
		OrderBy("v.current_rank ASC")

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	db := ctx.TimeScaleDB
	rows, err := db.Query(context.Background(), sql, args...)
	if err != nil {
		ErrorResponse(c, err)
		return
	}
	defer rows.Close()

	results := make([]LeaderBoardReferralResponse, 0)
	for rows.Next() {
		var r LeaderBoardReferralResponse
		err = rows.Scan(
			&r.ProfileId,
			&r.ExchangeId,
			&r.Volume,
			&r.PreviousRank,
			&r.CurrentRank,
			&r.Change,
			&r.InvitedCounter,
			&r.Wallet)

		if err != nil {
			ErrorResponse(c, err)
			return
		}
		results = append(results, r)
	}

	if err = rows.Err(); err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, results...)
}
