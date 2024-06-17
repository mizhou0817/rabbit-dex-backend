package api_client

import (
	"encoding/json"

	"github.com/strips-finance/rabbit-dex-backend/api"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

func (c *Client) CreateAirdrop(params *api.AirdropCreateRequest) (*Response[*model.Airdrop], error) {
	respBody, err := c.post(PATH_AIRDROP, params, nil)
	if err != nil {
		return nil, err
	}

	var resp Response[*model.Airdrop]
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) GetAirdrops() (*Response[[]*model.ProfileAirdrop], error) {
	queryParams := make(map[string]string)
	respBody, err := c.get(PATH_AIRDROP, queryParams, nil)
	if err != nil {
		return nil, err
	}

	var resp Response[[]*model.ProfileAirdrop]
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) AirdropClaim() (*Response[*model.ProfileAirdrop], error) {
	queryParams := make(map[string]string)
	respBody, err := c.post(PATH_AIRDROP_CLAIM, queryParams, nil)
	if err != nil {
		return nil, err
	}

	var resp Response[*model.ProfileAirdrop]
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) AirdropInit(params *api.ProfileAirdropInitRequest) (*Response[*model.ProfileAirdrop], error) {
	respBody, err := c.post(PATH_AIRDROP_INIT, params, nil)
	if err != nil {
		return nil, err
	}

	var resp Response[*model.ProfileAirdrop]
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// For testing only
func (c *Client) AirdropUpdateClaimable(params *api.UpdateClaimableRequest) (*Response[*model.ProfileAirdrop], error) {
	respBody, err := c.post(PATH_AIRDROP_UPDATE, params, nil)
	if err != nil {
		return nil, err
	}

	var resp Response[*model.ProfileAirdrop]
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
