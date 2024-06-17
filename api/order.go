package api

import (
	"fmt"
	"strings"

	"github.com/strips-finance/rabbit-dex-backend/api/types"

	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

type OrderListRequest struct {
	MarketId      string   `form:"market_id" binding:"omitempty"`
	TimeStamp     uint64   `form:"start_time,default=0" binding:"omitempty,min=0"`
	EndTime       uint64   `form:"end_time,default=0" binding:"omitempty,min=0"`
	Status        []string `form:"status" binding:"omitempty,dive,oneof=processing open closed rejected canceled canceling amending cancelingall placed"`
	OrderId       string   `form:"order_id" binding:"omitempty"`
	ClientOrderId string   `form:"client_order_id" binding:"omitempty"`
	OrderType     []string `form:"order_type" binding:"omitempty,dive,oneof=limit market stop_loss take_profit stop_loss_limit take_profit_limit stop_market stop_limit cancel amend"`
}

type OrderCreateRequest struct {
	MarketId      string   `json:"market_id" binding:"required"`
	Type          string   `json:"type" binding:"oneof=ping_limit limit market stop_loss take_profit stop_loss_limit take_profit_limit stop_market stop_limit cancel amend,required"`
	Side          string   `json:"side" binding:"required_unless=Type ping_limit Type stop_loss Type take_profit Type stop_loss_limit Type take_profit_limit,omitempty,oneof=short long"`
	Price         *float64 `json:"price" binding:"required_if=Type limit Type stop_limit Type stop_loss_limit Type take_profit_limit,omitempty"`
	Size          *float64 `json:"size" binding:"required_unless=Type stop_loss Type take_profit Type stop_loss_limit Type take_profit_limit,omitempty"`
	ClientOrderId *string  `json:"client_order_id" binding:"omitempty"`
	TriggerPrice  *float64 `json:"trigger_price" binding:"required_if=Type stop_loss Type take_profit Type stop_loss_limit Type take_profit_limit Type stop_market Type stop_limit,omitempty"`
	SizePercent   *float64 `json:"size_percent" binding:"required_if=Type stop_loss Type take_profit Type stop_loss_limit Type take_profit_limit,omitempty"`
	TimeInForce   *string  `json:"time_in_force" binding:"omitempty,oneof=good_till_cancel immediate_or_cancel fill_or_kill post_only"`
	IsPm          bool     `json:"is_pm" binding:"omitempty"`
}

type OrderAmendRequest struct {
	OrderId      string   `json:"order_id" binding:"required"`
	MarketId     string   `json:"market_id" binding:"required"`
	Price        *float64 `json:"price" binding:"omitempty"`
	Size         *float64 `json:"size" binding:"omitempty"`
	TriggerPrice *float64 `json:"trigger_price" binding:"omitempty"`
	SizePercent  *float64 `json:"size_percent" binding:"omitempty"`
}

type OrderCancelRequest struct {
	OrderId       string `json:"order_id" binding:"omitempty"`
	MarketId      string `json:"market_id" binding:"required"`
	ClientOrderId string `json:"client_order_id" binding:"omitempty"`
}

func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}

	text := strings.ToLower(err.Error())

	if strings.Contains(text, "rate_limit") {
		return true
	}

	return false
}

func HandleOrderCreate(c *gin.Context) {
	var request OrderCreateRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	//CHECKPROFILE
	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	// profile_id uint, market_id, order_tpye, side string, price, size float64
	ctx.Meta.SetPm(request.IsPm)
	res, err := apiModel.OrderCreate(c.Request.Context(),
		ctx.Profile.ProfileId,
		request.MarketId,
		request.Type,
		request.Side,
		request.Price,
		request.Size,
		request.ClientOrderId,
		request.TriggerPrice,
		request.SizePercent,
		request.TimeInForce,

		ctx.Meta,
	)
	if err != nil {
		if isRateLimitError(err) {
			RateLimitErrorResponse(c, err)
			return
		}
		ErrorResponse(c, err)
		return
	}

	logrus.
		Info("Order task created")

	SuccessResponse(c, res)
}

func HandleOrderCancel(c *gin.Context) {
	var request OrderCancelRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	//CHECKPROFILE
	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	// profile_id uint, market_id, order_type, side string, price, size float64

	res, err := apiModel.OrderCancel(c.Request.Context(),
		ctx.Profile.ProfileId,
		request.MarketId,
		request.OrderId,
		request.ClientOrderId)

	if err != nil {
		if isRateLimitError(err) {
			RateLimitErrorResponse(c, err)
			return
		}
		ErrorResponse(c, err)
		return
	}

	logrus.
		Info("Order cancel request sent")

	SuccessResponse(c, res)
}

func HandleOrderCancelAll(c *gin.Context) {
	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	err := apiModel.CancelAll(c.Request.Context(), ctx.Profile.ProfileId, false)
	if err != nil {
		if isRateLimitError(err) {
			RateLimitErrorResponse(c, err)
			return
		}
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, true)
}

func HandleOrderAmend(c *gin.Context) {
	var request OrderAmendRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	//CHECKPROFILE
	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	res, err := apiModel.OrderAmend(c.Request.Context(),
		ctx.Profile.ProfileId,
		request.MarketId,
		request.OrderId,
		request.Price,
		request.Size,
		request.TriggerPrice,
		request.SizePercent,
	)
	if err != nil {
		if isRateLimitError(err) {
			RateLimitErrorResponse(c, err)
			return
		}
		ErrorResponse(c, err)
		return
	}

	logrus.
		Info("Order amend request sent")

	SuccessResponse(c, res)
}

func HandleOrdersList(c *gin.Context) {
	var request OrderListRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)

	db := ctx.TimeScaleDB
	q := `SELECT "id", "profile_id", "market_id", "order_type", "status", "price", "size", "initial_size",
                 "total_filled_size", "side", "timestamp", "reason", "client_order_id",
				 "trigger_price", "size_percent", "time_in_force", "shard_id", "archive_id"
		  FROM app_order
          WHERE profile_id = @profile_id AND timestamp >= @timestamp
          %s
          ORDER BY updated_at ` + ctx.Pagination.Order
	limit := ` LIMIT @limit OFFSET @offset`

	filters := ""
	if len(request.Status) > 0 {
		filters += " AND status = ANY(@status)"
	}

	if request.MarketId != "" {
		filters += " AND market_id = @market_id"
	}

	if len(request.OrderType) > 0 {
		filters += " AND order_type = ANY(@order_type)"
	}

	if request.OrderId != "" {
		filters += " AND id = @order_id"
	}

	if request.ClientOrderId != "" {
		filters += " AND client_order_id = @client_order_id"
	}

	if request.EndTime > 0 {
		filters += " AND timestamp <= @end_time"
	}

	q = fmt.Sprintf(q, filters)
	args := pgx.NamedArgs{
		"profile_id":      ctx.Profile.ProfileId,
		"market_id":       request.MarketId,
		"timestamp":       request.TimeStamp,
		"status":          request.Status,
		"order_type":      request.OrderType,
		"order_id":        request.OrderId,
		"client_order_id": request.ClientOrderId,
		"end_time":        request.EndTime,
		"limit":           ctx.Pagination.Limit,
		"offset":          nil,
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

	results := make([]model.OrderData, 0)
	for rows.Next() {
		var r model.OrderData
		err = rows.Scan(
			&r.OrderId,
			&r.ProfileID,
			&r.MarketID,
			&r.OrderType,
			&r.Status,
			&r.Price,
			&r.Size,
			&r.InitialSize,
			&r.TotalFilledSize,
			&r.Side,
			&r.Timestamp,
			&r.Reason,
			&r.ClientOrderId,
			&r.TriggerPrice,
			&r.SizePercent,
			&r.TimeInForce,
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
