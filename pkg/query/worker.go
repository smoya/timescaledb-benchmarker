package query

import (
	"context"
	"github.com/smoya/timescaledb-benchmarker/pkg/run"
	"sync"
)

// Worker represents a worker of a WorkerPool.
type Worker interface {
	run.Startable
	run.Stoppable

	// ID returns the ID of the worker.
	ID() int

	// Enqueue enqueues a new query into the worker's processing line.
	Enqueue(context.Context, Query)
}

// DefaultWorker is a basic implementation of a Worker.
type DefaultWorker struct {
	Input       chan Query
	Output      chan Result
	queryRunner Runner
	id          int
	wg          sync.WaitGroup
	done        chan struct{}
}

// NewDefaultWorker creates a new DefaultWorker.
func NewDefaultWorker(ID int, runner Runner, input chan Query, output chan Result) *DefaultWorker {
	return &DefaultWorker{
		id:          ID,
		queryRunner: runner,
		Input:       input,
		Output:      output,
		done:        make(chan struct{}),
	}
}

// ID returns the ID of the worker. Implements the Worker interface.
func (w *DefaultWorker) ID() int {
	return w.id
}

// Enqueue enqueues a new query into the worker's processing line. Implements the Worker interface.
func (w *DefaultWorker) Enqueue(_ context.Context, query Query) {
	w.Input <- query // it is up to the caller decide if concurrency is needed in this operation.
}

// Start boots the Worker by starting its processing loop. Implements the Worker interface.
func (w *DefaultWorker) Start(ctx context.Context) error {
	for {
		select {
		case <-w.done:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		case q, ok := <-w.Input:
			if !ok {
				return nil
			}

			w.wg.Add(1)
			go func() {
				defer w.wg.Done()
				for _, r := range w.queryRunner(ctx, q) {
					w.Output <- r
				}
			}()
		}
	}
}

// Stop gracefully stops the worker by waiting for all executing queries. Implements the Worker interface.
func (w *DefaultWorker) Stop(_ context.Context) error {
	close(w.done)
	w.wg.Wait() // We wait for all in progress queries to finish
	return nil
}
