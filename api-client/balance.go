package api_client

import (
	"encoding/json"

	"github.com/strips-finance/rabbit-dex-backend/api"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

func (c *Client) Deposit(params *api.DepositRequest) (*Response[model.BalanceOps], error) {
	respBody, err := c.post(PATH_DEPOSIT, params, nil)
	if err != nil {
		return nil, err
	}

	var resp Response[model.BalanceOps]
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) Withdraw(params *api.WithdrawalRequest, pkSignature string, pkTimestamp int64) (*Response[model.BalanceOps], error) {
	respBody, err := c.postWithPkSignature(PATH_WITHDRAW, params, pkSignature, pkTimestamp)
	if err != nil {
		return nil, err
	}

	var resp Response[model.BalanceOps]
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) CancelWithdrawal() (*Response[string], error) {
	respBody, err := c.delete(PATH_CANCEL_WITHDRAWAL, nil)
	if err != nil {
		return nil, err
	}

	var resp Response[string]
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) ClaimWithdrawal() (*Response[model.WithdrawalResponse], error) {
	respBody, err := c.post(PATH_CLAIM_WITHDRAWAL, nil, nil)
	if err != nil {
		return nil, err
	}

	var resp Response[model.WithdrawalResponse]
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) ProcessingWithdrawal() (*Response[string], error) {
	respBody, err := c.post(PATH_CLAIM_WITHDRAWAL, nil, nil)
	if err != nil {
		return nil, err
	}

	var resp Response[string]
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
