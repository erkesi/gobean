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

func TestNewDataSourceOfPanic1(t *testing.T) {
	dataStream := NewDataStreamOfSlice(context.TODO(), []int{1, 2, 3})
	sink := NewMemorySink[int]()
	dataStream.Via(NewFilter(func(ctx context.Context, i int) (bool, error) {
		if i < 2 {
			return true, nil
		}
		panic("panic-2")
	})).To(sink)
	err := dataStream.State().Wait()
	if perr, ok := err.(*gerrors.PanicError); !ok {
		t.Fatal(perr)
	} else if !strings.Contains(perr.Error(), "panic-2") {
		t.Fatal(perr)
	}
}

func TestNewDataSourceOfPanic2(t *testing.T) {
	dataStream := NewDataStreamOfSlice(context.TODO(), []int{1, 2, 3})
	sink := NewMemorySink[string]()
	dataStream.Via(NewFlatMap(func(ctx context.Context, i int) ([]string, error) {
		if i < 2 {
			return []string{"1", "2"}, nil
		}
		panic("panic-2")
	})).To(sink)
	err := dataStream.State().Wait()
	if perr, ok := err.(*gerrors.PanicError); !ok {
		t.Fatal(perr)
	} else if !strings.Contains(perr.Error(), "panic-2") {
		t.Fatal(perr)
	}
}

func TestNewDataSourceOfPanic3(t *testing.T) {
	dataStream := NewDataStreamOfSlice(context.TODO(), []int{1, 2, 3})
	sink := NewMemorySink[string]()
	dataStream.Via(NewMap(func(ctx context.Context, i int) (string, error) {
		if i < 2 {
			return "2", nil
		}
		panic("panic-2")
	})).To(sink)
	err := dataStream.State().Wait()
	if perr, ok := err.(*gerrors.PanicError); !ok {
		t.Fatal(perr)
	} else if !strings.Contains(perr.Error(), "panic-2") {
		t.Fatal(perr)
	}
}

func TestNewDataSourceOfPanic4(t *testing.T) {
	dataStream := NewDataStreamOfSlice(context.TODO(), []int{1, 2, 3})
	sink := NewMemorySink[int]()
	dataStream.Via(NewReduce(func(ctx context.Context, t, i int) (int, error) {
		if i < 2 {
			return t + i, nil
		}
		panic("panic-2")
	})).To(sink)
	err := dataStream.State().Wait()
	if perr, ok := err.(*gerrors.PanicError); !ok {
		t.Fatal(perr)
	} else if !strings.Contains(perr.Error(), "panic-2") {
		t.Fatal(perr)
	}
}

func TestNewDataSourceOfReduce(t *testing.T) {
	dataStream := NewDataStreamOfSlice(context.TODO(), []int{1, 2, 3})
	sink := NewMemorySink[int]()
	dataStream.Via(NewReduce(func(ctx context.Context, t, i int) (int, error) {
		return t + i, nil
	})).To(sink)
	err := dataStream.State().Wait()
	if err != nil {
		t.Fatal(err)
	}
	if sink.Result()[len(sink.Result())-1] != 6 {
		t.Fatal(sink.Result())
	}
}

func TestNewDataSourceOf(t *testing.T) {
	sink := NewMemorySink[int]()
	dataStream := NewDataStreamOfSlice(context.TODO(), []int{1, 2, 3})
	dataStream.Via(NewFilter(func(ctx context.Context, i int) (bool, error) { return i < 2, nil })).To(sink)
	err := dataStream.State().Wait()
	if err != nil {
		t.Fatal(err)
	}
	for _, i := range sink.Result() {
		t.Log(i + 0)
	}
}

func TestNewDataSource(t *testing.T) {
	output := &tickerOutlet{FlowState: FlowStateWithContext(context.TODO())}
	output.init()

	dataStream := NewDataStream(output)

	a := &A{}
	transfers := FanOut(dataStream.Via(NewFlatMap(a.messageToStrs)), 2)
	var sinks []*stdoutSink
	for i, transfer := range transfers {
		if i == 0 {
			transfer = transfer.Via(NewFlatMap(func(ctx context.Context, s string) ([]string, error) { return []string{s + "f"}, nil }, 1))
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
