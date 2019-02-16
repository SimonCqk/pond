package pond

import (
	"fmt"
	"testing"
	"time"
)

func TestResizableChan(t *testing.T) {
	ch := NewResizableChan(3)
	for i := 0; i < 3; i++ {
		ch.In() <- &taskWrapper{}
	}
	go func() {
		select {
		case ch.In() <- &taskWrapper{}:
			t.Error("current size of chan is 3, no more input")
		default:
		}
	}()
	time.Sleep(100 * time.Millisecond)
	if ch.Len() != 3 {
		t.Error("len of ch should be 3")
	}
	ch.Resize(4)
	go func() {
		select {
		case ch.In() <- &taskWrapper{}:
		default:
			t.Error("current size of chan is 4, input should work")
		}
	}()
	time.Sleep(100 * time.Millisecond)
	if ch.Len() != 4 {
		t.Error("len of ch should be 4")
	}
	// pop 2 elements
	for i := 0; i < 2; i++ {
		<-ch.Out()
	}
	ch.Resize(2)
	go func() {
		select {
		case ch.In() <- &taskWrapper{}:
			t.Error("current size of chan is 2, no more input")
		default:
		}
	}()
	fmt.Println(ch.Len())
	if ch.Len() != 2 {
		t.Error("len of ch should be 2")
	}
	ch.Close()
}
