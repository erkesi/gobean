package gstream

import (
	"context"
	"sync"
)

type dataSource struct {
	FlowState
	in <-chan interface{}
}

func NewDataStream(out Outlet) Source {
	if out.State() == nil {
		panic("outlet state is nil")
	}
	ds := &dataSource{
		in: out.Out(),
	}
	ds.setState(out.State())
	return ds
}

type CursorNext[T any] func(ctx context.Context) (items []T, hasNext bool, err error)

func NewDataStreamOfCursor[T any](ctx context.Context, cursor func(ctx context.Context) CursorNext[T]) Source {
	state := NewState(ctx)
	in := make(chan interface{})
	go func() {
		defer close(in)
		next := cursor(ctx)
		for {
			items, hasNext, err := next(ctx)
			if err != nil {
				state.setErr(err)
				return
			}
			for _, item := range items {
				if state.hasErr() {
					return
				}
				in <- item
			}
			if !hasNext {
				return
			}
		}
	}()
	return &dataSource{
		FlowState: FlowState{_state: state},
		in:        in,
	}
}

func NewDataStreamOfSlice[T any](ctx context.Context, items []T) Source {
	state := NewState(ctx)
	in := make(chan interface{})
	go func() {
		for _, item := range items {
			if state.hasErr() {
				return
			}
			in <- item
		}
		close(in)
	}()
	return &dataSource{
		FlowState: FlowState{_state: state},
		in:        in,
	}
}

func (ds *dataSource) Out() <-chan interface{} {
	return ds.in
}

func (ds *dataSource) Via(flow Transfer) Transfer {
	flow.setState(ds.State())
	ds.doStream(ds, flow)
	return flow
}

// doStream streams data from the outlet to inlet.
func (ds *dataSource) doStream(outlet Outlet, inlet Inlet) {
	go func() {
		for element := range outlet.Out() {
			inlet.In() <- element
		}
		close(inlet.In())
	}()
}

// Split splits the stream into two flows according to the given boolean predicate.
func Split[T any](outlet Outlet, predicate func(T) bool) [2]Transfer {
	condTrue := newPassThrough()
	condTrue.setState(outlet.State())
	condFalse := newPassThrough()
	condFalse.setState(outlet.State())
	go func() {
		for element := range outlet.Out() {
			if predicate(element.(T)) {
				condTrue.In() <- element
			} else {
				condFalse.In() <- element
			}
		}
		close(condTrue.In())
		close(condFalse.In())
	}()

	return [...]Transfer{condTrue, condFalse}
}

// FanOut creates a number of identical flows from the single outlet.
// This can be useful when writing to multiple sinks is required.
func FanOut(outlet Outlet, magnitude int) []Transfer {
	var out []Transfer
	for i := 0; i < magnitude; i++ {
		pt := newPassThrough()
		pt.setState(outlet.State())
		out = append(out, pt)
	}

	go func() {
		for element := range outlet.Out() {
			for _, socket := range out {
				socket.In() <- element
			}
		}
		for i := 0; i < magnitude; i++ {
			close(out[i].In())
		}
	}()

	return out
}

// Merge merges multiple flows into a single flow.
func Merge(outlets ...Transfer) Transfer {
	merged := newPassThrough()
	merged.setState(outlets[0].State())
	var wg sync.WaitGroup
	wg.Add(len(outlets))

	for _, out := range outlets {
		go func(outlet Outlet) {
			for element := range outlet.Out() {
				merged.In() <- element
			}
			wg.Done()
		}(out)
	}

	// close the in channel on the last outlet close.
	go func(wg *sync.WaitGroup) {
		wg.Wait()
		close(merged.In())
	}(&wg)

	return merged
}
