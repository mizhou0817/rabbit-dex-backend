package model

import (
	"github.com/FZambia/tarantool"
)

type TestBroker struct {
	Conn *tarantool.Connection
}

type TestBrokerConfig struct {
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

func GetTestBroker(cfg TestBrokerConfig) (*TestBroker, error) {
	instance := new(TestBroker)

	//TODO: use config to setup all
	opts := tarantool.Opts{
		User: cfg.User,
	}
	conn, err := tarantool.Connect(cfg.Host, opts)
	if err != nil {
		return nil, err
	}
	instance.Conn = conn

	return instance, nil

}

func (broker *TestBroker) Close() error {
	err := broker.Conn.Close()
	if err != nil {
		return err
	}

	return nil
}
