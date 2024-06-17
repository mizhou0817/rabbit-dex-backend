package api_client

import (
	"encoding/json"
	"math/big"

	"github.com/sirupsen/logrus"
	"github.com/strips-finance/rabbit-dex-backend/api"
	"github.com/strips-finance/rabbit-dex-backend/auth"
	"github.com/strips-finance/rabbit-dex-backend/signer"
)

func (c *Client) Onboarding() (*Response[auth.OnboardMarketMakerResult], error) {
	signReq := &auth.MetamaskSignRequest{
		Message:   ONBOARDING_MESSAGE,
		Timestamp: c.getExpirationTimestamp(),
	}
	privateKey, err := c.Credentials.GetPrivateKey()
	if err != nil {
		return nil, err
	}

	signature, err := auth.EthSign(auth.EIP_191, signReq, privateKey, nil)
	if err != nil {
		return nil, err
	}

	wallet, err := c.Credentials.GetWallet()
	if err != nil {
		return nil, err
	}

	respBody, err := c.post(PATH_ONBOARDING, api.OnboardingRequest{
		IsClient:  false,
		Wallet:    wallet,
		Signature: signature,
	}, nil)
	if err != nil {
		return nil, err
	}

	var resp Response[auth.OnboardMarketMakerResult]

	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	//TODO: fix this test
	if len(resp.Result) <= 0 {
		panic("NEVER happened for resp.Result")
	}

	result := resp.Result[0]
	c.Credentials.Wallet = *result.Profile.Wallet
	c.Credentials.APIKey = result.APISecret.Key
	c.Credentials.APISecret = result.APISecret.Secret
	c.Credentials.Jwt = result.Jwt
	c.Credentials.ProfileID = result.Profile.ProfileID

	return &resp, nil
}

/*
Used only for internal testing of API, so it doesn't exported.
*/
func (c *Client) onboardingFrontend() (*Response[auth.OnboardFrontendResult], error) {
	signReq := &auth.MetamaskSignRequest{
		Message:   ONBOARDING_MESSAGE,
		Timestamp: c.getExpirationTimestamp(),
	}
	privateKey, err := c.Credentials.GetPrivateKey()
	if err != nil {
		return nil, err
	}

	encoder := signer.NewEIP712Encoder(
		"RabbitXId",
		"1",
		"",
		big.NewInt(int64(1)),
	)
	signature, err := auth.EthSign(auth.EIP_712, signReq, privateKey, encoder)
	logrus.Warnf("EIP712 Signature: %s", signature)

	// signature, err := auth.EthSign(auth.EIP_191, signReq, privateKey, nil)
	if err != nil {
		return nil, err
	}

	wallet, err := c.Credentials.GetWallet()
	if err != nil {
		return nil, err
	}

	respBody, err := c.post(PATH_ONBOARDING, api.OnboardingRequest{
		IsClient:  true,
		Wallet:    wallet,
		Signature: signature,
	}, nil)
	if err != nil {
		return nil, err
	}

	var resp Response[auth.OnboardFrontendResult]

	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	result := resp.Result[0]
	c.Credentials.Wallet = *result.Profile.Wallet
	c.Credentials.Jwt = result.Jwt
	c.Credentials.ProfileID = result.Profile.ProfileID
	c.Credentials.APISecret = result.RandomSecret

	return &resp, nil
}

func (c *Client) onboardingVault() (*Response[auth.OnboardFrontendResult], error) {
	signReq := &auth.MetamaskSignRequest{
		Message:   ONBOARDING_MESSAGE,
		Timestamp: c.getExpirationTimestamp(),
	}
	privateKey, err := c.Credentials.GetPrivateKey()
	if err != nil {
		return nil, err
	}

	// encoder := signer.NewEIP712Encoder(
	// 	"RabbitXId",
	// 	"1",
	// 	"",
	// 	big.NewInt(int64(31337)),
	// )
	// signature, err := auth.EthSign(auth.EIP_712, signReq, privateKey, encoder)
	signature, err := auth.EthSign(auth.EIP_191, signReq, privateKey, nil)
	if err != nil {
		return nil, err
	}

	wallet, err := c.Credentials.GetWallet()
	if err != nil {
		return nil, err
	}

	respBody, err := c.post(PATH_ONBOARDING, api.OnboardingRequest{
		IsClient:    true,
		Wallet:      wallet,
		Signature:   signature,
		ProfileType: "vault",
	}, nil)
	if err != nil {
		return nil, err
	}

	var resp Response[auth.OnboardFrontendResult]

	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	result := resp.Result[0]
	c.Credentials.Wallet = *result.Profile.Wallet
	c.Credentials.Jwt = result.Jwt
	c.Credentials.ProfileID = result.Profile.ProfileID
	c.Credentials.APISecret = result.RandomSecret

	return &resp, nil
}
