package pond

// Future associate with a Task instance and can be used to capture
// return value of task.
type Future interface {
	// Value return the value captured by Future.
	Value() (interface{}, error)

	// Then allows multiple functions chained together, the semantic is
	// provide the next action after this future done, and make the
	// function calls flow like stream.
	Then(next func(interface{}) (interface{}, error)) Future

	// OnSuccess provide the callback when future executed successfully.
	OnSuccess(f func(interface{})) Future

	// OnFailure provide the callback when future done with some error.
	OnFailure(f func(err error)) Future
}

type pondFuture struct {
	value interface{}
	err   error
	done  chan struct{}
}
