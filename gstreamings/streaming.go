package gstreamings

import (
	"context"
)

type _streaming struct {
	FlowState
	in <-chan interface{}
}

func NewStreaming(out Outlet) Streaming {
	if out.State() == nil {
		panic("outlet state is nil")
	}
	ds := &_streaming{
		in: out.Out(),
	}
	ds.setState(out.State())
	return ds
}

type CursorNext[T any] func(ctx context.Context) (items []T, hasNext bool, err error)

func NewStreamingOfCursor[T any](ctx context.Context, cursor func(ctx context.Context) CursorNext[T]) Streaming {
	state := newState(ctx)
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
	return &_streaming{
		FlowState: FlowState{state: state},
		in:        in,
	}
}

func NewStreamingOfSlice[T any](ctx context.Context, items []T) Streaming {
	state := newState(ctx)
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
	return &_streaming{
		FlowState: FlowState{state: state},
		in:        in,
	}
}

func (ds *_streaming) Out() <-chan interface{} {
	return ds.in
}

func (ds *_streaming) Via(flow Transfer) Transfer {
	flow.setState(ds.State())
	ds.doStream(ds, flow)
	return flow
}

// doStream streams data from the outlet to inlet.
func (ds *_streaming) doStream(outlet Outlet, inlet Inlet) {
	go func() {
		for element := range outlet.Out() {
			inlet.In() <- element
		}
		close(inlet.In())
	}()
}
