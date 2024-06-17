test: test_kms test_signer

test_kms:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestCreateKey" -v

test_signer:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestSigner" -v

test_show_bfx_address:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestGetBfxAddresses" -v

test_show_rbx_address:
	go test . -count=1 -tags=go_tarantool_msgpack_v5 -run "TestGetRbxAddresses" -v
