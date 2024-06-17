package vaultdata

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/strips-finance/rabbit-dex-backend/migrations"
)

func ApplyTestMigrations(t *testing.T, migrationConnStr string) {
	r := require.New(t)

	err := migrations.ApplyMigrations(migrationConnStr, "archiver", "archiver_db_version")
	r.NoError(err)

	err = migrations.ApplyMigrations(migrationConnStr, "referrals", "referrals_db_version")
	r.NoError(err)

	err = migrations.ApplyMigrations(migrationConnStr, "analytics", "analytics_db_version")
	r.NoError(err)

	err = migrations.ApplyMigrations(migrationConnStr, "dashboards", "dashboards_db_version")
	r.NoError(err)
}
