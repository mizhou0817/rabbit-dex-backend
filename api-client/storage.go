package api_client

import (
	"encoding/json"

	"github.com/strips-finance/rabbit-dex-backend/api"
)

func (c *Client) ReadStorage(jwt string) (*Response[[]byte], error) {
	queryParams := make(map[string]string)

	respBody, err := c.get(PATH_STORAGE, queryParams, &Secrets{jwt: jwt})
	if err != nil {
		return nil, err
	}

	var resp Response[[]byte]

	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) WriteStorage(jwt string, data *api.WriteStorageRequest) error {
	_, err := c.post(PATH_STORAGE, data, &Secrets{
		jwt: jwt,
	})
	if err != nil {
		return err
	}

	return nil
}
