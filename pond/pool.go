package pond

import (
	"runtime"
	"sync"
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
	workers  []Worker
	taskQ    chan *taskWrapper
	pause    chan struct{}
	close    chan struct{}
	lock     sync.RWMutex
	once     sync.Once
}

func newBasicPool(cap ...int) Pool {
	workersNum := defaultWorkersNumFactor * runtime.NumCPU()
	bp := &basicPool{
		capacity: append(cap, defaultPoolCapacityFactor*runtime.NumCPU())[0],
		taskQ:    make(chan *taskWrapper, defaultTaskBufferSizeFactor*runtime.NumCPU()),
		pause:    make(chan struct{}, 1), // make pause buffered
		close:    make(chan struct{}),
	}
	for i := 0; i < workersNum; i++ {
		bp.workers = append(bp.workers, newPondWorker(bp.taskQ))
	}
	go bp.purgeWorkers()
	return bp
}

// purgeWorkers purge idle workers periodically and recycle resource.
func (bp *basicPool) purgeWorkers() {
	ticker := time.NewTicker(defaultPurgeWorkersDuration)
	defer ticker.Stop()

	for {
		select {
		case <-bp.close:
			return
		case <-ticker.C:
			bp.lock.Lock()
			beg, end := 0, len(bp.workers)-1
			for beg < end {
				if bp.workers[beg].Idle() {
					// swap idle worker to the back
					bp.workers[beg], bp.workers[end] = bp.workers[end], bp.workers[beg]
					bp.workers[end].Close()
					bp.workers[end] = nil
					end--
				}
				if !bp.workers[beg].Idle() {
					beg++
				}
			}
			// handle edge case
			if bp.workers[end].Idle() {
				bp.workers = bp.workers[:end]
			} else {
				bp.workers = bp.workers[:end+1]
			}
			bp.lock.Unlock()
		default:
			bp.makePause()
		}
	}
}

func (bp *basicPool) Submit(task Task) (Future, error) {
	rc := make(chan taskResult)

	// check closed
	select {
	case <-bp.close:
		return nil, ErrPoolClosed
	default:
	}

	// check paused
	if len(bp.pause) > 0 {
		return nil, ErrPoolPaused
	}

	bp.taskQ <- &taskWrapper{t: task, resChan: rc}

	return newPondFuture(rc), nil
}

func (bp *basicPool) SubmitWithTimeout(task Task, timeout time.Duration) (Future, error) {
	rc := make(chan taskResult)

	// check closed
	select {
	case <-bp.close:
		return nil, ErrPoolClosed
	default:
	}

	// check paused
	if len(bp.pause) > 0 {
		return nil, ErrPoolPaused
	}

	select {
	case <-time.After(timeout):
		return nil, ErrTaskTimeout
	case bp.taskQ <- &taskWrapper{t: task, resChan: rc}:
	}

	return newPondFuture(rc), nil
}

func (bp *basicPool) Resize(newSize int) {
	bp.lock.Lock()
	defer bp.lock.Unlock()

	size := len(bp.workers)
	if size == newSize {
		return
	}

	if size < newSize {
		for i := size; i < newSize; i++ {
			bp.workers = append(bp.workers, newPondWorker(bp.taskQ))
		}
		return
	}

	if size > newSize {
		for i := newSize; i < size; i++ {
			bp.workers[i].Close()
			bp.workers[i] = nil
		}
		bp.workers = bp.workers[:newSize]
	}
}

func (bp *basicPool) Pause() {
	bp.pause <- struct{}{}
}

// makePause judge if pool is in paused state, if so, stop
// current action and wait for resume.
func (bp *basicPool) makePause() {
	if len(bp.pause) > 0 {
		// try send signal and it will block because buffer size
		// of pause channel is 1
		bp.pause <- struct{}{}
	}
}

func (bp *basicPool) Resume() {
	// clear pause signals
	for len(bp.pause) > 0 {
		<-bp.pause
	}
}

func (bp *basicPool) Close() {
	bp.once.Do(func() {
		close(bp.close)
		close(bp.pause)

		// clear workers
		for _, worker := range bp.workers {
			worker.Close()
		}
		bp.workers = nil

		close(bp.taskQ)
	})
}
