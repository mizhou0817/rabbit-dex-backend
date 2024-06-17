package dbtestsuite

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	migrationConnectionTemplate = "postgres://rabbitx:rabbitx@%s/rabbitx"
	connectionTemplate          = migrationConnectionTemplate + "?pool_max_conns=20"
)

type DBTestSuite struct {
	suite.Suite
	container testcontainers.Container
	db        *pgxpool.Pool
}

func (s *DBTestSuite) GetDB() *pgxpool.Pool {
	return s.db
}

func (s *DBTestSuite) BaseSetupSuite() {
	r := require.New(s.T())

	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "timescale/timescaledb:latest-pg14",
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		Env: map[string]string{
			"POSTGRES_PASSWORD": "rabbitx",
			"POSTGRES_DB":       "rabbitx",
			"POSTGRES_USER":     "rabbitx",
		},

		// Speedup DB with the following settings
		Mounts: nil,
		Tmpfs:  map[string]string{"/var/lib/postgresql/data": "rw"},
		Cmd: []string{
			"postgres",
			"-h", "0",
			"-c", "log_min_duration_statement=-1",
			"-c", "fsync=off",
			"-c", "synchronous_commit=off",
			"-c", "full_page_writes=off",
			"-c", "shared_buffers=256MB",
			"-c", "archive_mode=off",
		},
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	r.NoError(err)
	s.container = container

	endpoint, err := s.container.Endpoint(ctx, "")
	r.NoError(err)

	connStr := fmt.Sprintf(connectionTemplate, endpoint)
	s.db, err = pgxpool.New(ctx, connStr)
	r.NoError(err)
}

func (s *DBTestSuite) MigrationConnectionString() string {
	r := require.New(s.T())
	ctx := context.Background()

	endpoint, err := s.container.Endpoint(ctx, "")
	r.NoError(err)

	return fmt.Sprintf(migrationConnectionTemplate, endpoint)
}

func (s *DBTestSuite) BaseTearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
	if !s.container.IsRunning() {
		return
	}
	r := require.New(s.T())
	ctx := context.Background()
	err := s.container.Terminate(ctx)
	r.NoError(err)
}

func (s *DBTestSuite) DeleteTables(tables []string) {
	for _, tableName := range tables {
		s.DeleteTable(tableName)
	}
}

func (s *DBTestSuite) DeleteTable(tableName string) {
	s.Execute("DELETE FROM " + tableName)
}

func (s *DBTestSuite) Execute(query string, args ...any) {
	ctx := context.Background()
	_, err := s.db.Exec(ctx, query, args...)
	require.NoError(s.T(), err)
}

func (s *DBTestSuite) Count(tableName string, args ...any) int {
	return s.CountWhere(tableName, "", args...)
}

func (s *DBTestSuite) CountWhere(tableName string, whereClause string, args ...any) int {
	query := "SELECT count(*) FROM " + tableName
	if whereClause != "" {
		query += " WHERE " + whereClause
	}
	ctx := context.Background()
	row := s.db.QueryRow(ctx, query, args...)
	var result int
	err := row.Scan(&result)
	require.NoError(s.T(), err)
	return result
}

func (s *DBTestSuite) RefreshMaterializedView(viewName string) {
	s.Execute("REFRESH MATERIALIZED VIEW " + viewName)
}
