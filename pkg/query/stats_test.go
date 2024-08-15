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
	}

	collector := new(DefaultStatsCollector)
	for _, i := range input {
		collector.Add(i.startedAt, i.finishedAt)
	}

	stats := collector.Stats()
	assert.Equal(t, time.Minute+time.Second*50, stats.TimeAcrossAllQueries) // From 00:00:00 to 00:01:50
	assert.Equal(t, time.Second*10, stats.MinTime)
	assert.Equal(t, time.Second*30, stats.MedianTime)                            // From 00:01:00 to 00:01:30
	assert.Equal(t, time.Second*50, stats.MaxTime)                               // From 00:01:00 to 00:01:50
	assert.Equal(t, time.Duration(float64(time.Second)*(26.875)), stats.AvgTime) // 215s / 8 data points = 26,875s
}

func generateTime(timeStr string) time.Time {
	t, _ := time.Parse(time.DateTime, timeStr)
	return t
}
