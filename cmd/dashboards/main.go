package main

import (
	"context"
	"github.com/Code-Hex/sigctx"
	"github.com/sirupsen/logrus"
	"github.com/strips-finance/rabbit-dex-backend/dashboards"
	"github.com/strips-finance/rabbit-dex-backend/migrations"
	"syscall"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetReportCaller(true)

	logrus.Info("Starting Dashboards DB Job")

	ctx := sigctx.WithCancelSignals(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGHUP,
	)

	cfg, err := dashboards.ReadConfig()
	if err != nil {
		logrus.Panic(err)
	}

	err = migrations.ApplyMigrations(cfg.Service.MigrationsTimescaledbConnectionURI, "dashboards", "dashboards_db_version")
	if err != nil {
		logrus.Panic("Failed to apply migrations: ", err)
	}
	logrus.Info("Successfully applied migrations")

	<-ctx.Done()
}
