package gstream

type memorySink[T any] struct {
	StateConf
	in     chan interface{}
	result []T
}

func NewMemorySink[T any]() MemorySink[T] {
	ds := &memorySink[T]{
		in: make(chan interface{}),
	}
	ds.init()
	return ds
}

func (ds *memorySink[T]) init() {
	go func() {
		for elem := range ds.in {
			ds.result = append(ds.result, elem.(T))
		}
		ds.Done()
	}()
}

// In returns an input channel for receiving data
func (ds *memorySink[T]) In() chan<- interface{} {
	return ds.in
}

func (ds *memorySink[T]) Result() []T {
	return ds.result
}
