package pond

// Worker represents a executor broker for goroutine, do the real job
// and obtained by Pool.
type Worker interface {
	// Run start worker service listening for task coming.
	Run()

	// Running return whether worker is under working.
	Running() bool

	// Cancel cancel current under executing task immediately.
	Cancel()

	// Close close the worker and recycle resource, if it is
	// under working, waiting for task done synchronously.
	Close()
}

type pondWorker struct {
	taskQ  chan Task
	cancel chan struct{}
	close  chan struct{}
}

func newPondWorker() Worker {
	return &pondWorker{
		taskQ:  make(chan Task),
		cancel: make(chan struct{}),
		close:  make(chan struct{}),
	}
}

func (pw *pondWorker) Run() {
	for task := range pw.taskQ {

	}
}

func (pw *pondWorker) Running() bool {

}

func (pw *pondWorker) Cancel() {

}

func (pw *pondWorker) Close() {

}
