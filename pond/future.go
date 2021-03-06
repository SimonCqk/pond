package pond

import (
	"sync/atomic"
)

// Task represent a task to be executed. No args passed in because it
// can be easily achieved by closure, and return a Future instance to
// get return value.
type Task func() (interface{}, error)

type taskResult struct {
	val interface{}
	err error
}

type taskWrapper struct {
	t       Task
	resChan chan *taskResult
}

// Future associate with a Task instance and can be used to capture
// return value of task.
type Future interface {
	// Value synchronously return the value captured by Future.
	Value() (interface{}, error)

	// Then allows multiple functions chained together, the semantic is
	// provide the next action after this future done, and make the
	// function calls flow like stream.
	Then(next func(interface{}) (interface{}, error)) Future

	// OnSuccess register the callback when future executed successfully.
	// If task done with error, it is a no-op.
	OnSuccess(f func(interface{}))

	// OnFailure register the callback when future done with some error.
	// If task done with success, it is a no-op.
	OnFailure(f func(error))
}

// pond implementation of Future interface.
type pondFuture struct {
	value interface{}
	err   error
	done  chan *taskResult
	ready int32
}

func newPondFuture(doneC chan *taskResult) *pondFuture {
	return &pondFuture{done: doneC}
}

func (pf *pondFuture) Value() (interface{}, error) {
	if atomic.LoadInt32(&pf.ready) != 0 {
		return pf.value, pf.err
	}
	// block for done
	taskRes := <-pf.done

	// double checks. It may have multiply goroutines hanging and waiting
	// for done, when pf.done closed, all these goroutines will be free
	// and ready to get result.
	if atomic.LoadInt32(&pf.ready) != 0 {
		return pf.value, pf.err
	}

	pf.value, pf.err = taskRes.val, taskRes.err
	rscPool.PutTaskResult(taskRes)

	atomic.StoreInt32(&pf.ready, 1)
	close(pf.done)

	return pf.value, pf.err
}

func (pf *pondFuture) Then(next func(interface{}) (interface{}, error)) Future {
	doneC := make(chan *taskResult)
	f := newPondFuture(doneC)
	go func() {
		val, err := pf.Value()
		if err != nil {
			tr := rscPool.GetTaskResult(nil, err)
			doneC <- tr
			return
		}
		doneC <- rscPool.GetTaskResult(next(val))
	}()
	return f
}

func (pf *pondFuture) OnSuccess(f func(interface{})) {
	go func() {
		val, err := pf.Value()
		if err == nil {
			f(val)
		}
	}()
}

func (pf *pondFuture) OnFailure(f func(err error)) {
	go func() {
		_, err := pf.Value()
		if err != nil {
			f(err)
		}
	}()
}
