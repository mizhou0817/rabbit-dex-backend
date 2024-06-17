package tests

var TESTS map[string]map[string]string = map[string]map[string]string{
	"basic": {
		"json":  "./scenarios/orderbook_basic.json",
		"trace": "trace.json",
	},
	"advanced": {
		"json":  "./scenarios/orderbook_advanced.json",
		"trace": "trace.json",
	},
	"riskmanager_ordercheck1_atomic": {
		"json":  "./scenarios/riskmanager_ordercheck1_atomic.json",
		"trace": "trace.json",
	},
	"riskmanager_ordercheck2": {
		"json":  "./scenarios/riskmanager_ordercheck2.json",
		"trace": "trace.json",
	},
	"riskmanager_ordercheck2_selfCross": {
		"json":  "./scenarios/riskmanager_ordercheck2_selfCross.json",
		"trace": "trace.json",
	},
	"liquidation_w1w3": {
		"json":  "./scenarios/liquidation_w1w3.json",
		"trace": "trace.json",
	},
	"independent_w1": {
		"json":  "./scenarios/independent_w1.json",
		"trace": "trace.json",
	},
	"independent_w1_part1": {
		"json":  "./scenarios/independent_w1_part1.json",
		"trace": "trace.json",
	},
	"independent_w3_part1": {
		"json":  "./scenarios/independent_w3_part1.json",
		"trace": "trace.json",
	},
	"independent_w3": {
		"json":  "./scenarios/independent_w3.json",
		"trace": "trace.json",
	},
	"independent_w4_part1": {
		"json":  "./scenarios/independent_w4_part1.json",
		"trace": "trace.json",
	},
	"independent_w4": {
		"json":  "./scenarios/independent_w4.json",
		"trace": "trace.json",
	},
	"independent_w4_part2": {
		"json":  "./scenarios/independent_w4_part2.json",
		"trace": "trace.json",
	},
	"independent_w4_part3": {
		"json":  "./scenarios/independent_w4_part3.json",
		"trace": "trace.json",
	},
	"liquidation_w4_clawback1": {
		"json":  "./scenarios/liquidation_w4_clawback1.json",
		"trace": "trace.json",
	},
	"liquidation_w1w3_part2": {
		"json":  "./scenarios/liquidation_w1w3_part2.json",
		"trace": "trace.json",
	},
	"riskmanager_ordercheck3_reducePosition": {
		"json":  "./scenarios/riskmanager_ordercheck3_reducePosition.json",
		"trace": "trace.json",
	},
	"riskmanager_ordercheck4_SLTP": {
		"json":  "./scenarios/riskmanager_ordercheck4_SLTP.json",
		"trace": "trace.json",
	},
	"riskmanager_ordercheck5_stopLimit_stopMarket": {
		"json":  "./scenarios/riskmanager_ordercheck5_stopLimit_stopMarket.json",
		"trace": "trace.json",
	},
	"riskmanager_ordercheck6_FOK_IOC_postOnly": {
		"json":  "./scenarios/riskmanager_ordercheck6_FOK_IOC_postOnly.json",
		"trace": "trace.json",
	},
	"orders": {
		"json":  "./scenarios/orders.json",
		"trace": "trace.json",
	},
}
