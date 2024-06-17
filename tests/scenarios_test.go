package tests

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestBasicScenario(t *testing.T) {
	var SCENARIO string = "basic"
	logrus.Warnf("************************************************ EXECUTING scenario=%s", SCENARIO)

	broker := ClearAll(t)

	apimodel, market_ids, sequence := doScenario(t, broker, SCENARIO, false)
	checkResults(t, apimodel, market_ids, sequence, false)
}

func TestAdvancedScenario(t *testing.T) {
	var SCENARIO string = "advanced"
	logrus.Warnf("************************************************ EXECUTING scenario=%s", SCENARIO)

	broker := ClearAll(t)
	apimodel, market_ids, sequence := doScenario(t, broker, SCENARIO, false)
	checkResults(t, apimodel, market_ids, sequence, false)
}

func TestRiskmanagerOrdercheck1Atomic(t *testing.T) {
	var SCENARIO string = "riskmanager_ordercheck1_atomic"
	logrus.Warnf("************************************************ EXECUTING scenario=%s", SCENARIO)

	broker := ClearAll(t)
	apimodel, market_ids, sequence := doScenario(t, broker, SCENARIO, false)
	checkResults(t, apimodel, market_ids, sequence, false)
}

func TestRiskmanagerOrdercheck2(t *testing.T) {
	var SCENARIO string = "riskmanager_ordercheck2"
	logrus.Warnf("************************************************ EXECUTING scenario=%s", SCENARIO)

	broker := ClearAll(t)
	apimodel, market_ids, sequence := doScenario(t, broker, SCENARIO, false)
	checkResults(t, apimodel, market_ids, sequence, false)
}

func TestRiskSelfCross(t *testing.T) {
	var SCENARIO string = "riskmanager_ordercheck2_selfCross"
	logrus.Warnf("************************************************ EXECUTING scenario=%s", SCENARIO)

	broker := ClearAll(t)
	apimodel, market_ids, sequence := doScenario(t, broker, SCENARIO, false)
	checkResults(t, apimodel, market_ids, sequence, false)
}

func TestLiquidationW1W3(t *testing.T) {
	var SCENARIO string = "liquidation_w1w3"
	logrus.Warnf("************************************************ EXECUTING scenario=%s", SCENARIO)

	broker := ClearAll(t)
	apimodel, market_ids, sequence := doScenario(t, broker, SCENARIO, false)
	checkResults(t, apimodel, market_ids, sequence, false)
}

func TestLiquidationOnlyW1(t *testing.T) {
	var SCENARIO string = "independent_w1"
	logrus.Warnf("************************************************ EXECUTING scenario=%s", SCENARIO)

	broker := ClearAll(t)

	apimodel, market_ids, sequence := doScenario(t, broker, SCENARIO, false)
	doLiquidations(t, broker, market_ids, sequence)
	checkResults(t, apimodel, market_ids, sequence, false)
}

func TestLiquidationW1Part1(t *testing.T) {
	var SCENARIO string = "independent_w1_part1"
	logrus.Warnf("************************************************ EXECUTING scenario=%s", SCENARIO)

	broker := ClearAll(t)

	apimodel, market_ids, sequence := doScenario(t, broker, SCENARIO, false)
	doLiquidations(t, broker, market_ids, sequence)
	checkResults(t, apimodel, market_ids, sequence, false)
}

func TestLiquidationW3Part1(t *testing.T) {
	var SCENARIO string = "independent_w3_part1"
	logrus.Warnf("************************************************ EXECUTING scenario=%s", SCENARIO)

	broker := ClearAll(t)

	apimodel, market_ids, sequence := doScenario(t, broker, SCENARIO, false)
	doLiquidations(t, broker, market_ids, sequence)
	checkResults(t, apimodel, market_ids, sequence, false)
}

func TestLiquidationOnlyW3(t *testing.T) {
	var SCENARIO string = "independent_w3"
	logrus.Warnf("************************************************ EXECUTING scenario=%s", SCENARIO)

	broker := ClearAll(t)

	apimodel, market_ids, sequence := doScenario(t, broker, SCENARIO, false)
	doLiquidations(t, broker, market_ids, sequence)
	checkResults(t, apimodel, market_ids, sequence, false)
}

func TestLiquidationW4Part1(t *testing.T) {
	var SCENARIO string = "independent_w4_part1"
	logrus.Warnf("************************************************ EXECUTING scenario=%s", SCENARIO)

	broker := ClearAll(t)

	apimodel, market_ids, sequence := doScenario(t, broker, SCENARIO, false)
	doLiquidations(t, broker, market_ids, sequence)
	checkResults(t, apimodel, market_ids, sequence, false)
}

func TestIndependentW4(t *testing.T) {
	var SCENARIO string = "independent_w4"
	logrus.Warnf("************************************************ EXECUTING scenario=%s", SCENARIO)

	broker := ClearAll(t)

	apimodel, market_ids, sequence := doScenario(t, broker, SCENARIO, false)
	doLiquidations(t, broker, market_ids, sequence)
	checkResults(t, apimodel, market_ids, sequence, false)
}

func TestLiquidationW4Part2(t *testing.T) {
	var SCENARIO string = "independent_w4_part2"
	logrus.Warnf("************************************************ EXECUTING scenario=%s", SCENARIO)

	broker := ClearAll(t)

	apimodel, market_ids, sequence := doScenario(t, broker, SCENARIO, false)
	doLiquidations(t, broker, market_ids, sequence)
	checkResults(t, apimodel, market_ids, sequence, false)
}

func TestLiquidationW4Part3(t *testing.T) {
	var SCENARIO string = "independent_w4_part3"
	logrus.Warnf("************************************************ EXECUTING scenario=%s", SCENARIO)

	broker := ClearAll(t)

	apimodel, market_ids, sequence := doScenario(t, broker, SCENARIO, false)
	doLiquidations(t, broker, market_ids, sequence)
	checkResults(t, apimodel, market_ids, sequence, false)
}

func TestLiquidationW4Clawback1(t *testing.T) {
	var SCENARIO string = "liquidation_w4_clawback1"
	logrus.Warnf("************************************************ EXECUTING scenario=%s", SCENARIO)

	broker := ClearAll(t)

	apimodel, market_ids, sequence := doScenario(t, broker, SCENARIO, false)
	doLiquidations(t, broker, market_ids, sequence)
	checkResults(t, apimodel, market_ids, sequence, false)
}

func TestLiquidationW1W3Part2(t *testing.T) {
	var SCENARIO string = "liquidation_w1w3_part2"
	logrus.Warnf("************************************************ EXECUTING scenario=%s", SCENARIO)

	broker := ClearAll(t)

	apimodel, market_ids, sequence := doScenario(t, broker, SCENARIO, false)
	doLiquidations(t, broker, market_ids, sequence)
	checkResults(t, apimodel, market_ids, sequence, false)
}

func TestReducePosition(t *testing.T) {
	var SCENARIO string = "riskmanager_ordercheck3_reducePosition"
	logrus.Warnf("************************************************ EXECUTING scenario=%s", SCENARIO)

	broker := ClearAll(t)

	apimodel, market_ids, sequence := doScenario(t, broker, SCENARIO, false)
	doLiquidations(t, broker, market_ids, sequence)
	checkResults(t, apimodel, market_ids, sequence, false)
}

func TestSLTP(t *testing.T) {
	var SCENARIO string = "riskmanager_ordercheck4_SLTP"
	logrus.Warnf("************************************************ EXECUTING scenario=%s", SCENARIO)

	broker := ClearAll(t)

	apimodel, market_ids, sequence := doScenario(t, broker, SCENARIO, false)
	doLiquidations(t, broker, market_ids, sequence)
	checkResults(t, apimodel, market_ids, sequence, false)
}

func TestStopOrders(t *testing.T) {
	var SCENARIO string = "riskmanager_ordercheck5_stopLimit_stopMarket"
	logrus.Warnf("************************************************ EXECUTING scenario=%s", SCENARIO)

	broker := ClearAll(t)

	apimodel, market_ids, sequence := doScenario(t, broker, SCENARIO, false)
	doLiquidations(t, broker, market_ids, sequence)
	checkResults(t, apimodel, market_ids, sequence, false)
}

func TestTimeInForceFokIocPostOnly(t *testing.T) {
	var SCENARIO string = "riskmanager_ordercheck6_FOK_IOC_postOnly"
	logrus.Warnf("************************************************ EXECUTING scenario=%s", SCENARIO)

	broker := ClearAll(t)

	apimodel, market_ids, sequence := doScenario(t, broker, SCENARIO, false)
	doLiquidations(t, broker, market_ids, sequence)
	checkResults(t, apimodel, market_ids, sequence, false)
}

func TestOrders(t *testing.T) {
	var SCENARIO string = "orders"
	logrus.Warnf("************************************************ EXECUTING scenario=%s", SCENARIO)

	broker := ClearAll(t)

	apimodel, market_ids, sequence := doScenario(t, broker, SCENARIO, false)
	doLiquidations(t, broker, market_ids, sequence)
	checkResults(t, apimodel, market_ids, sequence, false)
}
