package settlement

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	"github.com/strips-finance/rabbit-dex-backend/pkg/log"

	"github.com/strips-finance/rabbit-dex-backend/model"
	"github.com/strips-finance/rabbit-dex-backend/tdecimal"

	"github.com/strips-finance/rabbit-dex-backend/settlement/bfx"
	"github.com/strips-finance/rabbit-dex-backend/settlement/deposit"
	"github.com/strips-finance/rabbit-dex-backend/settlement/rabbit_l1"
	"github.com/strips-finance/rabbit-dex-backend/vault/vault"
)

const (
	GAS_TOO_LOW                 = "max fee per gas less than block base fee"
	WITHDRAW_ACTION             = "withdraw"
	DEPOSIT_ACTION              = "deposit"
	CONSUME_MESSAGES_BATCH_SIZE = 500

	DEPOSIT_AND_STAKING_EVENT  = "deposit_and_staking"
	WITHDRAW_AND_UNSTAKE_EVENT = "withdraw_and_unstake"
	PROCESS_YIELD_EVENT        = "process_yield"

	BLAST_CHAIN_ID         = 81457
	BLAST_SEPOLIA_CHAIN_ID = 168587773
)

var (
	MAX_BLOCKS = big.NewInt(1999)
	ZERO       = big.NewInt(0)
	TWO        = big.NewInt(2)
)

type EthereumHandler struct {
	deposit_address      common.Address
	exchange_address     common.Address
	decimals             int32
	vaults               []common.Address
	providerUrl          string
	pkStr                string
	ethClient            *ethclient.Client
	bfxInstance          *bfx.Bfx
	privateKey           *ecdsa.PrivateKey
	walletAddress        common.Address
	exchangeAbi          abi.ABI
	depositAbi           abi.ABI
	vaultAbi             abi.ABI
	withdrawalReceiptID  common.Hash
	rabbitDepositID      common.Hash
	proxyDepositID       common.Hash
	vaultStakeID         common.Hash
	claimedYieldID       common.Hash
	latestBlock          *big.Int
	withdrawalBlockDelay *big.Int
	apiModel             IApiModel
	settlementService    *SettlementService
	defaultFromBlock     *big.Int
	blockConfirmations   *big.Int
	cancelInterval       time.Duration
	claimYield           bool
	exchangeId           string
	chainId              uint
}

func stripPrefix(input string, charsToRemove int) string {
	asRunes := []rune(input)
	return string(asRunes[charsToRemove:])
}

func NewEthereumHandler(
	exchange_address string,
	deposit_address string,
	decimals int32,
	vault_addresses []string,
	providerUrl string,
	pkStr string,
	apiModel IApiModel,
	withdrawalBlockDelay *big.Int,
	defaultFromBlock *big.Int,
	blockConfirmations *big.Int,
	cancelInterval int64,
	claimYield bool,
	exchangeId string,
	chainId uint,
) (*EthereumHandler, error) {

	if strings.HasPrefix(exchange_address, "0x") {
		exchange_address = stripPrefix(exchange_address, 2)
	}

	eh := &EthereumHandler{
		exchange_address:     common.HexToAddress(exchange_address),
		deposit_address:      common.HexToAddress(deposit_address),
		decimals:             decimals,
		providerUrl:          providerUrl,
		pkStr:                pkStr,
		apiModel:             apiModel,
		withdrawalBlockDelay: withdrawalBlockDelay,
		defaultFromBlock:     defaultFromBlock,
		blockConfirmations:   blockConfirmations,
		claimYield:           claimYield,
		exchangeId:           exchangeId,
		chainId:              chainId,
	}

	eh.cancelInterval = time.Duration(cancelInterval) * time.Second

	eh.vaults = make([]common.Address, 0, len(vault_addresses))
	for _, vault_address := range vault_addresses {
		if strings.HasPrefix(vault_address, "0x") {
			vault_address = stripPrefix(vault_address, 2)
			eh.vaults = append(eh.vaults, common.HexToAddress(vault_address))
		}
	}
	var abiStr string
	if chainId == BLAST_CHAIN_ID || chainId == BLAST_SEPOLIA_CHAIN_ID {
		abiStr = string(bfx.BfxMetaData.ABI)
	} else {
		abiStr = string(rabbit_l1.RabbitL1MetaData.ABI)
	}
	exchangeAbi, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return nil,
			fmt.Errorf("error retrieving L1 Rabbit contract ABI: %s", err.Error())
	}
	eh.exchangeAbi = exchangeAbi
	foundWithdrawalReceiptID := false
	foundRabbitDepositID := false
	foundClaimedYieldID := false
	for _, eventType := range eh.exchangeAbi.Events {
		if eventType.Name == "WithdrawalReceipt" {
			eh.withdrawalReceiptID = eventType.ID
			foundWithdrawalReceiptID = true
		} else if eventType.Name == "Deposit" {
			eh.rabbitDepositID = eventType.ID
			foundRabbitDepositID = true
		} else if eventType.Name == "ClaimedYield" {
			eh.claimedYieldID = eventType.ID
			foundClaimedYieldID = true
		}
		if foundWithdrawalReceiptID && foundRabbitDepositID && (foundClaimedYieldID || !eh.claimYield) {
			break
		}
	}

	if !foundWithdrawalReceiptID {
		return nil, errors.New("rabbit ABI does not define an event named WithdrawalReceipt")
	}

	if !foundRabbitDepositID {
		return nil, errors.New("rabbit ABI does not define an event named Deposit")
	}

	if eh.claimYield && !foundClaimedYieldID {
		return nil, errors.New("rabbit ABI does not define an event named ClaimedYield")
	}

	depositAbi, err := abi.JSON(strings.NewReader(
		string(deposit.DepositMetaData.ABI)))
	if err != nil {
		return nil,
			fmt.Errorf("errror retrieving L1 PoolDeposit contract ABI: %s", err.Error())
	}
	eh.depositAbi = depositAbi
	foundProxyDepositID := false
	for _, eventType := range eh.depositAbi.Events {
		if eventType.Name == "Deposit" {
			eh.proxyDepositID = eventType.ID
			foundProxyDepositID = true
			break
		}
	}
	if !foundProxyDepositID {
		return nil, errors.New("deposit ABI does not define an event named Deposit")
	}

	vaultAbi, err := abi.JSON(strings.NewReader(
		string(vault.VaultMetaData.ABI)))
	if err != nil {
		return nil,
			fmt.Errorf("errror retrieving Vault contract ABI: %s", err.Error())
	}
	eh.vaultAbi = vaultAbi
	foundStakeID := false
	for _, eventType := range eh.vaultAbi.Events {
		if eventType.Name == "Stake" {
			eh.vaultStakeID = eventType.ID
			foundStakeID = true
			break
		}
	}

	if !foundStakeID {
		return nil, errors.New("vault ABI does not define an event named Stake")
	}

	if eh.pkStr != "" && eh.claimYield {
		eh.privateKey, err = crypto.HexToECDSA(eh.pkStr)
		if err != nil {
			return nil, fmt.Errorf("error: <%s> parsing private key: \"%s\"", err.Error(), eh.pkStr)
		}
		publicKey := eh.privateKey.Public()
		publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("error casting public key: %v", publicKey)
		}
		eh.walletAddress = crypto.PubkeyToAddress(*publicKeyECDSA)
	} else {
		logrus.Warnf("CLAIM_YIELD_NOT_ACTIVE for exchange_id=%s chain_id=%d claimYield=%t", eh.exchangeId, eh.chainId, eh.claimYield)
	}

	logrus.Infof("created ethereum_handler, rabbit address: %s, pool deposit address %s, provider url: %s, withdrawal block delay: %s, default from block: %s, block confirmations: %s, cancel interval: %s, claim yield: %t",

		exchange_address, deposit_address, eh.providerUrl,
		eh.withdrawalBlockDelay.String(), eh.defaultFromBlock.String(),
		eh.blockConfirmations.String(), eh.cancelInterval.String(), eh.claimYield)

	return eh, nil
}

func (eh *EthereumHandler) Dial() error {
	eh.ethClient = nil
	sleepFor := 4 * time.Second
	var client *ethclient.Client
	var err error
	for i := 0; i < 5; i++ {
		client, err = ethclient.Dial(eh.providerUrl)
		if err == nil {
			break
		}
		time.Sleep(sleepFor)
		sleepFor = sleepFor * 2
	}
	if err != nil {
		return fmt.Errorf("error dialing eth client: %s", err.Error())
	}
	eh.ethClient = client
	if eh.chainId == BLAST_CHAIN_ID || eh.chainId == BLAST_SEPOLIA_CHAIN_ID {
		eh.bfxInstance, err = bfx.NewBfx(eh.exchange_address, eh.ethClient)
		if err != nil {
			return fmt.Errorf("error connecting to exchange contract: %s", err.Error())
		}
	}
	return nil
}

func (eh *EthereumHandler) processDepositsAndStaking(ctx context.Context) {
	if !eh.ensureConnected(ctx, eh.exchange_address.String(), DEPOSIT_AND_STAKING_EVENT, eh.chainId) {
		return
	}

	lastBlock, err := eh.apiModel.GetLastProcessedBlockNumber(ctx, eh.exchange_address.String(), eh.chainId, DEPOSIT_AND_STAKING_EVENT)
	if err != nil {
		logrus.Errorf("Settlement service, ethereumHandler.processDeposits, error reading last processed deposit block: %s",
			err.Error())
		return
	}
	fromBlock, toBlock, err := eh.getL1BlockNumbers(ctx, lastBlock)
	if err != nil {
		logrus.Errorf("Settlement service, error reading block numbers: %s", err.Error())
		return
	}

	logrus.Infof("process deposits from block %v, to block %v", fromBlock, toBlock)
	for fromBlock.Cmp(toBlock) != 1 {
		logrus.Infof("ethereum_handler processing deposits in block range %v to %v", fromBlock, toBlock)

		success := eh.processDepositEvents(ctx, fromBlock, toBlock)
		if !success {
			logrus.Errorf("processDepositEvents failed")
			return
		}

		success = eh.processStakeEvents(ctx, fromBlock, toBlock)
		if !success {
			logrus.Errorf("processStakeEvents failed")
			return
		}

		err = eh.apiModel.SetLastProcessedBlockNumber(ctx, toBlock, eh.exchange_address.String(), eh.chainId, DEPOSIT_AND_STAKING_EVENT)
		if err != nil {
			logrus.Errorf("Error setting last processed deposit block number: %s", err.Error())
			return
		}

		lastBlock, err := eh.apiModel.GetLastProcessedBlockNumber(ctx, eh.exchange_address.String(), eh.chainId, DEPOSIT_AND_STAKING_EVENT)
		if err != nil {
			logrus.Errorf("Settlement service, ethereumHandler.processDeposits, error reading last processed deposit block: %s",
				err.Error())
			return
		}
		fromBlock, toBlock, err = eh.getL1BlockNumbers(ctx, lastBlock)
		if err != nil {
			logrus.Errorf("Settlement service, error reading block numbers: %s", err.Error())
			return
		}
	}
	eh.ProcessDroppedDepositsAndStakes(ctx)
}

func (eh *EthereumHandler) DistributeYield(ctx context.Context) {
	if !eh.ensureConnected(ctx, eh.exchange_address.String(), PROCESS_YIELD_EVENT, eh.chainId) {
		return
	}

	lastBlock, err := eh.apiModel.GetLastProcessedBlockNumber(ctx, eh.exchange_address.String(), eh.chainId, PROCESS_YIELD_EVENT)
	if err != nil {
		logrus.Errorf("Settlement service, ethereumHandler.processYields, error reading last processed yield block: %s",
			err.Error())
		return
	}
	fromBlock, toBlock, err := eh.getL1BlockNumbers(ctx, lastBlock)
	if err != nil {
		logrus.Errorf("Settlement service, error reading block numbers: %s", err.Error())
		return
	}

	logrus.Infof("process yield from block %v, to block %v", fromBlock, toBlock)
	for fromBlock.Cmp(toBlock) != 1 {
		logrus.Infof("ethereum_handler processing yield in block range %v to %v", fromBlock, toBlock)

		success := eh.ProcessYieldEvents(ctx, fromBlock, toBlock)
		if !success {
			logrus.Errorf("ProcessYieldEvents failed")
			return
		}

		err = eh.apiModel.SetLastProcessedBlockNumber(ctx, toBlock, eh.exchange_address.String(), eh.chainId, PROCESS_YIELD_EVENT)
		if err != nil {
			logrus.Errorf("Error setting last processed yield block number: %s", err.Error())
			return
		}

		lastBlock, err := eh.apiModel.GetLastProcessedBlockNumber(ctx, eh.exchange_address.String(), eh.chainId, PROCESS_YIELD_EVENT)
		if err != nil {
			logrus.Errorf("Settlement service, ethereumHandler.processYield, error reading last processed yield block: %s",
				err.Error())
			return
		}
		fromBlock, toBlock, err = eh.getL1BlockNumbers(ctx, lastBlock)
		if err != nil {
			logrus.Errorf("Settlement service, error reading block numbers: %s", err.Error())
			return
		}
	}
}

func (eh *EthereumHandler) ProcessDroppedDepositsAndStakes(ctx context.Context) {
	pendingOps, err := eh.apiModel.GetPendingDeposits(ctx, eh.exchangeId, eh.chainId)
	if err != nil {
		logrus.Errorf("Settlement service, error retrieving pending deposits: %s", err.Error())
		return
	}
	pendingStakes, err := eh.apiModel.GetPendingStakes(ctx, eh.exchangeId, eh.chainId)
	if err != nil {
		logrus.Errorf("Settlement service, error retrieving pending stakes: %s", err.Error())
		return
	}
	pendingOps = append(pendingOps, pendingStakes...)
	if !eh.ensureConnected(ctx, eh.exchange_address.String(), DEPOSIT_AND_STAKING_EVENT, eh.chainId) {
		return
	}
	for _, op := range pendingOps {
		if isOlderThan(op.Timestamp, eh.cancelInterval) {
			txhash := op.Txhash
			tx, _, err := eh.ethClient.TransactionByHash(ctx, common.HexToHash(txhash))
			notFound := err != nil && strings.Contains(err.Error(), "not found")
			if notFound || (err == nil && tx == nil) {
				eh.apiModel.PendingDepositCanceled(ctx, op.OpsId)
			} else if err != nil {
				logrus.Errorf("Settlement service, error %s retrieving transaction %v by hash: %s",
					err.Error(), tx, txhash)
				return
			}
		}
	}
}

func isOlderThan(timestamp int64, duration time.Duration) bool {
	stampTime := time.Unix(timestamp/int64(time.Millisecond), 0)
	return time.Since(stampTime) > duration
}

func (eh *EthereumHandler) ensureConnected(ctx context.Context, address, event string, chainId uint) bool {
	if eh.ethClient == nil {
		err := eh.Dial()
		if err != nil {
			logrus.Errorf(
				"Settlement service, error connecting to go-ethereum: %s",
				err.Error(),
			)
			return false
		}
	}
	result := true
	if event != "" {
		err := eh.checkLastProcessedBlockNumbers(ctx, address, event, chainId)
		result = err == nil
	}
	return result
}

func (eh *EthereumHandler) checkLastProcessedBlockNumbers(ctx context.Context, address, event string, chainId uint) error {
	currentBlock, err := eh.getCurrentBlockNumber(ctx)
	if err != nil {
		return err
	}
	lastBlock, err := eh.apiModel.GetLastProcessedBlockNumber(ctx, address, chainId, event)
	if err != nil {
		logrus.Errorf("exchangeId=%s chain_id=%d Settlement service, checkLastProcessedBlockNumbers, error reading last processed block: %s",
			eh.exchangeId, eh.chainId, err.Error())
		lastBlock, err = eh.resetLastProcessedBlockNumber(ctx, address, event, chainId)
		if err != nil {
			return err
		}
	} else if (lastBlock.Cmp(currentBlock) == 1) || (lastBlock.Cmp(eh.defaultFromBlock) == -1) {
		lastBlock, err = eh.resetLastProcessedBlockNumber(ctx, address, event, chainId)
		if err != nil {
			return err
		}
	}

	logrus.Infof("exchangeId=%s chain_id=%d After checkLastProcessedBlockNumbers - currentBlock: %s, lastBlock: %s, defaultFromBlock: %s",
		eh.exchangeId,
		eh.chainId,
		currentBlock.String(),
		lastBlock.String(),
		eh.defaultFromBlock.String())
	return nil
}

func (eh *EthereumHandler) resetLastProcessedBlockNumber(ctx context.Context, address, event string, chainId uint) (*big.Int, error) {
	err := eh.apiModel.SetLastProcessedBlockNumber(ctx, eh.defaultFromBlock, address, chainId, event)
	if err != nil {
		logrus.Errorf("exchangeId=%s chain_id=%d checkLastProcessedBlockNumbers, error setting last processed block: %s",
			eh.exchangeId, eh.chainId, err.Error())
		return nil, err
	}
	logrus.Infof("exchangeId=%s chain_id=%d set last processed block to %s",
		eh.exchangeId, eh.chainId, eh.defaultFromBlock.String())
	return eh.defaultFromBlock, nil
}

func (eh *EthereumHandler) updatePendingWithdrawals(ctx context.Context) {
	if !eh.ensureConnected(ctx, eh.exchange_address.String(), WITHDRAW_AND_UNSTAKE_EVENT, eh.chainId) {
		return
	}
	currentBlock, err := eh.getCurrentBlockNumber(ctx)
	if err != nil {
		return
	}
	futureBlock := new(big.Int).Add(eh.withdrawalBlockDelay, currentBlock)

	logrus.Infof("****ADD_FUTURE: exchange_id=%s current_block=%s future_block=%s delay=%s",
		eh.exchangeId, currentBlock.String(), futureBlock.String(), eh.withdrawalBlockDelay.String())
	eh.apiModel.UpdatePendingWithdrawals(ctx, currentBlock, futureBlock, eh.exchange_address.String())
}

func (eh *EthereumHandler) completeWithdrawalsAndUnstakes(ctx context.Context) {
	if !eh.ensureConnected(ctx, eh.exchange_address.String(), WITHDRAW_AND_UNSTAKE_EVENT, eh.chainId) {
		return
	}
	lastBlock, err := eh.apiModel.GetLastProcessedBlockNumber(ctx, eh.exchange_address.String(), eh.chainId, WITHDRAW_AND_UNSTAKE_EVENT)
	if err != nil {
		logrus.Errorf("Settlement service, error reading last processed block: %s", err.Error())
		return
	}
	fromBlock, toBlock, err := eh.getL1BlockNumbers(ctx, lastBlock)
	if err != nil {
		logrus.Errorf("Settlement service, error reading block numbers: %s", err.Error())
		return
	}

	logrus.Debugf("withdrawal suspended %v, from block %v, to block %v, doing it %v", eh.settlementService.withdrawalSuspended, fromBlock, toBlock, !eh.settlementService.withdrawalSuspended && fromBlock.Cmp(toBlock) != 1)
	for !eh.settlementService.withdrawalSuspended && fromBlock.Cmp(toBlock) != 1 {
		logrus.Infof("ethereum_handler processing withdrawals in block range %v to %v", fromBlock, toBlock)

		success := eh.processWithdrawalReceiptEvents(ctx, fromBlock, toBlock)
		if !success {
			logrus.Errorf("processWithdrawalReceiptEvents failed")
			return
		}

		if eh.settlementService.withdrawalSuspended {
			return
		}

		err := eh.apiModel.SetLastProcessedBlockNumber(ctx, eh.latestBlock, eh.exchange_address.String(), eh.chainId, WITHDRAW_AND_UNSTAKE_EVENT)
		if err != nil {
			logrus.Errorf("Error setting last processed block number: %s", err.Error())
			return
		}

		lastBlock, err := eh.apiModel.GetLastProcessedBlockNumber(ctx, eh.exchange_address.String(), eh.chainId, WITHDRAW_AND_UNSTAKE_EVENT)
		if err != nil {
			logrus.Errorf("Settlement service, error reading last processed block: %s", err.Error())
			return
		}
		fromBlock, toBlock, err = eh.getL1BlockNumbers(ctx, lastBlock)
		if err != nil {
			logrus.Errorf("Settlement service, error reading block numbers: %s", err.Error())
			return
		}
	}
}

func (eh *EthereumHandler) getL1BlockNumbers(ctx context.Context, lastBlock *big.Int) (fromBlock *big.Int, toBlock *big.Int, err error) {

	if lastBlock != nil {
		fromBlock = new(big.Int).Add(lastBlock, ONE)
	} else {
		fromBlock = eh.defaultFromBlock
	}

	toBlock, err = eh.getCurrentBlockNumber(ctx)
	if err != nil {
		return fromBlock, nil, fmt.Errorf("error reading current block number: %s", err.Error())
	}
	// We need to handle only confirmed blocks
	toBlock = toBlock.Sub(toBlock, eh.blockConfirmations)

	fromPlusMaxBlocks := new(big.Int).Add(fromBlock, MAX_BLOCKS)
	if toBlock.Cmp(fromPlusMaxBlocks) == 1 {
		toBlock = fromPlusMaxBlocks
	}

	return fromBlock, toBlock, nil
}

type EventHandler func(eh *EthereumHandler, ctx context.Context, vLog types.Log, values []interface{}, accumulator interface{}) (newAccumulator interface{}, keepGoing bool)

type BatchHandler func(eh *EthereumHandler, ctx context.Context, accumulator interface{}, endedEarly bool) (success bool)

func (eh *EthereumHandler) processDepositEvents(ctx context.Context, fromBlock *big.Int, toBlock *big.Int) (success bool) {
	addresses := make([]common.Address, 0, 2)
	addresses = append(addresses, eh.exchange_address)
	addresses = append(addresses, eh.deposit_address)
	eventTypes := make(map[common.Hash]string, 2)
	eventTypes[eh.rabbitDepositID] = "Deposit"
	eventTypes[eh.proxyDepositID] = "Deposit"
	return eh.processEvents(ctx, fromBlock, toBlock, addresses, eventTypes, depositHandler, nil, nil)
}

func (eh *EthereumHandler) processStakeEvents(ctx context.Context, fromBlock *big.Int, toBlock *big.Int) (success bool) {
	eventTypes := make(map[common.Hash]string, 1)
	eventTypes[eh.vaultStakeID] = "Stake"
	return eh.processEvents(ctx, fromBlock, toBlock, eh.vaults, eventTypes, stakeHandler, nil, nil)
}

func (eh *EthereumHandler) ProcessYieldEvents(ctx context.Context, fromBlock *big.Int, toBlock *big.Int) (success bool) {
	addresses := make([]common.Address, 0, 1)
	addresses = append(addresses, eh.exchange_address)
	eventTypes := make(map[common.Hash]string, 1)
	eventTypes[eh.claimedYieldID] = "ClaimedYield"
	return eh.processEvents(ctx, fromBlock, toBlock, addresses, eventTypes, yieldHandler, nil, nil)
}

func (eh *EthereumHandler) processWithdrawalReceiptEvents(ctx context.Context, fromBlock *big.Int, toBlock *big.Int) (success bool) {
	withdrawalIds := make([]*model.WithdrawalTxInfo, 0, 100)
	addresses := make([]common.Address, 0, 1)
	addresses = append(addresses, eh.exchange_address)
	eventTypes := make(map[common.Hash]string, 1)
	eventTypes[eh.withdrawalReceiptID] = "WithdrawalReceipt"
	return eh.processEvents(ctx, fromBlock, toBlock, addresses, eventTypes, withdrawalReceiptHandler, withdrawalIds, withdrawalBatchHandler)
}

// Process events of the specified type with the supplied handler.
//
// The event handler is called on each event found. It accepts and returns
// an accumulator, which is passed on to the next invocation of the handler,
// and a boolean which indicates whether processing of events should continue.
//
// The accumulator can be used to build an accumulated result from the events.
// This result can then be processed with the batch handler when all events
// have been read. The batch handler returns a boolean indicating whether
// processing of the batch was successful.
//
// Both the accumulator and batch handler can be nil if batch processing
// of accumulated events is not required.
func (eh *EthereumHandler) processEvents(ctx context.Context, fromBlock *big.Int, toBlock *big.Int, addresses []common.Address, eventTypes map[common.Hash]string, handler EventHandler, accumulator interface{}, batchHandler BatchHandler) (success bool) {

	topics := make([]common.Hash, 0, len(eventTypes))
	for topic := range eventTypes {
		topics = append(topics, topic)
	}
	query := ethereum.FilterQuery{
		FromBlock: fromBlock,
		ToBlock:   toBlock,
		Addresses: addresses,
		Topics: [][]common.Hash{
			topics,
		},
	}

	logs, err := eh.ethClient.FilterLogs(ctx, query)
	if err != nil {
		logrus.WithField(log.AlertTag, log.AlertHigh).Errorf("Error filtering L1 rabbit event logs: %v", err)
		err = eh.Dial()
		if err != nil {
			logrus.WithField(log.AlertTag, log.AlertHigh).Errorf("Error re-dialing: %v", err)
			success = false
			return
		}
		logs, err = eh.ethClient.FilterLogs(ctx, query)
		if err != nil {
			logrus.WithField(log.AlertTag, log.AlertHigh).Errorf("Error on second attempt filtering L1 rabbit event log: %v", err)
			success = false
			return
		}
	}

	logrus.Infof("exchangeId=%s chain_id=%d processing events from exchange, found %v", eh.exchangeId, eh.chainId, len(logs))
	endedEarly := false
	for _, vLog := range logs {

		eventId := vLog.Topics[0]
		eventType, exists := eventTypes[eventId]

		if !exists {
			logrus.Errorf("Error: unexpected event ID: %s", eventId.Hex())
			continue
		}

		//protection from chain reorg
		if vLog.Removed {
			logrus.Errorf("Error: vLog removed eventType=%s", eventType)
			continue
		}

		//skip logs from pending blocks
		if vLog.BlockNumber == 0 {
			logrus.Errorf("Error: pending block detected eventType=%s", eventType)
			continue
		}

		var values []interface{}
		var err error
		var recognisedAddress = false
		if vLog.Address == eh.exchange_address {
			values, err = eh.exchangeAbi.Unpack(eventType, vLog.Data)
			recognisedAddress = true
		} else if vLog.Address == eh.deposit_address {
			values, err = eh.depositAbi.Unpack(eventType, vLog.Data)
			recognisedAddress = true
		} else {
			for _, vaultAddress := range eh.vaults {
				if vLog.Address == vaultAddress {
					values, err = eh.vaultAbi.Unpack(eventType, vLog.Data)
					recognisedAddress = true
					break
				}
			}
		}
		if !recognisedAddress {
			logrus.Errorf("Error: unexpected event address: %s", vLog.Address.Hex())
			continue
		}
		if err != nil {
			logrus.Errorf("Error unpacking %s event %s", eventType, err.Error())
			continue
		}

		if len(values) <= 0 {
			logrus.Warnf("No values after unpacking for event %s", eventType)
			continue

		}

		var keepGoing bool
		accumulator, keepGoing = handler(eh, ctx, vLog, values, accumulator)
		if !keepGoing {
			endedEarly = true
			break
		}
	}
	if batchHandler != nil {
		success = batchHandler(eh, ctx, accumulator, endedEarly)
	} else {
		success = !endedEarly
	}
	if success {
		eh.latestBlock = toBlock
	}
	return
}

func depositHandler(eh *EthereumHandler, ctx context.Context, vLog types.Log, values []interface{}, accumulator interface{}) (newAccumulator interface{}, keepGoing bool) {

	deposit_id := new(big.Int)
	deposit_id.SetBytes(vLog.Topics[1].Bytes())
	walletAddr := common.HexToAddress(vLog.Topics[2].String())
	// for pooled deposits a single USDT transfer covers multiple user
	// deposits, pooled deposits can only come from the deposit pool contract
	isPooledDeposit := false
	isFromDepositPool := vLog.Address == eh.deposit_address
	if isFromDepositPool {
		// we know this  deposit came from the pool contract, so we can
		// decode the pool id from the event - if the poolId is zero then
		// this is an individual deposit, otherwise it is a pooled deposit
		poolId := new(big.Int)
		poolId.SetBytes(vLog.Topics[3].Bytes())
		isPooledDeposit = poolId.Cmp(ZERO) != 0
	}
	// ignore rabbit contract deposit events for the deposit contract address
	if walletAddr == eh.deposit_address && !isFromDepositPool {
		return nil, true
	}

	wallet := model.GetWalletStringInRabbitTntStandardFormat(walletAddr.String())

	amount := tdecimal.TokenDecimalsToTDecimal(values[0].(*big.Int), eh.decimals)
	if amount.LessThanOrEqual(decimal.Zero) {
		logrus.Errorf("Wrong deposit amount %v found for wallet %s", amount, wallet)
		return nil, true
	}

	profile, err := eh.apiModel.GetProfileByWalletForExchangeId(ctx, wallet, eh.exchangeId)

	deposit := model.Deposit{
		Id:              fmt.Sprintf("d_%d", deposit_id.Uint64()),
		Wallet:          wallet,
		Amount:          amount,
		Tx:              vLog.TxHash.Hex(),
		ExchangeId:      eh.exchangeId,
		ChainId:         eh.chainId,
		ExchangeAddress: eh.exchange_address.String(),
	}
	logrus.Infof("..DEPOSIT decoded: wallet=%s  deposit_id=%s amount=%v", wallet, deposit.Id, deposit.Amount)
	if err == nil && profile != nil {
		logrus.Infof("processing deposit for %s: %v, is pooled %v", profile.Wallet, deposit.Amount, isPooledDeposit)
		err = eh.apiModel.ProcessDeposit(ctx, profile.ProfileId, deposit, isPooledDeposit)
		eh.apiModel.InvalidateCacheAndNotify(ctx, profile.ProfileId)
	} else if err.Error() == model.PROFILE_NOT_FOUND {
		logrus.Infof("processing deposit for unknown (%s): %v", wallet, deposit.Amount)
		err = eh.apiModel.ProcessDepositUnknown(ctx, deposit)
	}

	if err != nil {
		logrus.Errorf(
			"Error whilst processing deposit %s for wallet: %s, amount: %s, error: %s",
			fmt.Sprintf("d_%d", deposit_id),
			wallet,
			amount.String(),
			err.Error(),
		)
	}
	return nil, true
}

func stakeHandler(eh *EthereumHandler, ctx context.Context, vLog types.Log, values []interface{}, accumulator interface{}) (newAccumulator interface{}, keepGoing bool) {

	vaultWallet := model.GetWalletStringInRabbitTntStandardFormat(vLog.Address.String())

	stake_id := new(big.Int)
	stake_id.SetBytes(vLog.Topics[1].Bytes())
	stakerAddr := common.HexToAddress(vLog.Topics[2].String())
	stakerWallet := model.GetWalletStringInRabbitTntStandardFormat(stakerAddr.String())

	amount := tdecimal.TokenDecimalsToTDecimal(values[0].(*big.Int), eh.decimals)
	if amount.LessThanOrEqual(decimal.Zero) {
		logrus.Errorf("Wrong unstake amount %v found for staker %s on vault %s", amount, stakerWallet, vaultWallet)
		return nil, true
	}

	vaultProfile, err := eh.apiModel.GetProfileByWalletForExchangeId(ctx, vaultWallet, eh.exchangeId)
	if err != nil {
		logrus.Errorf("Error retrieving vault profile for stake %d wallet %s: %s", stake_id.Uint64(), vaultWallet, err.Error())
		return nil, true
	}
	if vaultProfile == nil {
		logrus.Errorf("Vault profile not found for stake %d wallet %s", stake_id.Uint64(), vaultWallet)
		return nil, true
	}
	if vaultProfile.Type != model.PROFILE_TYPE_VAULT {
		logrus.Errorf("Profile is not a vault, stake %d wallet %s", stake_id.Uint64(), vaultWallet)
		return nil, true
	}

	stakerProfile, err := eh.apiModel.GetProfileByWalletForExchangeId(ctx, stakerWallet, eh.exchangeId)
	if err != nil {
		logrus.Errorf("Error retrieving staker profile for stake %d wallet %s: %s", stake_id.Uint64(), stakerWallet, err.Error())
		return nil, true
	}
	if stakerProfile == nil {
		logrus.Errorf("Staker profile not found for stake %d wallet %s", stake_id.Uint64(), stakerWallet)
		return nil, true
	}

	vaultCache, err := eh.apiModel.InvalidateCache(ctx, vaultProfile.ProfileId)
	if err != nil {
		logrus.Errorf("Cache error for stake %d vault profile_id %d", stake_id.Uint64(), vaultProfile.ProfileId)
		return nil, true
	}

	stake := model.Stake{
		Id:             fmt.Sprintf("s_%d", stake_id.Uint64()),
		VaultProfileId: vaultProfile.ProfileId,
		VaultWallet:    vaultWallet,
		Amount:         amount,
		CurrentNav:     vaultCache.AccountEquity,
		Tx:             vLog.TxHash.Hex(),
	}
	logrus.Infof("..STAKE decoded: wallet=%s  deposit_id=%s amount=%v", stakerWallet, stake.Id, stake.Amount)
	logrus.Infof("processing stake %s to %s by %s: %v", stake.Id, vaultWallet, stakerWallet, stake.Amount)
	_, err = eh.apiModel.ProcessStake(ctx, stakerProfile.ProfileId, stake, false, eh.exchangeId)
	eh.apiModel.InvalidateCacheAndNotify(ctx, stakerProfile.ProfileId)
	eh.apiModel.InvalidateCacheAndNotify(ctx, vaultProfile.ProfileId)

	if err != nil {
		logrus.Errorf(
			"Error whilst processing stake %s to vault %s for wallet: %s, amount: %s, error: %s",
			stake.Id,
			vaultWallet,
			stakerWallet,
			amount.String(),
			err.Error(),
		)
	}
	return nil, true
}

func yieldHandler(eh *EthereumHandler, ctx context.Context, vLog types.Log, values []interface{}, accumulator interface{}) (newAccumulator interface{}, keepGoing bool) {

	yield_id := new(big.Int)
	// for a unique id use hash of tx hash (vLog.TxHash), block
	// index (vLog.Index) and chain id

	logIndexBytes := new(big.Int).SetUint64(uint64(vLog.Index)).Bytes()
	combinedBytes := append(vLog.TxHash.Bytes(), logIndexBytes...)
	chain_id_bytes := new(big.Int).SetUint64(uint64(eh.chainId)).Bytes()
	combinedBytes = append(combinedBytes, chain_id_bytes...)
	yield_id.SetBytes(crypto.Keccak256(combinedBytes))

	amount := tdecimal.TokenDecimalsToTDecimal(values[0].(*big.Int), eh.decimals)
	if amount.LessThanOrEqual(decimal.Zero) {
		if amount.LessThan(decimal.Zero) {
			logrus.Errorf("Negative yield amount %s", amount.String())
		}
		return nil, true
	}

	yield := model.Yield{
		Id:              fmt.Sprintf("y_%d", yield_id.Uint64()),
		Amount:          amount,
		Tx:              vLog.TxHash.Hex(),
		ExchangeId:      eh.exchangeId,
		ChainId:         eh.chainId,
		ExchangeAddress: eh.exchange_address.String(),
	}
	logrus.Infof("..yield decoded: yield_id=%s amount=%v", yield.Id, yield.Amount)
	err := eh.apiModel.ProcessYield(ctx, yield)

	if err != nil {
		logrus.Errorf(
			"Error whilst processing yield %s, amount: %s, error: %s",
			fmt.Sprintf("y_%d", yield_id),
			amount.String(),
			err.Error(),
		)
	}
	return nil, true
}

func withdrawalReceiptHandler(eh *EthereumHandler, ctx context.Context, vLog types.Log, values []interface{}, accumulator interface{}) (newAccumulator interface{}, keepGoing bool) {
	withdrawalId := new(big.Int).SetBytes(vLog.Topics[1].Bytes())
	wids, ok := accumulator.([]*model.WithdrawalTxInfo)
	if !ok {
		logrus.Errorf("withdrawalReceiptHandler accummulator expected []*model.WithdrawalTxInfo but got %v", accumulator)
		return nil, false
	}
	withdrawalIdStr := fmt.Sprintf("w_%d", withdrawalId.Uint64())
	logrus.Debugf("withdrawal id %s", withdrawalIdStr)
	txInfo := &model.WithdrawalTxInfo{
		Id:     withdrawalIdStr,
		TxHash: vLog.TxHash.Hex(),
	}
	return append(wids, txInfo), true
}

func withdrawalBatchHandler(eh *EthereumHandler, ctx context.Context, accumulator interface{}, endedEarly bool) (success bool) {
	wids, ok := accumulator.([]*model.WithdrawalTxInfo)
	if !ok {
		logrus.Errorf("withdrawalReceiptHandler accummulator expected []*model.WithdrawalTxInfo but got %v", accumulator)
		success = false
		return
	}
	err := eh.apiModel.CompletedWithdrawals(ctx, wids)
	if err != nil {
		logrus.Errorf("withdrawalBatchHandler error completing withdrawals: %s", err.Error())
		success = false
	} else {
		success = true
	}
	return
}

func (eh *EthereumHandler) getCurrentBlockNumber(ctx context.Context) (*big.Int, error) {
	header, err := eh.ethClient.HeaderByNumber(ctx, nil)
	if err != nil {
		logrus.Errorf("Error retrieving current block number: %s", err.Error())
		return nil, fmt.Errorf("error retrieving current block number: %s", err.Error())
	}
	return header.Number, nil
}

// alternative code for calling the claimYield (or any other) function
// can be useful when testing reverted transactions as it allows you to
// see the tx hash, which the geth abigen generated code approach doesn't

// func (eh *EthereumHandler) TestYieldClaim() (common.Hash, error) {

// 	functionSignature := []byte("claimYield()")
// 	hash := crypto.Keccak256Hash(functionSignature)
// 	functionSelector := hash.Bytes()[:4]

// 	nonce, err := eh.ethClient.PendingNonceAt(context.Background(), eh.walletAddress)
// 	if err != nil {
// 		logrus.Errorf("Failed to get nonce: %v", err)
// 	}

// 	gasPrice, err := eh.ethClient.SuggestGasPrice(context.Background())
// 	if err != nil {
// 		logrus.Errorf("Failed to suggest gas price: %v", err)
// 	}

// 	// Create the transaction
// 	tx := types.NewTransaction(nonce, eh.exchange_address, big.NewInt(0), uint64(300000), gasPrice, functionSelector)

// 	// Sign the transaction with the private key
// 	chainID, err := eh.ethClient.NetworkID(context.Background())
// 	if err != nil {
// 		logrus.Errorf("Failed to get network ID: %v", err)
// 	}

// 	signedTx, err := types.SignTx(tx, types.NewCancunSigner(chainID), eh.privateKey)
// 	// signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), eh.privateKey)
// 	if err != nil {
// 		logrus.Errorf("Failed to sign transaction: %v", err)
// 	}

// 	// Send the transaction
// 	err = eh.ethClient.SendTransaction(context.Background(), signedTx)
// 	if err != nil {
// 		logrus.Errorf("Failed to send transaction: %v", err)
// 	}

// 	fmt.Printf("Transaction sent! TX Hash: %s\n", signedTx.Hash().Hex())
// 	return signedTx.Hash(), nil

// }

func (eh *EthereumHandler) ClaimYield(ctx context.Context) error {
	if !eh.ensureConnected(ctx, eh.exchange_address.String(), "", eh.chainId) {
		return fmt.Errorf("failed to connect to: %s", eh.exchange_address.String())
	}

	signer, err := eh.getSigner(ctx)
	if err != nil {
		return fmt.Errorf("error creating blockchain signer: %s", err.Error())
	}

	tx, err := eh.bfxInstance.ClaimYield(signer)
	count := 0
	for err != nil && count < 5 && strings.Contains(err.Error(), GAS_TOO_LOW) {
		signer.GasPrice = signer.GasPrice.Mul(signer.GasPrice, TWO)
		fmt.Printf("doubled gas price to %v\n", signer.GasPrice)
		tx, err = eh.bfxInstance.ClaimYield(signer)
	}
	if err != nil {
		return fmt.Errorf("error claiming yield: %s", err.Error())
	}
	logrus.Infof("claim yield tx hash 0x%x\n", tx.Hash())
	return nil
}

func (eh *EthereumHandler) getSigner(ctx context.Context) (*bind.TransactOpts, error) {
	nonce, err := eh.ethClient.PendingNonceAt(ctx, eh.walletAddress)
	if err != nil {
		return nil, fmt.Errorf("error reading wallet nonce: %s", err.Error())
	}
	signer, err := bind.NewKeyedTransactorWithChainID(eh.privateKey, big.NewInt(int64(eh.chainId)))
	if err != nil {
		return nil, fmt.Errorf("error creating signer: %s", err.Error())
	}
	signer.Nonce = big.NewInt(int64(nonce))
	signer.Value = big.NewInt(0)
	// gasTipCap, err := eh.client.SuggestGasTipCap(ctx)
	gasPrice, err := eh.ethClient.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("error estimating L1 gas price: %s", err.Error())
	}
	signer.GasPrice = gasPrice
	// signer.GasTipCap = gasTipCap
	return signer, nil
}
