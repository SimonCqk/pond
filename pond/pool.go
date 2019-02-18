package pond

import (
	"runtime"
	"sync"
	"time"
)

func init() {
	resultPool = &sync.Pool{
		New: func() interface{} {
			return &taskResult{}
		},
	}
}

var resultPool *sync.Pool

// Pool interface defines the critical methods a pool must implement,
// it also represents the main methods exposed to users.
type Pool interface {
	// Submit is the main entry for submitting new tasks.
	Submit(task Task) (Future, error)

	// SubmitWithTimeout submit a new task and set expiration.
	SubmitWithTimeout(task Task, timeout time.Duration) (Future, error)

	// SetCapacity dynamically reset the capacity(number of workers) of pool.
	SetCapacity(newCap int)

	// SetTaskCapacity dynamically reset the capacity of task queue.
	SetTaskCapacity(newCap int)

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
	capacity      int
	workers       []Worker
	taskQ         *TaskQueue
	pause         chan struct{}
	close         chan struct{}
	mu            sync.RWMutex
	purgeDuration time.Duration
	purgeTicker   *time.Ticker
}

func newBasicPool(cap ...int) *basicPool {
	cores := runtime.NumCPU()
	bp := &basicPool{
		capacity:      append(cap, defaultPoolCapacityFactor*cores)[0],
		taskQ:         NewTaskQueue(defaultTaskQueueCapacity),
		pause:         make(chan struct{}, 1), // make pause buffered
		close:         make(chan struct{}),
		purgeDuration: defaultPurgeWorkersDuration,
		purgeTicker:   time.NewTicker(defaultPurgeWorkersDuration),
	}

	for i := 0; i < bp.capacity; i++ {
		bp.workers = append(bp.workers, newPondWorker(bp.taskQ))
	}
	go bp.purgeWorkers()
	return bp
}

// purgeWorkers purge idle workers periodically and recycle resource.
func (bp *basicPool) purgeWorkers() {
	for {
		select {
		case <-bp.close:
			return
		case <-bp.purgeTicker.C:
			bp.mu.Lock()
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
			bp.mu.Unlock()
		default:
			bp.tryPause()
		}
	}
}

func (bp *basicPool) Submit(task Task) (Future, error) {
	// not all callers hold the returned Future, so that there may no
	// receiver side which may cause block when worker send return values.
	rc := make(chan *taskResult, 1)

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

	bp.taskQ.In() <- &taskWrapper{t: task, resChan: rc}

	bp.scale()

	return newPondFuture(rc), nil
}

func (bp *basicPool) SubmitWithTimeout(task Task, timeout time.Duration) (Future, error) {
	// not all callers hold the returned Future, so that there may no
	// receiver side which may cause block when worker send return values.
	rc := make(chan *taskResult, 1)

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
	case bp.taskQ.In() <- &taskWrapper{t: task, resChan: rc}:
	}

	bp.scale()

	return newPondFuture(rc), nil
}

func (bp *basicPool) SetCapacity(newCap int) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	curCap := len(bp.workers)
	if curCap == newCap {
		return
	}

	if curCap < newCap {
		for i := curCap; i < newCap; i++ {
			bp.workers = append(bp.workers, newPondWorker(bp.taskQ))
		}
		return
	}

	if curCap > newCap {
		for i := newCap; i < curCap; i++ {
			bp.workers[i].Close()
			bp.workers[i] = nil
		}
		bp.workers = bp.workers[:newCap]
	}
}

func (bp *basicPool) SetTaskCapacity(newCap int) {
	if newCap == bp.taskQ.Size() {
		return
	}
	bp.taskQ.Resize(newCap)
}

func (bp *basicPool) Pause() {
	bp.pause <- struct{}{}
}

// tryPause judge if pool is in paused state, if so, stop
// current action and wait for resume.
func (bp *basicPool) tryPause() {
	if _, open := <-bp.close; open && len(bp.pause) > 0 {
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
	bp.mu.Lock()
	defer bp.mu.Unlock()

	close(bp.close)
	close(bp.pause)

	// clear workers
	for _, worker := range bp.workers {
		worker.Close()
	}
	bp.workers = nil

	bp.taskQ.Close()
	bp.purgeTicker.Stop()
}

// Capacity return current capacity of pool.
func (bp *basicPool) Capacity() int {
	bp.mu.RLock()
	defer bp.mu.RUnlock()
	return bp.capacity
}

// Workers return current number of under working workers.
func (bp *basicPool) Workers() int {
	bp.mu.RLock()
	defer bp.mu.RUnlock()
	return len(bp.workers)
}

// SetPurgeDuration set duration of pool recycling its idle workers.
func (bp *basicPool) SetPurgeDuration(dur time.Duration) {
	if dur != bp.purgeDuration {
		bp.mu.Lock()
		bp.purgeDuration = dur
		bp.purgeTicker = time.NewTicker(dur)
		bp.mu.Unlock()
	}
}

// scale expand number of workers when too many tasks accumulated.
func (bp *basicPool) scale() {
	if float32(bp.taskQ.Size())*autoScaleFactor > float32(bp.taskQ.Len()) {
		bp.SetCapacity(bp.capacity + bp.capacity/2)
	}
	// purgeWorkers() response for shrinking.
}
