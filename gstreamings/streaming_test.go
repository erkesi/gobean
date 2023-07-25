package gstreamings

import (
	"context"
	"fmt"
	"github.com/erkesi/gobean/gerrors"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewStreamingOfPanic1(t *testing.T) {
	sink := NewMemorySink[int]()
	err := NewStreamingOfSlice(context.TODO(), []int{1, 2, 3}).
		Via(NewFilter(func(ctx context.Context, i int) (bool, error) {
			if i < 2 {
				return true, nil
			}
			panic("panic-2")
		})).To(sink).Wait()
	if pErr, ok := err.(*gerrors.PanicError); !ok {
		t.Fatal(pErr)
	} else if !strings.Contains(pErr.Error(), "panic-2") {
		t.Fatal(pErr)
	}
}

func TestNewStreamingOfPanic2(t *testing.T) {
	sink := NewMemorySink[string]()
	err := NewStreamingOfSlice(context.TODO(), []int{1, 2, 3}).
		Via(NewFlatMap(func(ctx context.Context, i int) ([]string, error) {
			if i < 2 {
				return []string{"1", "2"}, nil
			}
			panic("panic-2")
		})).To(sink).Wait()
	if pErr, ok := err.(*gerrors.PanicError); !ok {
		t.Fatal(pErr)
	} else if !strings.Contains(pErr.Error(), "panic-2") {
		t.Fatal(pErr)
	}
}

func TestNewStreamingOfPanic3(t *testing.T) {
	sink := NewMemorySink[string]()
	err := NewStreamingOfSlice(context.TODO(), []int{1, 2, 3}).
		Via(NewMap(func(ctx context.Context, i int) (string, error) {
			if i < 2 {
				return "2", nil
			}
			panic("panic-2")
		})).To(sink).Wait()
	if pErr, ok := err.(*gerrors.PanicError); !ok {
		t.Fatal(pErr)
	} else if !strings.Contains(pErr.Error(), "panic-2") {
		t.Fatal(pErr)
	}
}

func TestNewStreamingOfPanic4(t *testing.T) {
	sink := NewMemorySink[int]()
	err := NewStreamingOfSlice(context.TODO(), []int{1, 2, 3}).Via(NewReduce(func(ctx context.Context, t, i int) (int, error) {
		if i < 2 {
			return t + i, nil
		}
		panic("panic-2")
	})).To(sink).Wait()
	if pErr, ok := err.(*gerrors.PanicError); !ok {
		t.Fatal(pErr)
	} else if !strings.Contains(pErr.Error(), "panic-2") {
		t.Fatal(pErr)
	}
}

func TestNewStreamingOfReduce(t *testing.T) {
	sink := NewMemorySink[int]()
	err := NewStreamingOfSlice(context.TODO(), []int{1, 2, 3}).
		Via(NewReduce(func(ctx context.Context, t, i int) (int, error) {
			return t + i, nil
		})).To(sink).Wait()
	if err != nil {
		t.Fatal(err)
	}
	if sink.Result()[len(sink.Result())-1] != 6 {
		t.Fatal(sink.Result())
	}
}

func TestNewStreamingOf(t *testing.T) {
	sink := NewMemorySink[int]()
	err := NewStreamingOfSlice(context.TODO(), []int{1, 2, 3}).
		Via(NewFilter(func(ctx context.Context, i int) (bool, error) { return i < 2, nil })).
		To(sink).Wait()
	if err != nil {
		t.Fatal(err)
	}
	for _, i := range sink.Result() {
		t.Log(i + 0)
	}
}

func TestNewStreaming(t *testing.T) {
	output := &tickerOutlet{FlowState: FlowStateWithContext(context.TODO())}
	output.init()
	a := &A{}
	streaming := NewStreaming(output)
	transfers := FanOut(streaming.Via(NewFlatMap(a.messageToStrs)), 2)
	var sinks []*stdoutSink
	for i, transfer := range transfers {
		if i == 0 {
			transfer = transfer.Via(NewFlatMap(func(ctx context.Context, s string) ([]string, error) { return []string{s + "f"}, nil }, 1))
		}
		sink := newStdoutSink(i)
		sinks = append(sinks, sink)
		transfer.To(sink)
	}
	err := streaming.Wait()
	if err.Error() != "test err" {
		t.Fatal(err)
	}
	for _, sink := range sinks {
		if atomic.LoadInt64(&sink.count) < 2 {
			t.Fatal(atomic.LoadInt64(&sink.count))
		}
	}

}

type A struct {
	lines []string
}

func (a *A) messageToStrs(ctx context.Context, item *message) ([]string, error) {
	return []string{item.content}, nil
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
