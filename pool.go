package pond

import (
	"github.com/pkg/errors"
	"time"
)

var (
	ErrPoolClosed  = errors.New("Pool has been closed, no more tasks submitted.")
	ErrPoolPaused  = errors.New("Pool has been paused, resume first please.")
	ErrTaskTimeout = errors.New("Task timeout.")
	ErrTaskNotDone = errors.New("Task has not been done properly.")
)

// Task represent a task to be executed. No args passed in because it
// can be easily achieved by closure, and return a Promise instance to
// get return value.
type Task func()

// Pool interface defines the critical methods a pool must implement,
// it also represents the main methods exposed to users.
type Pool interface {
	// Submit is the main entry for submitting new tasks.
	Submit(task Task) error

	// SubmitWithTimeout submit a new task and set expiration.
	SubmitWithTimeout(task Task, timeout time.Duration) error

	// Resize dynamically reset the capacity of pool.
	Resize(newSize int32)

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
