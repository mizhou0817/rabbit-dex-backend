//go:generate go run github.com/golang/mock/mockgen -destination=$PWD/mock/store.go -package=mock github.com/strips-finance/rabbit-dex-backend/profile/tsdb DB
package tsdb

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// interface hides usage of single DB or Pool
type DB interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}
