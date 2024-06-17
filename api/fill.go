package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/strips-finance/rabbit-dex-backend/api/types"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

type FillListRequest struct {
	MarketId  string `form:"market_id" binding:"omitempty"`
	TimeStamp uint64 `form:"start_time,default=0" binding:"omitempty,min=0"`
	EndTime   uint64 `form:"end_time,default=0" binding:"omitempty,min=0"`
}

type FillForOrderRequest struct {
	OrderId string `form:"order_id" binding:"required"`
}

func HandleFillsList(c *gin.Context) {
	var request FillListRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)

	db := ctx.TimeScaleDB
	q := `SELECT "id", "profile_id", "market_id", "order_id", "timestamp", "trade_id", "price", "size", "side",
                 "is_maker", "fee", "liquidation", "shard_id", "archive_id"
		  FROM app_fill
          WHERE profile_id = @profile_id AND timestamp >= @timestamp
		  %s
          ORDER BY timestamp ` + ctx.Pagination.Order
	limit := ` LIMIT @limit OFFSET @offset`

	filters := ""
	if request.MarketId != "" {
		filters += " AND market_id = @market_id"
	}

	if request.EndTime > 0 {
		filters += " AND timestamp <= @end_time"
	}

	q = fmt.Sprintf(q, filters)
	args := pgx.NamedArgs{
		"profile_id": ctx.Profile.ProfileId,
		"market_id":  request.MarketId,
		"timestamp":  request.TimeStamp,
		"order":      ctx.Pagination.Order,
		"limit":      ctx.Pagination.Limit,
		"end_time":   request.EndTime,
		"offset":     nil,
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

	results := make([]model.FillData, 0)
	for rows.Next() {
		var r model.FillData
		err = rows.Scan(
			&r.Id,
			&r.ProfileId,
			&r.MarketId,
			&r.OrderId,
			&r.Timestamp,
			&r.TradeId,
			&r.Price,
			&r.Size,
			&r.Side,
			&r.IsMaker,
			&r.Fee,
			&r.Liquidation,
			&r.ShardId,
			&r.ArchiveId)

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

func HandleFillsForOrder(c *gin.Context) {
	var request FillForOrderRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)

	db := ctx.TimeScaleDB
	q := `SELECT "id", "profile_id", "market_id", "order_id", "timestamp", "trade_id", "price", "size", "side",
                 "is_maker", "fee", "liquidation", "shard_id", "archive_id"
		  FROM app_fill
          WHERE profile_id = @profile_id AND order_id = @order_id`
	args := pgx.NamedArgs{
		"profile_id": ctx.Profile.ProfileId,
		"order_id":   request.OrderId,
	}

	rows, err := db.Query(c.Request.Context(), q, args)
	if err != nil {
		ErrorResponse(c, err)
		return
	}
	defer rows.Close()

	results := make([]model.FillData, 0)
	for rows.Next() {
		var r model.FillData
		err = rows.Scan(
			&r.Id,
			&r.ProfileId,
			&r.MarketId,
			&r.OrderId,
			&r.Timestamp,
			&r.TradeId,
			&r.Price,
			&r.Size,
			&r.Side,
			&r.IsMaker,
			&r.Fee,
			&r.Liquidation,
			&r.ShardId,
			&r.ArchiveId)

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
