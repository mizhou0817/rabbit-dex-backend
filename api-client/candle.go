package api_client

import (
	"encoding/json"
	"strconv"

	"github.com/strips-finance/rabbit-dex-backend/api"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

func (c *Client) CandleList(params *api.CandleListRequest) (*Response[model.CandleData], error) {
	queryParams := make(map[string]string)
	queryParams["market_id"] = params.MarketId
	queryParams["timestamp_from"] = strconv.Itoa(int(params.TimestampFrom))
	queryParams["timestamp_to"] = strconv.Itoa(int(params.TimestampTo))
	queryParams["period"] = strconv.Itoa(int(params.Period))

	respBody, err := c.get(PATH_CANDLES, queryParams, nil)
	if err != nil {
		return nil, err
	}

	var resp Response[model.CandleData]

	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
