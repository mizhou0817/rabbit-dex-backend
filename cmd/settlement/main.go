package main

import (
	"github.com/sirupsen/logrus"

	"github.com/strips-finance/rabbit-dex-backend/settlement"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetReportCaller(true)

	config, err := settlement.ReadConfig()
	if err != nil {
		logrus.Fatal(err)
	}

	settlementService, err := settlement.NewSettlementService(
		config.Handlers,
	)

	if err != nil {
		logrus.Fatal(err)
	}

	// No gracefull shutdown
	cancelf, err := settlementService.Run()
	defer cancelf()
	if err != nil {
		logrus.Fatal(err)
	}

	select {}
}
