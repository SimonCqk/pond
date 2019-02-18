package pond

// pond package level default pool constructor
// NewPool return a new basicPool instance
func NewPool(cap ...int) Pool {
	return newBasicPool(nil, cap...)
}

func NewFixedSizePool(cap, maxTasks int) Pool {
	return newFixedSizePool(cap, maxTasks, nil)
}

func NewFixedFuncPool(fixedFunc FixedFunc, cap ...int) *FixedFuncPool {
	return newFixedFuncPool(fixedFunc, nil, cap...)
}

// NewCustomizedWorkerPool create a pool with user-customized worker
// implementation, as user implement all asked interface.
func NewCustomizedWorkerPool(wc WorkerCtor, cap ...int) Pool {
	return newBasicPool(wc, cap...)
}
