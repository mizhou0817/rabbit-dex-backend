package profile

import (
	"errors"

	"github.com/shopspring/decimal"

	"github.com/strips-finance/rabbit-dex-backend/model"
)

type MarginV1LiquidateStrategy struct {
	ForcedMargin decimal.Decimal
}

func (s MarginV1LiquidateStrategy) Process(pc *model.ProfileCache) error {
	if pc.ProfileType == nil {
		return ErrNoProfileType
	}

	switch *pc.ProfileType {
	case model.PROFILE_TYPE_INSURANCE:
	default:
		if pc.AccountMargin.LessThan(s.ForcedMargin) {
			status := model.PROFILE_STATUS_LIQUIDATING
			pc.Status = &status
		}
	}

	return nil
}

type MarginV2LiquidateStrategy struct {}

func (s MarginV2LiquidateStrategy) Process(pc *model.ProfileCache) error {
	panic("not implemented")
}

var (
	ErrNoProfileType = errors.New("no profile type")
)
