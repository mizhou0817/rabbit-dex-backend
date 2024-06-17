package api_client

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
	"github.com/strips-finance/rabbit-dex-backend/api"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

func (c *Client) SecretCreate(params api.SecretCreateRequest, pkSignature string, pkTimestamp int64) (*Response[model.Secret], error) {
	respBody, err := c.postWithPkSignature(PATH_SECRETS, params, pkSignature, pkTimestamp)
	if err != nil {
		return nil, err
	}

	var resp Response[model.Secret]

	if err := json.Unmarshal(respBody, &resp); err != nil {
		logrus.Error(err)
		return nil, err
	}

	return &resp, nil
}

func (c *Client) ListSecrets(pkSignature string, pkTimestamp int64) (*Response[model.Secret], error) {
	queryParams := make(map[string]string)

	respBody, err := c.getWithPkSignature(PATH_SECRETS, queryParams, pkSignature, pkTimestamp)
	if err != nil {
		return nil, err
	}

	var resp Response[model.Secret]

	if err := json.Unmarshal(respBody, &resp); err != nil {
		logrus.Error(err)
		return nil, err
	}

	return &resp, nil
}
