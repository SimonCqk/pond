package pond

import "sync"

type resourcePool struct {
	resPool  *sync.Pool
	taskPool *sync.Pool
}

var rscPool *resourcePool

func (p *resourcePool) init() {
	p.resPool = &sync.Pool{
		New: func() interface{} {
			return &taskResult{}
		},
	}
	p.taskPool = &sync.Pool{
		New: func() interface{} {
			return &taskWrapper{}
		},
	}
}

func (p *resourcePool) GetTaskResult(val interface{}, err error) *taskResult {
	res := p.resPool.Get().(*taskResult)
	res.val, res.err = val, err
	return res
}

func (p *resourcePool) PutTaskResult(res *taskResult) {
	p.resPool.Put(res)
}

func (p *resourcePool) GetTask(t Task, resChan chan *taskResult) *taskWrapper {
	task := p.taskPool.Get().(*taskWrapper)
	task.t, task.resChan = t, resChan
	return task
}

func (p *resourcePool) PutTask(task *taskWrapper) {
	p.taskPool.Put(task)
}

func init() {
	rscPool = &resourcePool{}
	rscPool.init()
}
