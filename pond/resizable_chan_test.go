package pond

import (
	"testing"
	"time"
)

func TestResizableChan(t *testing.T) {
	ch := NewTaskQueue(3)
	for i := 0; i < 3; i++ {
		ch.In() <- &taskWrapper{}
	}

	select {
	case ch.In() <- &taskWrapper{}:
		t.Error("current size of chan is 3, no more input")
	default:
	}

	if ch.Size() != 3 {
		t.Error("size of ch should be 3")
	}
	ch.Resize(4)
	time.Sleep(time.Millisecond)

	select {
	case ch.In() <- &taskWrapper{}:
	default:
		t.Error("current size of chan is 4, input should work")
	}

	if ch.Size() != 4 {
		t.Error("size of ch should be 4")
	}
	// pop 2 elements
	for i := 0; i < 2; i++ {
		<-ch.Out()
	}
	ch.Resize(2)
	time.Sleep(time.Millisecond)

	select {
	case ch.In() <- &taskWrapper{}:
		t.Error("current size of chan is 2, no more input")
	default:
	}

	if ch.Size() != 2 {
		t.Error("size of ch should be 2")
	}
	ch.Close()
}
