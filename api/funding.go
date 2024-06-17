package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"github.com/strips-finance/rabbit-dex-backend/api/types"
)

type FundingRateListRequest struct {
	MarketId  string `form:"market_id" binding:"required"`
	TimeStamp uint64 `form:"start_time,default=0" binding:"omitempty,min=0"`
	EndTime   uint64 `form:"end_time,default=0" binding:"omitempty,min=0"`
}

type FundingRateData struct {
	MarketId    string          `json:"market_id"`
	TimeStamp   uint64          `json:"timestamp"`
	FundingRate decimal.Decimal `json:"funding_rate"`
}

func HandleFundingRateList(c *gin.Context) {
	var request FundingRateListRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)

	db := ctx.TimeScaleDB
	q := `SELECT "market_id", "timestamp", "funding_rate"
		  FROM funding_rate_hourly
          WHERE market_id = @market_id AND timestamp >= @timestamp
		  %s
          ORDER BY timestamp ` + ctx.Pagination.Order
	limit := ` LIMIT @limit OFFSET @offset`

	filters := ""
	if request.EndTime > 0 {
		filters += " AND timestamp <= @end_time"
	}

	q = fmt.Sprintf(q, filters)
	args := pgx.NamedArgs{
		"market_id": request.MarketId,
		"timestamp": request.TimeStamp,
		"order":     ctx.Pagination.Order,
		"limit":     ctx.Pagination.Limit,
		"end_time":  request.EndTime,
		"offset":    nil,
	}

	pagination := &types.PaginationResponse{
		Limit: ctx.Pagination.Limit,
		Page:  ctx.Pagination.Page,
		Order: ctx.Pagination.Order,
	}
	totalQuery := `SELECT COUNT(*) FROM (` + q + `) as t`
	db.QueryRow(c.Request.Context(), totalQuery, args).Scan(&pagination.Total)

	q = q + limit
	args["offset"] = ctx.Pagination.Limit * ctx.Pagination.Page
	rows, err := db.Query(c.Request.Context(), q, args)
	if err != nil {
		ErrorResponse(c, err)
		return
	}
	defer rows.Close()

	results := make([]FundingRateData, 0)
	for rows.Next() {
		var r FundingRateData
		err = rows.Scan(
			&r.MarketId,
			&r.TimeStamp,
			&r.FundingRate,
		)

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
	SuccessResponsePaginated(c, pagination, results...)
}
