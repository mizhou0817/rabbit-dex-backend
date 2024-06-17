package main

import (
	"github.com/sirupsen/logrus"

	"github.com/strips-finance/rabbit-dex-backend/slipstopper"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetReportCaller(true)

	cfg, err := slipstopper.ReadConfig()
	if err != nil {
		logrus.Fatalln(err)
	}

	wsClient := slipstopper.NewWSClient(cfg)
	readyChan := make(chan bool)
	wsClient.Run(readyChan)

	<-readyChan
	logrus.Info("We are ready!")

	select {}
}
