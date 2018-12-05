package pond

import (
	"time"
)

// Pool interface defines the critical methods a pool must implement,
// it also represents the main methods exposed to users.
type Pool interface {
	// Submit is the main entry for submitting new tasks.
	Submit(task Task) (Future, error)

	// SubmitWithTimeout submit a new task and set expiration.
	SubmitWithTimeout(task Task, timeout time.Duration) (Future, error)

	// Resize dynamically reset the capacity of pool.
	Resize(newSize int)

	// Pause will block the whole pool, util Resume is invoked.
	// Pool should wait for all under running tasks to be done, and
	// clear all idle workers.
	Pause()

	// Resume restart the paused pool, if pool is in running state,
	// this is a no-op.
	Resume()

	// Close close the pool and recycle the resource, users must invoke
	// this method if they do not use pool anymore.
	Close()
}

type basicPool struct {
	capacity int
	workers  []*Worker
	taskQ    chan<- *taskWrapper
	pause    chan struct{}
	close    chan struct{}
}

func (bp *basicPool) Submit(task Task) (Future, error) {
	rc := make(chan taskResult)

	// check paused or closed
	select {
	case <-bp.close:
		return nil, ErrPoolClosed
	case <-bp.pause:
		return nil, ErrPoolPaused
	}

	bp.taskQ <- &taskWrapper{
		t:       task,
		resChan: rc,
	}

	return newPondFuture(rc), nil
}

func (bp *basicPool) SubmitWithTimeout(task Task, timeout time.Duration) (Future, error) {
	rc := make(chan taskResult)

	// check paused or closed
	select {
	case <-bp.close:
		return nil, ErrPoolClosed
	case <-bp.pause:
		return nil, ErrPoolPaused
	}

	select {
	case <-time.After(timeout):
		return nil, ErrTaskTimeout
	case bp.taskQ <- &taskWrapper{t: task, resChan: rc}:
	}

	return newPondFuture(rc), nil
}
