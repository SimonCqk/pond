package pond

import "time"

// FixedSizePool has a fixed capacity and task queue length, once
// initialized, no more modification allowed over this two members.
type FixedSizePool struct {
	*basicPool
}

func newFixedSizePool(cap, maxTasks int, wc WorkerCtor) *FixedSizePool {
	bp := &basicPool{
		capacity:      cap,
		taskQ:         make(chan *taskWrapper, maxTasks),
		pause:         make(chan struct{}, 1), // make pause buffered
		close:         make(chan struct{}),
		purgeDuration: defaultPurgeWorkersDuration,
		purgeTicker:   time.NewTicker(defaultPurgeWorkersDuration),
	}
	if wc == nil {
		for i := 0; i < bp.capacity; i++ {
			bp.workers = append(bp.workers, newPondWorker(bp.taskQ))
		}
	} else {
		for i := 0; i < bp.capacity; i++ {
			bp.workers = append(bp.workers, wc(bp.taskQ))
		}
	}
	go bp.purgeWorkers()
	return &FixedSizePool{bp}
}

// SetCapacity do nothing, for overriding the SetCapacity impl of
// basicPool.
func (p *FixedSizePool) SetCapacity(newCap int) {}
