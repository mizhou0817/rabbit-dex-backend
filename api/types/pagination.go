package types

type PaginationResponse struct {
	Total int64  `json:"total"`
	Page  int64  `json:"page"`
	Limit int64  `json:"limit"`
	Order string `json:"order"`
}

type PaginationRequestParams struct {
	Page  int64
	Limit int64
	Order string
}
