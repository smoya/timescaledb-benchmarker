package main

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/smoya/timescaledb-benchmarker/cmd/benchmarker/cmd"
	"github.com/smoya/timescaledb-benchmarker/pkg/timescaledb"
	"github.com/urfave/cli/v2"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handleInterruptions(cancel)

	const envVarPrefix = "TIMESCALEDB_BENCHMARKER_BENCHMARK_"

	app := &cli.App{
		Name:                 "timescaledb_benchmarker",
		Usage:                "A benchmarker for TimescaleDB",
		EnableBashCompletion: true,
		Suggest:              true,
		Commands: []*cli.Command{
			{
				Name:    "benchmark",
				Aliases: []string{"b"},
				Usage:   "benchmark query performance across multiple workers/clients against a TimescaleDB instance",
				Subcommands: []*cli.Command{
					{
						Name:  "select",
						Usage: "benchmark SELECT queries",
						Flags: []cli.Flag{
							&cli.PathFlag{
								Name:      "file",
								Aliases:   []string{"f"},
								TakesFile: true,
								EnvVars:   []string{envVarPrefix + "FILE"},
								Usage:     "path to a csv file containing raw query fields",
							},
							&cli.StringFlag{
								Name:     "db_uri",
								EnvVars:  []string{envVarPrefix + "DB_URI"},
								Required: true,
								Usage:    "TimescaleDB Connection URI. I.e. postgres://username:password@localhost:5432/database_name",
							},
							&cli.UintFlag{
								Name:    "workers",
								Aliases: []string{"w"},
								EnvVars: []string{envVarPrefix + "WORKERS"},
								Value:   5,
								Usage:   "Number of query workers executing Queries concurrently. Different from Postgress pool size, which can be configured in parallel through the db_uri flag.",
							},
							&cli.DurationFlag{
								Name:    "timeout",
								Aliases: []string{"t"},
								EnvVars: []string{envVarPrefix + "TIMEOUT"},
								Value:   time.Millisecond * 100,
								Usage:   "Timeout for each query. A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix such as \"300ms\", \"-1.5h\" or \"2h45m",
							},
							&cli.BoolFlag{
								Name:    "debug",
								Aliases: []string{"d"},
								EnvVars: []string{envVarPrefix + "DEBUG"},
								Value:   false,
								Usage:   "Debug mode. Enable it for printing debug logs.",
							},
							&cli.StringFlag{
								Name:    "output_format",
								EnvVars: []string{envVarPrefix + "OUTPUT_FORMAT"},
								Value:   timescaledb.FormatHumanReadable,
								Usage:   "Output print format. By default, human readable output for printing in the console. Available formats: human,csv,tsv,md,html",
							},
						},
						Action: func(cCtx *cli.Context) error {
							benchmarkerConfig := &timescaledb.BenchmarkerConfig{
								DBConnectionURI: cCtx.String("db_uri"),
								NumWorkers:      cCtx.Uint("workers"),
								QueryTimeout:    cCtx.Duration("timeout"),
								Debug:           cCtx.Bool("debug"),
								OutputFormat:    cCtx.String("output_format"),
							}
							return cmd.BenchmarkTimescaleDBSelectQueries(cCtx.Context, cCtx.Path("file"), benchmarkerConfig)
						},
					},
				},
			},
		},
	}

	if err := app.RunContext(ctx, os.Args); err != nil {
		logrus.WithError(err).Fatal("CLI errored")
	}
}

func handleInterruptions(cancel context.CancelFunc) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-c
		logrus.WithField("signal", s).Info("Stopping CLI...", s)
		cancel()                // canceling this context will propagate this cancellation across the whole execution pipeline
		time.Sleep(time.Second) // grace time for printing debug/info if any
	}()
}
