package benchmark

import (
	"context"
	"github.com/smoya/timescaledb-benchmarker/pkg/query"
	"github.com/stretchr/testify/assert"
	"testing"
)

type testBenchmarker struct {
}

func (t testBenchmarker) Start(ctx context.Context) error {
	return nil
}

func (t testBenchmarker) Stop(ctx context.Context) (query.StatsCollector, error) {
	return nil, nil
}

func TestBenchmark(t *testing.T) {
	assert.Implements(t, (*Benchmarker)(nil), testBenchmarker{})
}
