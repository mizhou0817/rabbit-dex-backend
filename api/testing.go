package api

import (
	"crypto/ecdsa"
	"errors"
	"math"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	"github.com/strips-finance/rabbit-dex-backend/helpers"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

// Requests
type (
	CreateOrdersRequest struct {
		ProfileId uint                 `json:"profile_id" binding:"required"`
		Orders    []OrderCreateRequest `json:"orders"`
	}

	UpdateProfileRequest struct {
		ProfileId        uint   `json:"profile_id" binding:"required"`
		Credit           int    `json:"balance"`
		Leverage         int    `json:"leverage"`
		LeverageMarketId string `json:"leverage_market_id"`
	}

	UpdateMarketRequest struct {
		MarketId   string  `json:"market_id" binding:"required"`
		FairPrice  float64 `json:"fair_price" binding:"gte=0"`
		IndexPrice float64 `json:"index_price" binding:"gte=0"`
	}
)

// Responses
type (
	CreateAllResponse struct {
		Profile   *model.ProfileData `json:"profile"`
		SecretKey string             `json:"secret_key"`
		Orders    []*model.OrderData `json:"orders"`
	}

	CreateProfileResponse struct {
		Profile   *model.ProfileData `json:"profile"`
		SecretKey string             `json:"secret_key"`
	}

	UpdateProfileResponse struct {
		Profile    *model.ProfileData  `json:"profile"`
		BalanceOps []*model.BalanceOps `json:"balance_ops"`
	}
)

func HandleClear(c *gin.Context) {
	ctx := GetRabbitContext(c)
	instance := helpers.NewInstance(ctx.Broker)

	if err := instance.Reset([]string{}, ctx.Config.Service.Markets); err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, true)
}

func HandleCreateAll(c *gin.Context) {
	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	// Create profile
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	publicKey, ok := privateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		err = errors.New("cannot derive public key from private key")

		ErrorResponse(c, err)
		return
	}

	wallet := crypto.PubkeyToAddress(*publicKey).String()

	newProfile, err := apiModel.CreateProfile(c.Request.Context(), model.PROFILE_TYPE_TRADER, wallet, model.EXCHANGE_DEFAULT)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	// Add deposit
	_, err = apiModel.DepositCredit(c.Request.Context(), newProfile.ProfileId, 10000000000)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	// Create some orders for profile
	orders := make([]*model.OrderData, 0)

	for _, orderMarket := range ctx.Config.Service.Markets {
		for i := 0; i < 5; i += 1 {
			orderSide := model.LONG

			if i%2 == 1 {
				orderSide = model.SHORT
			}

			someId := "someid1123"
			somePrice := 1000.0
			someSize := 1.0
			newOrder, err := apiModel.OrderCreate(
				c.Request.Context(),
				newProfile.ProfileId,
				orderMarket,
				"market",
				orderSide,
				&somePrice,
				&someSize,
				&someId,
				nil,
				nil,
				nil,

				nil,
			)
			if err != nil {
				// TODO:  Invalid MsgPack - packet body
				ErrorResponse(c, err)
				return
			}

			time.Sleep(time.Millisecond * 100)

			newOrderData, err := apiModel.GetOrderById(c.Request.Context(), orderMarket, newOrder.OrderId)
			if err != nil {
				ErrorResponse(c, err)
				return
			}

			orders = append(orders, newOrderData)
		}
	}

	newProfileData, err := apiModel.GetProfileData(c.Request.Context(), newProfile.ProfileId)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, CreateAllResponse{
		Profile:   newProfileData,
		SecretKey: hexutil.Encode(crypto.FromECDSA(privateKey)),
		Orders:    orders,
	})
}

func HandleCreateProfile(c *gin.Context) {
	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	// Generate PrivateKey & PublicKey for profile
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	publicKey, ok := privateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		err = errors.New("cannot derive public key from private key")

		ErrorResponse(c, err)
		return
	}

	wallet := crypto.PubkeyToAddress(*publicKey).String()

	newProfile, err := apiModel.CreateProfile(c.Request.Context(), model.PROFILE_TYPE_TRADER, wallet, model.EXCHANGE_DEFAULT)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	newProfileData, err := apiModel.GetProfileData(c.Request.Context(), newProfile.ProfileId)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, CreateProfileResponse{
		Profile:   newProfileData,
		SecretKey: hexutil.Encode(crypto.FromECDSA(privateKey)),
	})
}

func HandleCreateOrders(c *gin.Context) {
	var request CreateOrdersRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)
	newOrders := make([]*model.OrderData, 0)

	for _, o := range request.Orders {
		newOrder, err := apiModel.OrderCreate(
			c.Request.Context(),
			request.ProfileId,
			o.MarketId,
			o.Type,
			o.Side,
			o.Price,
			o.Size,
			o.ClientOrderId,
			nil,
			nil,
			nil,

			nil,
		)
		if err != nil {
			ErrorResponse(c, err)
			return
		}

		time.Sleep(time.Millisecond * 100)

		newOrderData, err := apiModel.GetOrderById(c.Request.Context(), newOrder.MarketId, newOrder.OrderId)
		if err != nil {
			ErrorResponse(c, err)
			return
		}

		newOrders = append(newOrders, newOrderData)
	}

	SuccessResponse(c, newOrders...)
}

func HandleUpdateProfile(c *gin.Context) {
	var request UpdateProfileRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)
	balanceOps := make([]*model.BalanceOps, 0)

	if request.Credit > 0 {
		balanceOp, err := apiModel.DepositCredit(c.Request.Context(), request.ProfileId, math.Abs(float64(request.Credit)))
		if err != nil {
			ErrorResponse(c, err)
			return
		}

		balanceOps = append(balanceOps, balanceOp)
	}

	if request.Credit < 0 {
		balanceOp, err := apiModel.WithdrawCredit(c.Request.Context(), request.ProfileId, math.Abs(float64(request.Credit)))
		if err != nil {
			ErrorResponse(c, err)
			return
		}

		balanceOps = append(balanceOps, balanceOp)
	}

	if request.Leverage > 0 {
		_, err := apiModel.UpdateLeverage(c.Request.Context(), request.LeverageMarketId, request.ProfileId, uint(request.Leverage))

		if err != nil {
			ErrorResponse(c, err)
			return
		}
	}

	profileData, err := apiModel.GetProfileData(c.Request.Context(), request.ProfileId)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, UpdateProfileResponse{
		Profile:    profileData, // contains leverage and balance
		BalanceOps: balanceOps,
	})
}

func HandleUpdateMarket(c *gin.Context) {
	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	var request UpdateMarketRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	if (request.IndexPrice + request.FairPrice) <= 0 {
		err := errors.New("IndexPrice or FairPrice should be set")
		ErrorResponse(c, err)
		return
	}

	if request.FairPrice > 0 {
		_, _ = apiModel.TestUpdateFairPrice(c.Request.Context(), request.MarketId, request.FairPrice)
	}

	if request.IndexPrice > 0 {
		_ = apiModel.UpdateIndexPrice(c.Request.Context(), request.MarketId, request.IndexPrice)
	}

	marketData, err := apiModel.GetMarketData(c.Request.Context(), request.MarketId)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, marketData)
}
