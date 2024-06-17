package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/strips-finance/rabbit-dex-backend/api/types"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

type BalanceOpsListRequest struct {
	OpsType   []string `form:"ops_type" binding:"omitempty"`
	TimeStamp uint64   `form:"start_time,default=0" binding:"omitempty,min=0"`
	EndTime   uint64   `form:"end_time,default=0" binding:"omitempty,min=0"`
	Status    []string `form:"status" binding:"omitempty,dive,oneof=pending success processing unknown claimable claiming canceled requested"`
}

func HandleBalanceOpsList(c *gin.Context) {
	var request BalanceOpsListRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)

	db := ctx.TimeScaleDB
	q := `SELECT "id", "status", "reason", "txhash", "profile_id", "wallet", "ops_type", "ops_id2", "amount",
                 "timestamp", "shard_id", "archive_id"
		  FROM app_balance_operation
          WHERE profile_id = @profile_id AND timestamp >= @timestamp
		  %s
          ORDER BY timestamp ` + ctx.Pagination.Order
	limit := ` LIMIT @limit OFFSET @offset`

	filters := ""
	if len(request.OpsType) > 0 {
		filters += " AND ops_type = ANY (@ops_type)"
	}

	if request.EndTime > 0 {
		filters += " AND timestamp <= @end_time"
	}

	if len(request.Status) > 0 {
		filters += " AND status = ANY (@status)"
	}

	q = fmt.Sprintf(q, filters)
	args := pgx.NamedArgs{
		"profile_id": ctx.Profile.ProfileId,
		"timestamp":  request.TimeStamp,
		"order":      ctx.Pagination.Order,
		"limit":      ctx.Pagination.Limit,
		"ops_type":   request.OpsType,
		"end_time":   request.EndTime,
		"status":     request.Status,
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

	results := make([]model.BalanceOps, 0)
	for rows.Next() {
		var r model.BalanceOps
		err = rows.Scan(
			&r.OpsId,
			&r.Status,
			&r.Reason,
			&r.Txhash,
			&r.ProfileId,
			&r.Wallet,
			&r.Type,
			&r.Id2,
			&r.Amount,
			&r.Timestamp,
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
