test-all:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -v

test-l1-deposit:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestL1Deposit" -v

test-withdraw:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestWithdrawOnly" -v