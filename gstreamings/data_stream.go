package gstreamings

import (
	"context"
)

type dataStream struct {
	FlowState
	in <-chan interface{}
}

func NewDataStream(out Outlet) Source {
	if out.State() == nil {
		panic("outlet state is nil")
	}
	ds := &dataStream{
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
	return &dataStream{
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
	return &dataStream{
		FlowState: FlowState{_state: state},
		in:        in,
	}
}

func (ds *dataStream) Out() <-chan interface{} {
	return ds.in
}

func (ds *dataStream) Via(flow Transfer) Transfer {
	flow.setState(ds.State())
	ds.doStream(ds, flow)
	return flow
}

// doStream streams data from the outlet to inlet.
func (ds *dataStream) doStream(outlet Outlet, inlet Inlet) {
	go func() {
		for element := range outlet.Out() {
			inlet.In() <- element
		}
		close(inlet.In())
	}()
}
