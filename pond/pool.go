package pond

import (
	"runtime"
	"sync"
	"sync/atomic"
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

type basicPool struct {
	capacity int32
	workers  []Worker
	taskQ    chan *taskWrapper
	pause    chan struct{}
	close    chan struct{}
	lock     sync.RWMutex
	once     sync.Once
}

func newBasicPool(cap ...int32) Pool {
	workersNum := defaultWorkersNumFactor * runtime.NumCPU()
	bp := &basicPool{
		capacity: append(cap, int32(defaultPoolCapacity*runtime.NumCPU()))[0],
		taskQ:    make(chan *taskWrapper),
		pause:    make(chan struct{}, 1),
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
	for {
		select {
		case <-ticker.C:
			// TODO
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

	bp.taskQ <- &taskWrapper{
		t:       task,
		resChan: rc,
	}

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

func (bp *basicPool) Resize(newSize int32) {
	atomic.StoreInt32(&bp.capacity, newSize)

}

func (bp *basicPool) Pause() {
	bp.pause <- struct{}{}
}

func (bp *basicPool) Resume() {
	<-bp.pause
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
