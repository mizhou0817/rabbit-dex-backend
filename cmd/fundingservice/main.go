package main

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/strips-finance/rabbit-dex-backend/funding"
)

const (
	FUNDING_INTERVAL = 60 * time.Second
)

// TODO: split services for each market to different processes
func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetReportCaller(true)

	go func() {
		funding_service, err := funding.NewFundingService(FUNDING_INTERVAL)
		if err != nil {
			logrus.Fatalf("error <%s> creating funding service", err.Error())
		}
		cfunc, err := funding_service.Run()
		if err != nil {
			logrus.Fatalf("error <%s> starting funding service", err.Error())
		}
		defer cfunc()

		logrus.Info("..Started funding service")
		select {}
	}()

	select {}
}
