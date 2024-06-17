package main

import (
	"github.com/sirupsen/logrus"

	"github.com/strips-finance/rabbit-dex-backend/websocket"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetReportCaller(true)

	logrus.Info("Starting WebSocket Service")

	if err := websocket.Run(); err != nil {
		logrus.Panic(err)
	}
}
