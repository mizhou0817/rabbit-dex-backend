package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/strips-finance/rabbit-dex-backend/api"
	"github.com/strips-finance/rabbit-dex-backend/migrations"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetReportCaller(true)

	go func() {
		logrus.Info(http.ListenAndServe("localhost:6060", nil))
	}()

	logrus.Info("Starting REST-API Service")

	cfg, err := api.ReadConfig()
	if err != nil {
		logrus.Panic(err)
	}

	// remove query string, seems the migrator doesn't like it, it uses another connection under the hood?
	if !strings.EqualFold(cfg.Service.EnvMode, "dev") {
		err = migrations.ApplyMigrations(cfg.Service.MigrationsTimescaledbConnectionURI, "analytics", "analytics_db_version")
		if err != nil {
			logrus.Panic("Failed to apply migrations: ", err)
		}
	}
	// in order to be able to test api-client locally Add fake env variables
	if strings.EqualFold(cfg.Service.EnvMode, "dev") {
		logrus.Warnf("launching in dev mode")
		os.Setenv("FAKE", "0xea2dc6f116a4c3d6a15f06b4e8ad582a07c3dd9c")
		os.Setenv("MOCK", "mocked")
	}

	addr := fmt.Sprintf("%s:%d", cfg.Service.Host, cfg.Service.Port)

	r := api.Router()
	if err := r.Run(addr); err != nil {
		logrus.Error(err)
	}
}
