package liqengine

import (
	"math"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/strips-finance/rabbit-dex-backend/model"
	"github.com/strips-finance/rabbit-dex-backend/tdecimal"
)

type LiquidationEngine struct{}

const LIQUIDATION_MARGIN = 0.03
const TAKEOVER_MARGIN = 0.02

const ACTIONS_INITIAL_CAP = 20
const INSURANCE_MAX_ADV_FRAC = 0.001 // fo test: 0.1

// TODO: May 2023 we decided to increase it 10x to liquidate faster.
const TRADER_MAX_ADV_FRAC = 0.001 // for test: 0.01

func (le *LiquidationEngine) belowLiquidationMargin(margin float64) bool {
	return margin < LIQUIDATION_MARGIN
}

func (le *LiquidationEngine) isLiquidationEnding(profile *model.ProfileCache) bool {
	return *profile.Status == model.PROFILE_STATUS_LIQUIDATING && (profile.AccountMargin.InexactFloat64() >= LIQUIDATION_MARGIN)
}

func (le *LiquidationEngine) shouldLiquidationHaveMoreTime(profile *model.ProfileCache) bool {
	interval_passed := IsIntervalPassedForMicroseconds(*profile.LastLiqCheck, WATERFALL1_INTERVAL)
	return *profile.Status == model.PROFILE_STATUS_LIQUIDATING &&
		(profile.AccountMargin.InexactFloat64() < LIQUIDATION_MARGIN) &&
		(profile.AccountMargin.InexactFloat64() >= TAKEOVER_MARGIN) &&
		!interval_passed
}

func (le *LiquidationEngine) requiredActions(account *AccountData) ([]model.Action, []uint) {
	var actions []model.Action
	var liquidatedVaults []uint

	margin := account.Cache.AccountMargin.InexactFloat64()
	if margin < LIQUIDATION_MARGIN {
		if margin >= TAKEOVER_MARGIN {
			actions = le.waterfall1(account)
		} else {
			actions = le.waterfall3(account)
			if *account.Cache.ProfileType == model.PROFILE_TYPE_VAULT {
				liquidatedVaults = append(liquidatedVaults, account.Cache.ProfileID)
			}
		}
	}
	return actions, liquidatedVaults
}

func (le *LiquidationEngine) insuranceSelloffActions(insurance *AccountData) []model.Action {
	return le.generateSellOrders(insurance, INSURANCE_MAX_ADV_FRAC)
}

func (le *LiquidationEngine) waterfall1(account *AccountData) []model.Action {
	return le.generateSellOrders(account, TRADER_MAX_ADV_FRAC)
}

func (le *LiquidationEngine) waterfall3(account *AccountData) []model.Action {
	actions := make([]model.Action, 0, len(account.Positions))
	margin := account.Cache.AccountMargin.InexactFloat64()

	for _, pos := range account.Positions {
		zp := calcZp(pos, margin)
		logrus.Infof(".... pos market_id=%s trader_id=%d size=%f  margin=%.14f zp=%f side=%s", pos.MarketID, pos.ProfileID, pos.Size.InexactFloat64(), margin, zp, pos.Side)

		d_zp := tdecimal.NewDecimal(decimal.NewFromFloat(zp))
		d_size := tdecimal.NewDecimal(decimal.NewFromFloat(pos.Size.InexactFloat64()))
		actions = append(actions, model.Action{
			Kind:     model.AInsTakeover,
			TraderId: account.Cache.ProfileID,
			MarketId: pos.MarketID,
			Size:     *d_size,
			Price:    *d_zp,
		})
	}
	return actions
}

func calcUnrealizedPnl(posSize float64, entryPrice float64, exitPrice float64, posSide string) float64 {
	if posSide == model.LONG {
		return posSize * (exitPrice - entryPrice)
	} else {
		return posSize * (entryPrice - exitPrice)
	}
}

func (le *LiquidationEngine) generateSellOrders(account *AccountData, maxADVFrac float64) []model.Action {
	traderId := account.Cache.ProfileID
	orders := make([]model.Action, 0, len(account.Positions)*4)
	for _, pos := range account.Positions {
		pos_market := account.Markets[pos.MarketID]
		bestAsk := pos_market.BestAsk.InexactFloat64()
		bestBid := pos_market.BestBid.InexactFloat64()
		if bestAsk <= 0.0 && bestBid > 0.0 {
			bestAsk = bestBid
		} else if bestBid <= 0.0 && bestAsk > 0.0 {
			bestBid = bestAsk
		}
		if bestBid <= 0.0 {
			continue
		}
		//oneKSize := 2000.0 / (bestAsk + bestBid) //26.2467191601
		oneKSize := 1000.0 / pos_market.FairPrice.InexactFloat64()
		var orderSz float64

		pos_size := pos.Size.InexactFloat64() //50
		switch {
		case pos_size <= 0:
			continue
		case pos_size <= oneKSize:
			orderSz = pos_size
		case pos_size <= 10.0*oneKSize:
			orderSz = oneKSize
		default:
			orderSz = pos_size * 0.1
		}

		//TODO: be sure that AverageDailyVolume is rolling.
		maxLiqSz := pos_market.AverageDailyVolumeQ.InexactFloat64() * maxADVFrac //0.005004
		if orderSz > maxLiqSz {
			orderSz = maxLiqSz
		}

		min_order := pos_market.MinOrder.InexactFloat64()
		tick := pos_market.MinTick.InexactFloat64()
		if orderSz < min_order*5.0001 {
			orderSz = min_order * 5.0001 // 0.050001
		}

		// for SELL
		risk_limit_price := roundToNearestTick(0.99*pos_market.FairPrice.InexactFloat64(), tick)

		price1 := bestAsk + tick
		if price1 < risk_limit_price {
			price1 = risk_limit_price
		}

		price2 := bestAsk
		if price2 < risk_limit_price {
			price2 = risk_limit_price
		}

		price3 := bestBid + tick
		if price3 < risk_limit_price {
			price3 = risk_limit_price
		}

		price4 := bestBid
		if price4 < risk_limit_price {
			price4 = risk_limit_price
		}

		// for BUY
		if pos.Side == model.SHORT {
			risk_limit_price = roundToNearestTick(1.01*pos_market.FairPrice.InexactFloat64(), tick)

			price1 = bestBid - tick
			if price1 > risk_limit_price {
				price1 = risk_limit_price
			}

			price2 = bestBid
			if price2 > risk_limit_price {
				price2 = risk_limit_price
			}

			price3 = bestAsk - tick
			if price3 > risk_limit_price {
				price3 = risk_limit_price
			}

			price4 = bestAsk
			if price4 > risk_limit_price {
				price4 = risk_limit_price
			}
		}

		if orderSz <= pos_size {
			orderSlices := calcSellOrderSlices(orderSz, min_order)
			if orderSlices[0] > 0.0 {
				d_size := tdecimal.NewDecimal(decimal.NewFromFloat(orderSlices[0]))
				d_price := tdecimal.NewDecimal(decimal.NewFromFloat(price1))
				orders = append(orders,
					model.Action{
						Kind:     model.APlaceSellOrders,
						TraderId: traderId,
						MarketId: pos.MarketID,
						Size:     *d_size,
						Price:    *d_price,
					})
			}
			if orderSlices[1] > 0.0 {
				d_size := tdecimal.NewDecimal(decimal.NewFromFloat(orderSlices[1]))
				d_price := tdecimal.NewDecimal(decimal.NewFromFloat(price2))

				orders = append(orders,
					model.Action{
						Kind:     model.APlaceSellOrders,
						TraderId: traderId,
						MarketId: pos.MarketID,
						Size:     *d_size,
						Price:    *d_price,
					})
			}
			if orderSlices[2] > 0.0 {
				d_size := tdecimal.NewDecimal(decimal.NewFromFloat(orderSlices[2]))
				d_price := tdecimal.NewDecimal(decimal.NewFromFloat(price3))

				orders = append(orders,
					model.Action{
						Kind:     model.APlaceSellOrders,
						TraderId: traderId,
						MarketId: pos.MarketID,
						Size:     *d_size,
						Price:    *d_price,
					})

			}
			if orderSlices[3] > 0.0 {
				d_size := tdecimal.NewDecimal(decimal.NewFromFloat(orderSlices[3]))
				d_price := tdecimal.NewDecimal(decimal.NewFromFloat(price4))

				orders = append(orders,
					model.Action{
						Kind:     model.APlaceSellOrders,
						TraderId: traderId,
						MarketId: pos.MarketID,
						Size:     *d_size,
						Price:    *d_price,
					})

			}
		} else if pos_size >= min_order {
			d_price := tdecimal.NewDecimal(decimal.NewFromFloat(price4))
			orders = append(orders,
				model.Action{
					Kind:     model.APlaceSellOrders,
					TraderId: traderId,
					MarketId: pos.MarketID,
					Size:     pos.Size,
					Price:    *d_price,
				})
		}
	}
	return orders
}

func calcSellOrderSlices(orderSz float64, tick float64) [4]float64 {
	roundedFifth := roundDownToTick(orderSz*0.2, tick)
	totalSz := roundedFifth * 5
	deficit := roundDownToTick(orderSz-totalSz, tick)
	var slices [4]float64
	slices[0] = roundedFifth
	slices[1] = roundedFifth
	slices[2] = roundedFifth
	slices[3] = (roundedFifth * 2) + deficit
	return slices
}

func roundDownToTick(size float64, tick float64) float64 {
	if tick <= 0 {
		return size
	}
	milliTicks := int(math.Round(size * 1000 / tick))
	numTicks := float64(milliTicks / 1000)
	return numTicks * tick
}

func roundUpToTick(size float64, tick float64) float64 {
	if tick <= 0 {
		return size
	}
	milliTicks := int(math.Round(size * 1000 / tick))
	numTicks := float64(milliTicks / 1000)
	if milliTicks%1000 > 0 {
		numTicks += 1.0
	}
	return numTicks * tick
}

func roundToNearestTick(size float64, tick float64) float64 {
	if tick <= 0 {
		return size
	}
	numTicks := math.Round(size / tick)
	return numTicks * tick
}

func calcZp(pos *model.PositionData, margin float64) float64 {
	if pos.Side == model.LONG {
		return pos.FairPrice.InexactFloat64() * (1.0 - margin)
	} else {
		return pos.FairPrice.InexactFloat64() * (1.0 + margin)
	}
}

type traderCapacity struct {
	clawbackSize float64
	remaining    float64
}

func (le *LiquidationEngine) clawbackActions(
	insurance *AccountData,
	insurancePos *model.PositionData,
	winningTraders map[uint]*model.PositionData) []model.Action {

	min_order := insurance.Markets[insurancePos.MarketID].MinOrder.InexactFloat64()

	var totalRemaining float64
	clawbackMap := make(map[uint]*traderCapacity)
	var totalWinningSize float64
	for _, traderPos := range winningTraders {
		totalWinningSize += traderPos.Size.InexactFloat64()
	}
	deficit := insurancePos.Size.InexactFloat64()
	for traderId, traderPos := range winningTraders {
		if deficit < min_order {
			break
		}
		frac := traderPos.Size.InexactFloat64() / totalWinningSize
		clawbackSize := roundToNearestTick(insurancePos.Size.InexactFloat64()*frac, min_order)
		if clawbackSize > deficit {
			clawbackSize = deficit
		}
		var remaining float64
		if clawbackSize > traderPos.Size.InexactFloat64() {
			clawbackSize = traderPos.Size.InexactFloat64()
		} else {
			remaining = traderPos.Size.InexactFloat64() - clawbackSize
		}
		if clawbackSize > 0.0 || remaining > 0.0 {
			clawbackMap[traderId] = &traderCapacity{clawbackSize, remaining}
			deficit -= clawbackSize
			totalRemaining += remaining
		}
	}

	if deficit >= min_order {
		var remainingFrac float64
		if totalRemaining <= deficit {
			remainingFrac = 1.0
		} else if totalRemaining > 0 {
			remainingFrac = deficit / totalRemaining
		}

		for _, traderCapacity := range clawbackMap {
			remainingFrac := roundUpToTick(traderCapacity.remaining*remainingFrac, min_order)
			if remainingFrac > traderCapacity.remaining {
				remainingFrac = traderCapacity.remaining
			}

			traderCapacity.clawbackSize += remainingFrac
			traderCapacity.remaining -= remainingFrac
			deficit -= remainingFrac

			if deficit < min_order {
				break
			}
		}
	}

	zp := calcZp(insurancePos, insurance.Cache.AccountMargin.InexactFloat64())
	insClawbacks := make([]model.Action, 0, len(winningTraders))
	for traderId, traderCapacity := range clawbackMap {
		if deficit >= min_order && traderCapacity.remaining >= min_order {
			extra := math.Min(deficit, traderCapacity.remaining)
			traderCapacity.clawbackSize += extra
			traderCapacity.remaining -= extra
			deficit -= extra
		}

		if traderCapacity.clawbackSize > 0.0 {
			d_size := tdecimal.NewDecimal(decimal.NewFromFloat(traderCapacity.clawbackSize))
			d_price := tdecimal.NewDecimal(decimal.NewFromFloat(zp))

			insClawbacks = append(insClawbacks, model.Action{
				Kind:     model.AInsClawback,
				TraderId: traderId,
				MarketId: insurancePos.MarketID,
				Size:     *d_size,
				Price:    *d_price,
			})
		}
	}

	return insClawbacks
}
