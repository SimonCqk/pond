package pond

import (
	"fmt"
	"testing"
	"time"
)

func TestPondWorkerIdle(t *testing.T) {
	taskQueue := make(chan *taskWrapper)
	worker := newPondWorker(taskQueue)
	resChan := make(chan *taskResult, 1)
	sig := make(chan struct{})

	go func() {
		tw := &taskWrapper{
			t: func() (interface{}, error) {
				fmt.Println("This is a task")
				return nil, nil
			},
			resChan: resChan,
		}
		go func() {
			// avoid blocking when
			<-resChan
			sig <- struct{}{}
		}()
		taskQueue <- tw

		fmt.Println("sleep for [defaultIdleDuration+1] seconds...")
		time.Sleep(defaultIdleDuration + time.Second)

		sig <- struct{}{}
	}()

	select {
	case <-sig:
		if worker.Idle() {
			t.Fatal("worker should be under working when finish 1st task.\n")
		}
	}

	select {
	case <-sig:
		if !worker.Idle() {
			t.Fatal("worker should be idle when after IdleDuration.\n")
		}
	}
}
