package settlement

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/strips-finance/rabbit-dex-backend/model"
	"github.com/strips-finance/rabbit-dex-backend/tdecimal"
)

// Handles deposits and withdrawals.
// Interacts with Ethereum Rabbit Solidity contract via go-ethereum and Alchemy
// and with Starknet rabbit Cairo contract via contract_handler.go

type IApiModel interface {
	GetWithdrawalsSuspended(ctx context.Context) (bool, error)
	SuspendWithdrawals(ctx context.Context) error
	GetPendingWithdrawals(ctx context.Context, exchangeId string, chainId uint) ([]*model.BalanceOps, error)
	GetAllPendingWithdrawals(ctx context.Context) ([]*model.BalanceOps, error)
	AddContractMap(ctx context.Context, contract_address string, chain_id uint, exchange_id string) (*model.ContractMap, error)
	GetLastProcessedBlockNumber(ctx context.Context, forContract string, chainId uint, eventType string) (*big.Int, error)
	SetLastProcessedBlockNumber(ctx context.Context, lastProcessed *big.Int, forContract string, chainId uint, eventType string) error
	GetPendingDeposits(ctx context.Context, exchangeId string, chainId uint) ([]*model.BalanceOps, error)
	GetPendingStakes(ctx context.Context, exchangeId string, chainId uint) ([]*model.BalanceOps, error)
	PendingDepositCanceled(ctx context.Context, opsId string) (bool, error)
	UpdatePendingWithdrawals(ctx context.Context, currentBlock *big.Int, future_block *big.Int, for_contract string) error
	GetProfileByWalletForExchangeId(ctx context.Context, wallet, exchange_id string) (*model.Profile, error)
	InvalidateCacheAndNotify(ctx context.Context, profileId uint) (*model.ProfileCache, error)
	InvalidateCache(ctx context.Context, profileId uint) (*model.ProfileCache, error)
	ProcessDeposit(ctx context.Context, profileId uint, deposit model.Deposit, isPoolDeposit bool) error
	ProcessStake(ctx context.Context, stakerProfileId uint, stake model.Stake, fromBalance bool, exchange_id string) (*model.BalanceOps, error)
	ProcessDepositUnknown(ctx context.Context, deposit model.Deposit) error
	ProcessYield(ctx context.Context, yield model.Yield) error
	CompletedWithdrawals(ctx context.Context, ids []*model.WithdrawalTxInfo) error
	Rolling24hWds(ctx context.Context) (*tdecimal.Decimal, error)
}

type SettlementService struct {
	withdrawalSuspended bool
	withdrawMutex       sync.Mutex
	stopf               context.CancelFunc
	apiModel            IApiModel
	handlers            map[string]EthHandler // Map exchange contract address to handler
}

type EthHandler struct {
	ethereumHandler      *EthereumHandler
	cfg                  EthHandlerCfg
	depositInterval      time.Duration
	withdrawalInterval   time.Duration
	processYieldInterval time.Duration
}

const (
	INIITIAL_TRADERS_CAPACITY  = 100
	INIITIAL_RECEIPTS_CAPACITY = 100
)

var (
	ONE = big.NewInt(1)
	//MAX_WITHDRAWAL = big.NewInt(498799000000)
	DECIMALS_18 = int32(18)
)

func NewSettlementService(handlerCfgs map[string]EthHandlerCfg) (*SettlementService, error) {
	broker, err := model.GetBroker()
	if err != nil {
		return nil, fmt.Errorf("error obtaining Tarantool broker: %s", err.Error())
	}
	apiModel := model.NewApiModel(broker)
	return ConstructSettlementService(handlerCfgs, apiModel)
}

func ConstructSettlementService(handlerCfgs map[string]EthHandlerCfg, apiModel IApiModel) (*SettlementService, error) {
	s := &SettlementService{
		apiModel: apiModel,
		handlers: make(map[string]EthHandler, len(handlerCfgs)),
	}

	for _, config := range handlerCfgs {

		//AdHoc: Ensure that tarantool knows about addresses
		_, err := apiModel.AddContractMap(context.Background(),
			strings.ToLower(config.ExchangeAddress),
			config.ChainId,
			strings.ToLower(config.ExchangeId))
		if err != nil {
			return nil, err
		}

		depositIntervalSeconds, err := strconv.ParseInt(config.DepositInterval, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("exchange_id=%s can't read settlement deposit interval err=%s", config.ExchangeId, err.Error())
		}
		depositInterval := time.Second * time.Duration(depositIntervalSeconds)
		withdrawalIntervalSeconds, err := strconv.ParseInt(config.WithdrawalInterval, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("can't read settlement deposit interval err=%s", err.Error())
		}
		withdrawalInterval := time.Second * time.Duration(withdrawalIntervalSeconds)

		withdrawalBlockDelay, success := new(big.Int).SetString(config.WithdrawalBlockDelay, 10)
		if !success {
			return nil, fmt.Errorf("error parsing withdrawalBlockDelay: %s", config.WithdrawalBlockDelay)
		}

		defaultFromBlock, success := new(big.Int).SetString(config.DefaultFromBlock, 10)
		if !success {
			return nil, fmt.Errorf("error parsing defaultfromBlock: %s", config.DefaultFromBlock)
		}

		blockConfirmations, success := new(big.Int).SetString(config.ConfirmationBlocks, 10)
		if !success {
			return nil, fmt.Errorf("error parsing confirmation blocks: \"%s\"", config.ConfirmationBlocks)
		}

		if blockConfirmations.Cmp(big.NewInt(0)) != 1 {
			return nil, fmt.Errorf("error blockConfirmations amount: %v", blockConfirmations)
		}
		var cancelInterval int64
		cancelInterval, err = strconv.ParseInt(config.CancelInterval, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error %s parsing cancel interval: \"%s\"", err.Error(), config.CancelInterval)
		}

		var processYieldInterval time.Duration
		if config.ProcessYield {
			processYieldIntervalSeconds, err := strconv.ParseInt(config.ProcessYieldInterval, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("error %s parsing processYield interval: \"%s\"", err.Error(), config.ProcessYieldInterval)
			}
			processYieldInterval = time.Second * time.Duration(processYieldIntervalSeconds)
		}

		pkStr := config.ClaimerPk

		exchange_address := strings.ToLower(config.ExchangeAddress)
		deposit_address := strings.ToLower(config.DepositAddress)
		ethereumHandler, err := NewEthereumHandler(
			exchange_address,
			deposit_address,
			config.Decimals,
			config.Vaults,
			config.ProviderUrl,
			pkStr,
			apiModel,
			withdrawalBlockDelay,
			defaultFromBlock,
			blockConfirmations,
			cancelInterval,
			config.ProcessYield,
			config.ExchangeId,
			config.ChainId,
		)
		if err != nil {
			return nil, fmt.Errorf("can't create Ethereum handler err=%s", err.Error())
		}

		s.handlers[config.ExchangeAddress] = EthHandler{
			cfg:                  config,
			ethereumHandler:      ethereumHandler,
			depositInterval:      depositInterval,
			withdrawalInterval:   withdrawalInterval,
			processYieldInterval: processYieldInterval,
		}

		ethereumHandler.settlementService = s

		logrus.Infof("exchange_id=%s chain_id=%d ETHhandler created", config.ExchangeId, config.ChainId)
	}

	return s, nil
}

func (s *SettlementService) Run() (context.CancelFunc, error) {
	ctx, cancelf := context.WithCancel(context.Background())
	s.stopf = cancelf

	for eId, handler := range s.handlers {
		logrus.Infof("Starting settlement service for exchangeId:%s exchangeAddress:%s deposit interval %s, withdrawal interval %s, yield interval %s", handler.cfg.ExchangeId, handler.cfg.ExchangeAddress, handler.depositInterval, handler.withdrawalInterval, handler.processYieldInterval)
		pause(ctx, time.Second*15) // give tarantool a chance to get going
		go func(eId string) {

			ethHandler := s.handlers[eId]
			depositTicker := time.NewTicker(ethHandler.depositInterval)
			defer depositTicker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-depositTicker.C:
					s.ProcessDepositsAndStaking(ctx, ethHandler.ethereumHandler)
				}
			}
		}(eId)

		//try to make deposit, withdrawal and process yield run at different times
		var withdrawalStartDelay, processYieldStartDelay time.Duration
		if handler.depositInterval > handler.withdrawalInterval {
			withdrawalStartDelay = handler.withdrawalInterval / 2
		} else {
			withdrawalStartDelay = handler.depositInterval / 2
		}
		processYieldStartDelay = (withdrawalStartDelay * 3) / 2
		ethHandler := s.handlers[eId]
		go func(delay time.Duration) {
			pause(ctx, delay)
			withdrawalTicker := time.NewTicker(ethHandler.withdrawalInterval)
			defer withdrawalTicker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-withdrawalTicker.C:
					s.ProcessWithdrawal(ctx, ethHandler.ethereumHandler)
				}
			}
		}(withdrawalStartDelay)
		if ethHandler.processYieldInterval > 0 {
			go func(delay time.Duration) {
				pause(ctx, delay)
				s.ClaimYield(ctx, ethHandler.ethereumHandler)
				pause(ctx, 5 * time.Minute)
				s.DistributeYield(ctx, ethHandler.ethereumHandler)
				yieldTicker := time.NewTicker(ethHandler.processYieldInterval)
				defer yieldTicker.Stop()
				for {
					select {
					case <-ctx.Done():
						return
					case <-yieldTicker.C:
						s.ClaimYield(ctx, ethHandler.ethereumHandler)
						pause(ctx, 5 * time.Minute)
						s.DistributeYield(ctx, ethHandler.ethereumHandler)
					}
				}
			}(processYieldStartDelay)
		}
	}

	return cancelf, nil
}

func pause(ctx context.Context, duration time.Duration) {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return
	case <-timer.C:
		return
	}
}

func (s *SettlementService) Stop() {
	if s.stopf != nil {
		s.stopf()
	}
}

func (s *SettlementService) ProcessDepositsAndStaking(ctx context.Context, handler *EthereumHandler) {
	logrus.Infof("..TICK*** happened: exchangeId=%s chainId=%d ProcessDepositsAndStaking time=%d", handler.exchangeId, handler.chainId, time.Now().Unix())
	handler.processDepositsAndStaking(ctx)
}

func (s *SettlementService) ProcessWithdrawal(ctx context.Context, handler *EthereumHandler) {
	logrus.Infof("..TICK*** happened exchangeId=%s chainId=%d ProcessWithdrawal time=%d", handler.exchangeId, handler.chainId, time.Now().Unix())
	s.checkWithdrawalSuspendedOnDB(ctx)
	if s.withdrawalSuspended {
		logrus.Infof("settlement_service processing of withdrawals is suspended")
		return
	}
	s.withdrawMutex.Lock()
	s.processPendingWithdrawals(ctx, handler)
	s.withdrawMutex.Unlock()

	handler.completeWithdrawalsAndUnstakes(ctx)
}

func (s *SettlementService) ClaimYield(ctx context.Context, handler *EthereumHandler) {
	logrus.Infof("..TICK*** happened exchangeId=%s chainId=%d ClaimYield time=%d", handler.exchangeId, handler.chainId, time.Now().Unix())
	handler.ClaimYield(ctx)
}

func (s *SettlementService) DistributeYield(ctx context.Context, handler *EthereumHandler) {
	logrus.Infof("..TICK*** happened exchangeId=%s chainId=%d DistributeYield time=%d", handler.exchangeId, handler.chainId, time.Now().Unix())
	handler.DistributeYield(ctx)
}

func (s *SettlementService) checkWithdrawalSuspendedOnDB(ctx context.Context) {
	withdrawalSuspended, err := s.apiModel.GetWithdrawalsSuspended(ctx)
	if err == nil {
		s.withdrawalSuspended = withdrawalSuspended
	} else {
		logrus.Errorf("Settlement service, error checking if withdrawals are suspended: %s", err.Error())
		s.withdrawalSuspended = true
	}
}

func (s *SettlementService) SuspendWithdrawals(ctx context.Context) {
	s.withdrawalSuspended = true
	err := s.apiModel.SuspendWithdrawals(ctx)
	if err != nil {
		logrus.Errorf("Settlement service, error suspending withdrawals: %s", err.Error())
	}
}

func (s *SettlementService) GetWithdrawalsSuspended(ctx context.Context) bool {
	return s.withdrawalSuspended
}

func (s *SettlementService) processPendingWithdrawals(ctx context.Context, handler *EthereumHandler) {
	withdrawals, err := s.apiModel.GetPendingWithdrawals(ctx, handler.exchangeId, handler.chainId)
	if err != nil {
		logrus.Errorf("Settlement service, error retrieving pending withdrawals: %s", err.Error())
		return
	}
	if len(withdrawals) == 0 {
		logrus.Debugf("Settlement service, found no pending withdrawals")
		return
	}

	MAX_WITHDRAWAL_18 := new(big.Int)
	MAX_WITHDRAWAL_18, ok := MAX_WITHDRAWAL_18.SetString("4000000000000000000000000", 10)
	if !ok {
		logrus.Error("CAN'T init MAX_WITHDRAWAL_18")
		return
	}

	rolling24h, _ := s.apiModel.Rolling24hWds(ctx)
	totalPending := big.NewInt(0)
	totalPending.Add(totalPending, tdecimal.TDecimalToTokenDecimals(rolling24h, DECIMALS_18))

	if totalPending.Cmp(MAX_WITHDRAWAL_18) == 1 {
		s.SuspendWithdrawals(ctx)
		logrus.Errorf("Settlement service, total pending check, suspending withdrawals: %s", totalPending.String())
		return
	}
	handler.updatePendingWithdrawals(ctx)
}

// Will convert all to max possible 18 decimals and compare with max
func (s *SettlementService) sumUniversalPendingWithdrawals(withdrawals []*model.BalanceOps) *big.Int {
	totalWithdrawal := big.NewInt(0)
	for _, w := range withdrawals {
		totalWithdrawal.Add(totalWithdrawal, tdecimal.TDecimalToTokenDecimals(&w.Amount, DECIMALS_18))
	}
	return totalWithdrawal
}

/*
func (s *SettlementService) sumPendingWithdrawalsPerDecimal(withdrawals []*model.BalanceOps, decimals int32) *big.Int {
	totalWithdrawal := big.NewInt(0)
	for _, w := range withdrawals {
		totalWithdrawal.Add(totalWithdrawal, tdecimal.TDecimalToTokenDecimals(&w.Amount, decimals))
	}
	return totalWithdrawal
}*/
