package tests

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"
)

type marketItem struct {
	MarketId         string  `json:"marketID"`
	MinOrder         float64 `json:"minOrder"`
	MinTick          float64 `json:"minTick"`
	InitialMargin    float64 `json:"initialMargin"`
	ForcedMargin     float64 `json:"forcedMargin"`
	LiquidatedMargin float64 `json:"liquidatedMargin"`
	FairPrice        float64 `json:"fairPrice"`
}

type sequenceItem struct {
	Action       string  `json:"action"`
	MarketId     string  `json:"marketID"`
	Amount       float64 `json:"amount,omitempty"`
	TraderId     uint    `json:"traderID,omitempty"`
	OrderType    string  `json:"orderType,omitempty"`
	Price        float64 `json:"price,omitempty"`
	Side         int     `json:"side,omitempty"`
	Size         float64 `json:"size,omitempty"`
	Leverage     float64 `json:"leverage,omitempty"` // Cheng's generated json not correctly generate uint
	OrderId      uint    `json:"orderID,omitempty"`
	FairPrice    float64 `json:"fairPrice,omitempty"`
	TriggerPrice float64 `json:"triggerPrice,omitempty"`
	SizePercent  float64 `json:"percentPosition,omitempty"`
	TimeInForce  string  `json:"timeInForce,omitempty"`

	// for liquidate action
	UserId uint `json:"userID,omitempty"`
}

type fillItem struct {
	MarketId   string  `json:"marketID"`
	Price      float64 `json:"price"`
	Size       float64 `json:"size"`
	TakerId    float64 `json:"takerID"` // Cheng's generated json not correctly generate uint
	MakerId    float64 `json:"makerID"` // Cheng's generated json not correctly generate uint
	TakerFee   float64 `json:"takerFee"`
	MakerFee   float64 `json:"makerFee"`
	BidOrderId uint    `json:"bidOrderID"`
	AskOrderId uint    `json:"askOrderID"`
}

type bookStateItem struct {
	MarketId    string  `json:"marketID"`
	TotalBids   float64 `json:"totalBids"`
	BidMin      float64 `json:"bidMin"`
	BidMax      float64 `json:"bidMax"`
	TotalOffers float64 `json:"totalOffers"`
	OfferMin    float64 `json:"offerMin"`
	OfferMax    float64 `json:"offerMax"`
}

type traderAccountItem struct {
	TraderId      uint    `json:"traderID"`
	CumVolume     float64 `json:"cumVolume"`
	WalletBalance float64 `json:"walletBalance"`
	UnrealizedPnl float64 `json:"unrealizedPnL"`
	AccountEquity float64 `json:"AE"`
	Margin        float64 `json:"margin"`
	Withdrawable  float64 `json:"withdrawable"`
}

type exchangeItem struct {
	TradingFee                 float64 `json:"tradingFee"`
	ExchangeBalanceExInsurance float64 `json:"exchangeBalanceExInsurance"`
	CumulativeVolume           float64 `json:"cumulativeVolume"`
	CumulativeVolume_Q         float64 `json:"cumulativeVolume_Q"`
}

type inv3Item struct {
	Status                       bool    `json:"status"`
	SumAE                        float64 `json:"sumAE"`
	ExchangeBal_insuranceDeposit float64 `json:"exchangeBal_insuranceDeposit"`
}

type orderQueueItem struct {
	Sequence        uint    `json:"sequence"`
	TraderID        uint    `json:"traderID"`
	MarketID        string  `json:"market"`
	Price           float64 `json:"price"`
	TriggerPrice    float64 `json:"triggerPrice"`
	PercentPosition float64 `json:"percentPosition"`
	OrderType       string  `json:"orderType"`
	Status          string  `json:"status"`
	StopID          uint    `json:"orderID"`
	Limit           uint    `json:"limit"`
}

type expectedItem struct {
	Fills         []fillItem          `json:"fills"`
	Orderbook     []bookStateItem     `json:"orderbook"`
	TraderAccount []traderAccountItem `json:"traderAccount"`
	Exchange      []exchangeItem      `json:"exchange"`
	INV3          []inv3Item          `json:"INV3"`
	OrderQueue    []orderQueueItem    `json:"orderQueue"`
}

type testJson struct {
	Markets  []marketItem   `json:"market"`
	Sequence []sequenceItem `json:"sequence"`
	Expected expectedItem   `json:"expected"`
}

func LoadJson(path string) *testJson {
	jsonFile, err := os.Open(path)
	if err != nil {
		logrus.Fatalf("JSON file opn error: %s", err.Error())
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var parsedJson testJson
	err = json.Unmarshal(byteValue, &parsedJson)
	if err != nil {
		logrus.Fatalf("JSON parse error: %s", err.Error())
	}
	return &parsedJson
}
