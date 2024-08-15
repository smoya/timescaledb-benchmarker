package timescaledb

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"github.com/smoya/timescaledb-benchmarker/pkg/benchmark"
	"github.com/smoya/timescaledb-benchmarker/pkg/query"
	"slices"
	"strings"
	"time"
)

type OutputFormat = string

const (
	FormatHumanReadable OutputFormat = "human"
	FormatCSV           OutputFormat = "csv"
	FormatTSV           OutputFormat = "tsv"
	FormatMarkdown      OutputFormat = "md"
	FormatHTML          OutputFormat = "html"
)

var allowedFormats = []OutputFormat{FormatHumanReadable, FormatCSV, FormatTSV, FormatMarkdown, FormatHTML}

// Benchmarker benchmarks TimescaleDB queries.
type Benchmarker struct {
	benchmark.Benchmarker
	workerPool query.WorkerPool
	dbConnPool *pgxpool.Pool
	stats      query.StatsCollector
}

// BenchmarkerConfig holds the configuration for the Benchmarker.
type BenchmarkerConfig struct {
	// Parsed according to https://github.com/jackc/pgx/blob/master/pgxpool/pool.go#L281-L297 respecting https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING
	DBConnectionURI string

	// Num of workers executing queries concurrently
	NumWorkers uint

	// QueryTimeout is the desired max query time. By default, no timeout is applied.
	QueryTimeout time.Duration

	// Debug enables/disables debug mode. Used for printing debug log lines.
	Debug bool

	// OutputFormat is the printing format. [human,csv,markdown], defaults to human.
	OutputFormat OutputFormat
}

// Validate validates the config.
func (c BenchmarkerConfig) Validate() error {
	var errs []error

	// extended parsing and validation of the DB Connection URI is done by the driver later on.
	if c.DBConnectionURI == "" {
		errs = append(errs, errors.New("DBConnectionURI is a required value"))
	}

	if c.NumWorkers == 0 {
		errs = append(errs, errors.New("NumWorkers should be 1 or greater"))
	}

	if c.OutputFormat != "" && !slices.Contains(allowedFormats, c.OutputFormat) {
		errs = append(errs, fmt.Errorf("invalid OutputFormat %s. Allowed values: %s", c.OutputFormat, strings.Join(allowedFormats, ",")))
	}

	return errors.Join(errs...)
}

// NewBenchmarker creates a new Benchmarker.
func NewBenchmarker(ctx context.Context, c *BenchmarkerConfig, input chan query.Query, output chan query.Result) (*Benchmarker, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	if c.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	dbConf, err := pgxpool.ParseConfig(c.DBConnectionURI)
	if err != nil {
		return nil, err
	}

	dbPool, err := pgxpool.NewWithConfig(ctx, dbConf)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %s", err)
	}

	statsCollector := new(query.DefaultStatsCollector)
	runner := query.WithStats(NewDBRunner(dbPool), statsCollector)
	if c.QueryTimeout > 0 {
		runner = query.WithContextTimeout(runner, c.QueryTimeout)
	}

	pool := query.NewShardedWorkerPool(runner, c.NumWorkers, query.ShardingFNV1aWorkerAssigner, input, output)
	return &Benchmarker{workerPool: pool, stats: statsCollector, dbConnPool: dbPool}, nil
}

// Start starts the Benchmarker. Implements the run.Startable interface.
func (b *Benchmarker) Start(ctx context.Context) error {
	return b.workerPool.Start(ctx)
}

// Stop stops the Benchmarker, producing the benchmark stats.

func (b *Benchmarker) Stop(ctx context.Context) (query.StatsCollector, error) {
	defer b.dbConnPool.Close()
	return b.stats, b.workerPool.Stop(ctx)
}
