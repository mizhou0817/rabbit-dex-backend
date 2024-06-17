package liqengine

import (
	"context"

	"github.com/strips-finance/rabbit-dex-backend/model"
)

type WinningPos struct {
	TraderId      uint
	Size          float64
	AvgEntryPrice float64
	Side          int
}

type Assistant interface {
	Queue(ctx context.Context, actions []model.Action) error
	LiquidatedVaults(ctx context.Context, vaults []uint) error
	CompletedLiquidation(ctx context.Context, traderId uint) error
	CancelExistingOrders(ctx context.Context, traderId uint) error
	UpdateLastChecked(ctx context.Context, traderId uint) error
	GetNextLiqBatch(ctx context.Context, last_id *uint, limit int) ([]*model.ProfileCache, error)
	GetInsuranceData(ctx context.Context, insurance_id uint) (*AccountData, error)
	GetAccountData(ctx context.Context, profile *model.ProfileCache) (*AccountData, error)
	ClawbackRequired(ctx context.Context) bool
	GetWinningTraderPostns(ctx context.Context, marketId string, side string, atPrice float64, insuranceId uint) (map[uint]*model.PositionData, error)
	GetNextLiquidationServiceId() ServiceId
	GetOrCreateInsurance(ctx context.Context) (uint, error)
	WaitForCancellAllAccepted(ctx context.Context, traderId uint) error
}
