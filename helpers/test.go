package helpers

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/FZambia/tarantool"
	"github.com/sirupsen/logrus"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

type space struct {
	Id         uint        `msgpack:"id"`
	Owner      uint        `msgpack:"owner"`
	Name       string      `msgpack:"name"`
	Engine     string      `msgpack:"engine"`
	FieldCount uint        `msgpack:"field_count"`
	Flags      interface{} `msgpack:"flags"`
	Format     interface{} `msgpack:"format"`
}

type sequence struct {
	A    interface{} `msgpack:"a"`
	B    interface{} `msgpack:"b"`
	Name string      `msgpack:"name"`
	C    interface{} `msgpack:"c"`
	D    interface{} `msgpack:"d"`
	E    interface{} `msgpack:"e"`
	F    interface{} `msgpack:"f"`
	G    interface{} `msgpack:"g"`
	H    interface{} `msgpack:"h"`
}

type TarantoolInstance struct {
	broker *model.Broker
}

func NewInstance(broker *model.Broker) *TarantoolInstance {
	return &TarantoolInstance{
		broker: broker,
	}
}

func (ti *TarantoolInstance) Reset(skipSpaces, skipInstances []string) error {
	logrus.Info("Removing spaces on all instances:")
	for title, connection := range ti.broker.Pool {
		logrus.Infof("...cleaning instance = %s", title)

		if slices.Contains(skipInstances, title) {
			logrus.Warnf(".. skip cleaning instance = %s", title)
			continue
		}

		command := "box.space._space:select"
		var spaces [][]space
		err := connection.ExecTypedContext(context.Background(), tarantool.Call(command, []interface{}{}), &spaces)
		if err != nil {
			logrus.Error(err)
			return err
		}

		for _, space := range spaces[0] {
			if strings.HasPrefix(space.Name, "_") {
				continue
			}

			if slices.Contains(skipSpaces, space.Name) {
				logrus.Warnf(".. skip cleaning space = %s", space.Name)
				continue
			}

			command = fmt.Sprintf("box.space.%s:truncate", space.Name)
			_, err := connection.Exec(tarantool.Call(command, []interface{}{}))
			if err != nil {
				logrus.Error(err)
				return err
			}
		}

		command = "box.space._sequence:select"
		var sequences [][]sequence
		err = connection.ExecTypedContext(context.Background(), tarantool.Call(command, []interface{}{}), &sequences)
		if err != nil {
			logrus.Error(err)
		}

		for _, seq := range sequences[0] {
			if !strings.HasPrefix(seq.Name, "_") {

				if slices.Contains(skipSpaces, seq.Name) {
					logrus.Warnf(".. skip cleaning sequence = %s", seq.Name)
					continue
				}

				command = fmt.Sprintf("box.sequence.%s:reset", seq.Name)
				_, err := connection.Exec(tarantool.Call(command, []interface{}{}))
				if err != nil {
					logrus.Error(err)
					return err
				}
			}
		}
	}

	logrus.Info("TarantoolInstance RESET success")
	return nil
}
