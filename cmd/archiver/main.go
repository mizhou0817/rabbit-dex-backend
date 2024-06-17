package main

import (
	"context"
	"syscall"

	"github.com/Code-Hex/sigctx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"github.com/strips-finance/rabbit-dex-backend/archiver"
	"github.com/strips-finance/rabbit-dex-backend/migrations"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetReportCaller(true)

	logrus.Info("Starting Archiver Service")

	ctx := sigctx.WithCancelSignals(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGHUP,
	)

	cfg, err := archiver.ReadConfig()
	if err != nil {
		logrus.Panic(err)
	}

	err = migrations.ApplyMigrations(cfg.Service.MigrationsTimescaledbConnectionURI, "archiver", "archiver_db_version")
	if err != nil {
		logrus.Panic("Failed to apply migrations: ", err)
	}

	dbpool, err := pgxpool.New(ctx, cfg.Service.TimescaledbConnectionURI)
	if err != nil {
		logrus.Panic("Unable to connect to database: ", err)
	}
	defer dbpool.Close()

	err = archiver.New(cfg).Run(ctx, dbpool)
	if err != nil {
		logrus.Panic("Failed to run service: ", err)
	}
}
