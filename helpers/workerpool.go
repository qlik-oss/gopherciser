package helpers

import (
	"fmt"

	"github.com/pkg/errors"
)

type (
	WorkerPool struct {
		taskChan     chan func() error
		resultChan   chan error
		addedTasks   SyncCounter
		handledTasks SyncCounter
		total        int
	}
)

func (pool *WorkerPool) worker(id int) {
	i := 0
	for task := range pool.taskChan {
		i++
		fmt.Printf("Worker<%d> processing task<%d>\n", id, i) // TODO remove
		pool.resultChan <- task()
		if pool.handledTasks.Inc() >= pool.total {
			close(pool.resultChan)
		}
	}
}

// NewWorkerPool creates channels based on Concurrency and Total tasks.
// Add tasks with AddTask function, once all tasks has been added,
// Results can be read from Results(), once Results is closed all tasks are finished
// Exactly the "total" amount needs to be added or Results will never be closed
func NewWorkerPool(concurrency, total int) (*WorkerPool, error) {
	if concurrency < 1 {
		return nil, errors.Errorf("no concurrency set")
	}
	if total < 1 {
		return nil, errors.Errorf("total tasks not set")
	}
	pool := &WorkerPool{
		taskChan:   make(chan func() error, concurrency),
		resultChan: make(chan error, total),
		total:      total,
	}

	// Start the workers
	for i := 1; i <= concurrency; i++ {
		go pool.worker(i)
	}

	return pool, nil
}

// AddTask to worker pool (Needs to be exactly the amount of tasks defined by "Total" or pool will never close results)
func (pool *WorkerPool) AddTask(f func() error) error {
	addedTasks := pool.addedTasks.Inc()

	if addedTasks > pool.total {
		return errors.Errorf("cannot add more than %d tasks (defined by total when starting worker pool)", pool.total)
	}
	pool.taskChan <- f
	if addedTasks == pool.total {
		close(pool.taskChan)
	}
	return nil
}

// Results from tasks, once this channel is closed all tasks are finished
func (pool *WorkerPool) Results() <-chan error {
	return pool.resultChan
}
