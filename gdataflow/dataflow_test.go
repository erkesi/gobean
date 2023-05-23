package gdataflow

import (
	"fmt"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewDataSource(t *testing.T) {
	output := &tickerOutlet{}
	output.init()
	source := NewDataSource(output)
	flows := FanOut(source.Via(NewFlatMap(messageToStrs, 1)), 2)
	var sinks []*stdoutSink
	for i, flow := range flows {
		if i == 0 {
			flow = flow.Via(NewFlatMap(func(s string) []string { return []string{s + "f"} }, 1))
		}
		sink := newStdoutSink(i)
		sinks = append(sinks, sink)
		flow.To(sink)
	}
	source.State().Wait()
	for _, sink := range sinks {
		if atomic.LoadInt64(&sink.count) != 4 {
			t.Fatal(atomic.LoadInt64(&sink.count))
		}
	}
	if source.State().Err().Error() != "test err" {
		t.Fatal(source.State().Err())
	}
}

func messageToStrs(item *message) []string {
	return []string{item.content}
}

type tickerOutlet struct {
	FlowState
	out <-chan interface{}
}

func (to *tickerOutlet) Out() <-chan interface{} {
	return to.out
}

type message struct {
	content string
}

func (to *tickerOutlet) init() {
	ticker := time.NewTicker(10 * time.Millisecond)
	oc := ticker.C
	nc := make(chan interface{})
	go func() {
		i := 0
		for range oc {
			i++
			if to.State().IsErr() {
				nc <- &message{content: "err end"}
				close(nc)
				return
			}
			if i == 5 {
				nc <- &message{content: "finish"}
				close(nc)
				return
			}
			nc <- &message{strconv.FormatInt(int64(i), 10)}
		}
	}()
	to.out = nc
}

type stdoutSink struct {
	FlowState
	in    chan interface{}
	i     int
	count int64
}

// NewStdoutSink returns a new stdoutSink instance.
func newStdoutSink(i int) *stdoutSink {
	sink := &stdoutSink{
		in: make(chan interface{}),
		i:  i,
	}
	sink.init()
	return sink
}

func (stdout *stdoutSink) init() {
	go func() {
		for elem := range stdout.in {
			fmt.Printf("%d-%s\n", stdout.i, elem)
			atomic.AddInt64(&stdout.count, 1)
			if elem == "3" {
				stdout.State().SetErr(fmt.Errorf("test err"))
				continue
			}
		}
		stdout.Done()
	}()
}

// In returns an input channel for receiving data
func (stdout *stdoutSink) In() chan<- interface{} {
	return stdout.in
}
