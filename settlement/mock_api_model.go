package settlement

import (
	"context"
	"math/big"

	"github.com/strips-finance/rabbit-dex-backend/model"
	"github.com/strips-finance/rabbit-dex-backend/tdecimal"
)

type MockApiModel struct {
	ProcessedYield []model.Yield
}

func (m *MockApiModel) GetWithdrawalsSuspended(ctx context.Context) (bool, error) {
	return false, nil
}

func (m *MockApiModel) SuspendWithdrawals(ctx context.Context) error {
	return nil
}

func (m *MockApiModel) GetPendingWithdrawals(ctx context.Context, exchangeId string, chainId uint) ([]*model.BalanceOps, error) {
	return []*model.BalanceOps{}, nil
}

func (m *MockApiModel) GetAllPendingWithdrawals(ctx context.Context) ([]*model.BalanceOps, error) {
	return []*model.BalanceOps{}, nil
}

func (m *MockApiModel) AddContractMap(ctx context.Context, contract_address string, chain_id uint, exchange_id string) (*model.ContractMap, error) {
	return nil, nil
}

func (m *MockApiModel) GetLastProcessedBlockNumber(ctx context.Context, forContract string, chainId uint, eventType string) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (m *MockApiModel) SetLastProcessedBlockNumber(ctx context.Context, lastProcessed *big.Int, forContract string, chainId uint, eventType string) error {
	return nil
}

func (m *MockApiModel) GetPendingDeposits(ctx context.Context, exchangeId string, chainId uint) ([]*model.BalanceOps, error) {
	return []*model.BalanceOps{}, nil
}

func (m *MockApiModel) GetPendingStakes(ctx context.Context, exchangeId string, chainId uint) ([]*model.BalanceOps, error) {
	return []*model.BalanceOps{}, nil
}

func (m *MockApiModel) PendingDepositCanceled(ctx context.Context, opsId string) (bool, error) {
	return false, nil
}

func (m *MockApiModel) UpdatePendingWithdrawals(ctx context.Context, currentBlock *big.Int, future_block *big.Int, for_contract string) error {
	return nil
}

func (m *MockApiModel) ProcessDepositUnknown(ctx context.Context, deposit model.Deposit) error {
	return nil
}

func (m *MockApiModel) ProcessYield(ctx context.Context, yield model.Yield) error {
	m.ProcessedYield = append(m.ProcessedYield, yield)
	return nil
}

func (m *MockApiModel) CompletedWithdrawals(ctx context.Context, ids []*model.WithdrawalTxInfo) error {
	return nil
}
func (m *MockApiModel) GetProfileByWalletForExchangeId(ctx context.Context, wallet, exchange_id string) (*model.Profile, error) {
	return nil, nil
}

func (m *MockApiModel) InvalidateCacheAndNotify(ctx context.Context, profileId uint) (*model.ProfileCache, error) {
	return nil, nil
}

func (m *MockApiModel) InvalidateCache(ctx context.Context, profileId uint) (*model.ProfileCache, error) {
	return nil, nil
}

func (m *MockApiModel) ProcessDeposit(ctx context.Context, profileId uint, deposit model.Deposit, isPoolDeposit bool) error {
	return nil
}

func (m *MockApiModel) ProcessStake(ctx context.Context, stakerProfileId uint, stake model.Stake, fromBalance bool, exchange_id string) (*model.BalanceOps, error) {
	return nil, nil
}

func (m *MockApiModel) Rolling24hWds(ctx context.Context) (*tdecimal.Decimal, error) {
	return nil, nil
}
