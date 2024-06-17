package main

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/strips-finance/rabbit-dex-backend/pricing"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetReportCaller(true)

	pricingService := pricing.NewPricingService()

	config, err := pricing.ReadConfig()
	logrus.Infof("pricing service config=%+v", config)
	if err != nil {
		logrus.Fatalf("Can't read pricing service config err=%s", err.Error())
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pricingService.Run(ctx, config)
	select {}
}
