package api_client

import (
	"encoding/json"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

func (c *Client) MarketList(marketIDs []string) (*Response[model.MarketData], error) {
	params := make(map[string]string)

	if marketIDs != nil && len(marketIDs) > 0 {
		params["marketID"] = strings.Join(marketIDs, ",")
	}

	respBody, err := c.get(PATH_MARKETS, params, nil)
	if err != nil {
		return nil, err
	}

	var resp Response[model.MarketData]

	if err := json.Unmarshal(respBody, &resp); err != nil {
		logrus.Error(err)
		return nil, err
	}

	return &resp, nil
}
