package query

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type testQuery struct {
	entityID        string
	expectedResults []Result
}

func (t testQuery) String() string {
	return ""
}

func (t testQuery) EntityID() string {
	return t.entityID
}

func TestWithStats(t *testing.T) {
	results := []Result{
		{EntityID: "abc"},
		{EntityID: "bcd"},
	}
	r := func(_ context.Context, query Query) []Result {
		return results
	}

	collector := new(DefaultStatsCollector)
	wrapped := WithStats(r, collector)
	assert.Equal(t, results, wrapped(context.Background(), testQuery{}))
	assert.Equal(t, results, wrapped(context.Background(), testQuery{}))

	stats := collector.Stats()
	assert.Equal(t, 2, stats.TotalQueries)
}

func TestWithContextTimeout(t *testing.T) {
	r := func(ctx context.Context, query Query) []Result {
		<-ctx.Done()
		return []Result{{Err: ctx.Err()}}
	}

	wrapped := WithContextTimeout(r, time.Nanosecond)
	result := wrapped(context.Background(), testQuery{})
	require.Len(t, result, 1)
	assert.ErrorIs(t, result[0].Err, context.DeadlineExceeded)
}
