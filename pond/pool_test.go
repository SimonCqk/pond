package pond

import (
	"testing"
	"time"
)

func foo() (interface{}, error) {
	time.Sleep(10 * time.Millisecond)
	return nil, nil
}

func TestBasicPoolSubmit(t *testing.T) {
	pool := NewPool()
	future, _ := pool.Submit(foo)
	_, _ = future.Value()
	pool.Close()
}

func TestBasicPoolBatchSubmit(t *testing.T) {
	pool := NewPool()
	var future Future
	for i := 0; i < 10; i++ {
		future, _ = pool.Submit(foo)
		_, _ = future.Value()
	}
	pool.Close()
}

func TestBasicPoolPause(t *testing.T) {
	pool := NewPool()
	_, _ = pool.Submit(foo)
	pool.Pause()
	_, err := pool.Submit(foo)
	if err == nil || err != ErrPoolPaused {
		t.Error("pool has paused, no more tasks submitted!")
	}
	pool.Resume()
	_, err = pool.Submit(foo)
	if err != nil {
		t.Error("pool has resumed and should works well.")
	}
	pool.Close()
}

func TestBasicPoolClose(t *testing.T) {
	pool := NewPool()
	pool.Close()
	_, err := pool.Submit(foo)
	if err == nil || err != ErrPoolClosed {
		t.Error("pool has closed, no more task submitted and should return ErrPoolClosed")
	}
}

func TestBasicPoolSetCapacity(t *testing.T) {
	pool := NewPool()
	pool.SetCapacity(1)
	if pool.Workers() != 1 {
		t.Error("new number of workers should be 1")
	}
	pool.SetCapacity(10)
	if pool.Workers() != 10 {
		t.Error("new number of workers should be 10")
	}
	pool.SetCapacity(5)
	if pool.Workers() != 5 {
		t.Error("new number of workers should be 5")
	}
	pool.Close()
}
