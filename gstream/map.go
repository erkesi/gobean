package gstream

import "context"

// MapFunction represents a Map transformation function.
type MapFunction[T, R any] func(context.Context, T) R

// Map takes one element and produces one element.
//
// in  -- 1 -- 2 ---- 3 -- 4 ------ 5 --
//
// [ ---------- MapFunction ---------- ]
//
// out -- 1' - 2' --- 3' - 4' ----- 5' -
type Map[T, R any] struct {
	FlowState
	mapFunction MapFunction[T, R]
	in          chan interface{}
	out         chan interface{}
	parallelism uint
}

// Verify Map satisfies the Transfer interface.
var _ Transfer = (*Map[any, any])(nil)

// NewMap returns a new Map instance.
//
// mapFunction is the Map transformation function.
// parallelism is the flow parallelism factor. In case the events order matters, use parallelism = 1.
func NewMap[T, R any](mapFunction MapFunction[T, R], parallelism uint) *Map[T, R] {
	mapFlow := &Map[T, R]{
		mapFunction: mapFunction,
		in:          make(chan interface{}),
		out:         make(chan interface{}),
		parallelism: parallelism,
	}
	go mapFlow.doStream()
	return mapFlow
}

// Via streams data through the given flow
func (m *Map[T, R]) Via(flow Transfer) Transfer {
	flow.setState(m.State())
	go m.transmit(flow)
	return flow
}

// To streams data to the given sink
func (m *Map[T, R]) To(sink Sink) {
	sink.SetSinkState(m.State())
	go m.transmit(sink)
}

// Out returns an output channel for sending data
func (m *Map[T, R]) Out() <-chan interface{} {
	return m.out
}

// In returns an input channel for receiving data
func (m *Map[T, R]) In() chan<- interface{} {
	return m.in
}

func (m *Map[T, R]) transmit(inlet Inlet) {
	for element := range m.Out() {
		inlet.In() <- element
	}
	close(inlet.In())
}

func (m *Map[T, R]) doStream() {
	sem := make(chan struct{}, m.parallelism)
	for elem := range m.in {
		sem <- struct{}{}
		go func(element T) {
			defer func() { <-sem }()
			result := m.mapFunction(m.Context(), element)
			m.out <- result
		}(elem.(T))
	}
	for i := 0; i < int(m.parallelism); i++ {
		sem <- struct{}{}
	}
	close(m.out)
}
