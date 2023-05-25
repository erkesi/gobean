package gstream

import "context"

// FilterPredicate represents a filter predicate (boolean-valued function).
type FilterPredicate[T any] func(context.Context, T) (bool, error)

// Filter filters incoming elements using a filter predicate.
// If an element matches the predicate, the element is passed downstream.
// If not, the element is discarded.
//
// in  -- 1 -- 2 ---- 3 -- 4 ------ 5 --
//
// [ -------- FilterPredicate -------- ]
//
// out -- 1 -- 2 ------------------ 5 --
type Filter[T any] struct {
	FlowState
	filterPredicate FilterPredicate[T]
	in              chan interface{}
	out             chan interface{}
	parallelism     uint
}

// Verify Filter satisfies the Transfer interface.
var _ Transfer = (*Filter[any])(nil)

// NewFilter returns a new Filter instance.
//
// filterPredicate is the boolean-valued filter function.
// parallelism is the flow parallelism factor. In case the events order matters, use parallelism = 1.
func NewFilter[T any](filterPredicate FilterPredicate[T], parallelism uint) *Filter[T] {
	filter := &Filter[T]{
		filterPredicate: filterPredicate,
		in:              make(chan interface{}),
		out:             make(chan interface{}),
		parallelism:     parallelism,
	}
	go filter.doStream()

	return filter
}

// Via streams data through the given flow
func (f *Filter[T]) Via(flow Transfer) Transfer {
	flow.setState(f.State())
	go f.transmit(flow)
	return flow
}

// To streams data to the given sink
func (f *Filter[T]) To(sink Sink) {
	sink.SetSinkState(f.State())
	go f.transmit(sink)
}

// Out returns an output channel for sending data
func (f *Filter[T]) Out() <-chan interface{} {
	return f.out
}

// In returns an input channel for receiving data
func (f *Filter[T]) In() chan<- interface{} {
	return f.in
}

func (f *Filter[T]) transmit(inlet Inlet) {
	for element := range f.Out() {
		inlet.In() <- element
	}
	close(inlet.In())
}

// doStream discards items that don't match the filter predicate.
func (f *Filter[T]) doStream() {
	sem := make(chan struct{}, f.parallelism)
	for elem := range f.in {
		sem <- struct{}{}
		go func(element T) {
			defer func() { <-sem }()
			ok, err := f.filterPredicate(f.Context(), element)
			if err != nil {
				f.SetStateErr(err)
				return
			}
			if ok {
				f.out <- element
			}
		}(elem.(T))
	}
	for i := 0; i < int(f.parallelism); i++ {
		sem <- struct{}{}
	}
	close(f.out)
}
