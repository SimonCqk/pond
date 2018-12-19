package pond

import (
	"fmt"
	"github.com/pkg/errors"
	"testing"
	"time"
)

func TestFutureValue(t *testing.T) {
	f := func() (interface{}, error) {
		s := "value"
		return s, nil
	}
	c := make(chan taskResult, 1)
	future := newPondFuture(c)
	go func() {
		time.Sleep(100 * time.Millisecond)
		val, err := f()
		c <- taskResult{val: val, err: err}
	}()

	val, err := future.Value()
	fmt.Printf("future: %v, %v\n", val, err)
}

func TestFutureOnSuccess(t *testing.T) {
	f := func() (interface{}, error) {
		s := "future"
		return s, nil
	}
	c := make(chan taskResult, 1)
	future := newPondFuture(c)

	future.OnSuccess(func(i interface{}) {
		fmt.Println("success with ", i)
	})

	go func() {
		time.Sleep(100 * time.Millisecond)
		val, err := f()
		c <- taskResult{val: val, err: err}
	}()

	val, err := future.Value()
	time.Sleep(100 * time.Millisecond)
	fmt.Printf("on success: %v, %v\n", val, err)
}

func TestFutureOnFailure(t *testing.T) {
	f := func() (interface{}, error) {
		s := "future"
		return s, errors.New("failure")
	}
	c := make(chan taskResult, 1)
	future := newPondFuture(c)

	future.OnFailure(func(err error) {
		fmt.Println("error: ", err)
	})

	go func() {
		time.Sleep(100 * time.Millisecond)
		val, err := f()
		c <- taskResult{val: val, err: err}
	}()

	val, err := future.Value()
	time.Sleep(100 * time.Millisecond)
	fmt.Printf("on failure: %v, %v\n", val, err)
}

func TestFutureThen(t *testing.T) {
	f := func() (interface{}, error) {
		s := "future"
		return s, nil
	}
	c := make(chan taskResult, 1)
	future := newPondFuture(c)

	f1 := future.Then(func(i interface{}) (interface{}, error) {
		fmt.Printf("1. %v\n", i)
		return i, nil
	})
	f2 := f1.Then(func(i interface{}) (interface{}, error) {
		fmt.Printf("2. %v\n", i)
		return i, nil
	})
	f3 := f2.Then(func(i interface{}) (interface{}, error) {
		fmt.Printf("3. %v\n", i)
		return nil, errors.New("stop")
	})

	go func() {
		val, err := f()
		c <- taskResult{val: val, err: err}
	}()

	val, err := f3.Value()
	fmt.Printf("finally: %v, %v\n", val, err)
}
