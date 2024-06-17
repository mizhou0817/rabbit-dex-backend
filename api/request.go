package api

import "errors"

type (
	OffsetPagination struct {
		Offset uint `json:"offset" form:"offset"`
		Limit  uint `json:"limit" form:"limit"`
	}

	BasePagination struct {
		Page    uint `json:"page" form:"page"`
		PerPage uint `json:"per_page" form:"per_page"`
	}

	PaginationRequest struct {
		*OffsetPagination
		*BasePagination
	}
)

func (p *PaginationRequest) Validate() error {
	if offsetPagination := p.OffsetPagination; offsetPagination != nil {
		if !(offsetPagination.Offset <= 100 && offsetPagination.Offset >= 0) {
			return errors.New("offset value should be between 0 and 100")
		}
	}

	if basePagination := p.BasePagination; basePagination != nil {
		if !(basePagination.PerPage <= 100 && basePagination.PerPage >= 0) {
			return errors.New("offset value should be between 0 and 100")
		}
	}

	return nil
}

func (p *PaginationRequest) Offset() uint {
	var offset uint

	if p.OffsetPagination != nil {
		offset = p.OffsetPagination.Offset
	} else if p.BasePagination != nil && offset == 0 {
		page := uint(1)

		if p.BasePagination.Page >= 1 {
			page = p.BasePagination.Page
		}

		offset = (page - 1) * p.BasePagination.PerPage
	}

	return offset
}

func (p *PaginationRequest) Limit() uint {
	var limit uint

	if p.OffsetPagination != nil {
		limit = p.OffsetPagination.Limit
	} else if p.BasePagination != nil && limit == 0 {
		limit = p.BasePagination.PerPage
	}

	return limit
}
