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

	// default pool capacity is defaultPoolCapacityFactor * NumCPU
	defaultPoolCapacityFactor = 4

	// pool auto expand its capacity when its len(tasksQueue) / cap(taskQueue) equals or
	// greater than autoExpandFactor
	autoExpandFactor = 0.75
)
