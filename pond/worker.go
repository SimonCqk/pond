package pond

import (
	"time"
)

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
	// taskQ is a replication of Pool.taskQ, workers preempt for tasks
	// over Pool.taskQ, if no more task comes, worker will be asleep.
	taskQ  chan *taskWrapper
	cancel chan struct{}
	close  chan struct{}
	idle   bool
}

func newPondWorker(tq chan *taskWrapper) Worker {
	pw := &pondWorker{
		taskQ: tq,
		// when other goroutine call Cancel, it will not block
		cancel: make(chan struct{}, 1),
		close:  make(chan struct{}),
		idle:   false,
	}
	go pw.run()
	return pw
}

func (pw *pondWorker) run() {
	timer := time.NewTimer(defaultIdleDuration)

	for {
		select {
		case <-pw.close:
			return
		case <-pw.cancel:
			// if cancel signal arrive earlier than task, then deprecate the
			// coming task, else if no pending task now, it's a no-op.
			_, _ = <-pw.taskQ
		case task := <-pw.taskQ:
			pw.idle = false
			val, err := task.t()
			task.resChan <- taskResult{val: val, err: err}
			timer.Reset(defaultIdleDuration)
		case <-timer.C:
			pw.idle = true
			timer.Reset(defaultIdleDuration)
		}
	}
}

func (pw *pondWorker) Idle() bool {
	return pw.idle
}

func (pw *pondWorker) Close() {
	close(pw.close)
}
