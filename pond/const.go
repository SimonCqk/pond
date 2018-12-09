package pond

import (
	"errors"
	"time"
)

var (
	ErrPoolClosed  = errors.New("pool: pool has been closed, no more tasks submitted")
	ErrPoolPaused  = errors.New("pool: pool has been paused, resume first please")
	ErrTaskTimeout = errors.New("task: task timeout")
)

// constraints for pool
const (
	// purge idle workers to recycle resource every defaultPurgeWorkersDuration seconds.
	defaultPurgeWorkersDuration = 32 * time.Second
	// make task queue a buffered channel, so as to avoid accidently blocking main
	// goroutine when submit, BUT BLOCK MAY HAPPEN SOME TIME.
	defaultTaskBufferSizeFactor = 16
	// default pool capacity is defaultPoolCapacityFactor * NumCPU
	defaultPoolCapacityFactor = 4
	// pool auto expand its capacity when its len(tasksQueue) / cap(taskQueue) equals or
	// greater than autoExpandFactor
	autoExpandFactor = 0.8
)

// constraints for workers
const (
	// if a worker doesn't preempt a task after defaultIdleDuration,
	// it will be flagged as idle.
	defaultIdleDuration = 8 * time.Second
)
