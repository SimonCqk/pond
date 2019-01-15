package pond

import (
	"testing"
)

func fooPow(num interface{}) (i interface{}, err error) {
	n := num.(int)
	return n * n, nil
}

func TestFixedFuncPoolSubmit(t *testing.T) {
	pool := NewFixedFuncPool(fooPow)
	future, _ := pool.Submit(2)
	val, _ := future.Value()
	if val = val.(int); val != 4 {
		t.Error("execution result should be 4!")
	}
	pool.Close()
}

func TestFixedFuncPoolSetNewFixedFunc(t *testing.T) {
	pool := NewFixedFuncPool(fooPow)
	future, _ := pool.Submit(2)
	val, _ := future.Value()
	if val = val.(int); val != 4 {
		t.Error("execution result should be 4!")
	}
	newFunc := func(num interface{}) (i interface{}, err error) {
		n := num.(int)
		return n * n * n, nil
	}
	pool.SetNewFixedFunc(newFunc)
	future, _ = pool.Submit(2)
	val, _ = future.Value()
	if val = val.(int); val != 8 {
		t.Error("execution result should be 8!")
	}
	pool.Close()
}
