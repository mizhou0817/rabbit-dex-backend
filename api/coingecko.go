package api

import (
	"github.com/imroc/req/v3"
	neturl "net/url"
)

const COINGECKO_API_URL = "https://pro-api.coingecko.com/api/v3"
const COINGECKO_API_KEY = "CG-eS5rw4sQxgbxd8HP3CEtzgNr"

type CoinGecko struct {
	apiUrl string
	apiKey string
}

func NewCoinGecko() *CoinGecko {
	return &CoinGecko{
		apiUrl: COINGECKO_API_URL,
		apiKey: COINGECKO_API_KEY,
	}
}

func (c *CoinGecko) buildURL(path string) (string, error) {
	return neturl.JoinPath(c.apiUrl, path)
}

func (c *CoinGecko) GetMarkets(coinIds string) ([]interface{}, error) {
	client := req.C()
	client.DisableDump()

	params := map[string]string{
		"vs_currency": "usd",
		"ids":         coinIds,
		"order":       "market_cap_desc",
		"per_page":    "100",
		"page":        "1",
		"sparkline":   "false",
		"locale":      "en",
	}

	url, err := c.buildURL("/coins/markets")
	if err != nil {
		return nil, err
	}

	r, err := client.R().
		SetHeader("x-cg-pro-api-key", c.apiKey).
		SetQueryParams(params).
		Get(url)

	var resp []interface{}
	if r.IsSuccessState() {
		err = r.Into(&resp)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	return resp, nil
}
