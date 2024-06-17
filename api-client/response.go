package api_client

type Response[T any] struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Result  []T    `json:"result"`
}
