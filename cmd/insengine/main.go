package main

import (
	"github.com/sirupsen/logrus"

	"github.com/strips-finance/rabbit-dex-backend/liqengine"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetReportCaller(true)

	broker, err := model.GetBroker()
	if err != nil {
		logrus.Fatal(err)
	}

	as, err := liqengine.NewTntAssistant(broker, "0", 0)
	if err != nil {
		logrus.Fatal(err)
	}

	insurance_service := liqengine.NewInsuranceService(1, as)

	cancelf := insurance_service.Run()
	defer cancelf()

	select {}
}
