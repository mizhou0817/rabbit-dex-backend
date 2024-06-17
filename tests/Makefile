.PHONY: test
test: test_market_fee test_api test_api_liq test_archived_spaces test_new_liquidation_methods test_reject test_periodics test_insurance_unwind test_admin_api

test_market_fee:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestMarketFee" -v

test_profile_fee:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestProfileFee" -v

.PHONY: test_api
test_api: test_api_public test_api_test test_api_liq test_api_depositwithdraw test_api_airdrop

test_api_public:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestAPIPublic" -v

test_api_order:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestAPIPublicOrder" -v

test_api_test:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestAPITest" -v

test_api_liq:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestAPILiq" -v

test_api_depositwithdraw:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestAPIDepositWithdraw" -v

test_api_airdrop:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestAPIAirdrop" -v

test_admin_api:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestAdminApi" -v

test_tiers:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestTiers" -v

test_missing_data:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestMissingData" -v

test_cex_auth_api:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestCexAuth" -v

test_ping_limit:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestPingLimit" -v


#fix
#test_latency:
#	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestEngineLatency" -v

#fix
#test_liq_methods:
#	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestLiqMethods" -v

test_client_order_id:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestOrderCreateClientOrderId" -v

all_sequences: test_basic test_advanced test_risk_atomic test_risk_self_cross test_sltp test_stop_orders test_fok_ioc_postonly

#TODO: this test will be failed until we will have adv inside tarantool
risk_2: test_risk_2
new_features: test_client_order_id test_reduce_position

test_orders:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestOrders" -v

test_sltp:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestSLTP" -v

test_stop_orders:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestStopOrders" -v

test_fok_ioc_postonly:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestTimeInForceFokIocPostOnly" -v

test_basic:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestBasicScenario" -v

test_advanced:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestAdvancedScenario" -v

test_risk_atomic:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestRiskmanagerOrdercheck1Atomic" -v

test_risk_2:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestRiskmanagerOrdercheck2" -v

test_risk_self_cross:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestRiskSelfCross" -v

#fix
#test_liquidation_w1w3:
#	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestLiquidationW1W3" -v

#fix
#test_liquidation_only_w1:
#	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestLiquidationOnlyW1" -v

#fix
#test_liquidation_part1_w1:
#	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestLiquidationW1Part1" -v

#fix
#test_liquidation_part1_w3:
#	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestLiquidationW3Part1" -v

#fix
#test_liquidation_only_w3:
#	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestLiquidationOnlyW3" -v

#fix
#test_liquidation_part1_w4:
#	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestLiquidationW4Part1" -v

#fix
#test_independent_w4:
#	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestIndependentW4" -v

#fix
#test_liquidation_part2_w4:
#	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestLiquidationW4Part2" -v

#fix
#test_liquidation_part3_w4:
#	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestLiquidationW4Part3" -v

#fix
#test_liquidation_w4_clawback1:
#	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestLiquidationW4Clawback1" -v

#fix
#test_liquidation_w1w3_part2:
#	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestLiquidationW1W3Part2" -v

test_archived_spaces:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestArchivedSpaces" -v

#fix
#test_rolling:
#	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestRolling" -v

test_new_liquidation_methods:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestNewLiquidationMethods" -v

#runs forever fix it
#test_broker:
#	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestBroker" -v

test_reject:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestRejectPostMatch" -v

test_periodics:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestPeriodics" -v

#fix
#test_sequence:
#	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestSequence" -v

test_insurance_unwind:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestInsuranceUnwind" -v

test_reduce_position:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestReducePosition" -v

test_deadman:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestDeadman" -v

test_vault_permissions:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestVaultPermissions" -v
