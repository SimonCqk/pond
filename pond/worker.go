package pond

// Worker represents a executor broker for goroutine, do the real job
// and obtained by Pool.
type Worker interface {
	// run start worker service listening for task coming, it should be
	// invoked instantly when worker created.
	run()

	// Idle return whether worker is in long-idle state which indicate
	// can be recycled.
	Idle() bool

	// Close close the worker and recycle resource, if it is
	// under working, waiting for task done synchronously.
	Close()
}

type pondWorker struct {
	// taskQ is a replication of Pool.taskQ, pool will dispatch task to
	// idle workers, otherwise worker will be asleep.
	taskQ chan chan *taskWrapper
	selfQ chan *taskWrapper
	close chan struct{}
	idle  bool
}

func newPondWorker(tq chan chan *taskWrapper) Worker {
	pw := &pondWorker{
		taskQ: tq,
		selfQ: make(chan *taskWrapper),
		close: make(chan struct{}, 1),
		idle:  false,
	}
	tq <- pw.selfQ
	go pw.run()
	return pw
}

func (pw *pondWorker) run() {

	for {
		select {
		case <-pw.close:
			return
		case task := <-pw.selfQ:
			pw.idle = false

			taskRes := resultPool.Get().(*taskResult)
			taskRes.val, taskRes.err = task.t()
			task.resChan <- taskRes

			// check closing
			select {
			case <-pw.close:
				return
			default:
				// resend self queue into pool's task queue for dispatching.
				pw.taskQ <- pw.selfQ
			}

			pw.idle = true
		}
	}
}

func (pw *pondWorker) Idle() bool {
	return pw.idle
}

func (pw *pondWorker) Close() {
	close(pw.close)
}
