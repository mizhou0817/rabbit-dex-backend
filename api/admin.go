package api

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

type ProfileCacheRequest struct {
	ProfileId uint `form:"profile_id"`
}

type BalanceOpsRequest struct {
	ProfileId uint `form:"profile_id"`
	Offset    uint `form:"offset"`
	Limit     uint `form:"limit"`
}

func HandleExchangeTotalRequest(c *gin.Context) {
	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	res, err := apiModel.GetExchangeData(c.Request.Context())
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, res)

}

func HandleBalanceOpsRequest(c *gin.Context) {
	var request BalanceOpsRequest
	ctx := GetRabbitContext(c)

	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	apiModel := model.NewApiModel(ctx.Broker)

	ops, err := apiModel.BalanceOpsList(c.Request.Context(), request.ProfileId, request.Offset, request.Limit)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, ops)
}

func HandleProfileCacheRequest(c *gin.Context) {
	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)
	cache, err := apiModel.InvalidateCache(c.Request.Context(), ctx.Profile.ProfileId)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, cache)
}

func HandleInv3Request(c *gin.Context) {
	ctx := GetRabbitContext(c)

	apiModel := model.NewApiModel(ctx.Broker)

	valid, err := apiModel.IsInv3Valid(c.Request.Context(), 0)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, valid)
}

type ChangeIconUrlRequest struct {
	MarketId string `json:"market_id" binding:"required"`
	Url      string `json:"url" binding:"required"`
}

func HandleChangeIconUrl(c *gin.Context) {
	var request ChangeIconUrlRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	//CHECKPROFILE
	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	res, err := apiModel.MarketUpdateIconUrl(c.Request.Context(),
		request.MarketId,
		request.Url)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, res)
}

type ChangeMarketTitle struct {
	MarketId string `json:"market_id" binding:"required"`
	Title    string `json:"title" binding:"required"`
}

func HandleChangeMarketTitle(c *gin.Context) {
	var request ChangeMarketTitle

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	//CHECKPROFILE
	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	res, err := apiModel.MarketUpdateTitle(c.Request.Context(),
		request.MarketId,
		request.Title)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, res)
}

/*
type Tier struct {
	Tier      uint             `msgpack:"tier" json:"tier"`
	Title     string           `msgpack:"title" json:"title"`
	MakerFee  tdecimal.Decimal `msgpack:"maker_fee" json:"maker_fee"`
	TakerFee  tdecimal.Decimal `msgpack:"taker_fee" json:"taker_fee"`
	MinVolume tdecimal.Decimal `msgpack:"min_volume" json:"min_volume"`
	MinAssets tdecimal.Decimal `msgpack:"min_assets" json:"min_assets"`
}

type SpecialTier struct {
	Tier     uint             `msgpack:"tier" json:"tier"`
	Title    string           `msgpack:"title" json:"title"`
	MakerFee tdecimal.Decimal `msgpack:"maker_fee" json:"maker_fee"`
	TakerFee tdecimal.Decimal `msgpack:"taker_fee" json:"taker_fee"`
}

type ProfileTier struct {
	ProfileID uint `msgpack:"profile_id" json:"profile_id"`
	TierID    uint `msgpack:"tier_id" json:"tier_id"`
}

superAdminAuthRequired.POST("/tiers", HandleAddTier)
superAdminAuthRequired.POST("/tiers/special", HandleAddSpecialTier)
superAdminAuthRequired.POST("/tiers/profile", HandleAddProfileTier)
superAdminAuthRequired.DELETE("/tiers/profile", HandleRemoveProfileTier)

*/

// Tiers management
type AddTierRequest struct {
	MarketId  string  `json:"market_id" binding:"required"`
	TierId    uint    `json:"tier_id" binding:"required"`
	Title     string  `json:"title" binding:"required"`
	MakerFee  float64 `json:"maker_fee"`
	TakerFee  float64 `json:"taker_fee"`
	MinVolume float64 `json:"min_volume"`
	MinAssets float64 `json:"min_assets"`
}

type AddSpecialTierRequest struct {
	MarketId string  `json:"market_id" binding:"required"`
	TierId   uint    `json:"tier_id" binding:"required"`
	Title    string  `json:"title" binding:"required"`
	MakerFee float64 `json:"maker_fee"`
	TakerFee float64 `json:"taker_fee"`
}

type AddProfileToTierRequest struct {
	MarketId  string `json:"market_id" binding:"required"`
	ProfileId uint   `json:"profile_id" binding:"required"`
	TierId    uint   `json:"tier_id" binding:"required"`
}

type RemoveAnyTierRequest struct {
	MarketId string `json:"market_id" binding:"required"`
	TierId   uint   `json:"tier_id" binding:"required"`
}

type RemoveProfileTierRequest struct {
	MarketId  string `json:"market_id" binding:"required"`
	ProfileId uint   `json:"profile_id" binding:"required"`
}

type GetAllTiersRequest struct {
	MarketId string `form:"market_id" binding:"required"`
}

type WhichTierRequest struct {
	ProfileId uint `form:"profile_id"`
}

func HandleGetTiers(c *gin.Context) {
	var request GetAllTiersRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	//CHECKPROFILE
	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	res, err := apiModel.GetTiers(c.Request.Context(),
		request.MarketId)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, res)
}

func HandleGetSpecialTiers(c *gin.Context) {
	var request GetAllTiersRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	//CHECKPROFILE
	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	res, err := apiModel.GetSpecialTiers(c.Request.Context(),
		request.MarketId)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, res)
}

func HandleGetProfileTiers(c *gin.Context) {
	var request GetAllTiersRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	//CHECKPROFILE
	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	res, err := apiModel.GetProfileTiers(c.Request.Context(),
		request.MarketId)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, res)
}

func HandleRemoveTier(c *gin.Context) {
	var request RemoveAnyTierRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	if request.TierId == 0 {
		ErrorResponse(c, errors.New("ZERO_TIER_DENIED"))
		return
	}

	//CHECKPROFILE
	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	res, err := apiModel.RemoveTier(c.Request.Context(),
		request.MarketId,
		request.TierId)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, res)
}

func HandleRemoveSpecialTier(c *gin.Context) {
	var request RemoveAnyTierRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	if request.TierId == 0 {
		ErrorResponse(c, errors.New("ZERO_TIER_DENIED"))
		return
	}

	//CHECKPROFILE
	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	res, err := apiModel.RemoveSpecialTier(c.Request.Context(),
		request.MarketId,
		request.TierId)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, res)
}

func HandleRemoveProfileTier(c *gin.Context) {
	var request RemoveProfileTierRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	//CHECKPROFILE
	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	res, err := apiModel.RemoveProfileTier(c.Request.Context(),
		request.MarketId,
		request.ProfileId)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, res)
}

func HandleAddTier(c *gin.Context) {
	var request AddTierRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	//CHECKPROFILE
	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	res, err := apiModel.AddTier(c.Request.Context(),
		request.MarketId,
		request.TierId,
		request.Title,
		request.MakerFee,
		request.TakerFee,
		request.MinVolume,
		request.MinAssets)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, res)
}

func HandleAddSpecialTier(c *gin.Context) {
	var request AddSpecialTierRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	//CHECKPROFILE
	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	res, err := apiModel.AddSpecialTier(c.Request.Context(),
		request.MarketId,
		request.TierId,
		request.Title,
		request.MakerFee,
		request.TakerFee)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, res)
}

func HandleAddProfileTier(c *gin.Context) {
	var request AddProfileToTierRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	//CHECKPROFILE
	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	res, err := apiModel.AddProfileToTier(c.Request.Context(),
		request.MarketId,
		request.ProfileId,
		request.TierId)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, res)
}

func HandleEditTier(c *gin.Context) {
	var request AddTierRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	//CHECKPROFILE
	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	res, err := apiModel.EditTier(c.Request.Context(),
		request.MarketId,
		request.TierId,
		request.Title,
		request.MakerFee,
		request.TakerFee,
		request.MinVolume,
		request.MinAssets)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, res)
}

func HandleEditSpecialTier(c *gin.Context) {
	var request AddSpecialTierRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	//CHECKPROFILE
	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	res, err := apiModel.EditSpecialTier(c.Request.Context(),
		request.MarketId,
		request.TierId,
		request.Title,
		request.MakerFee,
		request.TakerFee)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, res)
}

func HandleWhichTier(c *gin.Context) {
	var request WhichTierRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	//CHECKPROFILE
	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	response := make([]model.TierData, 0)

	for _, market := range ctx.Config.Service.Markets {
		tier, err := apiModel.WhichTier(c.Request.Context(), market, request.ProfileId)
		if err != nil {
			ErrorResponse(c, err)
			return
		}

		response = append(response, model.TierData{
			MarketId: market,
			TierData: tier,
		})
	}

	SuccessResponse(c, response...)
}
