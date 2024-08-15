package query

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type collectorAddInput struct {
	startedAt  time.Time
	finishedAt time.Time
}

func TestDefaultStatsCollector_Stats(t *testing.T) {
	input := []collectorAddInput{
		{startedAt: generateTime("2022-12-05 00:00:00"), finishedAt: generateTime("2022-12-05 00:00:10")}, // 10s
		{startedAt: generateTime("2022-12-05 00:00:01"), finishedAt: generateTime("2022-12-05 00:00:11")}, // 10s
		{startedAt: generateTime("2022-12-05 00:00:02"), finishedAt: generateTime("2022-12-05 00:00:12")}, // 10s
		{startedAt: generateTime("2022-12-05 00:00:05"), finishedAt: generateTime("2022-12-05 00:00:15")}, // 10s
		{startedAt: generateTime("2022-12-05 00:01:00"), finishedAt: generateTime("2022-12-05 00:01:30")}, // 30s
		{startedAt: generateTime("2022-12-05 00:00:00"), finishedAt: generateTime("2022-12-05 00:00:45")}, // 45s
		{startedAt: generateTime("2022-12-05 00:00:00"), finishedAt: generateTime("2022-12-05 00:00:50")}, // 50s
		{startedAt: generateTime("2022-12-05 00:01:00"), finishedAt: generateTime("2022-12-05 00:01:50")}, // 50s
		{startedAt: generateTime("2022-12-05 00:01:00"), finishedAt: generateTime("2022-12-05 00:01:52")}, // 52s
		{startedAt: generateTime("2022-12-05 00:01:00"), finishedAt: generateTime("2022-12-05 00:01:55")}, // 55s
	}

	collector := new(DefaultStatsCollector)
	for _, i := range input {
		collector.Add(i.startedAt, i.finishedAt)
	}

	stats := collector.Stats()
	assert.Equal(t, time.Minute+time.Second*55, stats.TimeAcrossAllQueries) // From 00:00:00 to 00:01:55
	assert.Equal(t, time.Second*10, stats.MinTime)
	assert.Equal(t, time.Second*45, stats.MedianTime)                                     // From 00:01:00 to 00:01:45
	assert.Equal(t, time.Second*55, stats.MaxTime)                                        // From 00:00:00 to 00:00:55
	assert.Equal(t, 32200*time.Millisecond, stats.AvgTime)                                // 322 / 10 data points = 32.2s
	assert.Equal(t, time.Duration(19.197916553*float64(time.Second)), stats.StdDeviation) // Deviation is 19.197916553s (manually calculated)
	assert.Equal(t, time.Second*52, stats.Percentile95th)                                 // The 95% of queries took 52s or less. The rest took more than 52s.
}

func generateTime(timeStr string) time.Time {
	t, _ := time.Parse(time.DateTime, timeStr)
	return t
}
