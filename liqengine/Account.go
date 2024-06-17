package liqengine

import (
	"github.com/strips-finance/rabbit-dex-backend/model"
)

type AccountData struct {
	Cache     *model.ProfileCache
	Positions []*model.PositionData
	Markets   map[string]*model.MarketData
}

func FlipSide(side string) string {
	if side == model.LONG {
		return model.SHORT
	}
	return model.LONG
}
