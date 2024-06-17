package volume

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"

	"github.com/strips-finance/rabbit-dex-backend/model"
)

type VolumeRequest struct {
	StartDate int64 `binding:"gt=0,required" form:"start_date"`
	EndDate   int64 `binding:"gt=0,required" form:"end_date"`
}

type VolumeResponse struct {
	Volume decimal.Decimal `json:"volume"`
}

func HandleBfxVolume(ctx context.Context, db *pgxpool.Pool, request VolumeRequest) (*VolumeResponse, error) {
	return getVolumeForExchange(ctx, db, model.EXCHANGE_BFX, request.StartDate, request.EndDate)
}

func HandleRbxVolume(
	ctx context.Context,
	db *pgxpool.Pool,
	request VolumeRequest,
) (*VolumeResponse, error) {
	return getVolumeForExchange(ctx, db, model.EXCHANGE_RBX, request.StartDate, request.EndDate)
}

func getVolumeForExchange(
	ctx context.Context,
	db *pgxpool.Pool,
	exchangeId string,
	startDate, endDate int64,
) (*VolumeResponse, error) {
	sql := `
SELECT COALESCE(SUM(f.volume), 0) AS volume
FROM app_fill_1h AS f 
JOIN app_profile AS p ON p.id = f.profile_id
WHERE p.exchange_id = @exchange_id
AND f.timestamp >= time_bucket(3600000000::bigint, @timestamp_from)
AND f.timestamp <= time_bucket(3600000000::bigint, @timestamp_to)
`
	args := pgx.NamedArgs{
		"exchange_id":    exchangeId,
		"timestamp_from": startDate,
		"timestamp_to":   endDate,
	}

	var r VolumeResponse
	err := db.QueryRow(ctx, sql, args).Scan(&r.Volume)
	if err != nil {
		return nil, errors.Wrap(err, "execute query")
	}

	return &r, nil
}
