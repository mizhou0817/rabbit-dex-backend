package archiver

import (
	"context"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type tsdb struct {
	db *pgxpool.Pool

	instance     string
	table        string
	batchBuilder tsdbSqlBatchBuilder
	sqlBuilder   sq.StatementBuilderType
}

func newTsDB(db *pgxpool.Pool, instance, table string, builder tsdbSqlBatchBuilder) *tsdb {
	return &tsdb{
		db:           db,
		instance:     instance,
		table:        table,
		batchBuilder: builder,
		sqlBuilder:   sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (db *tsdb) String() string {
	return fmt.Sprintf("instance=%s, table=%s", db.instance, db.table)
}

func (db *tsdb) getLastArchiveId(ctx context.Context) (uint64, error) {
	sql, args, err := db.sqlBuilder.
		Select().
		Column(
			sq.Expr("last_shard_archive_id(?,?,?)", "public", db.table, getShardId(db.instance)),
		).
		ToSql()
	if err != nil {
		return 0, errors.Wrap(err, "build query")
	}

	var max uint64
	if err := db.db.QueryRow(ctx, sql, args...).Scan(&max); err != nil {
		return 0, errors.Wrap(err, "scan query")
	}

	if overrider, ok := db.batchBuilder.(LastArchiveIdOverrider); ok {
		max = overrider.OverrideLastArchiveId(max)
	}

	return max, nil
}

func (db *tsdb) getLastArchiveIdRaw(ctx context.Context) (uint64, error) {
	sql, args, err := db.sqlBuilder.
		Select().
		Column("COALESCE(MAX(archive_id), 0)").
		From(db.table).
		Where(sq.Eq{"shard_id": getShardId(db.instance)}).
		ToSql()
	if err != nil {
		return 0, errors.Wrap(err, "build query")
	}

	var max uint64
	if err := db.db.QueryRow(ctx, sql, args...).Scan(&max); err != nil {
		return 0, errors.Wrap(err, "scan query")
	}

	return max, nil
}

func (db *tsdb) sync(ctx context.Context, res *batchResponse) (err error) {
	conn, err := db.db.Acquire(ctx)
	if err != nil {
		return errors.Wrap(err, "acquire db conn")
	}
	defer conn.Release()

	conn.Conn().TypeMap().RegisterType(&pgtype.Type{
		Name:  "jsonb",
		OID:   pgtype.JSONBOID,
		Codec: &anyMapJsonbCodec{},
	})

	const size = 1000 // TODO: do it configurable

	tx, err := conn.Begin(ctx)
	if err != nil {
		return errors.Wrap(err, "begin tx")
	}
	defer func() {
		if err != nil {
			err = multierr.Append(err, tx.Rollback(ctx))
		}
	}()

	var (
		sql  string
		args []any
	)
	for i, l := 0, len(res.data()); i < l; i += size {
		sql, args, err = db.batchBuilder.Build(db.table, res.getPart(i, i+size))
		if err != nil {
			return errors.Wrap(err, "build query")
		}

		if _, err := tx.Exec(ctx, sql, args...); err != nil {
			return errors.Wrap(err, "exec query")
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return errors.Wrap(err, "commit tx")
	}

	return nil
}

type tsdbSqlBatchBuilder interface {
	Build(table string, res *batchResponse) (sql string, args []any, err error)
}

type tsdbLiveSqlBuilder struct {
	uniqueId   []string
	sqlBuilder sq.StatementBuilderType
}

func newLiveSqlBuilder(uniqueId []string) *tsdbLiveSqlBuilder {
	return &tsdbLiveSqlBuilder{
		uniqueId:   uniqueId,
		sqlBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (b *tsdbLiveSqlBuilder) Build(table string, res *batchResponse) (string, []any, error) {
	if res.size() == 0 {
		return "", nil, nil
	}

	columns := res.getColumns()
	columns = append(columns, "archive_timestamp")

	var updateFieldsClause []string
	for _, column := range columns {
		updateFieldsClause = append(updateFieldsClause, column+"=EXCLUDED."+column)
	}

	builder := b.sqlBuilder.
		Insert(table).
		Columns(strings.Join(columns, ",")).
		Suffix("ON CONFLICT (" + strings.Join(b.uniqueId, ",") + ") DO UPDATE SET " + strings.Join(updateFieldsClause, ","))

	for _, row := range res.data() {
		values := row.([]any)
		values = append(values, res.timestamp())
		builder = builder.Values(values...)
	}

	return builder.ToSql()
}

type tsdbSnapshotSqlBuilder struct {
	sqlBuilder sq.StatementBuilderType
}

func newSnapshotSqlBuilder() *tsdbSnapshotSqlBuilder {
	return &tsdbSnapshotSqlBuilder{
		sqlBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (b *tsdbSnapshotSqlBuilder) Build(table string, res *batchResponse) (string, []any, error) {
	if res.size() == 0 {
		return "", nil, nil
	}

	columns := res.getColumns()
	columns = append(columns, "archive_timestamp")

	builder := b.sqlBuilder.
		Insert(table).
		Columns(strings.Join(columns, ","))

	for _, row := range res.data() {
		values := row.([]any)
		values = append(values, res.timestamp())
		builder = builder.Values(values...)
	}

	return builder.ToSql()
}

type LastArchiveIdOverrider interface {
	OverrideLastArchiveId(oldValue uint64) uint64
}

type tsdbFullSnapshotSqlBuilder struct {
	*tsdbSnapshotSqlBuilder
}

func newFullSnapshotSqlBuilder() *tsdbFullSnapshotSqlBuilder {
	return &tsdbFullSnapshotSqlBuilder{
		tsdbSnapshotSqlBuilder: newSnapshotSqlBuilder(),
	}
}

func (b *tsdbFullSnapshotSqlBuilder) OverrideLastArchiveId(oldValue uint64) uint64 {
	return 0
}
