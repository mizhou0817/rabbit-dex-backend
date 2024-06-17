package archiver

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/strips-finance/rabbit-dex-backend/pkg/log"

	"github.com/strips-finance/rabbit-dex-backend/model"
)

const (
	maxBatcherLimit = 1000000
	batchLimit      = 1000 // recomendation of tarantool team
)

type incrementalBatcher struct {
	broker *model.Broker

	space     string
	instance  string
	batchSize uint64
}

func newIncrementalBatcher(broker *model.Broker, space, instance string, batchSize uint64) *incrementalBatcher {
	if batchSize == 0 || batchSize > batchLimit {
		batchSize = batchLimit
	}
	return &incrementalBatcher{
		broker: broker,

		space:     space,
		instance:  instance,
		batchSize: batchSize,
	}
}

func (b *incrementalBatcher) String() string {
	return fmt.Sprintf("space=%s, instance=%s, batchSize=%d", b.space, b.instance, b.batchSize)
}

func (b *incrementalBatcher) getNextBatch(ctx context.Context, lastArchiveId uint64) (batchResponse, error) {
	res, columns, err := b.broker.SelectUntyped(ctx, b.instance, b.space, "archive_id", lastArchiveId, model.IterGt, uint32(b.batchSize))
	if err != nil {
		return batchResponse{}, errors.Wrap(err, "get next incremental batch")
	}

	tm := unixMicro(time.Now())
	return batchResponse{batch: res, columns: columns, ts: tm}, nil
}

type snapshotBatcher struct {
	broker *model.Broker

	space     string
	instance  string
	batchSize uint64
}

func newSnapshotBatcher(broker *model.Broker, space, instance string, batchSize uint64) *snapshotBatcher {
	if batchSize == 0 || batchSize > batchLimit {
		batchSize = batchLimit
	}
	return &snapshotBatcher{
		broker: broker,

		space:     space,
		instance:  instance,
		batchSize: batchSize,
	}
}

func (b *snapshotBatcher) String() string {
	return fmt.Sprintf("space=%s, instance=%s, batchSize=%d", b.space, b.instance, b.batchSize)
}

func (b *snapshotBatcher) getNextBatch(ctx context.Context) (batchResponse, error) {
	data, columns, err := b.broker.SelectUntyped(ctx, b.instance, b.space, "primary", nil, model.IterAll, uint32(b.batchSize))
	if err != nil {
		return batchResponse{}, errors.Wrap(err, "select untyped")
	}

	tm := unixMicro(time.Now())
	size := uint64(len(data))

	if size < b.batchSize {
		return batchResponse{batch: data, columns: columns, ts: tm}, nil
	}

	for {
		//TODO: do refactoring of ugly code
		key := data[size-1].([]any)[0] // awaiting id is the first column

		res, _, err := b.broker.SelectUntyped(ctx, b.instance, b.space, "primary", key, model.IterGt, uint32(b.batchSize))
		if err != nil {
			return batchResponse{}, errors.Wrap(err, "select untyped")
		}
		data = append(data, res...)
		size = uint64(len(data))

		if uint64(len(res)) < b.batchSize {
			break
		}
		if size > maxBatcherLimit {
			logrus.WithField(log.AlertTag, log.AlertMid).Warn("batcher rows limit reached")
			break
		}
	}

	return batchResponse{batch: data, columns: columns, ts: tm}, nil
}

type batchResponse struct {
	batch   []any
	columns []string
	ts      uint64
}

func (b *batchResponse) getColumns() []string {
	columns := make([]string, len(b.columns))
	copy(columns, b.columns)
	return columns
}

func (b *batchResponse) size() int {
	return len(b.batch)
}

func (b *batchResponse) timestamp() uint64 {
	return b.ts
}

func (b *batchResponse) data() []any {
	return b.batch
}

func (b *batchResponse) getPart(left, right int) *batchResponse {
	if right > len(b.batch) {
		right = len(b.batch)
	}
	if left > right {
		left = right
	}
	data := b.batch[left:right]

	res := batchResponse{
		columns: b.columns,
		batch:   data,
		ts:      b.ts,
	}

	return &res
}
