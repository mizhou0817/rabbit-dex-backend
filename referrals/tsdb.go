package referrals

import (
	"context"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"time"
)

// tables
const PAYOUT_TABLE = "referral_payout"
const RUNNER_TABLE = "app_referral_runner"
const RUNNER_PROC_VOLUME = "volume"
const RUNNER_PROC_LEADERBOARD = "leaderboard"
const RUNNER_PROC_CREATE_PAYOUT = "create_payout"
const RUNNER_PROC_PROCESS_PAYOUT = "process_payout"

// functions
const VOLUMES_FN = "referral_get_volumes('%s')"
const FILLS_FN = "referral_get_fills('%s')"
const REFRESH_LEADERBOARD_RANK_FN = "referral_refresh_leaderboard_rank('%s', '%s')"

type volumeRow struct {
	ProfileId      uint64
	Volume         decimal.Decimal
	ExistingVolume decimal.Decimal
}

type volumeRes struct {
	Volume         decimal.Decimal
	ExistingVolume decimal.Decimal
	Model          string
}

type windowRes struct {
	ShardId        string
	ArchiveIdStart uint64
	ArchiveIdEnd   uint64
}

type referralFillRow struct {
	ReferrerId      *uint64
	InvitedId       *uint64
	ProfileId       uint64
	TradeId         string
	Fee             decimal.Decimal
	IsMaker         bool
	Model           *string
	ModelFeePercent *decimal.Decimal
	ProfileVolume   decimal.Decimal
}

type referralPayoutRow struct {
	Id        string
	ProfileId uint64
	MarketId  string
	Amount    decimal.Decimal
}

type tsdb struct {
	db         *pgxpool.Pool
	sqlBuilder sq.StatementBuilderType
}

func newTSDB(dbPool *pgxpool.Pool) *tsdb {
	return &tsdb{
		db:         dbPool,
		sqlBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (t *tsdb) updateVolume(tx pgx.Tx, profileId uint64, volume decimal.Decimal) error {
	insertBuilder := t.sqlBuilder.
		Insert("referral_volumes AS rv").
		Columns("profile_id", "volume").
		Values(profileId, volume).
		Suffix(`
			ON CONFLICT (profile_id)
			DO UPDATE
			SET volume      = rv.volume + EXCLUDED.volume,
				updated_at  = unix_now();
		`)

	sql, args, err := insertBuilder.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.Exec(context.Background(), sql, args...)
	if err != nil {
		return err
	}

	return nil
}

func (t *tsdb) saveWindowPosition(tx pgx.Tx, table string, shardId string, archiveStartId uint64, archiveEndId uint64) error {
	if archiveStartId == 0 && archiveEndId == 0 {
		return nil
	}

	insertBuilder := t.sqlBuilder.
		Insert(table).
		Columns("shard_id", "archive_id_start", "archive_id_end").
		Values(shardId, archiveStartId, archiveEndId)

	sql, args, err := insertBuilder.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.Exec(context.Background(), sql, args...)
	if err != nil {
		return err
	}

	return nil
}

func (t *tsdb) saveVolumePosition(tx pgx.Tx, shardId string, archiveStartId uint64, archiveEndId uint64) error {
	return t.saveWindowPosition(tx, "referral_volumes_integrity", shardId, archiveStartId, archiveEndId)
}

func (t *tsdb) saveFillsPosition(tx pgx.Tx, shardId string, archiveStartId uint64, archiveEndId uint64) error {
	return t.saveWindowPosition(tx, "referral_fills_integrity", shardId, archiveStartId, archiveEndId)
}

func (t *tsdb) calculateVolumes(tx pgx.Tx, shardId string) (map[uint64]volumeRes, *windowRes, error) {
	selectBuilder := t.sqlBuilder.
		Select("v.profile_id", "v.volume", "v.existing_volume", "v.archive_id", "c.model").
		From(fmt.Sprintf(VOLUMES_FN, shardId) + " as v").
		LeftJoin("app_referral_code c ON c.profile_id = v.profile_id")

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, nil, err
	}

	rows, err := tx.Query(context.Background(), sql, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	m := make(map[uint64]volumeRes)

	var archiveId uint64
	var archiveIdStart uint64
	var archiveIdEnd uint64
	var model string
	for rows.Next() {
		var r volumeRow

		err = rows.Scan(
			&r.ProfileId,
			&r.Volume,
			&r.ExistingVolume,
			&archiveId,
			&model)

		if err != nil {
			return nil, nil, err
		}

		_, ok := m[r.ProfileId]
		if !ok {
			v := volumeRes{
				Volume:         decimal.Zero,
				ExistingVolume: r.ExistingVolume,
				Model:          model,
			}

			m[r.ProfileId] = v
		}

		v := m[r.ProfileId]
		v.Volume = v.Volume.Add(r.Volume)
		m[r.ProfileId] = v

		if archiveIdStart == 0 {
			archiveIdStart = archiveId
		}
		archiveIdEnd = archiveId
	}

	if err = rows.Err(); err != nil {
		return nil, nil, err
	}

	vRes := &windowRes{ShardId: shardId, ArchiveIdEnd: archiveIdEnd, ArchiveIdStart: archiveIdStart}
	return m, vRes, nil
}

func (t *tsdb) createReferralPayout(tx pgx.Tx, profileId uint64, marketId string, amount decimal.Decimal) error {
	id := "payout-" + uuid.New().String()
	insertBuilder := t.sqlBuilder.
		Insert(PAYOUT_TABLE).
		Columns("id", "profile_id", "market_id", "amount").
		Values(id, profileId, marketId, amount)

	sql, args, err := insertBuilder.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.Exec(context.Background(), sql, args...)
	if err != nil {
		return err
	}

	return nil
}

func (t *tsdb) createBonusPayoutIntegrity(tx pgx.Tx, profileId uint64, lvl uint64) error {
	insertBuilder := t.sqlBuilder.
		Insert("referral_payout_bonus_integrity").Columns("profile_id", "level").
		Values(profileId, lvl)

	sql, args, err := insertBuilder.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.Exec(context.Background(), sql, args...)
	if err != nil {
		return err
	}

	return nil
}

func (t *tsdb) getReferralFills(tx pgx.Tx, shardId string) ([]referralFillRow, *windowRes, error) {
	selectBuilder := t.sqlBuilder.
		Select("referrer_id",
			"invited_id",
			"profile_id",
			"trade_id",
			"fee",
			"is_maker",
			"model",
			"model_fee_percent",
			"profile_volume",
			"archive_id").
		From(fmt.Sprintf(FILLS_FN, shardId))

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, nil, err
	}

	rows, err := tx.Query(context.Background(), sql, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var archiveId uint64
	var archiveIdStart uint64
	var archiveIdEnd uint64
	res := make([]referralFillRow, 0)
	for rows.Next() {
		var r referralFillRow
		err = rows.Scan(
			&r.ReferrerId,
			&r.InvitedId,
			&r.ProfileId,
			&r.TradeId,
			&r.Fee,
			&r.IsMaker,
			&r.Model,
			&r.ModelFeePercent,
			&r.ProfileVolume,
			&archiveId)

		if err != nil {
			return nil, nil, err
		}

		if archiveIdStart == 0 {
			archiveIdStart = archiveId
		}
		archiveIdEnd = archiveId

		res = append(res, r)
	}

	if err = rows.Err(); err != nil {
		return nil, nil, err
	}

	vRes := &windowRes{ShardId: shardId, ArchiveIdEnd: archiveIdEnd, ArchiveIdStart: archiveIdStart}
	return res, vRes, nil
}

func (t *tsdb) getUnProcessedPayouts(tx pgx.Tx) ([]referralPayoutRow, error) {
	selectBuilder := t.sqlBuilder.
		Select("id", "profile_id", "market_id", "amount").
		From(PAYOUT_TABLE).
		Where("processed IS FALSE")

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(context.Background(), sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]referralPayoutRow, 0)
	for rows.Next() {
		var r referralPayoutRow
		err = rows.Scan(
			&r.Id,
			&r.ProfileId,
			&r.MarketId,
			&r.Amount)

		if err != nil {
			return nil, err
		}

		res = append(res, r)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (t *tsdb) setToProcessed(tx pgx.Tx, ids []string) error {
	sqlBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	sql, args, err := sqlBuilder.
		Update(PAYOUT_TABLE).
		Set("processed", true).
		Where(sq.Eq{"id": ids}).
		ToSql()

	if err != nil {
		return err
	}

	_, err = tx.Exec(context.Background(), sql, args...)
	if err != nil {
		return err
	}

	return nil
}

func (t *tsdb) refreshLeaderBoard(exchangeId, period string) error {
	selectBuilder := t.sqlBuilder.
		Select("*").
		From(fmt.Sprintf(REFRESH_LEADERBOARD_RANK_FN, exchangeId, period))

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return err
	}

	_, err = t.db.Exec(context.Background(), sql, args...)
	if err != nil {
		return err
	}

	return nil
}

func (t *tsdb) setupRunners() error {
	sqlBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	sql, args, err := sqlBuilder.
		Insert(RUNNER_TABLE).
		Values(RUNNER_PROC_VOLUME).
		Values(RUNNER_PROC_LEADERBOARD).
		Values(RUNNER_PROC_CREATE_PAYOUT).
		Values(RUNNER_PROC_PROCESS_PAYOUT).
		Suffix("ON CONFLICT (proc_name) DO NOTHING").
		ToSql()

	if err != nil {
		return err
	}

	_, err = t.db.Exec(context.Background(), sql, args...)
	if err != nil {
		return err
	}

	return nil
}

func (t *tsdb) getRunners() (map[string]time.Time, error) {

	sqlBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	sql, args, err := sqlBuilder.
		Select("proc_name", "timestamp").
		From(RUNNER_TABLE).
		ToSql()

	if err != nil {
		return nil, err
	}

	rows, err := t.db.Query(context.Background(), sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var procName string
	var timestamp time.Time
	res := make(map[string]time.Time)
	for rows.Next() {
		err = rows.Scan(
			&procName,
			&timestamp)

		if err != nil {
			return nil, err
		}

		res[procName] = timestamp
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (t *tsdb) saveRunner(procName string) error {
	sqlBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	sql, args, err := sqlBuilder.
		Insert(RUNNER_TABLE).
		Values(procName).
		Suffix("ON CONFLICT (proc_name) DO UPDATE SET timestamp = CURRENT_TIMESTAMP").
		ToSql()

	if err != nil {
		return err
	}

	_, err = t.db.Exec(context.Background(), sql, args...)
	if err != nil {
		return err
	}

	return nil
}
