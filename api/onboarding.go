package api

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/strips-finance/rabbit-dex-backend/auth"
	"github.com/strips-finance/rabbit-dex-backend/model"
	"github.com/strips-finance/rabbit-dex-backend/referrals"
)

type OnboardingRequestMetaItem struct {
	UTMSource   string `json:"utm_source,omitempty"`
	UTMMedium   string `json:"utm_medium,omitempty"`
	UTMCampaign string `json:"utm_campaign,omitempty"`
	UTMTerm     string `json:"utm_term,omitempty"`
	UTMContent  string `json:"utm_content,omitempty"`
	TS          string `json:"ts,omitempty"`
	HTTPReferer string `json:"http_referer,omitempty"`
}

type OnboardingRequestMeta struct {
	Campaign          []OnboardingRequestMetaItem `json:"campaign,omitempty"`
	ReferrerShortCode string                      `json:"referrer_short_code,omitempty"`
}

type OnboardingRequest struct {
	IsClient    bool                   `json:"is_client,omitempty"`
	Wallet      string                 `json:"wallet,omitempty" binding:"len=42,required"`
	Signature   string                 `json:"signature,omitempty" binding:"len=132,required"`
	ProfileType string                 `json:"profile_type" binding:"omitempty,oneof=vault trader"`
	Meta        *OnboardingRequestMeta `json:"meta,omitempty"`
}

func HandleOnboarding(c *gin.Context) {
	var (
		request  OnboardingRequest
		response interface{}
		jwt      string
	)

	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	wallet := model.GetWalletStringInRabbitTntStandardFormat(request.Wallet)

	if request.ProfileType == "" {
		request.ProfileType = model.PROFILE_TYPE_TRADER
	}

	ctx := GetRabbitContext(c)
	broker, err := model.GetBroker()
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	cfg := ctx.Config

	verifyRequest := &auth.MetamaskVerifyRequest{
		Wallet:        wallet,
		Timestamp:     ctx.Timestamp,
		Signature:     request.Signature,
		ProfileType:   request.ProfileType,
		EIP712Encoder: ctx.EIP712Encoder,
	}

	logrus.
		WithField("Wallet", wallet).
		WithField("Timestamp", ctx.Timestamp).
		WithField("Signature", request.Signature).
		WithField("ProfileType", request.ProfileType).
		Info("received verifyRequest")

	var invitedId uint64
	var newProfile bool
	if request.IsClient {
		oldJwt, _ := c.Cookie("jwt")
		onboardFrontendResult, err := auth.OnboardFrontend(
			broker,
			oldJwt,
			ctx.Config.Service.HMACSecret,
			ctx.Config.Service.JwtLifetime,
			ctx.Config.Service.RefreshTokenLifetime,
			verifyRequest,
			cfg.Service.EnvMode,
			ctx.ExchangeId,
			ctx.ExchangeCfg.OnboardingMessages,
		)
		if err != nil {
			ErrorResponse(c, err)
			return
		}

		response = onboardFrontendResult
		jwt = onboardFrontendResult.Jwt
		invitedId = uint64(onboardFrontendResult.Profile.ProfileID)
		newProfile = onboardFrontendResult.NewProfile
	} else {
		onboardMarketMakerResult, err := auth.OnboardMarketMaker(
			broker,
			ctx.Config.Service.HMACSecret,
			ctx.Config.Service.JwtLifetime,
			verifyRequest,
			cfg.Service.EnvMode,
			ctx.ExchangeId,
			ctx.ExchangeCfg.OnboardingMessages,
		)
		if err != nil {
			ErrorResponse(c, err)
			return
		}

		response = onboardMarketMakerResult
		jwt = onboardMarketMakerResult.Jwt
		invitedId = uint64(onboardMarketMakerResult.Profile.ProfileID)
		newProfile = onboardMarketMakerResult.NewProfile
	}

	// check for referrals.
	if request.Meta != nil {
		if request.Meta.ReferrerShortCode != "" && newProfile {
			go referrals.CreateReferralLink(ctx.TimeScaleDB, request.Meta.ReferrerShortCode, invitedId)
		}
	}

	c.SetCookie("jwt", jwt, int(ctx.Config.Service.JwtLifetime), "/", "", true, true)
	SuccessResponse(c, response)
}
