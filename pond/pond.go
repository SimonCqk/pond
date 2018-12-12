package pond

// pond package level default pool constructor
// NewPool return a new basicPool instance
func NewPool(cap ...int) *basicPool {
	return newBasicPool(cap...)
}
