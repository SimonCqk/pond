package pond

import (
	"time"
)

// Worker represents a executor broker for goroutine, do the real job
// and obtained by Pool.
type Worker interface {
	// Init do some initial working before worker launch.
	Init()

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
	// taskQ is a replication of Pool.taskQ, workers preempt tasks over
	// task queue, and it is the main communicate entry for workers and
	// the pool.
	taskQ chan *taskWrapper
	close chan struct{}
	idle  bool
}

// WorkCtor is a worker constructor and return a new worker instance,
// Workers preempt tasks over task queue, and it is the main communicate
// entry for workers and the pool.
type WorkerCtor func(tq chan *taskWrapper) Worker

func newPondWorker(tq chan *taskWrapper) Worker {
	pw := &pondWorker{
		taskQ: tq,
		close: make(chan struct{}, 1),
		idle:  false,
	}
	go pw.run()
	return pw
}

func (pw *pondWorker) Init() {}

func (pw *pondWorker) run() {
	timer := time.NewTimer(defaultWorkerIdleDuration)
	defer timer.Stop()

	for {
		select {
		case <-pw.close:
			return
		case task := <-pw.taskQ:
			if task == nil {
				continue
			}
			pw.idle = false

			task.resChan <- rscPool.GetTaskResult(task.t())
			rscPool.PutTask(task)

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
