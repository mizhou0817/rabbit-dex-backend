package portfolio

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

type PortfolioRequest struct {
	Range string `form:"range" binding:"oneof=1h 1d 1w 1m 1y all"`
}

type PortfolioData struct {
	Time  int64           `json:"time"`
	Value decimal.Decimal `json:"value"`
}

type mappingEntry struct {
	timestampInterval string
	tableName         string
}

var mapping = map[string]mappingEntry{
	"1h":  {timestampInterval: "1 hour", tableName: "app_profile_cache_1m"},
	"1d":  {timestampInterval: "1 day", tableName: "app_profile_cache_15m"},
	"1w":  {timestampInterval: "7 days", tableName: "app_profile_cache_30m"},
	"1m":  {timestampInterval: "28 days", tableName: "app_profile_cache_1h"},
	"1y":  {timestampInterval: "365 days", tableName: "app_profile_cache_1d"},
	"all": {timestampInterval: "365 days", tableName: "app_profile_cache_1d"},
}

func HandlePortfolioList(ctx context.Context, db *pgxpool.Pool, request PortfolioRequest, profileId uint) ([]PortfolioData, error) {
	entry, found := mapping[request.Range]
	if !found {
		return nil, fmt.Errorf("RANGE_NOT_FOUND %s", request.Range)
	}
	timestampInterval := entry.timestampInterval
	tableName := entry.tableName

	q := fmt.Sprintf(`SELECT
			h.archive_timestamp,
			h.account_equity
			FROM %s as h
			WHERE h.id = @id
			AND h.archive_timestamp >= (EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - @timestamp_interval::interval)) * 1000000)::bigint
			ORDER BY 1 ASC;`, tableName)

	args := pgx.NamedArgs{
		"id":                 profileId,
		"timestamp_interval": timestampInterval,
	}

	rows, err := db.Query(ctx, q, args)
	if err != nil {
		return nil, errors.Wrap(err, "query error")
	}
	defer rows.Close()

	results := make([]PortfolioData, 0)
	for rows.Next() {
		var r PortfolioData
		err = rows.Scan(
			&r.Time,
			&r.Value)

		if err != nil {
			return nil, errors.Wrap(err, "scan row error")
		}
		results = append(results, r)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows error")
	}
	return results, nil
}
