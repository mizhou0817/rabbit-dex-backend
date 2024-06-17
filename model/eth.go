package model

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
)

type EthHelper struct {
	client *ethclient.Client
	providerUrl string
	initialRedialDelay time.Duration
}

func NewEthHelper(providerUrl string, initialRedialDelay time.Duration) *EthHelper {
	if initialRedialDelay <= 0 {
		initialRedialDelay = 2 * time.Second
	}
	return &EthHelper{
		providerUrl: providerUrl,
		initialRedialDelay: initialRedialDelay,
	}
}

func (eh *EthHelper) GetCurrentBlockNumber(ctx context.Context) (*big.Int, error) {
	err := eh.ensureConnected()
	if err != nil {
		return nil, err
	}
	header, err := eh.client.HeaderByNumber(ctx, nil)
	if err != nil {
		eh.client = nil
		logrus.Errorf("Error retrieving current block number: %s", err.Error())
		return nil, fmt.Errorf("error retrieving current block number: %s", err.Error())
	}
	return header.Number, nil
}

func (eh *EthHelper) dial() error {
	sleepFor := eh.initialRedialDelay
	var err error
	for i := 0; i < 5; i++ {
		eh.client, err = ethclient.Dial(eh.providerUrl)
		if err == nil {
			break
		}
		eh.client = nil
		time.Sleep(sleepFor)
		sleepFor = sleepFor * 2
	}
	if err != nil {
		return fmt.Errorf("error dialing eth client: %s", err.Error())
	}
	return nil
}

func (eh *EthHelper) ensureConnected() error {
	var err error
	if eh.client == nil {
		err = eh.dial()
		if err != nil {
			logrus.Errorf(
				"REST API service, error connecting to go-ethereum: %s",
				err.Error(),
			)
		}
	}
	return err
}