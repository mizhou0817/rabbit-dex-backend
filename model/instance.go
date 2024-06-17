package model

import (
	"errors"

	"golang.org/x/exp/slices"
)

type InstanceGetter struct {
	cfg *BrokerConfig
}

func GetInstance() *InstanceGetter {
	cfg, err := ReadClusterConfig()
	if err != nil {
		panic(err)
	}

	return &InstanceGetter{cfg: cfg}
}

func (ig InstanceGetter) getFirst(
	f func(config *InstanceConfig) bool,
) (*InstanceConfig, error) {
	for _, instance := range ig.cfg.Instances {
		if f(&instance) {
			return &instance, nil
		}
	}

	return nil, errors.New("NOT_FOUND")
}

func (ig InstanceGetter) getList(
	f func(config *InstanceConfig) bool,
) ([]*InstanceConfig, error) {
	result := make([]*InstanceConfig, 0)

	for _, instance := range ig.cfg.Instances {
		if f(&instance) {
			result = append(result, &instance)
		}
	}

	if len(result) == 0 {
		return nil, errors.New("INSTANCE_NOT_FOUND")
	}

	return result, nil
}

func (ig *InstanceGetter) ByMarketID(marketID string) (*InstanceConfig, error) {
	return ig.getFirst(func(i *InstanceConfig) bool {
		return slices.Contains(i.Tags, marketID) || i.Title == marketID
	})
}

func (ig *InstanceGetter) ByTitle(title string) (*InstanceConfig, error) {
	return ig.getFirst(func(i *InstanceConfig) bool {
		return i.Title == title
	})
}
