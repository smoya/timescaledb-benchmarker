package timescaledb

import (
	"context"
	"github.com/smoya/timescaledb-benchmarker/pkg/benchmark"
	"github.com/smoya/timescaledb-benchmarker/pkg/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type testWorkerPool struct {
	started bool
	stopped bool
}

func (t *testWorkerPool) Start(ctx context.Context) error {
	t.started = true
	return nil
}

func (t *testWorkerPool) Stop(ctx context.Context) error {
	t.stopped = true
	return nil
}

func TestNewBenchmarker(t *testing.T) {
	conf := &BenchmarkerConfig{
		DBConnectionURI: "postgres://postgres:postgres@localhost:5432/test",
		NumWorkers:      5,
	}

	input := make(chan query.Query)
	output := make(chan query.Result)
	b, err := NewBenchmarker(context.Background(), conf, input, output)
	assert.Nil(t, err)
	assert.IsType(t, &Benchmarker{}, b)
	assert.Implements(t, (*benchmark.Benchmarker)(nil), b)
}

func TestNewBenchmarker_LifeCycle_loop(t *testing.T) {
	conf := &BenchmarkerConfig{
		DBConnectionURI: "postgres://postgres:postgres@localhost:5432/test",
		NumWorkers:      5,
	}

	input := make(chan query.Query)
	output := make(chan query.Result)
	b, err := NewBenchmarker(context.Background(), conf, input, output)
	require.Nil(t, err)

	pool := &testWorkerPool{}
	b.workerPool = pool

	err = b.Start(context.Background())
	assert.NoError(t, err)
	assert.True(t, pool.started)
	assert.False(t, pool.stopped)

	statsCollector, err := b.Stop(context.Background())
	assert.NoError(t, err)
	assert.True(t, pool.stopped)
	assert.Implements(t, (*query.StatsCollector)(nil), statsCollector)
}

func TestBenchmarkerConfig_Validate(t *testing.T) {
	tests := []struct {
		name           string
		config         BenchmarkerConfig
		expectedErrStr string
	}{
		{
			name: "Valid config",
			config: BenchmarkerConfig{
				DBConnectionURI: "postgres://postgres:postgres@localhost:5432/test",
				NumWorkers:      5,
				QueryTimeout:    time.Second,
				Debug:           false,
			},
		},
		{
			name: "Invalid config - missing DBConnectionURI",
			config: BenchmarkerConfig{
				DBConnectionURI: "",
				NumWorkers:      5,
				QueryTimeout:    time.Second,
				Debug:           false,
			},
			expectedErrStr: "DBConnectionURI is a required value",
		},

		{
			name: "Invalid config - missing DBConnectionURI and numWorkers",
			config: BenchmarkerConfig{
				DBConnectionURI: "",
				NumWorkers:      0,
				QueryTimeout:    time.Second,
				Debug:           false,
			},
			expectedErrStr: "DBConnectionURI is a required value\nNumWorkers should be 1 or greater",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.config.Validate()
			if test.expectedErrStr != "" {
				assert.EqualError(t, err, test.expectedErrStr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewBenchmarker_Error_Invalid_Config(t *testing.T) {
	invalidConf := &BenchmarkerConfig{}
	b, err := NewBenchmarker(context.Background(), invalidConf, make(chan query.Query), make(chan query.Result))
	assert.EqualError(t, err, invalidConf.Validate().Error())
	assert.Nil(t, b)
}

func TestNewBenchmarker_Error_Invalid_URI_format(t *testing.T) {
	conf := &BenchmarkerConfig{
		DBConnectionURI: "invalid-uri",
		NumWorkers:      5,
	}

	input := make(chan query.Query)
	output := make(chan query.Result)
	b, err := NewBenchmarker(context.Background(), conf, input, output)

	assert.EqualError(t, err, "cannot parse `invalid-uri`: failed to parse as keyword/value (invalid keyword/value)")
	assert.Nil(t, b)
}
