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

.PHONY: gen_mocks
gen_mocks:
	go generate ./...

.PHONY: proto-deps
proto-deps:
	# protoc compiler is also required
	# linux users can set it from packages (eg. apt install protobuf-compiler)
	# mac users from homebrew (brew install protobuf)
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.4.0

.PHONY: proto-gen
proto-gen: proto-deps
	protoc --go_out=paths=source_relative:pkg/centrifugo --go-grpc_out=paths=source_relative:pkg/centrifugo --proto_path=proto/centrifugo api.proto

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
