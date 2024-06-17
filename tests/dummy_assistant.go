package tests

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"github.com/strips-finance/rabbit-dex-backend/liqengine"
	"github.com/strips-finance/rabbit-dex-backend/model"
	"github.com/strips-finance/rabbit-dex-backend/tdecimal"
)

type DummyAssistant struct{
	alreadySentBatch bool
	VaultsSeen []uint
}

func (da *DummyAssistant) Queue(ctx context.Context, actions []model.Action) error {
	return nil
}

func (da *DummyAssistant) LiquidatedVaults(ctx context.Context, vaults []uint) error {
	da.VaultsSeen = append(da.VaultsSeen, vaults...)
	return nil
}

func (da *DummyAssistant) CompletedLiquidation(ctx context.Context, traderId uint) error {
	return nil
}

func (da *DummyAssistant) CancelExistingOrders(ctx context.Context, traderId uint) error {
	return nil
}

func (da *DummyAssistant) UpdateLastChecked(ctx context.Context, traderId uint) error {
	return nil
}

func (da *DummyAssistant) GetNextLiqBatch(ctx context.Context, last_id *uint, limit int) ([]*model.ProfileCache, error) {

	if da.alreadySentBatch {
		return []*model.ProfileCache{}, nil
	}
	vaultType := "vault"
	traderType := "trader"
	activeStatus := "active"
	liquidatingStatus := "liquidating"
	badMargin := tdecimal.NewDecimal(decimal.NewFromFloat(0.01)) 
	betterMargin := tdecimal.NewDecimal(decimal.NewFromFloat(0.021)) 
	goodMargin := tdecimal.NewDecimal(decimal.NewFromFloat(0.031)) 
	lastCheck := time.Now().Add(-5 * time.Minute).Unix()
    vaultProfile1 := model.ProfileCache{}
	vaultProfile1.ProfileID = 123
	vaultProfile1.ProfileType = &vaultType
	vaultProfile1.AccountMargin = badMargin
	zero := tdecimal.NewDecimal(decimal.NewFromFloat(0.0))
	vaultProfile1.AccountEquity = zero
	vaultProfile1.TotalNotional = zero
	vaultProfile1.LastLiqCheck = &lastCheck
	vaultProfile1.Status = &activeStatus

	vaultProfile2 := model.ProfileCache{}
	vaultProfile2.ProfileID = 124
	vaultProfile2.ProfileType = &vaultType
	vaultProfile2.AccountMargin = betterMargin
	vaultProfile2.AccountEquity = zero
	vaultProfile2.TotalNotional = zero
	vaultProfile2.LastLiqCheck = &lastCheck
	vaultProfile2.Status = &activeStatus

	vaultProfile3 := model.ProfileCache{}
	vaultProfile3.ProfileID = 125
	vaultProfile3.ProfileType = &vaultType
	vaultProfile3.AccountMargin = goodMargin
	vaultProfile3.AccountEquity = zero
	vaultProfile3.TotalNotional = zero
	vaultProfile3.LastLiqCheck = &lastCheck
	vaultProfile3.Status = &liquidatingStatus

	traderProfile1 := model.ProfileCache{}
	traderProfile1.ProfileID = 126
	traderProfile1.ProfileType = &traderType
	traderProfile1.AccountMargin = badMargin
	traderProfile1.AccountEquity = zero
	traderProfile1.TotalNotional = zero
	traderProfile1.LastLiqCheck = &lastCheck
	traderProfile1.Status = &activeStatus

	traderProfile2 := model.ProfileCache{}
	traderProfile2.ProfileID = 127
	traderProfile2.ProfileType = &traderType
	traderProfile2.AccountMargin = betterMargin
	traderProfile2.AccountEquity = zero
	traderProfile2.TotalNotional = zero
	traderProfile2.LastLiqCheck = &lastCheck
	traderProfile2.Status = &activeStatus

	traderProfile3 := model.ProfileCache{}
	traderProfile3.ProfileID = 128
	traderProfile3.ProfileType = &traderType
	traderProfile3.AccountMargin = goodMargin
	traderProfile3.AccountEquity = zero
	traderProfile3.TotalNotional = zero
	traderProfile3.LastLiqCheck = &lastCheck
	traderProfile3.Status = &liquidatingStatus

	da.alreadySentBatch = true

	return []*model.ProfileCache {&vaultProfile1, &traderProfile1, &vaultProfile2, &traderProfile2, &vaultProfile3, &traderProfile3, }, nil
}

func (da *DummyAssistant) GetInsuranceData(ctx context.Context, insurance_id uint) (*liqengine.AccountData, error) {
	return &liqengine.AccountData{}, nil
}

func (da *DummyAssistant) GetAccountData(ctx context.Context, profile *model.ProfileCache) (*liqengine.AccountData, error) {
	acData := liqengine.AccountData{}
	acData.Cache = profile
	return &acData, nil
}

func (da *DummyAssistant) ClawbackRequired(ctx context.Context) bool {
	return false
}

func (da *DummyAssistant) GetWinningTraderPostns(ctx context.Context, marketId string, side string, atPrice float64, insuranceId uint) (map[uint]*model.PositionData, error) {
	return make(map[uint]*model.PositionData), nil
}

func (da *DummyAssistant) GetNextLiquidationServiceId() liqengine.ServiceId {
	return 0
}

func (da *DummyAssistant) GetOrCreateInsurance(ctx context.Context) (uint, error) {
	return 0, nil
}

func (da *DummyAssistant) WaitForCancellAllAccepted(ctx context.Context, traderId uint) error {
	return nil
}

