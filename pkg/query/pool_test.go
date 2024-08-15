package query

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"testing"
)

type testWorker struct {
	id int
}

func (t testWorker) Start(_ context.Context) error {
	return nil
}

func (t testWorker) Stop(_ context.Context) error {
	return nil
}

func (t testWorker) ID() int {
	return t.id
}

func (t testWorker) Enqueue(_ context.Context, _ Query) {}

func TestShardedWorkerPool_LifeCycle(t *testing.T) {
	r := func(ctx context.Context, query Query) []Result {
		return query.(testQuery).expectedResults
	}

	assigner := func(query Query, workers []Worker) Worker {
		return workers[0] // for simplicity’s sake
	}

	input := make(chan Query)
	output := make(chan Result)
	pool := NewShardedWorkerPool(r, 5, assigner, input, output)

	var g errgroup.Group
	g.Go(func() error {
		return pool.Start(context.Background())
	})

	results := []Result{{EntityID: "host_a"}, {EntityID: "host_b"}}
	go func() {
		input <- testQuery{
			expectedResults: results,
		}
	}()

	assert.Equal(t, results[0], <-output)
	assert.Equal(t, results[1], <-output)

	require.NoError(t, pool.Stop(context.Background()))
	assert.NoError(t, g.Wait()) // Start loop should be terminated at this point. This wait just proves that.
}

func TestNewShardedWorkerPool(t *testing.T) {
	r := func(ctx context.Context, query Query) []Result {
		return nil
	}

	assigner := func(query Query, workers []Worker) Worker {
		return workers[0] // for simplicity’s sake
	}

	input := make(chan Query)
	output := make(chan Result)
	pool := NewShardedWorkerPool(r, 5, assigner, input, output)
	assert.IsType(t, &ShardedWorkerPool{}, pool)
	assert.Implements(t, (*WorkerPool)(nil), pool)
}

func TestShardingFNV1aWorkerAssigner(t *testing.T) {
	tests := []struct {
		workers          []Worker
		entityID         string // For simplicity, the worker's index in the slice
		expectedWorkerID int
	}{
		{workers: createWorkers(1), entityID: "entity_a", expectedWorkerID: 0},
		{workers: createWorkers(2), entityID: "entity_a", expectedWorkerID: 0},
		{workers: createWorkers(3), entityID: "entity_a", expectedWorkerID: 2},
		{workers: createWorkers(5), entityID: "entity_a", expectedWorkerID: 2},
		{workers: createWorkers(50), entityID: "entity_a", expectedWorkerID: 12},
		{workers: createWorkers(1), entityID: "entity_b", expectedWorkerID: 0},
		{workers: createWorkers(2), entityID: "entity_b", expectedWorkerID: 1},
		{workers: createWorkers(3), entityID: "entity_b", expectedWorkerID: 0},
		{workers: createWorkers(5), entityID: "entity_b", expectedWorkerID: 3},
		{workers: createWorkers(50), entityID: "entity_b", expectedWorkerID: 43},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("EntityID %s should be assigned to worker id %d of %d workers", test.entityID, test.expectedWorkerID, len(test.workers)), func(t *testing.T) {
			require.Greater(t, len(test.workers), test.expectedWorkerID, "The expectedWorkerID is greater than the len of available workers")
			assignedWorker := ShardingFNV1aWorkerAssigner(testQuery{entityID: test.entityID}, test.workers)
			assert.Equal(t, test.workers[test.expectedWorkerID], assignedWorker)
		})
	}
}

func createWorkers(amount int) []Worker {
	workers := make([]Worker, amount)
	for i := 0; i < amount; i++ {
		workers[i] = testWorker{id: i}
	}

	return workers
}
