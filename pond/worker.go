package pond

import "time"

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
	taskQ *ResizableChan
	close chan struct{}
	idle  bool
}

func newPondWorker(tq *ResizableChan) Worker {
	pw := &pondWorker{
		taskQ: tq,
		close: make(chan struct{}, 1),
		idle:  false,
	}
	go pw.run()
	return pw
}

func (pw *pondWorker) run() {
	timer := time.NewTimer(defaultWorkerIdleDuration)
	defer timer.Stop()

	for {
		select {
		case <-pw.close:
			return
		case task := <-pw.taskQ.Out():
			pw.idle = false

			taskRes := resultPool.Get().(*taskResult)
			taskRes.val, taskRes.err = task.t()
			task.resChan <- taskRes

			// check closing
			select {
			case <-pw.close:
				return
			default:
				timer.Reset(defaultWorkerIdleDuration)
			}
		case <-timer.C:
			pw.idle = true
			timer.Reset(defaultWorkerIdleDuration)
		}
	}
}

func (pw *pondWorker) Idle() bool {
	return pw.idle
}

func (pw *pondWorker) Close() {
	close(pw.close)
}
