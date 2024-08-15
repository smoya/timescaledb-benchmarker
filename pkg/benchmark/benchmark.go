package benchmark

import (
	"context"
	"github.com/smoya/timescaledb-benchmarker/pkg/query"
	"github.com/smoya/timescaledb-benchmarker/pkg/run"
)

// Benchmarker benchmarks DB queries.
type Benchmarker interface {
	run.Startable
	Stop(ctx context.Context) (query.StatsCollector, error)
}
