package api_client

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/strips-finance/rabbit-dex-backend/api"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

func (c *Client) BalanceOpsList(params *api.BalanceOpsRequest) (*Response[[]*model.BalanceOps], error) {
	queryParams := make(map[string]string)
	queryParams["profile_id"] = strconv.Itoa(int(params.ProfileId))
	queryParams["offset"] = strconv.Itoa(int(params.Offset))
	queryParams["limit"] = strconv.Itoa(int(params.Limit))

	QUERY_PATH := fmt.Sprintf("%s/balanceops/list", PATH_ADMIN)
	respBody, err := c.get(QUERY_PATH, queryParams, nil)
	if err != nil {
		return nil, err
	}

	var resp Response[[]*model.BalanceOps]
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) ProfileCache(params *api.ProfileCacheRequest) (*Response[model.ProfileCache], error) {
	queryParams := make(map[string]string)
	queryParams["profile_id"] = strconv.Itoa(int(params.ProfileId))

	QUERY_PATH := fmt.Sprintf("%s/profile", PATH_ADMIN)
	respBody, err := c.get(QUERY_PATH, queryParams, nil)
	if err != nil {
		return nil, err
	}

	var resp Response[model.ProfileCache]
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) IsInv3Valid() (*Response[bool], error) {
	QUERY_PATH := fmt.Sprintf("%s/inv3", PATH_ADMIN)
	respBody, err := c.get(QUERY_PATH, nil, nil)
	if err != nil {
		return nil, err
	}

	var resp Response[bool]
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) ExchangeTotals() (*Response[model.ExchangeData], error) {
	QUERY_PATH := fmt.Sprintf("%s/exchange/totals", PATH_ADMIN)
	respBody, err := c.get(QUERY_PATH, nil, nil)
	if err != nil {
		return nil, err
	}

	var resp Response[model.ExchangeData]
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) AddTier(params *api.AddTierRequest) (*Response[*model.Tier], error) {
	QUERY_PATH := fmt.Sprintf("%s/tiers", PATH_SUPER_ADMIN)

	respBody, err := c.post(QUERY_PATH, params, nil)
	if err != nil {
		return nil, err
	}

	var resp Response[*model.Tier]
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) WhichTier(params *api.WhichTierRequest) (*Response[model.TierData], error) {
	queryParams := make(map[string]string)
	queryParams["profile_id"] = strconv.Itoa(int(params.ProfileId))

	QUERY_PATH := fmt.Sprintf("%s/tiers/which", PATH_SUPER_ADMIN)
	respBody, err := c.get(QUERY_PATH, queryParams, nil)
	if err != nil {
		return nil, err
	}

	var resp Response[model.TierData]
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
