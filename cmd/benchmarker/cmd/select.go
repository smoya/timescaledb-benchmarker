package cmd

import (
	"context"
	"encoding/csv"
	"errors"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/sirupsen/logrus"
	"github.com/smoya/timescaledb-benchmarker/pkg/query"
	"github.com/smoya/timescaledb-benchmarker/pkg/timescaledb"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// BenchmarkTimescaleDBSelectQueries is the function executed by the "benchmark SELECT queries" command.
func BenchmarkTimescaleDBSelectQueries(ctx context.Context, filePath string, config *timescaledb.BenchmarkerConfig) error {
	queries := make(chan query.Query)
	if err := readCSV(filePath, queries); err != nil {
		return err
	}

	result := make(chan query.Result)
	benchmarker, err := timescaledb.NewBenchmarker(ctx, config, queries, result)
	if err != nil {
		return err
	}

	// Fancy and cool table writer.
	t := createTableWriter()

	var resultsWg sync.WaitGroup
	resultsWg.Add(1)
	go func() {
		defer resultsWg.Done()
		readResultsLoop(ctx, result, t)
	}()

	if err := benchmarker.Start(ctx); err != nil {
		return err
	}

	statsCollector, err := benchmarker.Stop(ctx)
	if err != nil {
		return err
	}

	// No more results to enqueue.
	close(result)

	// Wait for all results to be processed.
	resultsWg.Wait()

	// Print results and stats.
	render(t, statsCollector.Stats(), config)

	return nil
}

func render(t table.Writer, stats query.Stats, config *timescaledb.BenchmarkerConfig) {
	t.AppendFooter(table.Row{"", "", "# Queries", stats.TotalQueries})
	t.AppendFooter(table.Row{"", "", "Total Query time (across all queries)", stats.TimeAcrossAllQueries})
	t.AppendFooter(table.Row{"", "", "Min query time", stats.MinTime})
	t.AppendFooter(table.Row{"", "", "Median query time", stats.MedianTime})
	t.AppendFooter(table.Row{"", "", "Average query time", stats.AvgTime})
	t.AppendFooter(table.Row{"", "", "Max query time", stats.MaxTime})

	switch strings.ToLower(config.OutputFormat) {
	case timescaledb.FormatCSV:
		t.RenderCSV()
	case timescaledb.FormatTSV:
		t.RenderTSV()
	case timescaledb.FormatMarkdown:
		t.RenderMarkdown()
	case timescaledb.FormatHTML:
		t.RenderHTML()
	case timescaledb.FormatHumanReadable:
		fallthrough
	default:
		t.Render()
	}
}

func createTableWriter() table.Writer {
	t := table.NewWriter()
	t.SetStyle(table.StyleColoredBright)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"TS", "HOST", "MAX CPU USAGE", "MIN CPU USAGE"})
	t.SortBy([]table.SortBy{
		{Name: "TS", Mode: table.Asc},
		{Name: "HOST", Mode: table.Asc},
	})
	return t
}

func readResultsLoop(ctx context.Context, result chan query.Result, t table.Writer) {
	for {
		select {
		case <-ctx.Done():
			return
		case r, ok := <-result:
			if !ok {
				return
			}

			if r.Err != nil {
				if errors.Is(r.Err, context.DeadlineExceeded) {
					logrus.WithError(r.Err).Fatal("Timeout reached. Please consider setting a greater timeout if makes sense.")
				}
				logrus.WithError(r.Err).Fatal("Query errored")
			}

			t.AppendRow([]interface{}{r.TS, r.EntityID, r.Max, r.Min})
		}
	}
}

func readCSV(filepath string, dest chan<- query.Query) error {
	var r io.Reader
	if filepath != "" {
		file, err := os.Open(filepath)
		if err != nil {
			return err
		}

		defer func() {
			_ = file.Close()
		}()
		r = file
	} else {
		// try with STDIN
		fi, _ := os.Stdin.Stat()
		if (fi.Mode() & os.ModeCharDevice) != 0 {
			return errors.New("STDIN for this command is only supported via pipe and not terminal. Please pipe a file through the redirection operator (I.e. command < file.csv)")
		}
		r = os.Stdin
	}

	readRowsWG := &sync.WaitGroup{}
	if err := readFromReader(r, readRowsWG, dest); err != nil {
		return err
	}

	go func() {
		// Close the dest channel once all lines have been read.
		readRowsWG.Wait()
		close(dest)
	}()

	return nil
}

func readFromReader(r io.Reader, readRowsWG *sync.WaitGroup, dest chan<- query.Query) error {
	csvReader := csv.NewReader(r)

	// Skipping headers.
	if _, err := csvReader.Read(); err != nil {
		return err
	}

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if len(record) < 3 {
			return errors.New("CSV row should contain only 3 values: hostname,start_time,end_time")
		}

		hostname := record[0]
		if hostname == "" {
			return errors.New("hostname should be present on each CSV row")
		}

		from, err := time.Parse(time.DateTime, record[1])
		if err != nil {
			return err
		}
		to, err := time.Parse(time.DateTime, record[2])
		if err != nil {
			return err
		}

		readRowsWG.Add(1)
		go func() {
			defer readRowsWG.Done()
			dest <- timescaledb.Query{
				EntityIDValue:   hostname,
				EntityIDColumn:  "host",
				BenchmarkColumn: "usage",
				Table:           "cpu_usage",
				PeriodFrom:      from,
				PeriodTo:        to,
				BucketInterval:  "1 minute",
				BucketTSColumn:  "ts",
			}
		}()
	}
	return nil
}
