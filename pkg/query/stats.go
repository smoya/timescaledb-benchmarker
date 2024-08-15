package query

import (
	"math"
	"slices"
	"sync"
	"time"
)

// Stats are stats collected during Query executions.
type Stats struct {
	// TotalQueries is the # of executed queries
	TotalQueries int

	// TimeAcrossAllQueries is the execution time since the very first query until the last one.
	TimeAcrossAllQueries time.Duration

	// MinTime is the lowest query execution time.
	MinTime time.Duration

	// MedianTime is by the duration in the middle, where one half of the durations are higher and the other half are slower.
	MedianTime time.Duration

	// AvgTime is the sum of all durations divided by the total number of queries.
	AvgTime time.Duration

	// MaxTime is the highest query execution time.
	MaxTime time.Duration
}

// StatsCollector collect stats from query executions.
type StatsCollector interface {
	Add(startedAt, finishedAt time.Time)
	Stats() Stats
}

// DefaultStatsCollector collects Stats from query executions.
type DefaultStatsCollector struct {
	sync.Mutex // Not using RWMutex because in both ops we need to block on R and W.
	wg         sync.WaitGroup

	startedAts  []time.Time
	finishedAts []time.Time
	durations   []time.Duration
}

// Add stores a new data point.
func (c *DefaultStatsCollector) Add(startedAt, finishedAt time.Time) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.Lock()
		defer c.Unlock()

		c.startedAts = append(c.startedAts, startedAt)
		c.finishedAts = append(c.finishedAts, finishedAt)
		c.durations = append(c.durations, finishedAt.Sub(startedAt))
	}()
}

// Stats generates a new Stats struct based on sent data points.
func (c *DefaultStatsCollector) Stats() Stats {
	c.wg.Wait()

	c.Lock()
	defer c.Unlock()
	slices.SortFunc(c.startedAts, func(a, b time.Time) int {
		return a.Compare(b)
	})
	slices.SortFunc(c.finishedAts, func(a, b time.Time) int {
		return a.Compare(b)
	})

	timeAcrossAllQueries := c.finishedAts[len(c.finishedAts)-1].Sub(c.startedAts[0])

	// Required for getting the Median later on. Doing it here in order to calculate later on the max, and min at once avoiding extra iterations.
	slices.Sort(c.durations)

	queriesLen := len(c.durations)
	maxV := c.durations[queriesLen-1] // No need to use slices.Max since we have sorted the slice
	minV := c.durations[0]            // No need to use slices.Min since we have sorted the slice
	halfIndex, _ := math.Modf(float64(queriesLen / 2))
	median := c.durations[int(halfIndex)]

	var sum time.Duration
	for _, t := range c.durations {
		sum += t
	}
	avg := time.Duration(int(sum) / queriesLen)

	return Stats{
		TotalQueries:         queriesLen,
		TimeAcrossAllQueries: timeAcrossAllQueries,
		MinTime:              minV,
		MedianTime:           median,
		AvgTime:              avg,
		MaxTime:              maxV,
	}
}
