package api

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

const TEST_JWT = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIwIiwiZXhwIjo1MjYyNjUyMDEwfQ.x_245iYDEvTTbraw1gt4jmFRFfgMJb-GJ-hsU9HuDik"
const TEST_JWT_11 = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMSIsImV4cCI6NTI2MjY1MjAxMH0.5ZOMK_kMBqO70HRvRmKLBluODTkzFByiOEeEvCG3WUU"

type SequenceRequest struct {
	Sequence uint    `form:"sequence"`
	Bid      float64 `form:"bid"`
	BidSize  float64 `form:"bid_size"`
	Ask      float64 `form:"ask"`
	AskSize  float64 `form:"ask_size"`
}

type TradeCreateRequest struct {
	MarketId string  `form:"market_id"`
	Time     int64   `form:"time"`
	Price    float64 `form:"price"`
	Size     float64 `form:"size"`
}

type WhiteListRequest struct {
	ProfileId uint `form:"profile_id"`
}

type ProfileRequest struct {
	ProfileId uint `form:"profile_id"`
}

type FairPriceRequest struct {
	MarketId  string  `form:"market_id"`
	FairPrice float64 `form:"fair_price"`
}

type DepositInsuranceRequest struct {
	Amount float64 `form:"amount"`
}

type ReplayRequest struct {
	FromBlock string `form:"from_block"`
}

func HandleDebugWhitelist(c *gin.Context) {
	var request WhiteListRequest
	ctx := GetRabbitContext(c)

	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	apiModel := model.NewApiModel(ctx.Broker)
	res := apiModel.WhiteListProfile(c.Request.Context(),
		request.ProfileId,
	)

	SuccessResponse[interface{}](c, res)
}

func HandleDebugTradeCreate(c *gin.Context) {
	var request TradeCreateRequest
	ctx := GetRabbitContext(c)

	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	debugModel := model.NewDebugModel(ctx.Broker)
	_, err := debugModel.CreateCandle(c.Request.Context(),
		request.MarketId,
		request.Time,
		request.Price,
		request.Size,
	)

	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse[int](c)
}

func HandleDebugPushProfile(c *gin.Context) {
	var request ProfileRequest
	ctx := GetRabbitContext(c)

	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	debugModel := model.NewDebugModel(ctx.Broker)
	_, err := debugModel.PushProfile(c.Request.Context(), request.ProfileId)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse[int](c)
}

/*
func HandleDebugOrderList(c *gin.Context) {
	var request ProfileRequest
	ctx := GetRabbitContext(c)

	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	apiModel := model.NewApiModel(ctx.Broker)

	orders, err := apiModel.GetOpenOrders(context.Background(), request.ProfileId)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, orders)
}

func HandleDebugPositionList(c *gin.Context) {
	var request ProfileRequest
	ctx := GetRabbitContext(c)

	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	apiModel := model.NewApiModel(ctx.Broker)

	positions, err := apiModel.GetOpenPositions(context.Background(), request.ProfileId, 0, 100)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, positions)
}*/

func HandleDebugInfo(c *gin.Context) {
	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)
	logrus.Info(apiModel)

	logrus.Info(ctx.Broker.Pool)
}

func HandleDebugSequence(c *gin.Context) {
	var request SequenceRequest
	ctx := GetRabbitContext(c)

	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	debugModel := model.NewDebugModel(ctx.Broker)
	_, err := debugModel.PushOrderbook(c.Request.Context(), model.TEST_MARKET_TICKER, request.Sequence, request.Bid, request.BidSize, request.Ask, request.AskSize)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse[int](c)
}

func HandleDebugOrderCreate(c *gin.Context) {

}

func HandleDebugStats(c *gin.Context) {
	ctx := GetRabbitContext(c)

	apiModel := model.NewApiModel(ctx.Broker)
	response := make([]interface{}, 0)

	res, err := apiModel.GetApiStats(c.Request.Context())
	if err != nil {
		ErrorResponse(c, err)
		return
	}
	response = append(response, res)

	for _, market := range ctx.Config.Service.Markets {
		res, err := apiModel.GetMarketStats(c.Request.Context(), market)
		if err != nil {
			ErrorResponse(c, err)
			return
		}
		response = append(response, res)
	}

	SuccessResponse(c, response...)
}

func HandleSetFairPrice(c *gin.Context) {
	var request FairPriceRequest
	ctx := GetRabbitContext(c)

	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	apiModel := model.NewApiModel(ctx.Broker)
	new_price, err := apiModel.TestUpdateFairPrice(c.Request.Context(),
		request.MarketId,
		request.FairPrice,
	)

	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse[string](c, new_price.String())

}

func HandleDepositInsurance(c *gin.Context) {
	var request DepositInsuranceRequest
	ctx := GetRabbitContext(c)

	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	apiModel := model.NewApiModel(ctx.Broker)
	ops, err := apiModel.DepositCredit(c.Request.Context(), 0, request.Amount)

	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse[*model.BalanceOps](c, ops)
}

func HandleNowUTC(c *gin.Context) {
	SuccessResponse(c, time.Now().UnixMicro())
}

// Settlement debug endpoints, can't be fully removed in prod
// Need to have a way to do below operations

func HandleSettlementState(c *gin.Context) {
	ctx := GetRabbitContext(c)

	apiModel := model.NewApiModel(ctx.Broker)
	res, err := apiModel.GetSettlementState(c.Request.Context())
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, res)
}

func HandleReplayDeposit(c *gin.Context) {
	ErrorResponse(c, fmt.Errorf("REMOVED"))
	return
}

func HandleReplayWithdraw(c *gin.Context) {
	ErrorResponse(c, fmt.Errorf("REMOVED"))
	return
}

func HandleMergeUnknownOps(c *gin.Context) {
	ctx := GetRabbitContext(c)

	apiModel := model.NewApiModel(ctx.Broker)
	total, err := apiModel.MergeUnknown(c.Request.Context())

	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, total)
}

func HandleShowProcessingOps(c *gin.Context) {
	ctx := GetRabbitContext(c)

	apiModel := model.NewApiModel(ctx.Broker)
	res, err := apiModel.GetProcessingOps(c.Request.Context())

	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, res)
}

func HandleDeleteUnknown(c *gin.Context) {
	ctx := GetRabbitContext(c)

	apiModel := model.NewApiModel(ctx.Broker)
	total, err := apiModel.DeleteUnknown(c.Request.Context())

	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, total)
}

func HandleShowMeta(c *gin.Context) {
	ctx := GetRabbitContext(c)

	SuccessResponse(c, ctx.Meta)
}

func HandleSrvTimestamp(c *gin.Context) {
	currentTimestamp := time.Now().Unix()

	SuccessResponse(c, currentTimestamp)
}
