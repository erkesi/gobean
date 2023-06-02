package gstream

import "context"

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
