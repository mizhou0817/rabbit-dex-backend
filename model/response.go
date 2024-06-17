package model

import (
	"context"
	"errors"

	"github.com/shopspring/decimal"
	"github.com/strips-finance/rabbit-dex-backend/tdecimal"
)

type ProfileNotification struct {
	NotifType   string  `msgpack:"type" json:"type"`
	Title       *string `msgpack:"title" json:"title,omitempty"`
	Description *string `msgpack:"description" json:"description,omitempty"`
}

type ProfileMeta struct {
	ProfileID           uint             `msgpack:"profile_id" json:"profile_id"`
	MarketID            string           `msgpack:"market_id" json:"market_id"`
	Status              string           `msgpack:"status" json:"status"`
	CumUnrealizedPnl    tdecimal.Decimal `msgpack:"cum_unrealized_pnl" json:"cum_unrealized_pnl"`
	TotalNotional       tdecimal.Decimal `msgpack:"total_notional" json:"total_notional"`
	TotalPositionMargin tdecimal.Decimal `msgpack:"total_position_margin" json:"total_position_margin"`
	TotalOrderMargin    tdecimal.Decimal `msgpack:"total_order_margin" json:"total_order_margin"`
	InitialMargin       tdecimal.Decimal `msgpack:"initial_margin" json:"initial_margin"`
	MarketLeverage      tdecimal.Decimal `msgpack:"market_leverage" json:"market_leverage"`
	Balance             tdecimal.Decimal `msgpack:"balance" json:"balance"`
	CumTradingVolume    tdecimal.Decimal `msgpack:"cum_trading_volume" json:"cum_trading_volume"`
	Timestamp           int64            `msgpack:"timestamp" json:"timestamp"`
}

type ProfileCache struct {
	ProfileID           uint                         `msgpack:"id" json:"id"`
	ProfileType         *string                      `msgpack:"profile_type" json:"profile_type,omitempty"`
	Status              *string                      `msgpack:"status" json:"status,omitempty"`
	Wallet              *string                      `msgpack:"wallet" json:"wallet,omitempty"`
	LastUpdate          *int64                       `msgpack:"last_update" json:"last_update,omitempty"`
	Balance             *tdecimal.Decimal            `msgpack:"balance" json:"balance,omitempty"`
	AccountEquity       *tdecimal.Decimal            `msgpack:"account_equity" json:"account_equity,omitempty"`
	TotalPositionMargin *tdecimal.Decimal            `msgpack:"total_position_margin" json:"total_position_margin,omitempty"`
	TotalOrderMargin    *tdecimal.Decimal            `msgpack:"total_order_margin" json:"total_order_margin,omitempty"`
	TotalNotional       *tdecimal.Decimal            `msgpack:"total_notional" json:"total_notional,omitempty"`
	AccountMargin       *tdecimal.Decimal            `msgpack:"account_margin" json:"account_margin,omitempty"`
	WithdrawableBalance *tdecimal.Decimal            `msgpack:"withdrawable_balance" json:"withdrawable_balance,omitempty"`
	CumUnrealizedPnl    *tdecimal.Decimal            `msgpack:"cum_unrealized_pnl" json:"cum_unrealized_pnl,omitempty"`
	Health              *tdecimal.Decimal            `msgpack:"health" json:"health,omitempty"`
	AccountLeverage     *tdecimal.Decimal            `msgpack:"account_leverage" json:"account_leverage,omitempty"`
	CumTradingVolume    *tdecimal.Decimal            `msgpack:"cum_trading_volume" json:"cum_trading_volume,omitempty"`
	Leverage            map[string]*tdecimal.Decimal `msgpack:"leverage" json:"leverage,omitempty"`
	LastLiqCheck        *int64                       `msgpack:"last_liq_check" json:"last_liq_check,omitempty"`
	ShardId             string                       `msgpack:"shard_id" json:"-"`
	ArchiveId           int                          `msgpack:"archive_id" json:"-"`
}

type ProfileData struct {
	ProfileCache
	Positions     []*PositionData        `msgpack:"positions" json:"positions,omitempty"`
	Orders        []*OrderData           `msgpack:"orders" json:"orders,omitempty"`
	Notifications []*ProfileNotification `msgpack:"notifications" json:"profile_notifications,omitempty"`
}

type ProfileCacheMetas struct {
	Cache *ProfileCache           `msgpack:"cache"`
	Metas map[string]*ProfileMeta `msgpack:"metas"`
}

type DummyTier struct {
	Tier      uint            `json:"tier"`
	Title     string          `json:"title"`
	MinVolume decimal.Decimal `json:"min_volume"`
}

type TierStatusData struct {
	Current      DummyTier        `json:"current"`
	Next         *DummyTier       `json:"next,omitempty"`
	NeededVolume *decimal.Decimal `json:"needed_volume,omitempty"`
}

type ExtendedProfileData struct {
	ProfileCache
	Positions     []*ExtendedPositionData `msgpack:"positions" json:"positions,omitempty"`
	Orders        []*OrderData            `msgpack:"orders" json:"orders,omitempty"`
	Notifications []*ProfileNotification  `msgpack:"notifications" json:"profile_notifications,omitempty"`
}

type ExtendedProfileTierStatusData struct {
	ExtendedProfileData
	TierStatusData TierStatusData `json:"tier_status"`
}

type OrderData struct {
	OrderId         string            `msgpack:"id" json:"id"`
	ProfileID       uint              `msgpack:"profile_id" json:"profile_id"`
	MarketID        string            `msgpack:"market_id" json:"market_id"`
	OrderType       string            `msgpack:"order_type" json:"order_type"`
	Status          string            `msgpack:"status" json:"status"`
	Price           *tdecimal.Decimal `msgpack:"price" json:"price,omitempty"`
	Size            *tdecimal.Decimal `msgpack:"size" json:"size,omitempty"`
	InitialSize     *tdecimal.Decimal `msgpack:"initial_size" json:"initial_size,omitempty"`
	TotalFilledSize *tdecimal.Decimal `msgpack:"total_filled_size" json:"total_filled_size,omitempty"`
	Side            string            `msgpack:"side" json:"side"`
	Timestamp       int64             `msgpack:"timestamp" json:"timestamp"`
	Reason          string            `msgpack:"reason" json:"reason"`
	ClientOrderId   *string           `msgpack:"client_order_id" json:"client_order_id,omitempty"`
	TriggerPrice    *tdecimal.Decimal `msgpack:"trigger_price" json:"trigger_price,omitempty"`
	SizePercent     *tdecimal.Decimal `msgpack:"size_percent" json:"size_percent,omitempty"`
	TimeInForce     string            `msgpack:"time_in_force" json:"time_in_force"`
	CreatedAt       int64             `msgpack:"created_at" json:"created_at"`
	UpdatedAt       int64             `msgpack:"updated_at" json:"updated_at"`
	ShardId         string            `msgpack:"shard_id" json:"-"`
	ArchiveId       int               `msgpack:"archive_id" json:"-"`
}

type PositionData struct {
	PositionID        string            `msgpack:"id" json:"id"`
	MarketID          string            `msgpack:"market_id" json:"market_id"`
	ProfileID         uint              `msgpack:"profile_id" json:"profile_id"`
	Size              tdecimal.Decimal  `msgpack:"size" json:"size"`
	Side              string            `msgpack:"side" json:"side"`
	EntryPrice        tdecimal.Decimal  `msgpack:"entry_price" json:"entry_price"`
	UnrealizedPnlFair *tdecimal.Decimal `msgpack:"unrealized_pnl" json:"unrealized_pnl,omitempty"`
	NotionalFair      *tdecimal.Decimal `msgpack:"notional" json:"notional,omitempty"`
	Margin            *tdecimal.Decimal `msgpack:"margin" json:"margin,omitempty"`
	LiquidationPrice  *tdecimal.Decimal `msgpack:"liquidation_price" json:"liquidation_price,omitempty"`
	FairPrice         *tdecimal.Decimal `msgpack:"fair_price" json:"fair_price,omitempty"`
	ShardId           string            `msgpack:"shard_id" json:"-"`
	ArchiveId         int               `msgpack:"archive_id" json:"-"`
}

type ExtendedPositionData struct {
	PositionData
	StopLoss   *OrderData `msgpack:"stop_loss" json:"stop_loss"`
	TakeProfit *OrderData `msgpack:"take_profit" json:"take_profit"`
}

type MarketData struct {
	MarketID          string            `msgpack:"id" json:"id"`
	Status            *string           `msgpack:"status" json:"status,omitempty"`
	MinInitialMargin  *tdecimal.Decimal `msgpack:"min_initial_margin" json:"min_initial_margin,omitempty"`
	ForcedMargin      *tdecimal.Decimal `msgpack:"forced_margin" json:"forced_margin,omitempty"`
	LiquidationMargin *tdecimal.Decimal `msgpack:"liquidation_margin" json:"liquidation_margin,omitempty"`
	MinTick           *tdecimal.Decimal `msgpack:"min_tick" json:"min_tick,omitempty"`
	MinOrder          *tdecimal.Decimal `msgpack:"min_order" json:"min_order,omitempty"`
	BestBid           *tdecimal.Decimal `msgpack:"best_bid" json:"best_bid,omitempty"`
	BestAsk           *tdecimal.Decimal `msgpack:"best_ask" json:"best_ask,omitempty"`
	MarketPrice       *tdecimal.Decimal `msgpack:"market_price" json:"market_price,omitempty"`
	IndexPrice        *tdecimal.Decimal `msgpack:"index_price" json:"index_price,omitempty"`
	LastTradePrice    *tdecimal.Decimal `msgpack:"last_trade_price" json:"last_trade_price,omitempty"`
	FairPrice         *tdecimal.Decimal `msgpack:"fair_price" json:"fair_price,omitempty"`

	InstantFundingRate *tdecimal.Decimal `msgpack:"instant_funding_rate" json:"instant_funding_rate,omitempty"`
	LastFundingRate    *tdecimal.Decimal `msgpack:"last_funding_rate_basis" json:"last_funding_rate_basis,omitempty"`

	LastUpdateTime        int64             `msgpack:"last_update_time" json:"last_update_time,omitempty"`
	LastUpdateSequence    int64             `msgpack:"last_update_sequence" json:"last_update_sequence,omitempty"`
	AverageDailyVolumeQ   *tdecimal.Decimal `msgpack:"average_daily_volume_q" json:"average_daily_volume_q,omitempty"`
	LastFundingUpdateTime int64             `msgpack:"last_funding_update_time" json:"last_funding_update_time,omitempty"`
	IconUrl               string            `msgpack:"icon_url" json:"icon_url"`
	MarketTitle           string            `msgpack:"market_title" json:"market_title"`

	ShardId   string `msgpack:"shard_id" json:"-"`
	ArchiveId int    `msgpack:"archive_id" json:"-"`
}

type TradeData struct {
	TradeId     string           `msgpack:"id" json:"id"`
	MarketId    string           `msgpack:"market_id" json:"market_id"`
	Timestamp   uint64           `msgpack:"timestamp" json:"timestamp"`
	Price       tdecimal.Decimal `msgpack:"price" json:"price"`
	Size        tdecimal.Decimal `msgpack:"size" json:"size"`
	Liquidation bool             `msgpack:"liquidation"  json:"liquidation"`
	TakerSide   string           `msgpack:"taker_side"  json:"taker_side"`
	ShardId     string           `msgpack:"shard_id" json:"-"`
	ArchiveId   uint64           `msgpack:"archive_id" json:"-"`
}

type FillData struct {
	Id            string           `msgpack:"id" json:"id"`
	ProfileId     uint             `msgpack:"profile_id" json:"profile_id"`
	MarketId      string           `msgpack:"market_id" json:"market_id"`
	OrderId       string           `msgpack:"order_id" json:"order_id"`
	Timestamp     int64            `msgpack:"timestamp" json:"timestamp"`
	TradeId       string           `msgpack:"trade_id" json:"trade_id"`
	Price         tdecimal.Decimal `msgpack:"price" json:"price"`
	Size          tdecimal.Decimal `msgpack:"size" json:"size"`
	Side          string           `msgpack:"side"  json:"side"`
	IsMaker       bool             `msgpack:"is_maker"  json:"is_maker"`
	Fee           tdecimal.Decimal `msgpack:"fee" json:"fee"`
	Liquidation   bool             `msgpack:"liquidation"  json:"liquidation"`
	ClientOrderId *string          `msgpack:"client_order_id" json:"client_order_id,omitempty"`

	ShardId   string `msgpack:"shard_id" json:"-"`
	ArchiveId int    `msgpack:"archive_id" json:"-"`
}

type OrderbookData struct {
	MarketID  string               `msgpack:"market_id" json:"market_id"`
	Bids      [][]tdecimal.Decimal `msgpack:"bids" json:"bids,omitempty"`
	Asks      [][]tdecimal.Decimal `msgpack:"asks" json:"asks,omitempty"`
	Sequence  uint                 `msgpack:"sequence" json:"sequence"`
	Timestamp int64                `msgpack:"timestamp" json:"timestamp"`
}

type UntypedOrderbookData struct {
	MarketID  string      `msgpack:"market_id" json:"market_id"`
	Bids      interface{} `msgpack:"bids" json:"bids,omitempty"`
	Asks      interface{} `msgpack:"asks" json:"asks,omitempty"`
	Sequence  uint        `msgpack:"sequence" json:"sequence"`
	Timestamp int64       `msgpack:"timestamp" json:"timestamp"`
}

type Profile struct {
	ProfileId  uint   `msgpack:"profile_id" json:"id"`
	Type       string `msgpack:"profile_type" json:"profile_type"`
	Status     string `msgpack:"status" json:"status"`
	Wallet     string `msgpack:"wallet" json:"wallet"`
	CreatedAt  int64  `msgpack:"created_at" json:"created_at"`
	ExchangeId string `msgpack:"exchange_id" json:"-"`
	ShardId    string `msgpack:"shard_id" json:"-"`
	ArchiveId  int    `msgpack:"archive_id" json:"-"`
}

type ExtendedProfile struct {
	Profile
	Balance tdecimal.Decimal `msgpack:"balance" json:"balance"`
}

type BalanceOps struct {
	OpsId           string           `msgpack:"id" json:"id"`
	Status          string           `msgpack:"status" json:"status"`
	Reason          string           `msgpack:"reason" json:"reason"`
	Txhash          string           `msgpack:"txhash" json:"txhash"`
	ProfileId       uint             `msgpack:"profile_id" json:"profile_id"`
	Wallet          string           `msgpack:"wallet" json:"wallet"`
	Type            string           `msgpack:"ops_type" json:"ops_type"`
	Id2             string           `msgpack:"ops_id2" json:"ops_id2"`
	Amount          tdecimal.Decimal `msgpack:"amount" json:"amount"`
	Timestamp       int64            `msgpack:"timestamp" json:"timestamp"`
	DueBlock        uint             `msgpack:"due_block" json:"due_block"`
	ExchangeId      string           `msgpack:"exchange_id" json:"-"`
	ChainId         uint             `msgpack:"chain_id" json:"-"`
	ContractAddress string           `msgpack:"contract_address" json:"-"`
	ShardId         string           `msgpack:"shard_id" json:"shard_id"`
	ArchiveId       int              `msgpack:"archive_id" json:"-"`
}

type VaultInfo struct {
	ProfileId          uint             `msgpack:"vault_profile_id" json:"profile_id"`
	ManagerProfileId   uint             `msgpack:"manager_profile_id" json:"manager_profile_id"`
	TreasurerProfileId uint             `msgpack:"treasurer_profile_id" json:"treasurer_profile_id"`
	PerformanceFee     tdecimal.Decimal `msgpack:"performance_fee" json:"performance_fee"`
	Status             string           `msgpack:"status" json:"status"`
	TotalShares        tdecimal.Decimal `msgpack:"total_shares" json:"total_shares"`
	VaultName          string           `msgpack:"vault_name" json:"vault_name"`
	ManagerName        string           `msgpack:"manager_name" json:"manager_name"`
	InitialisedAt      int64            `msgpack:"initialised_at" json:"initialised_at"`
	ShardId            string           `msgpack:"shard_id" json:"shard_id"`
	ArchiveId          int              `msgpack:"archive_id" json:"-"`
}

type VaultHoldingInfo struct {
	VaultProfileId  uint             `msgpack:"vault_profile_id" json:"vault_profile_id"`
	StakerProfileId uint             `msgpack:"staker_profile_id" json:"staker_profile_id"`
	Shares          tdecimal.Decimal `msgpack:"shares" json:"shares"`
	EntryNav        tdecimal.Decimal `msgpack:"entry_nav" json:"entry_nav"`
	EntryPrice      tdecimal.Decimal `msgpack:"entry_price" json:"entry_price"`
	ShardId         string           `msgpack:"shard_id" json:"shard_id"`
	ArchiveId       int              `msgpack:"archive_id" json:"-"`
}

type CandleData struct {
	Time   int64            `msgpack:"time" json:"time"`
	Low    tdecimal.Decimal `msgpack:"low" json:"low"`
	High   tdecimal.Decimal `msgpack:"high" json:"high"`
	Open   tdecimal.Decimal `msgpack:"open" json:"open"`
	Close  tdecimal.Decimal `msgpack:"close" json:"close"`
	Volume tdecimal.Decimal `msgpack:"volume" json:"volume"`
}

type ExchangeData struct {
	Id           int64            `msgpack:"id" json:"id"`
	TradingFee   tdecimal.Decimal `msgpack:"trading_fee" json:"trading_fee"`
	TotalBalance tdecimal.Decimal `msgpack:"total_balance" json:"total_balance"`
}

type StatsData struct {
	Timestamp    int64  `msgpack:"timestamp" json:"timestamp"`
	MetricName   string `msgpack:"metric_name" json:"metric_name"`
	Count        uint   `msgpack:"count" json:"count"`
	InstanceName string `msgpack:"instance_name" json:"instance_name"`
}

type DataResponse[T any] struct {
	Res   T      `msgpack:"res" json:"res"`
	Error string `msgpack:"error" json:"error"`
}

func (r DataResponse[C]) Request(ctx context.Context, instance string, broker *Broker, fn string, params []interface{}) (C, error) {
	var res []DataResponse[C]
	err := broker.Execute(instance, ctx, fn, params, &res)
	if err != nil {
		return *new(C), err
	}

	if len(res) == 0 {
		return *new(C), errors.New("UNKNOWN ERROR")
	}

	if res[0].Error != "" {
		return *new(C), errors.New(res[0].Error)
	}

	return res[0].Res, nil
}

type EmptyResponse struct {
	Error string `msgpack:"error"`
}

func (r EmptyResponse) Request(ctx context.Context, instance string, broker *Broker, fn string, params []interface{}) error {
	var res []EmptyResponse
	err := broker.Execute(instance, ctx, fn, params, &res)
	if err != nil {
		return err
	}

	if len(res) == 0 {
		return errors.New("UNKNOWN ERROR")
	}

	if res[0].Error != "" {
		return errors.New(res[0].Error)
	}

	return nil
}
