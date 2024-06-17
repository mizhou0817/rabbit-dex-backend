test-all:
	go test -count=1 -tags=go_tarantool_msgpack_v5

test-custom:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "$(title)" -v

test-withdraw-flow:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestBalanceSuite" -v

test-admin:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestAdminSuite" -v

test-airdrop:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestAirdropSuite" -v

test-secrets:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestSecretsSuite" -v

test-vault-onboarding:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestOnboardingVaultSuite" -v

test-storage:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestStorageSuite" -v

test-mm-onboarding:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestOnboardingMarketMakerSuite" -v

test-frontend-onboarding:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestOnboardingFrontendSuite" -v

test-profile-data:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestProfileData" -v
