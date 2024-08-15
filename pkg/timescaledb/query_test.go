package timescaledb

import (
	"context"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/smoya/timescaledb-benchmarker/pkg/query"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var mockQuery = Query{
	EntityIDValue:   "host_a",
	EntityIDColumn:  "host",
	BenchmarkColumn: "usage",
	Table:           "cpu_usage",
	PeriodFrom:      generateTime("2022-12-05 00:00:00"),
	PeriodTo:        generateTime("2022-12-05 01:00:00"),
	BucketInterval:  "1 minute",
	BucketTSColumn:  "ts",
}

func TestNewDBRunner_WithResults(t *testing.T) {
	db, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	expectedResults := []query.Result{
		{EntityID: "host_a", TS: time.Now(), Max: 122, Min: 80},
		{EntityID: "host_a", TS: time.Now(), Max: 345, Min: 188},
	}
	r := NewDBRunner(db)

	rows := pgxmock.NewRows([]string{"TS", "max", "min"}).
		AddRow(expectedResults[0].TS, expectedResults[0].Max, expectedResults[0].Min).
		AddRow(expectedResults[1].TS, expectedResults[1].Max, expectedResults[1].Min)

	db.ExpectQuery(mockQuery.String()).WillReturnRows(rows).RowsWillBeClosed()

	results := r(context.Background(), mockQuery)
	assert.Equal(t, expectedResults, results)
}

func TestNewDBRunner_With_No_Results(t *testing.T) {
	db, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	r := NewDBRunner(db)

	rows := pgxmock.NewRows([]string{"TS", "max", "min"})
	db.ExpectQuery(mockQuery.String()).WillReturnRows(rows) // No results.
	results := r(context.Background(), mockQuery)
	assert.Empty(t, results)
}

func TestQuery_String(t *testing.T) {
	tests := []struct {
		query       Query
		expectedStr string
	}{
		{
			query:       mockQuery,
			expectedStr: "SELECT time_bucket('1 minute', ts) as bucket, max(usage), min(usage) FROM cpu_usage WHERE host = 'host_a' AND ts BETWEEN '2022-12-05 00:00:00' AND '2022-12-05 01:00:00' GROUP BY bucket, host ORDER BY bucket ASC;",
		},
		{
			query: Query{
				EntityIDValue:   "device_a",
				EntityIDColumn:  "device",
				BenchmarkColumn: "consumption",
				Table:           "power_consumption",
				PeriodFrom:      generateTime("2019-02-26 00:00:00"),
				PeriodTo:        generateTime("2022-02-27 00:00:00"),
				BucketInterval:  "1 hour",
				BucketTSColumn:  "readAt",
			},
			expectedStr: "SELECT time_bucket('1 hour', readAt) as bucket, max(consumption), min(consumption) FROM power_consumption WHERE device = 'device_a' AND readAt BETWEEN '2019-02-26 00:00:00' AND '2022-02-27 00:00:00' GROUP BY bucket, host ORDER BY bucket ASC;",
		},
	}
	for _, test := range tests {
		t.Run(t.Name(), func(t *testing.T) {
			assert.Equal(t, test.expectedStr, test.query.String())
			assert.Equal(t, test.query.EntityID(), test.query.EntityIDValue)
		})
	}
}

func generateTime(timeStr string) time.Time {
	t, _ := time.Parse(time.DateTime, timeStr)
	return t
}
