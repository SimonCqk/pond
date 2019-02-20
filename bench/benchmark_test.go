package bench

import (
	"math/rand"
	"pond/pond"
	"sync"
	"testing"
	"time"
)

func benchSleepUnit() {
	time.Sleep(5 * time.Millisecond)
}

func benchComputeUnit() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	l := make([]int, 0)
	for i := 0; i < 10000; i++ {
		l = append(l, r.Int())
	}
}

func BenchmarkSleepUnitGo(t *testing.B) {
	wg := sync.WaitGroup{}
	t.StartTimer()
	for i := 0; i < t.N; i++ {
		wg.Add(1)
		go func() {
			benchComputeUnit()
			wg.Done()
		}()
	}
	wg.Wait()
	t.StopTimer()
}

func BenchmarkSleepUnitPool(t *testing.B) {
	p := pond.NewPool(5000)
	defer p.Close()
	t.StartTimer()
	for i := 0; i < t.N; i++ {
		_, _ = p.Submit(func() (i interface{}, e error) {
			benchComputeUnit()
			return nil, nil
		})
	}
	t.StopTimer()
}
