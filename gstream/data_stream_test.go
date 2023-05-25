package gstream

import (
	"context"
	"fmt"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewDataSourceOf(t *testing.T) {
	dataStream := NewDataStreamOf(context.TODO(), []int{1, 2, 3})
	sink := NewMemorySink[int]()
	dataStream.Via(NewFilter(func(ctx context.Context, i int) bool { return i < 2 }, 1)).To(sink)
	err := dataStream.State().Wait()
	if err != nil {
		t.Fatal(err)
	}
	for _, i := range sink.Result() {
		t.Log(i + 0)
	}
}

func TestNewDataSource(t *testing.T) {
	output := &tickerOutlet{}
	output.init()

	dataStream := NewDataStream(context.TODO(), output)

	a := &A{}
	transfers := FanOut(dataStream.Via(NewFlatMap(a.messageToStrs, 1)), 2)
	var sinks []*stdoutSink
	for i, transfer := range transfers {
		if i == 0 {
			transfer = transfer.Via(NewFlatMap(func(ctx context.Context, s string) []string { return []string{s + "f"} }, 1))
		}
		sink := newStdoutSink(i)
		sinks = append(sinks, sink)
		transfer.To(sink)
	}

	err := dataStream.State().Wait()
	if err.Error() != "test err" {
		t.Fatal(dataStream.State().error())
	}

	for _, sink := range sinks {
		if atomic.LoadInt64(&sink.count) != 4 {
			t.Fatal(atomic.LoadInt64(&sink.count))
		}
	}

}

type A struct {
	lines []string
}

func (a *A) messageToStrs(ctx context.Context, item *message) []string {
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
	ticker := time.NewTicker(1 * time.Second)
	oc := ticker.C
	nc := make(chan interface{})
	go func() {
		i := 0
		for range oc {
			i++
			if to.HasStateErr() {
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
			fmt.Printf("sink%d-%s\n", stdout.i, elem)
			atomic.AddInt64(&stdout.count, 1)
			if elem == "3" {
				stdout.SetStateErr(fmt.Errorf("test err"))
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
