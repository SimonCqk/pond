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

	// worker will be set to idle after defaultWorkerIdleDuration.
	defaultWorkerIdleDuration = defaultPurgeWorkersDuration / 2

	// default pool capacity is defaultPoolCapacityFactor * NumCPU
	defaultPoolCapacityFactor = 16

	// default task queue size, each worker hold 128 buffered tasks.
	defaultTaskQueueSize = defaultPoolCapacityFactor * 128

	// default task queue capacity.
	defaultTaskQueueCapacity = 1024

	// pool auto expand its capacity when its len(tasksQueue) / cap(taskQueue) equals or
	// greater than autoScaleFactor
	autoScaleFactor = 0.75
)
