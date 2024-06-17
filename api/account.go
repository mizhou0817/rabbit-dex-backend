package api

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

type AccountSetLeverageRequest struct {
	MarketId string `json:"market_id" binding:"required"`
	Leverage uint   `json:"leverage" binding:"oneof=1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20,required"`
}

type AccountValidateRequest struct {
	Jwt string `form:"jwt"`
}

func HandleAccount(c *gin.Context) {
	ctx := GetRabbitContext(c)
	broker, err := model.GetBroker()
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	apiModel := model.NewApiModel(broker)

	profileData, err := apiModel.GetProfileData(c.Request.Context(), ctx.Profile.ProfileId)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, profileData)
}

func HandleAccountValidate(c *gin.Context) {
	var request AccountValidateRequest
	ctx := GetRabbitContext(c)
	jwt_ := c.GetHeader("RBT-JWT")

	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	if j := request.Jwt; j != "" {
		jwt_ = j
	}

	verifiedJwt, err := jwt.Parse(jwt_, func(_ *jwt.Token) (interface{}, error) {
		hmacSecret := []byte(ctx.Config.Service.HMACSecret)

		return hmacSecret, nil
	})
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	claimsMap, ok := verifiedJwt.Claims.(jwt.MapClaims)
	if !ok {
		ErrorResponse(c, jwt.ErrTokenInvalidClaims)
		return
	}

	claimsData, err := json.Marshal(claimsMap)
	if err != nil {
		ErrorResponse(c, jwt.ErrTokenInvalidClaims)
		return
	}

	var claims jwt.RegisteredClaims

	if err = json.Unmarshal(claimsData, &claims); err != nil {
		ErrorResponse(c, jwt.ErrTokenInvalidClaims)
		return
	}

	profileID, err := strconv.ParseUint(claims.Subject, 10, 32)
	if err != nil {
		ErrorResponse(c, jwt.ErrTokenInvalidClaims)
		return
	}

	apiModel := model.NewApiModel(ctx.Broker)

	profile, err := apiModel.GetProfileById(c.Request.Context(), uint(profileID))
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	profileData, err := apiModel.GetProfileData(c.Request.Context(), profile.ProfileId)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, profileData)
}

func HandleAccountSetLeverage(c *gin.Context) {
	var request AccountSetLeverageRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)
	broker, err := model.GetBroker()
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	apiModel := model.NewApiModel(broker)
	_, err = apiModel.UpdateLeverage(
		c.Request.Context(),
		request.MarketId,
		ctx.Profile.ProfileId,
		request.Leverage,
	)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	profile, err := apiModel.InvalidateCacheAndNotify(c.Request.Context(), ctx.Profile.ProfileId)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, profile)
}
