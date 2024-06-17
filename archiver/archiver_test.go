package archiver

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/FZambia/tarantool"
	"github.com/stretchr/testify/assert"

	"github.com/strips-finance/rabbit-dex-backend/model"
)

const (
	instanceName = "BTC-USD"
)

func setUp() {
}

func tearDown() {
}

func TestArchiverFullSnapshot100k(t *testing.T) {
	// this ugly test shows we can get snapshot with multiple selects with positions updating in the middle of "snapshoting"
	ctx := context.Background()

	broker, err := model.GetBroker()
	assert.NoError(t, err)

	conn := broker.Pool[instanceName]
	assert.NotNil(t, conn)

	toSet := func(a []any) (m map[string]struct{}, key string) {
		m = make(map[string]struct{}, len(a))
		for _, v := range a {
			key = v.([]any)[0].(string)
			m[key] = struct{}{}
		}
		return m, key
	}

	cmd := `
	local pos = require('app.engine.position')
	local cfg = require('app.config')
	require('app.config.constants')

	for i=1,30000 do
		local _, err = pos.create('BTC-USD', i, ONE, cfg.params.LONG, ONE)
		if err ~= nil then
			error(err)
		end
	end
	for i=70001,100000 do
		local _, err = pos.create('BTC-USD', i, ONE, cfg.params.LONG, ONE)
		if err ~= nil then
			error(err)
		end
	end
	`
	_, err = conn.ExecContext(context.Background(), tarantool.Eval(cmd, []any{}))
	assert.NoError(t, err)

	const limit = 5000
	var data []any

	// read first piece 10k with two calls
	res, _, err := broker.SelectUntyped(ctx, instanceName, "position", "primary", nil, model.IterAll, limit)
	assert.NoError(t, err)
	data = append(data, res...)
	assert.Equal(t, limit, len(res))

	key := res[len(res)-1].([]any)[0] // last primary id
	res, _, err = broker.SelectUntyped(ctx, instanceName, "position", "primary", key, model.IterGt, limit)
	assert.NoError(t, err)
	assert.Equal(t, limit, len(res))

	data = append(data, res...)
	assert.Equal(t, 2*limit, len(data))

	cmd = `
	local pos = require('app.engine.position')
	local cfg = require('app.config')
	require('app.config.constants')

	for i=30001,70000 do
		local _, err = pos.create('BTC-USD', i, ONE, cfg.params.LONG, ONE)
		if err ~= nil then
			error(err)
		end
	end
	`
	_, err = conn.ExecContext(context.Background(), tarantool.Eval(cmd, []any{}))
	assert.NoError(t, err)

	const limit2 = 90000
	key = res[len(res)-1].([]any)[0] // last primary id

	// read the rest with one call
	res, _, err = broker.SelectUntyped(ctx, instanceName, "position", "primary", key, model.IterGt, 10*limit2)
	assert.NoError(t, err)
	assert.Equal(t, limit2, len(res))

	data = append(data, res...)

	// check we have all positions in snapshot
	set, _ := toSet(data)
	for i := range data {
		posKey := fmt.Sprintf("pos-BTC-USD-tr-%d", i+1)
		_, ok := set[posKey]
		assert.True(t, ok, posKey)
	}
}

func TestMain(m *testing.M) {
	setUp()
	ret := m.Run()
	tearDown()
	os.Exit(ret)
}
