package query

import (
	"context"
	"fmt"
	"time"
)

// Query represents a query in a DB.
type Query interface {
	fmt.Stringer

	// EntityID returns the ID representing the entity being queried. Useful for sharding, debugging, etc.
	EntityID() string
}

// Result is the product of an executed Query.
type Result struct {
	Err      error
	EntityID string
	TS       time.Time
	Max      float64
	Min      float64
}

// Runner runs queries in a DB. Runners are Context aware, including Context timeouts.
type Runner func(context.Context, Query) []Result

// WithStats is a Runner wrapper that collect stats through a StatsCollector.
func WithStats(r Runner, statsCollector StatsCollector) Runner {
	return func(ctx context.Context, query Query) []Result {
		startTime := time.Now()
		results := r(ctx, query)
		statsCollector.Add(startTime, time.Now())
		return results
	}
}

// WithContextTimeout is a Runner wrapper that sets a timeout in the context of the execution.
func WithContextTimeout(r Runner, maxDuration time.Duration) Runner {
	return func(ctx context.Context, query Query) []Result {
		ctx, cancel := context.WithTimeout(ctx, maxDuration)
		defer cancel()
		return r(ctx, query)
	}
}
