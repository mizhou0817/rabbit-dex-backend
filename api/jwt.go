package api

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/strips-finance/rabbit-dex-backend/auth"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

type JwtRequest struct {
	IsClient     bool   `json:"is_client"`
	RefreshToken string `json:"refresh_token"`
}

type JwtFrontendResponse struct {
	Jwt          string `json:"jwt"`
	RandomSecret string `json:"random_secret"`
	RefreshToken string `json:"refresh_token"`
}

type JwtMarketMakerResponse struct {
	Jwt string `json:"jwt"`
}

func HandleJwt(c *gin.Context) {
	var request JwtRequest

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

	if request.IsClient {
		profileID, err := auth.JwtToProfileID(request.RefreshToken, ctx.Config.Service.HMACSecret)
		if err != nil {
			ErrorResponse(c, err)
			return
		}

		frontendSecretModel := model.NewFrontendSecretModel(broker)

		// Fetch actual frontendSecret from Tarantool
		frontendSecret, err := frontendSecretModel.GetByRefreshToken(c.Request.Context(), request.RefreshToken)
		if err != nil {
			ErrorResponse(c, err)
			return
		}

		if err = ctx.Payload.Verify(ctx.Signature, frontendSecret.RandomSecret, ctx.Config.Service.EnvMode); err != nil {
			ErrorResponse(c, err)
			return
		}

		// Replace old frontendSecret with it one-time RefreshToken by jwt from current one
		frontendSecret, err = frontendSecretModel.CreateFromProfileID(
			c.Request.Context(),
			profileID,
			frontendSecret.Jwt,
			ctx.Config.Service.HMACSecret,
			ctx.Config.Service.JwtLifetime,
			ctx.Config.Service.RefreshTokenLifetime,
		)
		if err != nil {
			ErrorResponse(c, err)
			return
		}

		c.SetCookie("jwt", frontendSecret.Jwt, int(ctx.Config.Service.JwtLifetime), "/", "", true, true)

		SuccessResponse(c, JwtFrontendResponse{
			Jwt:          frontendSecret.Jwt,
			RandomSecret: frontendSecret.RandomSecret,
			RefreshToken: frontendSecret.RefreshToken,
		})
	} else {
		if ctx.MarketMakerAPIKey == "" {
			ErrorResponse(c, errors.New("API-KEY header required"))
			return
		}

		apiSecretModel := model.NewApiSecretModel(broker)

		apiSecret, err := apiSecretModel.GetByKey(c.Request.Context(), ctx.MarketMakerAPIKey)
		if err != nil {
			ErrorResponse(c, err)
			return
		}

		if err = ctx.Payload.Verify(ctx.Signature, apiSecret.Secret, ctx.Config.Service.EnvMode); err != nil {
			ErrorResponse(c, err)
			return
		}

		expiresAt := time.Now().Add(time.Second * time.Duration(ctx.Config.Service.JwtLifetime)).Unix()

		jwt, err := apiSecretModel.GenerateJwt(
			apiSecret.ProfileID,
			ctx.Config.Service.HMACSecret,
			expiresAt,
		)
		if err != nil {
			ErrorResponse(c, err)
			return
		}

		err = apiSecretModel.UpdateApiSecretExpire(c.Request.Context(), ctx.MarketMakerAPIKey, expiresAt)
		if err != nil {
			ErrorResponse(c, err)
			return
		}

		SuccessResponse(c, JwtMarketMakerResponse{
			Jwt: jwt,
		})
	}
}
