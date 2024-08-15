package query

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"testing"
)

func TestNewDefaultWorker(t *testing.T) {
	output := make(chan Result)
	_, input, worker := createWorker(output, nil)

	require.IsType(t, &DefaultWorker{}, worker)
	require.Implements(t, (*Worker)(nil), worker)
	assert.Equal(t, input, worker.Input)
	assert.Equal(t, output, worker.Output)
	assert.Equal(t, 99, worker.ID())
}

func TestDefaultWorker_Enqueue(t *testing.T) {
	output := make(chan Result)
	_, input, worker := createWorker(output, nil)
	q := &testQuery{}
	go worker.Enqueue(context.Background(), q)
	receivedQ := <-input
	assert.Same(t, q, receivedQ)
}

func TestDefaultWorker_LifeCycle_ExecuteQueries_Stop_By_Closing_Input_Channel(t *testing.T) {
	testtestWorkerLifeCycle(t, context.Background(), func(worker *DefaultWorker) {
		close(worker.Input)
	}, nil)
}

func TestDefaultWorker_LifeCycle_ExecuteQueries_Stop_By_Stopping_Worker(t *testing.T) {
	testtestWorkerLifeCycle(t, context.Background(), func(worker *DefaultWorker) {
		assert.NoError(t, worker.Stop(context.Background()))
	}, nil)
}

func TestDefaultWorker_LifeCycle_ExecuteQueries_Stop_By_Canceling_Context(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	testtestWorkerLifeCycle(t, ctx, func(worker *DefaultWorker) {
		cancel()
	}, context.Canceled)
}

// testtestWorkerLifeCycle tests DefaultWorker.Start method allowing to pass a function that will stop the Start loop.
func testtestWorkerLifeCycle(t *testing.T, ctx context.Context, stopper func(worker *DefaultWorker), expectedErr error) {
	tQuery := testQuery{}
	testResult := Result{
		EntityID: "test",
	}

	r := func(context.Context, Query) []Result {
		return []Result{
			{
				EntityID: "test",
			},
		}
	}

	output := make(chan Result)
	_, input, worker := createWorker(output, r)

	g := errgroup.Group{}
	g.Go(func() error {
		return worker.Start(ctx)
	})

	go func() {
		input <- tQuery
	}()

	result := <-output
	assert.Equal(t, testResult, result)

	// stopper stops the worker
	stopper(worker)

	assert.ErrorIs(t, expectedErr, g.Wait())
}

func createWorker(output chan Result, r Runner) (Runner, chan Query, *DefaultWorker) {
	if r == nil {
		r = func(context.Context, Query) []Result {
			return nil
		}
	}
	input := make(chan Query)
	return r, input, NewDefaultWorker(99, r, input, output)
}
