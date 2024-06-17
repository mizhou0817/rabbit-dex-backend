package api

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/strips-finance/rabbit-dex-backend/auth"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

type SecretListResponse struct {
	Secrets []*model.Secret `json:"secrets"`
}

type SecretCreateRequest struct {
	Tag           string   `json:"tag" binding:"required"`
	Expiration    int64    `json:"expiration" binding:"required"`
	AllowedIpList []string `json:"allowed_ip_list"`
}

type SecretDeleteRequest struct {
	Key string `json:"key" binding:"required"`
}

type SecretRefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func HandleSecretsSessionRemove(c *gin.Context) {
	ctx := GetRabbitContext(c)
	apiSecretModel := model.NewApiSecretModel(ctx.Broker)

	err := apiSecretModel.RemoveSessionSecrets(c.Request.Context(), ctx.Profile.ProfileId)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, true)
}

func HandleSecretCreate(c *gin.Context) {
	var request SecretCreateRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	expiresAt := time.Now().Add(time.Second * time.Duration(request.Expiration)).Unix()
	if expiresAt < time.Now().Unix() {
		ErrorResponse(c, errors.New("EXPIRATION_IN_THE_PAST"))
		return
	}

	//CHECKPROFILE
	ctx := GetRabbitContext(c)
	apiSecretModel := model.NewApiSecretModel(ctx.Broker)

	jwt, err := apiSecretModel.GenerateJwt(ctx.Profile.ProfileId, ctx.Config.Service.HMACSecret, expiresAt)
	if err != nil {
		logrus.Error(err)
		ErrorResponse(c, errors.New("JWT_GENERATION_ERROR"))
		return
	}

	refreshToken, err := apiSecretModel.GenerateJwt(ctx.Profile.ProfileId, ctx.Config.Service.HMACSecret, expiresAt*2)
	if err != nil {
		logrus.Error(err)
		ErrorResponse(c, errors.New("REFRESH_TOKEN_ERROR"))
		return
	}

	res, err := apiSecretModel.CreatePairFromProfileID(
		c.Request.Context(),
		request.Tag,
		ctx.Profile.ProfileId,
		&expiresAt,
		jwt,
		refreshToken,
		request.AllowedIpList)

	if err != nil {
		logrus.Error(err)
		ErrorResponse(c, errors.New("PAIR_GENERATION_ERROR"))
		return
	}

	eCfg := ctx.Config.Service.Exchanges[ctx.ExchangeId]

	SuccessResponse(c, model.Secret{
		APISecret:     res,
		JwtPrivate:    jwt,
		JwtPublic:     eCfg.JwtPublic,
		RefreshToken:  refreshToken,
		AllowedIpList: request.AllowedIpList,
		CreatedAt:     time.Now().UnixMicro(),
	})

}

func HandleSecretDelete(c *gin.Context) {
	var request SecretDeleteRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	//CHECKPROFILE
	ctx := GetRabbitContext(c)
	apiSecretModel := model.NewApiSecretModel(ctx.Broker)

	err := apiSecretModel.DeletePairForProfile(c.Request.Context(), ctx.Profile.ProfileId, request.Key)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, true)

}

func HandleListSecrets(c *gin.Context) {
	ctx := GetRabbitContext(c)
	apiSecretModel := model.NewApiSecretModel(ctx.Broker)

	secrets, err := apiSecretModel.GetSecrets(c.Request.Context(), ctx.Profile.ProfileId)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, secrets...)
}

func HandleSecretRefresh(c *gin.Context) {
	var request SecretRefreshRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)
	apiSecretModel := model.NewApiSecretModel(ctx.Broker)

	// Check refresh token
	profileID, err := auth.JwtToProfileID(request.RefreshToken, ctx.Config.Service.HMACSecret)
	if err != nil {
		ErrorResponse(c, err)
		c.Abort()
		return
	}

	if profileID != ctx.Profile.ProfileId {
		ErrorResponse(c, errors.New("NOT_YOUR_REFRESH_TOKEN"))
		c.Abort()
		return
	}

	expiresAt := time.Now().Add(time.Second * time.Duration(30*24*60*60)).Unix()

	newJwt, err := apiSecretModel.GenerateJwt(ctx.Profile.ProfileId, ctx.Config.Service.HMACSecret, expiresAt)
	if err != nil {
		logrus.Error(err)
		ErrorResponse(c, errors.New("JWT_GENERATION_ERROR"))
		return
	}

	newRefreshToken, err := apiSecretModel.GenerateJwt(ctx.Profile.ProfileId, ctx.Config.Service.HMACSecret, expiresAt*2)
	if err != nil {
		logrus.Error(err)
		ErrorResponse(c, errors.New("REFRESH_TOKEN_ERROR"))
		return
	}

	new_secret, err := apiSecretModel.RefreshSecret(c.Request.Context(),
		ctx.Profile.ProfileId,
		request.RefreshToken,
		newJwt,
		newRefreshToken,
		expiresAt,
	)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, new_secret)

}
