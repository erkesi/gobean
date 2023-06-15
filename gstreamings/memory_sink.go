package gstreamings

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
