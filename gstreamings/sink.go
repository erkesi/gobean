package gstreamings

import "context"

type memorySink[T any] struct {
	FlowState
	in     chan interface{}
	result []T
}

func NewMemorySink[T any]() MemorySink[T] {
	sink := &memorySink[T]{
		in: make(chan interface{}),
	}
	sink.init()
	return sink
}

func (sink *memorySink[T]) init() {
	go func() {
		for elem := range sink.in {
			sink.result = append(sink.result, elem.(T))
		}
		sink.Done()
	}()
}

// In returns an input channel for receiving data
func (sink *memorySink[T]) In() chan<- interface{} {
	return sink.in
}

func (sink *memorySink[T]) Result() []T {
	return sink.result
}

type storeSink[T any] struct {
	FlowState
	in    chan interface{}
	store func(context.Context, []T) error
	batch int
}

func NewStoreSink[T any](batch int, store func(ctx context.Context, ts []T) error) StoreSink[T] {
	sink := &storeSink[T]{
		in:    make(chan interface{}),
		batch: batch,
		store: store,
	}
	sink.init()
	return sink
}

func (sink *storeSink[T]) init() {
	go func() {
		var ts []T
		for elem := range sink.in {
			ts = append(ts, elem.(T))
			if len(ts) >= sink.batch {
				err := sink.store(sink.Context(), ts)
				if err != nil {
					sink.SetStateErr(err)
				}
				ts = ts[0:0]
			}
		}
		if len(ts) > 0 {
			err := sink.store(sink.Context(), ts)
			if err != nil {
				sink.SetStateErr(err)
			}
		}
		sink.Done()
	}()
}

// In returns an input channel for receiving data
func (sink *storeSink[T]) In() chan<- interface{} {
	return sink.in
}
