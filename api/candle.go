package api

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

type CandleListRequest struct {
	MarketId      string `form:"market_id" binding:"required"`
	TimestampFrom int64  `form:"timestamp_from" binding:"gt=0,required"`
	TimestampTo   int64  `form:"timestamp_to" binding:"gt=0,required"`
	Period        uint   `form:"period" binding:"oneof=1 5 15 30 60 240 1440,required"`
}

func HandleCandleList(c *gin.Context) {
	var request CandleListRequest

	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	if request.TimestampFrom > request.TimestampTo {
		err := errors.New("timestampTo should be greater than timestampFrom")
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)
	candles, err := apiModel.GetCandles(c.Request.Context(), request.MarketId, request.Period, request.TimestampFrom, request.TimestampTo)

	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, candles...)
}
