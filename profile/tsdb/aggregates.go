package tsdb

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

type CumVolume struct {
	ProfileId ProfileId       `db:"profile_id"`
	CumVolume decimal.Decimal `db:"volume"`
}

func (s *Store) GetVolumesAggregatesLast30d(ctx context.Context) ([]CumVolume, error) {
	columns := []string{"profile_id", "SUM(volume)"}

	now := time.Now()

	builder := s.builder.
		Select(columns...).From("app_fill_1d").
		Where(sq.GtOrEq{"bucket": now.Add(-30 * 24 * time.Hour).Truncate(24 * time.Hour).UnixMicro()}).
		GroupBy("profile_id")
	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "build query")
	}

	rows, err := s.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "run query")
	}
	defer rows.Close()

	var volumes []CumVolume
	for rows.Next() {
		var profileId ProfileId
		var volume float64

		err := rows.Scan(
			&profileId,
			&volume,
		)
		if err != nil {
			return nil, errors.Wrap(err, "scan rows")
		}

		volumes = append(volumes, CumVolume{
			ProfileId: profileId,
			CumVolume: decimal.NewFromFloat(volume),
		})
	}

	return volumes, nil
}
