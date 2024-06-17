package api_client

import (
	"encoding/json"

	"github.com/strips-finance/rabbit-dex-backend/api"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

func (c *Client) Account() (*Response[model.ProfileData], error) {
	queryParams := make(map[string]string)

	respBody, err := c.get(PATH_ACCOUNT, queryParams, nil)
	if err != nil {
		return nil, err
	}

	var resp Response[model.ProfileData]

	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) AccountValidate(jwt string) (*Response[model.ProfileData], error) {
	queryParams := make(map[string]string)
	queryParams["jwt"] = jwt

	respBody, err := c.get(PATH_ACCOUNT_VALIDATE, queryParams, nil)
	if err != nil {
		return nil, err
	}

	var resp Response[model.ProfileData]

	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) AccountUpdateLeverage(params *api.AccountSetLeverageRequest) (*Response[model.ProfileData], error) {
	respBody, err := c.put(PATH_ACCOUNT_LEVERAGE, params)
	if err != nil {
		return nil, err
	}

	var resp Response[model.ProfileData]
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
