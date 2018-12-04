package pond

// Worker represents a executor broker for goroutine, do the real job
// and obtained by Pool.
type Worker interface {
	// Do ask for a new task from Pool instance and execute it.
	Do()

	// Running return whether worker is under working.
	Running() bool

	// Cancel cancel current under executing task immediately.
	Cancel()

	// Close close the worker and recycle resource, if it is
	// under working, waiting for task done synchronously.
	Close()
}
