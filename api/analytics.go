package api

import (
	"context"
	"sync"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/emirpasic/gods/queues/arrayqueue"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

const QueueMaxLimit = 2000
const QueueBatchSize = 400

var collectorInstance *AnalyticsCollector = nil
var collectorMutex sync.Mutex

type AnalyticEvent struct {
	ProfileId          uint   `json:"profile_id,omitempty"`
	ClientHeader       string `json:"client_header,omitempty"`
	ClientIPAddress    string `json:"client_ip_address,omitempty"`
	URLPath            string `json:"url_path,omitempty"`
	HTTPMethod         string `json:"http_method,omitempty"`
	RequestBody        string `json:"request_body,omitempty"`
	ResponseStatusCode int    `json:"response_status_code,omitempty"`
	ResponseBody       string `json:"response_body,omitempty"`
}

type AnalyticsCollector struct {
	queue *arrayqueue.Queue
	mu    sync.Mutex
	db    *pgxpool.Pool
}

func GetAnalyticsCollector(db *pgxpool.Pool) (*AnalyticsCollector, error) {
	collectorMutex.Lock()
	defer collectorMutex.Unlock()

	if collectorInstance != nil {
		return collectorInstance, nil
	}

	collectorInstance = New(db)
	go collectorInstance.Run()

	return collectorInstance, nil
}

func New(db *pgxpool.Pool) *AnalyticsCollector {
	return &AnalyticsCollector{
		queue: arrayqueue.New(),
		db:    db,
	}
}

func (c *AnalyticsCollector) SyncBatch() error {
	sqlBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	insertBuilder := sqlBuilder.
		Insert("analytic_event").
		Columns("event")

	// get a batch
	var vals []any
	for i := 0; i < QueueBatchSize; i++ {
		c.mu.Lock()
		el, success := c.queue.Dequeue()
		c.mu.Unlock()
		if success == false {
			// nothing left in the queue.
			break
		}

		vals = []any{el}
		insertBuilder = insertBuilder.Values(vals...)
	}

	if len(vals) > 0 {
		sql, args, err := insertBuilder.ToSql()
		if err != nil {
			return err
		}

		_, err = c.db.Exec(context.TODO(), sql, args...)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *AnalyticsCollector) Run() {
	logrus.Info("--- Running AnalyticsCollector ---")
	for {
		err := c.SyncBatch()
		if err != nil {
			logrus.Warn("error while syncing analytic queue batch: ", err)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func (c *AnalyticsCollector) Push(event any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	size := c.queue.Size()
	if size >= QueueMaxLimit {
		logrus.Warn("queue size exceeded: ", size, QueueMaxLimit)
		return
	}
	c.queue.Enqueue(event)
}
