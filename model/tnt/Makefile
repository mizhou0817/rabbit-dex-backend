.PHONY: lib
lib:
	cargo build --manifest-path rbx/Cargo.toml \
	&& ls rbx/target/debug | egrep '\.(so|dylib)' | xargs -I{} cp -f rbx/target/debug/{} . \
	&& mkdir -p .rocks/share/tarantool/rbx && cp -fR rbx/lua/* .rocks/share/tarantool/rbx/

.PHONY: build
build:
	cartridge build

.PHONY: test
test: build
	.rocks/bin/luatest -v ./app/test/luaparts_test.lua
	.rocks/bin/luatest -v ./app/test/util_test.lua
	.rocks/bin/luatest -v ./app/test/ping_limit_test.lua
	.rocks/bin/luatest -v ./app/test/archiver_test.lua
	.rocks/bin/luatest -v ./app/test/engine_notif_test.lua
	.rocks/bin/luatest -v ./app/test/profile_notif_test.lua
	.rocks/bin/luatest -v ./app/test/rolling_test.lua
	.rocks/bin/luatest -v ./app/test/trade_total_volume_test.lua
	.rocks/bin/luatest -v ./app/test/airdrop_test.lua
	.rocks/bin/luatest -v ./app/test/tier_test.lua
	.rocks/bin/luatest -v ./app/test/coid_getter_test.lua
	.rocks/bin/luatest -v ./app/test/clear_coid_test.lua
	.rocks/bin/luatest -v ./app/test/api_migration_test.lua
	.rocks/bin/luatest -v ./app/test/balance_test.lua
	.rocks/bin/luatest -v ./app/test/stake_test.lua
	.rocks/bin/luatest -v ./app/test/integrity_test.lua
	.rocks/bin/luatest -v ./app/test/tick_test.lua
	.rocks/bin/luatest -v ./app/test/revert_test.lua
	.rocks/bin/luatest -v ./app/test/auth_test.lua
	.rocks/bin/luatest -v ./app/test/position_enrich_test.lua
	.rocks/bin/luatest -v ./app/test/pqueue_test.lua
	.rocks/bin/luatest -v ./app/test/maintenance_test.lua
	.rocks/bin/luatest -v ./app/test/cancel_from_the_queue_test.lua
	.rocks/bin/luatest -v ./app/test/api_secret_refresh_test.lua
	.rocks/bin/luatest -v ./app/test/cancel_from_queue_with_no_priority.lua
	.rocks/bin/luatest -v ./app/test/referral_payout_test.lua
	.rocks/bin/luatest -v ./app/test/exchange_id_related_migrations_test.lua
	.rocks/bin/luatest -v ./app/test/is_valid_signer_test.lua
	.rocks/bin/luatest -v ./app/test/vaults_migration_test.lua
	.rocks/bin/luatest -v ./app/test/private_matching_test.lua
	.rocks/bin/luatest -v ./app/test/wdm_test.lua
	.rocks/bin/luatest -v ./app/test/profile_conn_test.lua
	.rocks/bin/luatest -v ./app/test/deadman_test.lua
	.rocks/bin/luatest -v ./app/test/yield_test.lua

	
engine_notif_test: build
	.rocks/bin/luatest -v -c ./app/test/engine_notif_test.lua

rolling_test: build
	.rocks/bin/luatest -v -c ./app/test/rolling_test.lua

market_migration_test: build
	.rocks/bin/luatest -v -c ./app/test/market_migration_test.lua

trade_total_volume_test: build
	.rocks/bin/luatest -v -c ./app/test/trade_total_volume_test.lua

airdrop_test: build
	.rocks/bin/luatest -v -c ./app/test/airdrop_test.lua

tier_test: build
	.rocks/bin/luatest -v -c ./app/test/tier_test.lua

coid_getter_test: build
	.rocks/bin/luatest -v -c ./app/test/coid_getter_test.lua

clear_coid_test: build
	.rocks/bin/luatest -v -c ./app/test/clear_coid_test.lua

api_migration_test: build
	.rocks/bin/luatest -v -c ./app/test/api_migration_test.lua

balance_test:
	.rocks/bin/luatest -v -c ./app/test/balance_test.lua

integrity_test:
	.rocks/bin/luatest -v -c ./app/test/integrity_test.lua
	
revert_test:
	.rocks/bin/luatest -v -c ./app/test/revert_test.lua

auth_test:
	.rocks/bin/luatest -v -c ./app/test/auth_test.lua

position_enrich_test:
	.rocks/bin/luatest -v -c ./app/test/position_enrich_test.lua

pqueue_test:
	.rocks/bin/luatest -v -c ./app/test/pqueue_test.lua

maintenance_test:
	.rocks/bin/luatest -v -c ./app/test/maintenance_test.lua

cancel_queue_test:
	.rocks/bin/luatest -v -c ./app/test/cancel_from_the_queue_test.lua

cancel_queue_no_priority_test:
	.rocks/bin/luatest -v -c ./app/test/cancel_from_queue_with_no_priority.lua

ping_limit_test:
	.rocks/bin/luatest -v -c ./app/test/ping_limit_test.lua

api_secret_refresh_test:
	.rocks/bin/luatest -v -c ./app/test/api_secret_refresh_test.lua

profile_conn_test:
	.rocks/bin/luatest -v -c ./app/test/profile_conn_test.lua

deadman_test:
	.rocks/bin/luatest -v -c ./app/test/deadman_test.lua

referral_payout_test:
	.rocks/bin/luatest -v -c ./app/test/referral_payout_test.lua

eid_migrations_test:
	.rocks/bin/luatest -v -c ./app/test/exchange_id_related_migrations_test.lua

valid_signer_test:
	.rocks/bin/luatest -v -c ./app/test/is_valid_signer_test.lua

vaults_migration_test: build
	.rocks/bin/luatest -v -c ./app/test/vaults_migration_test.lua

stake_test:
	.rocks/bin/luatest -v -c ./app/test/stake_test.lua

yield_test:
	.rocks/bin/luatest -v -c ./app/test/yield_test.lua

pm_test:
	.rocks/bin/luatest -v -c ./app/test/private_matching_test.lua

wdm_test:
	.rocks/bin/luatest -v -c ./app/test/wdm_test.lua
