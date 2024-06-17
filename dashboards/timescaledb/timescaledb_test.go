package timescaledb

import (
	"testing"

	"github.com/stretchr/testify/suite"
	common_timescaledb "github.com/strips-finance/rabbit-dex-backend/dashboards/common/timescaledb"
	"github.com/strips-finance/rabbit-dex-backend/dbtestsuite"
)

type dbTestSuite struct {
	dbtestsuite.DBTestSuite
}

func Test_dbTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timescaledb integration test")
	}
	testSuite := new(dbTestSuite)
	suite.Run(t, testSuite)
}

func (s *dbTestSuite) SetupSuite() {
	s.BaseSetupSuite()
	common_timescaledb.ApplyTestMigrations(s.T(), s.MigrationConnectionString())
}

func (s *dbTestSuite) TearDownSuite() {
	s.BaseTearDownSuite()
}

func (s *dbTestSuite) TestMigrations() {
}
