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

const (
	// purge idle workers to recycle resource every 30s.
	defaultPurgeWorkersDuration = 30 * time.Second
	// make task queue a buffered channel, so as to avoid accidently blocking main
	// goroutine, BUT IT MAY HAPPEN SOME TIME.
	defaultTaskBufferSize = 16
	// default number of goroutines launched when pool initialized is
	// defaultWorkersNumFactor * NumCPU
	defaultWorkersNumFactor = 2
	// default pool capacity is defaultPoolCapacity * NumCPU
	defaultPoolCapacity = 8
)
