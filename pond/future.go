package pond

// Future associate with a Task instance and can be used to capture
// return value of task.
type Future interface {
	// Value synchronously return the value captured by Future.
	Value() (interface{}, error)

	// Then allows multiple functions chained together, the semantic is
	// provide the next action after this future done, and make the
	// function calls flow like stream.
	Then(next func(interface{}) (interface{}, error)) Future

	// OnSuccess provide the callback when future executed successfully.
	// If task done with error, it is a no-op.
	OnSuccess(f func(interface{}))

	// OnFailure provide the callback when future done with some error.
	// If task done with success, it is a no-op.
	OnFailure(f func(error))
}

// pond implementation of Future interface.
type pondFuture struct {
	value interface{}
	err   error
	done  chan struct{}
}

func newPondFuture(next func() (interface{}, error)) *pondFuture {
	f := pondFuture{done: make(chan struct{})}
	go func() {
		f.value, f.err = next()
		f.done <- struct{}{}
	}()
	return &f
}

func (pf *pondFuture) Value() (interface{}, error) {
	// block for done
	<-pf.done
	return pf.value, pf.err
}

func (pf *pondFuture) Then(next func(interface{}) (interface{}, error)) Future {
	return newPondFuture(func() (interface{}, error) {
		val, err := pf.Value()
		if err != nil {
			return val, err
		}
		return next(val)
	})
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
