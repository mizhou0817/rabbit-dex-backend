package auth

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

type OnboardFrontendResult struct {
	Profile      *model.ProfileData `json:"profile"`
	Jwt          string             `json:"jwt"`
	RefreshToken string             `json:"refreshToken"`
	RandomSecret string             `json:"randomSecret"`
	NewProfile   bool               `json:"-"`
}

func OnboardFrontend(
	broker *model.Broker,
	oldJwt string,
	hmacSecret string,
	jwtLifetime uint64,
	refreshTokenLifetime uint64,
	verifyRequest *MetamaskVerifyRequest,
	envMode string,
	exchangeId string,
	messages []string,
) (*OnboardFrontendResult, error) {
	if err := VerifyProfile(EIP_712, verifyRequest, TRADER_ROLE, messages); err != nil {
		return nil, err
	}

	apiModel := model.NewApiModel(broker)

	profile, err := apiModel.GetProfileByWalletForExchangeId(context.Background(), verifyRequest.Wallet, exchangeId)
	newProfile := false
	if err != nil {
		if err.Error() == model.PROFILE_NOT_FOUND_ERROR {
			profile, err = apiModel.CreateProfile(context.Background(), verifyRequest.ProfileType, verifyRequest.Wallet, exchangeId)
			if err != nil {
				return nil, err
			}
			newProfile = true
		} else {
			return nil, err
		}
	}

	frontendSecretModel := model.NewFrontendSecretModel(broker)

	frontendSecret, err := frontendSecretModel.CreateFromProfileID(
		context.Background(),
		profile.ProfileId,
		oldJwt,
		hmacSecret,
		jwtLifetime,
		refreshTokenLifetime,
	)
	if err != nil {
		return nil, err
	}

	profileData, err := apiModel.GetProfileData(context.Background(), profile.ProfileId)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	return &OnboardFrontendResult{
		Profile:      profileData,
		Jwt:          frontendSecret.Jwt,
		RefreshToken: frontendSecret.RefreshToken,
		RandomSecret: frontendSecret.RandomSecret,
		NewProfile:   newProfile,
	}, nil
}
