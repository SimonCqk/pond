package pond

import "github.com/Simoncqk/go-containers"

type ResizableChan struct {
	in, out  chan *taskWrapper
	resizeCh chan int
	buffer   *containers.Queue
	size     int
}

func NewResizableChan(initSize int) *ResizableChan {
	ch := &ResizableChan{
		in:       make(chan *taskWrapper),
		out:      make(chan *taskWrapper),
		resizeCh: make(chan int),
		buffer:   containers.NewQueue(),
		size:     initSize,
	}
	go ch.autoResize()
	return ch
}

func (ch *ResizableChan) In() chan<- *taskWrapper {
	return ch.in
}

func (ch *ResizableChan) Out() <-chan *taskWrapper {
	return ch.out
}

func (ch *ResizableChan) Resize(size int) {
	if size == ch.size {
		return
	}
	if size <= 0 {
		panic("invalid size of ResizableChan")
	}
	ch.resizeCh <- size
}

func (ch *ResizableChan) Size() int {
	return ch.size
}

func (ch *ResizableChan) Len() int {
	return ch.buffer.Size()
}

func (ch *ResizableChan) Close() {
	close(ch.in)
}

func (ch *ResizableChan) autoResize() {

	var (
		input, output chan *taskWrapper
		nextTask      *taskWrapper
	)

	input = ch.in

	for input != nil || output != nil {
		select {
		case income, open := <-input:
			if open {
				ch.buffer.Push(income)
			} else {
				input = nil
			}
		case output <- nextTask:
			ch.buffer.Pop()
		case ch.size = <-ch.resizeCh:
		}

		// no more outgoing
		if ch.buffer.Size() == 0 {
			output = nil
			nextTask = nil
		} else {
			output = ch.out
			nextTask = ch.buffer.Peek().(*taskWrapper)
		}
		// ch is full
		if ch.buffer.Size() >= ch.size {
			input = nil
		} else {
			input = ch.in
		}
	}

	close(ch.out)
	close(ch.resizeCh)
}
