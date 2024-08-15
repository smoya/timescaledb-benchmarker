package timescaledb

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/smoya/timescaledb-benchmarker/pkg/query"
	"time"
)

// DB represents a TimescaleDB database. It is based on pgx.Conn and pgxpool.Pool, which is a driver for PostgreSQL.
type DB interface {
	Query(ctx context.Context, sql string, optionsAndArgs ...interface{}) (pgx.Rows, error)
}

// Query represents a query in a TimescaleDB DB.
type Query struct {
	BucketInterval  string // Postgres interval string representation.
	BucketTSColumn  string
	EntityIDValue   string
	EntityIDColumn  string
	BenchmarkColumn string
	Table           string
	PeriodFrom      time.Time
	PeriodTo        time.Time
}

func (q Query) EntityID() string {
	return q.EntityIDValue
}

// String returns a string representation of the Query. It implements fmt.Stringer interface.
func (q Query) String() string {
	return fmt.Sprintf(
		"SELECT time_bucket('%s', %s) as bucket, max(%s), min(%s) FROM %s WHERE %s = '%s' AND %s BETWEEN '%s' AND '%s' GROUP BY bucket, host ORDER BY bucket ASC;",
		q.BucketInterval, q.BucketTSColumn, q.BenchmarkColumn, q.BenchmarkColumn, q.Table, q.EntityIDColumn, q.EntityIDValue, q.BucketTSColumn, q.PeriodFrom.Format(time.DateTime), q.PeriodTo.Format(time.DateTime),
	)
}

// NewDBRunner creates a new TimescaleDB query.Runner
func NewDBRunner(dbPool DB) query.Runner {
	return func(ctx context.Context, q query.Query) []query.Result {
		rows, err := dbPool.Query(ctx, q.String())
		if err != nil {
			return []query.Result{{Err: err}}
		}

		var ts time.Time
		var maxV float64
		var minV float64

		var results []query.Result
		_, err = pgx.ForEachRow(rows, []any{&ts, &maxV, &minV}, func() error {
			results = append(results, query.Result{
				EntityID: q.EntityID(),
				TS:       ts,
				Max:      maxV,
				Min:      minV,
			})
			return nil
		})

		if err != nil {
			return []query.Result{{Err: err}}
		}

		return results
	}
}
