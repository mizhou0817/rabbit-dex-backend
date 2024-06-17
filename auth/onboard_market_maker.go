package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/strips-finance/rabbit-dex-backend/model"
)

type OnboardMarketMakerResult struct {
	Profile    *model.ProfileData `json:"profile"`
	APISecret  *model.APISecret   `json:"apiSecret"`
	Jwt        string             `json:"jwt"`
	NewProfile bool               `json:"-"`
}

func OnboardMarketMaker(
	broker *model.Broker,
	hmacSecret string,
	jwtLifetime uint64,
	verifyRequest *MetamaskVerifyRequest,
	envMode string,
	exchangeId string,
	messages []string,
) (*OnboardMarketMakerResult, error) {
	if err := VerifyProfile(EIP_191, verifyRequest, TRADER_ROLE, messages); err != nil {
		return nil, err
	}

	apiModel := model.NewApiModel(broker)
	apiSecretModel := model.NewApiSecretModel(broker)

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

	apiSecrets, err := apiSecretModel.GetOrRefreshSecretByProfileID(context.Background(), profile.ProfileId)
	if err != nil {
		return nil, err
	}

	if len(apiSecrets) == 0 {
		return nil, fmt.Errorf("SECRET_GENERATE_ERROR")
	}

	expiresAt := time.Now().Add(time.Second * time.Duration(jwtLifetime)).Unix()

	jwt, err := apiSecretModel.GenerateJwt(profile.ProfileId, hmacSecret, expiresAt)
	if err != nil {
		return nil, err
	}

	profileData, err := apiModel.GetProfileData(context.Background(), profile.ProfileId)
	if err != nil {
		return nil, err
	}

	return &OnboardMarketMakerResult{
		Profile:    profileData,
		APISecret:  apiSecrets[0],
		Jwt:        jwt,
		NewProfile: newProfile,
	}, nil
}
