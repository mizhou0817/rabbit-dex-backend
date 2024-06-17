.PHONY: test
test:
	@(cd archiver && make test_archiver)
	@(cd model/tnt && make test)

.PHONY: test-ci-cd
test-ci-cd:
	@(cd dashboards && make test)
	@(cd vaultdata && make test)
	@(cd portfolio && make test)
	@(cd volume && make test)
	@(cd gameassets && make test)

gen_mocks:
	go generate ./...

run_api:
	go run -tags=go_tarantool_msgpack_v5 cmd/api/main.go

run_websocket:
	go run -tags=go_tarantool_msgpack_v5 cmd/websocket/main.go

run_centrifugo:
	centrifugo -c docker/centrifugo/config.json

run_liqengine:
	go run -tags=go_tarantool_msgpack_v5 cmd/liqengine/main.go

run_insengine:
	go run -tags=go_tarantool_msgpack_v5 cmd/insengine/main.go

run_settlement:
	go run -tags=go_tarantool_msgpack_v5 cmd/settlement/main.go

run_pricing:
	go run -tags=go_tarantool_msgpack_v5 cmd/pricingservice/main.go

run_funding:
	go run -tags=go_tarantool_msgpack_v5 cmd/fundingservice/main.go

run_profile_periodics:
	go run -tags=go_tarantool_msgpack_v5 cmd/profile/periodics/main.go
