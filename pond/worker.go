package pond

// Worker represents a executor broker for goroutine, do the real job
// and obtained by Pool.
type Worker interface {
	// run start worker service listening for task coming, it should be
	// invoked instantly when worker created.
	run()

	// Running return whether worker is under working.
	Running() bool

	// Cancel cancel next to-be-executed task immediately. If no pending
	// task now, it is a no-op.
	Cancel()

	// Close close the worker and recycle resource, if it is
	// under working, waiting for task done synchronously.
	Close()
}

type pondWorker struct {
	// taskQ is a replication of Pool.taskQ, workers preempt for tasks
	// over Pool.taskQ, if no more task comes, worker will be asleep.
	taskQ  chan *taskWrapper
	cancel chan struct{}
	close  chan struct{}
	idle   bool
}

func newPondWorker(tq chan *taskWrapper) Worker {
	pw := &pondWorker{
		taskQ:  tq,
		cancel: make(chan struct{}),
		close:  make(chan struct{}),
		idle:   true,
	}
	pw.run()
	return pw
}

func (pw *pondWorker) run() {
	for {
		select {
		case <-pw.close:
			return
		case <-pw.cancel:
			// if cancel signal arrive earlier than task, then deprecate the
			// coming task, else if no pending task now, do nothing cause we
			// will be failed to read taskQ.
			_, _ = <-pw.taskQ
		case task := <-pw.taskQ:
			pw.idle = false
			val, err := task.t()
			task.resChan <- taskResult{val: val, err: err}
			pw.idle = true
		}
	}
}

func (pw *pondWorker) Running() bool {
	return !pw.idle
}

func (pw *pondWorker) Cancel() {
	pw.cancel <- struct{}{}
}

func (pw *pondWorker) Close() {
	close(pw.close)
}
