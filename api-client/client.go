package api_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"
	"time"

	"github.com/strips-finance/rabbit-dex-backend/auth"
)

type Secrets struct {
	jwt string
}

type Client struct {
	apiUrl      string
	httpClient  http.Client
	Credentials *ClientCredentials

	// onboardFrontendResp user credentials
}

func NewClient(apiUrl string, credentials *ClientCredentials) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	if credentials == nil {
		panic("nil ClientCredentials isn't acceptable")
	}

	httpClient := http.Client{Jar: jar}
	client := &Client{apiUrl: apiUrl, httpClient: httpClient, Credentials: credentials}

	return client, nil
}

func (c *Client) setCustomHeader(req *http.Request, title, value string) {
	req.Header.Set("Content-Type", "application/json")

	req.Header.Set(title, value)
}

func (c *Client) setHeaders(req *http.Request, signature string) {
	req.Header.Set("Content-Type", "application/json")

	expirationTimestamp := strconv.FormatInt(c.getExpirationTimestamp(), 10)
	req.Header.Set("RBT-TS", expirationTimestamp)

	if apiKey := c.Credentials.APIKey; len(apiKey) > 0 {
		req.Header.Set("RBT-API-KEY", apiKey)
	}

	if len(signature) > 0 {
		req.Header.Set("RBT-SIGNATURE", signature)
	}
}

func (c *Client) signRequest(req *http.Request) (string, error) {
	payload, err := c.parsePayload(req)
	if err != nil {
		return "", err
	}

	// Method is public or it's /onboarding
	if c.Credentials.APISecret == "" {
		return "", nil
	}

	// In case method is public
	if payload == nil {
		return "", nil
	}

	signature, err := payload.Sign(c.Credentials.APISecret)
	if err != nil {
		return "", err
	}

	return signature, nil
}

func (c *Client) getCurrentTimestamp() int64 {
	return time.Now().Unix()
}

func (c *Client) getExpirationTimestamp() int64 {
	return c.getCurrentTimestamp() + SIGNATURE_LIFETIME
}

func (c *Client) getWithPkSignature(path string, params map[string]string, pkSignature string, pkTimestamp int64) ([]byte, error) {
	// Prepare request
	url := fmt.Sprintf("%s%s", c.apiUrl, path)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	pkTimestampStr := strconv.FormatInt(pkTimestamp, 10)

	c.setHeaders(req, "")
	c.setCustomHeader(req, "RBT-PK-SIGNATURE", pkSignature)
	c.setCustomHeader(req, "RBT-PK-TS", pkTimestampStr)

	// Process query params
	q := req.URL.Query()

	for paramKey, paramValue := range params {
		q.Add(paramKey, paramValue)
	}

	req.URL.RawQuery = q.Encode()

	// Send request to server
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (c *Client) get(path string, params map[string]string, secret *Secrets) ([]byte, error) {
	// Prepare request
	url := fmt.Sprintf("%s%s", c.apiUrl, path)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if secret != nil {
		req.AddCookie(&http.Cookie{Name: "jwt", Value: secret.jwt})
	}

	c.setHeaders(req, "")

	// Process query params
	q := req.URL.Query()

	for paramKey, paramValue := range params {
		q.Add(paramKey, paramValue)
	}

	req.URL.RawQuery = q.Encode()

	// Send request to server
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (c *Client) postWithPkSignature(path string, body interface{}, pkSignature string, pkTimestamp int64) ([]byte, error) {
	// Prepare request
	url := fmt.Sprintf("%s%s", c.apiUrl, path)
	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	// Send request to server
	signature, err := c.signRequest(req)
	if err != nil {
		return nil, err
	}

	pkTimestampStr := strconv.FormatInt(pkTimestamp, 10)

	c.setHeaders(req, signature)
	c.setCustomHeader(req, "RBT-PK-SIGNATURE", pkSignature)
	c.setCustomHeader(req, "RBT-PK-TS", pkTimestampStr)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (c *Client) post(path string, body interface{}, secret *Secrets) ([]byte, error) {
	// Prepare request
	url := fmt.Sprintf("%s%s", c.apiUrl, path)
	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	if secret != nil {
		req.AddCookie(&http.Cookie{Name: "jwt", Value: secret.jwt})
	}

	// Send request to server
	signature, err := c.signRequest(req)
	if err != nil {
		return nil, err
	}

	c.setHeaders(req, signature)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w; status: %s", err, resp.Status)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d; body: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (c *Client) put(path string, body interface{}) ([]byte, error) {
	// Prepare request
	url := fmt.Sprintf("%s%s", c.apiUrl, path)
	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	// Send request to server
	signature, err := c.signRequest(req)
	if err != nil {
		return nil, err
	}

	c.setHeaders(req, signature)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (c *Client) delete(path string, body interface{}) ([]byte, error) {
	// Prepare request
	url := fmt.Sprintf("%s%s", c.apiUrl, path)
	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodDelete, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	// Send request to server
	signature, err := c.signRequest(req)
	if err != nil {
		return nil, err
	}

	c.setHeaders(req, signature)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (c *Client) parsePayload(req *http.Request) (*auth.Payload, error) {
	var data map[string]json.RawMessage

	rMethod := req.Method
	if !(rMethod == http.MethodPost || rMethod == http.MethodPut || rMethod == http.MethodDelete) {
		return nil, nil
	}

	// Read raw request body to convert later
	jsonData, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	// Make it readable again for remain api handlers
	req.Body = io.NopCloser(bytes.NewReader(jsonData))
	if err = json.Unmarshal(jsonData, &data); err != nil {
		return nil, err
	}

	// Convert json.RawMessage to string
	payloadData := map[string]string{}
	for k, v := range data {
		payloadData[k] = strings.Trim(string(v), "\"")
	}

	payloadData["method"] = rMethod
	payloadData["path"] = req.URL.Path

	return auth.NewPayload(c.getExpirationTimestamp(), payloadData)
}
