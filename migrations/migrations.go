package migrations

import (
	"database/sql"
	"embed"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
)

//go:embed archiver/*.sql
//go:embed analytics/*.sql
//go:embed referrals/*.sql
//go:embed dashboards/*.sql
var embedMigrations embed.FS

func ApplyMigrations(dsn string, path, version string) error {
	goose.SetBaseFS(embedMigrations)
	goose.SetTableName(version)

	err := goose.SetDialect("postgres")
	if err != nil {
		return errors.Wrap(err, "set migrations dialect")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return errors.Wrap(err, "set migrations open migrations connection")
	}
	defer db.Close()

	err = goose.Up(db, path)
	if err != nil {
		return errors.Wrap(err, "apply migration")
	}

	return nil
}
