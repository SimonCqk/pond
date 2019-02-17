package pond

import (
	"context"
	"runtime"
	"time"
)

// FixedFuncPool hold a function and accept different payload
// arguments, do execution and pass the result. See it as a
// executor.
type FixedFuncPool struct {
	pool *basicPool
	f    FixedFunc
}

type FixedFunc func(interface{}) (interface{}, error)

func NewFixedFuncPool(f FixedFunc, cap ...int) *FixedFuncPool {
	cores := runtime.NumCPU()
	bp := &basicPool{
		capacity:      append(cap, defaultPoolCapacityFactor*cores)[0],
		taskQ:         NewResizableChan(defaultTaskQueueCapacity),
		pause:         make(chan struct{}, 1), // make pause buffered
		close:         make(chan struct{}),
		purgeDuration: defaultPurgeWorkersDuration,
		purgeTicker:   time.NewTicker(defaultPurgeWorkersDuration),
	}

	for i := 0; i < bp.capacity; i++ {
		bp.workers = append(bp.workers, newPondWorker(bp.taskQ))
	}
	go bp.purgeWorkers()
	return &FixedFuncPool{
		pool: bp,
		f:    f,
	}
}

func (p *FixedFuncPool) Submit(arg interface{}) (Future, error) {
	// not all callers hold the returned Future, so that there may no
	// receiver side which may cause block when worker send return values.
	rc := make(chan *taskResult, 1)

	// check closed
	select {
	case <-p.pool.close:
		return nil, ErrPoolClosed
	default:
	}

	// check paused
	if len(p.pool.pause) > 0 {
		return nil, ErrPoolPaused
	}

	p.pool.taskQ.In() <- &taskWrapper{
		t:       func() (interface{}, error) { return p.f(arg) },
		resChan: rc,
	}

	p.pool.scale()

	return newPondFuture(rc), nil
}

func (p *FixedFuncPool) SubmitWithTimeout(arg interface{}, timeout time.Duration) (Future, error) {
	// not all callers hold the returned Future, so that there may no
	// receiver side which may cause block when worker send return values.
	rc := make(chan *taskResult, 1)

	// check closed
	select {
	case <-p.pool.close:
		return nil, ErrPoolClosed
	default:
	}

	// check paused
	if len(p.pool.pause) > 0 {
		return nil, ErrPoolPaused
	}

	task := &taskWrapper{
		t:       func() (interface{}, error) { return p.f(arg) },
		resChan: rc,
	}

	select {
	case <-time.After(timeout):
		return nil, ErrTaskTimeout
	case p.pool.taskQ.In() <- task:
	}

	p.pool.scale()

	return newPondFuture(rc), nil
}

func (p *FixedFuncPool) SetCapacity(newCap int) {
	p.pool.SetCapacity(newCap)
}

func (p *FixedFuncPool) SetTaskCapacity(newCap int) {
	p.pool.SetTaskCapacity(newCap)
}

// SetNewFixedFunc dynamically set new fixed function hold inside
// pool, it will paused util old tasks done.
func (p *FixedFuncPool) SetNewFixedFunc(newFunc FixedFunc) {
	// pause the service, equivalent to a LOCK.
	p.Pause()
	empty := make(chan struct{})

	emptyChecker := func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				curLen := p.pool.taskQ.Len()
				if curLen == 0 {
					empty <- struct{}{}
					return
				}
				// control the checking-frequency
				time.Sleep(time.Millisecond * time.Duration(curLen))
			}
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	go emptyChecker(ctx)
	// wait for old tasks done
	<-empty
	cancel()
	// set new function
	p.f = newFunc
	// resume the service
	p.Resume()
}

func (p *FixedFuncPool) Pause() {
	p.pool.Pause()
}

func (p *FixedFuncPool) Resume() {
	p.pool.Resume()
}

func (p *FixedFuncPool) Close() {
	p.pool.Close()
	p.f = nil
}

func (p *FixedFuncPool) SetPurgeDuration(dur time.Duration) {
	p.pool.SetPurgeDuration(dur)
}

func (p *FixedFuncPool) Capacity() int {
	return p.pool.Capacity()
}

func (p *FixedFuncPool) Workers() int {
	return p.pool.Workers()
}
