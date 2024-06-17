package api_client

import (
	"encoding/json"

	"github.com/strips-finance/rabbit-dex-backend/api"
)

func (c *Client) JwtUpdate(params *api.JwtRequest) (*Response[api.JwtMarketMakerResponse], error) {
	respBody, err := c.post(PATH_JWT, params, nil)
	if err != nil {
		return nil, err
	}

	var resp Response[api.JwtMarketMakerResponse]
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	if resp.Result != nil && len(resp.Result) > 0 {
		c.Credentials.Jwt = resp.Result[0].Jwt
	}

	return &resp, nil
}
