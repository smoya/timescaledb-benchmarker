package query

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/smoya/timescaledb-benchmarker/pkg/run"
	"golang.org/x/sync/errgroup"
	"hash/fnv"
	"sync"
)

// WorkerPool is a pool of Worker
type WorkerPool interface {
	run.Startable
	run.Stoppable
}

// ShardedWorkerPool is a WorkerPool that uses standard sharding in order to distribute queries across workers.
// The hashing algorithm is implemented on WorkerAssigner.
type ShardedWorkerPool struct {
	Input                  chan Query
	Output                 chan Result
	Workers                []Worker
	WorkerAssigner         WorkerAssigner
	wg                     sync.WaitGroup
	workerChannels         []chan Query
	workerAssignationCache map[string]Worker
	done                   chan struct{}
}

func (s *ShardedWorkerPool) assignWorker(q Query) Worker {
	worker, ok := s.workerAssignationCache[q.EntityID()]
	if !ok {
		worker = s.WorkerAssigner(q, s.Workers)
		s.workerAssignationCache[q.EntityID()] = worker
		logrus.WithFields(logrus.Fields{"host": q.EntityID(), "worker": worker.ID()}).Debug("Worker assignation")
	}

	return worker
}

// Start initiates the pool and enters into a read loop.
func (s *ShardedWorkerPool) Start(ctx context.Context) error {
	for _, w := range s.Workers {
		go func() {
			_ = w.Start(ctx)
		}()
	}

	for {
		select {
		case <-s.done:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		case q, ok := <-s.Input:
			if !ok {
				return nil
			}
			s.wg.Add(1)
			worker := s.assignWorker(q)
			go func() {
				defer s.wg.Done()
				worker.Enqueue(ctx, q)
			}()
		}
	}
}

// Stop waits for all queries to be executed and then stop workers and close their channels.
func (s *ShardedWorkerPool) Stop(ctx context.Context) error {
	close(s.done) // No more queries allowed. Exit the Start loop if it was still running.
	s.wg.Wait()   // Wait until all queries have been enqueued

	var g *errgroup.Group
	g, ctx = errgroup.WithContext(ctx)
	for _, w := range s.Workers {
		g.Go(func() error {
			return w.Stop(ctx)
		})
	}

	// Wait until all workers let finish their queries
	if err := g.Wait(); err != nil {
		return err
	}

	// Closing channels, so workers can free their loops once queries are done.
	// Responsibility of the pool, which is the one who created them.
	for _, c := range s.workerChannels {
		close(c)
	}

	return nil
}

// NewShardedWorkerPool creates a ShardedWorkerPool.
func NewShardedWorkerPool(queryRunner Runner, numOfWorkers uint, workerAssigner WorkerAssigner, input chan Query, output chan Result) *ShardedWorkerPool {
	var workers = make([]Worker, numOfWorkers)
	var workerChannels = make([]chan Query, numOfWorkers)
	for i := range workers {
		workerInput := make(chan Query) // A different channel for each worker, otherwise fan out will be happening instead of sharding
		workerChannels[i] = workerInput
		workers[i] = NewDefaultWorker(i, queryRunner, workerInput, output) // For simplicity, Worker's ID is the slice index.
	}

	return &ShardedWorkerPool{
		Workers:                workers,
		Input:                  input,
		Output:                 output,
		WorkerAssigner:         workerAssigner,
		workerChannels:         workerChannels,
		workerAssignationCache: make(map[string]Worker),
		done:                   make(chan struct{}),
	}
}

// WorkerAssigner assigns a Worker based on a Query and available workers.
type WorkerAssigner func(Query, []Worker) Worker

// ShardingFNV1aWorkerAssigner is a WorkerAssigner that foll
func ShardingFNV1aWorkerAssigner(q Query, workers []Worker) Worker {
	h := fnv.New32a()
	_, _ = h.Write([]byte(q.EntityID()))          // Hashing the entityID so we get a deterministic number
	i := uint(h.Sum32() % (uint32(len(workers)))) // Consistent Hashing algorithm -> modulo(hashed entityID, num workers)
	return workers[i]
}
