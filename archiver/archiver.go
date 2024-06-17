package archiver

import (
	"context"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/strips-finance/rabbit-dex-backend/pkg/log"

	"github.com/strips-finance/rabbit-dex-backend/model"
)

type Archiver struct {
	cfg *Config
}

func New(cfg *Config) *Archiver {
	return &Archiver{
		cfg: cfg,
	}
}

func (a *Archiver) Run(ctx context.Context, db *pgxpool.Pool) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	broker, err := model.GetBroker()
	if err != nil {
		return errors.Wrap(err, "get broker")
	}

	g := &errgroup.Group{}
	for space, spaceCfg := range a.cfg.Service.TarantoolSpaces {
		randomStartInterval := time.Duration(rand.Float64() * float64(time.Second))
		for _, instance := range spaceCfg.Instances {
			space, spaceCfg, instance, mode := space, spaceCfg, instance, spaceCfg.Mode // change loop scope
			interval := time.Duration(spaceCfg.SyncInterval) * time.Second
			logrus.Infof("launching space=%s, instance=%s, mode=%s, interval=%v", space, instance, mode, interval)

			//TODO: refactoring is needed
			makeArchiveFn := func() (func(context.Context) (int, error), error) {
				switch spaceCfg.Mode {
				case ArchiveLiveDataMode:
					return func(ctx context.Context) (int, error) {
						return archiveUpdatedSpace(
							ctx,
							newIncrementalBatcher(broker, space, instance, spaceCfg.BatchSize),
							newTsDB(db, instance, spaceCfg.TimescaledbTable, newLiveSqlBuilder(spaceCfg.UniqueId)),
						)
					}, nil
				case ArchiveLiveRawDataMode:
					return func(ctx context.Context) (int, error) {
						return archiveUpdatedSpaceRaw(
							ctx,
							newIncrementalBatcher(broker, space, instance, spaceCfg.BatchSize),
							newTsDB(db, instance, spaceCfg.TimescaledbTable, newLiveSqlBuilder(spaceCfg.UniqueId)),
						)
					}, nil
				case ArchiveSnapshotDataMode:
					return func(ctx context.Context) (int, error) {
						return archiveUpdatedSpace(
							ctx,
							newIncrementalBatcher(broker, space, instance, spaceCfg.BatchSize),
							newTsDB(db, instance, spaceCfg.TimescaledbTable, newSnapshotSqlBuilder()),
						)
					}, nil
				case ArchiveFullSnapshotDataMode:
					return func(ctx context.Context) (int, error) {
						return archiveFullSpace(
							ctx,
							newSnapshotBatcher(broker, space, instance, spaceCfg.BatchSize),
							newTsDB(db, instance, spaceCfg.TimescaledbTable, newFullSnapshotSqlBuilder()),
						)
					}, nil
				}
				return nil, ErrUnknownArchiveMode
			}

			g.Go(func() error {
				startTicker := time.NewTicker(randomStartInterval)
				select {
				case <-ctx.Done():
					startTicker.Stop()
					return nil
				case <-startTicker.C:
					startTicker.Stop()
				}

				archiveSpace, err := makeArchiveFn()
				if err != nil {
					return errors.Wrapf(err, "make archive func for space=%s instance=%s", space, instance)
				}
				ticker := time.NewTicker(interval)
				defer ticker.Stop()

				for {
					logrus.Infof("starting archive space=%s, instance=%s, mode=%s", space, instance, mode)
					n, err := archiveSpace(ctx)
					if err != nil {
						select {
						case <-ctx.Done():
							return nil
						default:
							cancel()
							logrus.WithField(log.AlertTag, log.AlertMid).Errorf("failed to archive space=%s, instance=%s: %v", space, instance, err)
							return err
						}
					}
					logrus.Infof("archived space=%s, instance=%s, rows_count=%d, mode=%s", space, instance, n, mode)
					select {
					case <-ctx.Done():
						return nil
					case <-ticker.C:
						ticker.Reset(time.Duration(spaceCfg.SyncInterval) * time.Second)
					}
				}
			})
		}
	}
	return g.Wait()
}

func archiveUpdatedSpace(ctx context.Context, b *incrementalBatcher, db *tsdb) (int, error) {
	tm0 := time.Now()
	archiveId, err := db.getLastArchiveId(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "get last archive-id")
	}
	logrus.Infof("last archive-id: %d, %s, elapsed=%v", archiveId, db, time.Now().Sub(tm0))

	tm0 = time.Now()
	batch, err := b.getNextBatch(ctx, archiveId)
	if err != nil {
		return 0, errors.Wrapf(err, "get next batch: %d", archiveId)
	}
	logrus.Infof("get next batch: %s, rows_count=%d, elapsed=%v", b, batch.size(), time.Now().Sub(tm0))

	tm0 = time.Now()
	if err := db.sync(ctx, &batch); err != nil {
		return 0, errors.Wrap(err, "sync tsdb")
	}
	logrus.Infof("db sync: %s, elapsed=%v", b, time.Now().Sub(tm0))

	return batch.size(), nil
}

func archiveUpdatedSpaceRaw(ctx context.Context, b *incrementalBatcher, db *tsdb) (int, error) {
	tm0 := time.Now()
	archiveId, err := db.getLastArchiveIdRaw(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "get last archive-id")
	}
	logrus.Infof("last archive-id: %d, %s, elapsed=%v", archiveId, db, time.Now().Sub(tm0))

	tm0 = time.Now()
	batch, err := b.getNextBatch(ctx, archiveId)
	if err != nil {
		return 0, errors.Wrapf(err, "get next batch: %d", archiveId)
	}
	logrus.Infof("get next batch: %s, rows_count=%d, elapsed=%v", b, batch.size(), time.Now().Sub(tm0))

	tm0 = time.Now()
	if err := db.sync(ctx, &batch); err != nil {
		return 0, errors.Wrap(err, "sync tsdb")
	}
	logrus.Infof("db sync: %s, elapsed=%v", b, time.Now().Sub(tm0))

	return batch.size(), nil
}

func archiveFullSpace(ctx context.Context, b *snapshotBatcher, db *tsdb) (int, error) {
	tm0 := time.Now()
	batch, err := b.getNextBatch(ctx)
	if err != nil {
		return 0, errors.Wrapf(err, "get next batch")
	}
	logrus.Infof("get next batch: %s, rows_count=%d, elapsed=%v", b, batch.size(), time.Now().Sub(tm0))

	tm0 = time.Now()
	if err := db.sync(ctx, &batch); err != nil {
		return 0, errors.Wrap(err, "sync tsdb")
	}
	logrus.Infof("db sync: %s, elapsed=%v", b, time.Now().Sub(tm0))

	return batch.size(), nil
}

type ArchiveMode string

const (
	ArchiveLiveDataMode         = "live"
	ArchiveLiveRawDataMode      = "live-raw"
	ArchiveSnapshotDataMode     = "snapshot"
	ArchiveFullSnapshotDataMode = "full-snapshot"
)

func (m *ArchiveMode) UnmarshalYAML(unmarshal func(any) error) error {

	var mode string
	if err := unmarshal(&mode); err != nil {
		return err
	}

	switch mode {
	case ArchiveLiveDataMode, ArchiveLiveRawDataMode, ArchiveSnapshotDataMode, ArchiveFullSnapshotDataMode:
	default:
		return errors.Wrapf(ErrUnknownArchiveMode, "unmarshal %s", mode)
	}

	*m = ArchiveMode(mode)
	return nil
}

var (
	ErrUnknownArchiveMode = errors.New("unknown archive mode")
)
