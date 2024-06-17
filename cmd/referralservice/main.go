package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"github.com/strips-finance/rabbit-dex-backend/migrations"
	"github.com/strips-finance/rabbit-dex-backend/referrals"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetReportCaller(true)

	logrus.Info("Starting Referral Service")

	cfg, err := referrals.ReadConfig()
	if err != nil {
		logrus.Panic(err)
	}

	err = migrations.ApplyMigrations(cfg.Service.MigrationsTimescaledbConnectionURI, "referrals", "referrals_db_version")
	if err != nil {
		logrus.Panic("Failed to apply migrations: ", err)
	}

	dbpool, err := pgxpool.New(context.Background(), cfg.Service.TimescaledbConnectionURI)
	if err != nil {
		logrus.Panic("Unable to connect to database: ", err)
	}
	defer dbpool.Close()

	svc := referrals.New(cfg, dbpool)
	svc.Run()
}
