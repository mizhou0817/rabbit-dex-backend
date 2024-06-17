package model

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/FZambia/tarantool"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/strips-finance/rabbit-dex-backend/pkg/log"
)

type Iterator uint32

const (
	IterEq            = Iterator(0)
	IterReq           = Iterator(1)
	IterAll           = Iterator(2)
	IterLt            = Iterator(3)
	IterLe            = Iterator(4)
	IterGe            = Iterator(5)
	IterGt            = Iterator(6)
	IterBitsAllSet    = Iterator(7)
	IterBitsAnySet    = Iterator(8)
	IterBitsAllNotSet = Iterator(9)
)

// Singleton, thread-safe broker
// Incapsulate communication with DB and inner services
// Manage connection pool if required
// Implements queue and pub/sub interfaces
type Broker struct {
	Pool map[string]*tarantool.Connection
	cfg  BrokerConfig
}

var reconnectMutex sync.RWMutex
var mutex sync.Mutex
var broker_instance *Broker = nil

const RECONNECT_TRY_COUNT = 32

// TODO: implement context
func GetBroker() (*Broker, error) {
	mutex.Lock()
	defer mutex.Unlock()
	if broker_instance != nil {
		return broker_instance, nil
	}
	broker_instance = new(Broker)

	cfg, err := ReadClusterConfig()
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	logrus.Info("Broker config data:")
	logrus.Info(cfg)

	broker_instance.Pool = make(map[string]*tarantool.Connection)

	//Create connections to instances
	for _, instance := range cfg.Instances {
		addr := fmt.Sprintf("%s:%d", instance.Host, instance.Port)
		logrus.Info(addr)

		// reconnect logic in connect
		tryCount := 0

		for {
			conn, err := tarantool.Connect(addr, tarantool.Opts{
				User:     instance.User,
				Password: instance.Password,
			})

			if err != nil && tryCount > RECONNECT_TRY_COUNT {
				return nil, err
			}

			if !shouldReconnect(err) {
				broker_instance.Pool[instance.Title] = conn
				break
			}

			tryCount += 1

			time.Sleep(time.Second * 5)
		}
	}

	broker_instance.cfg = *cfg

	return broker_instance, nil
}

func (broker *Broker) Close() error {
	mutex.Lock()
	defer mutex.Unlock()
	if broker_instance == nil {
		return errors.New("No broker")
	}

	for _, conn := range broker_instance.Pool {
		err := conn.Close()
		if err != nil {
			return err
		}
	}
	broker_instance = nil
	return nil
}

func (broker *Broker) Execute(
	instance string,
	ctx context.Context,
	fn string,
	params []interface{},
	res interface{},
) error {
	return broker.exec(ctx, instance, func(conn *tarantool.Connection) error {
		err := conn.ExecTypedContext(ctx, tarantool.Call(fn, params), &res)
		if err != nil {
			logrus.Error(err, instance, params)
			debug.PrintStack()
		}
		return err
	})
}

func (broker *Broker) SelectUntyped(
	ctx context.Context,
	instance string,
	space string,
	index any,
	key any,
	iterator Iterator,
	limit uint32,
) ([]any, []string, error) {
	var data []any
	var columns []string

	err := broker.exec(ctx, instance, func(conn *tarantool.Connection) error {
		offset := uint32(0) // see Warning https://www.tarantool.io/en/doc/latest/reference/reference_lua/box_index/select/

		keyArg := []any{}
		if key != nil {
			keyArg = []any{key}
		}
		req := tarantool.Select(space, index, offset, limit, uint32(iterator), keyArg)
		res, err := conn.ExecContext(ctx, req)
		if err != nil {
			return errors.Wrap(err, "tarantool select")
		}
		data = res

		schema, ok := conn.Schema().Spaces[space]
		if !ok {
			return errors.New("space not found")
		}

		for i := uint32(0); i < uint32(len(schema.FieldsByID)); i++ {
			f := schema.FieldsByID[i]
			columns = append(columns, f.Name)
		}
		return nil
	})
	return data, columns, err
}

func (broker *Broker) exec(
	ctx context.Context,
	instance string,
	fn func(*tarantool.Connection) error,
) error {
	var conn *tarantool.Connection
	var err error

	for {
		reconnectMutex.RLock()
		var ok bool
		if conn, ok = broker.Pool[instance]; !ok {
			text := fmt.Sprintf("Instance=%s not found", instance)
			reconnectMutex.RUnlock()
			return errors.New(text)
		}

		if conn == nil {
			reconnectMutex.RUnlock()
			logrus.WithField(log.AlertTag, log.AlertCrit).Error("reconnecting: nil connection found")
			goto reconnect
		}

		// reconnect logic in execution
		err = fn(conn)
		reconnectMutex.RUnlock()

		if !shouldReconnect(err) {
			return err
		}

		//TODO: this is not thread safe. Need to fix that
	reconnect:
		i, err := GetInstance().ByTitle(instance)
		if err != nil {
			return err
		}

		new_conn, err := NewConnection(i)
		if err != nil {
			logrus.WithField(log.AlertTag, log.AlertCrit).Error(err)
			time.Sleep(time.Second * 5)
			goto reconnect
		}
		if new_conn == nil {
			logrus.WithField(log.AlertTag, log.AlertCrit).Error("new connection is nil")
			time.Sleep(time.Second * 5)
			goto reconnect
		}
		reconnectMutex.Lock()
		broker.Pool[instance] = new_conn
		reconnectMutex.Unlock()
	}
}

func NewConnection(instance *InstanceConfig) (*tarantool.Connection, error) {
	addr := fmt.Sprintf("%s:%d", instance.Host, instance.Port)
	opts := tarantool.Opts{
		User:     instance.User,
		Password: instance.Password,
	}
	conn, err := tarantool.Connect(addr, opts)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func shouldReconnect(err error) bool {
	targetErrors := []string{
		"read tcp",
		"network is unreachable",
		"closed connection",
	}

	if err == nil {
		return false
	}

	for _, e := range targetErrors {
		if strings.Contains(err.Error(), e) {
			logrus.WithField("reconnect", true).Error(err)

			return true
		}
	}

	logrus.WithField("reconnect", false).Error(err)

	return false
}
