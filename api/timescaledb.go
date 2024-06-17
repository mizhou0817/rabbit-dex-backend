package api

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"sync"
)

var mutexDb sync.Mutex
var timescaledbInstance *pgxpool.Pool = nil

func GetTimescaleDbPool(uri string) (*pgxpool.Pool, error) {
	mutexDb.Lock()
	defer mutexDb.Unlock()

	if timescaledbInstance != nil {
		return timescaledbInstance, nil
	}

	pool, err := pgxpool.New(context.Background(), uri)
	if err != nil {
		return nil, err
	}

	timescaledbInstance = pool
	return timescaledbInstance, nil
}
