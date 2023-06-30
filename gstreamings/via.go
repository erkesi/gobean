package gstreamings

import (
	"context"
	"github.com/erkesi/gobean/grecovers"
	"sync"
)

// MapFunction represents a Map transformation function.
type MapFunction[T, R any] func(context.Context, T) (R, error)

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
func NewMap[T, R any](mapFunction MapFunction[T, R], parallelism ...uint) *Map[T, R] {
	mapFlow := &Map[T, R]{
		mapFunction: mapFunction,
		in:          make(chan interface{}),
		out:         make(chan interface{}),
		parallelism: parallel(parallelism),
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
	sink.setSinkState(m.State())
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

			if m.HasStateErr() {
				return
			}
			result, err := grecovers.RecoverVGFn(func() (interface{}, error) {
				return m.mapFunction(m.Context(), element)
			})()
			if err != nil {
				m.SetStateErr(err)
				return
			}
			m.out <- result
		}(elem.(T))
	}
	for i := 0; i < int(m.parallelism); i++ {
		sem <- struct{}{}
	}
	close(m.out)
}

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
func NewFilter[T any](filterPredicate FilterPredicate[T], parallelism ...uint) *Filter[T] {
	filter := &Filter[T]{
		filterPredicate: filterPredicate,
		in:              make(chan interface{}),
		out:             make(chan interface{}),
		parallelism:     parallel(parallelism),
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
	sink.setSinkState(f.State())
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
			if f.HasStateErr() {
				return
			}
			ok, err := grecovers.RecoverVGFn(func() (interface{}, error) {
				return f.filterPredicate(f.Context(), element)
			})()
			if err != nil {
				f.SetStateErr(err)
				return
			}
			if ok.(bool) {
				f.out <- element
			}
		}(elem.(T))
	}
	for i := 0; i < int(f.parallelism); i++ {
		sem <- struct{}{}
	}
	close(f.out)
}

// FlatMapFunction represents a FlatMap transformation function.
type FlatMapFunction[T, R any] func(context.Context, T) ([]R, error)

// FlatMap takes one element and produces zero, one, or more elements.
//
// in  -- 1 -- 2 ---- 3 -- 4 ------ 5 --
//
// [ -------- FlatMapFunction -------- ]
//
// out -- 1' - 2' -------- 4'- 4" - 5' -
type FlatMap[T, R any] struct {
	FlowState
	flatMapFunction FlatMapFunction[T, R]
	in              chan interface{}
	out             chan interface{}
	parallelism     uint
}

// Verify FlatMap satisfies the Transfer interface.
var _ Transfer = (*FlatMap[any, any])(nil)

// NewFlatMap returns a new FlatMap instance.
//
// flatMapFunction is the FlatMap transformation function.
// parallelism is the flow parallelism factor. In case the events order matters, use parallelism = 1.
func NewFlatMap[T, R any](flatMapFunction FlatMapFunction[T, R], parallelism ...uint) *FlatMap[T, R] {
	flatMap := &FlatMap[T, R]{
		flatMapFunction: flatMapFunction,
		in:              make(chan interface{}),
		out:             make(chan interface{}),
		parallelism:     parallel(parallelism),
	}
	go flatMap.doStream()

	return flatMap
}

// Via streams data through the given flow
func (fm *FlatMap[T, R]) Via(flow Transfer) Transfer {
	flow.setState(fm.State())
	go fm.transmit(flow)
	return flow
}

// To streams data to the given sink
func (fm *FlatMap[T, R]) To(sink Sink) {
	sink.setSinkState(fm.State())
	go fm.transmit(sink)
}

// Out returns an output channel for sending data
func (fm *FlatMap[T, R]) Out() <-chan interface{} {
	return fm.out
}

// In returns an input channel for receiving data
func (fm *FlatMap[T, R]) In() chan<- interface{} {
	return fm.in
}

func (fm *FlatMap[T, R]) transmit(inlet Inlet) {
	for element := range fm.Out() {
		inlet.In() <- element
	}
	close(inlet.In())
}

func (fm *FlatMap[T, R]) doStream() {
	sem := make(chan struct{}, fm.parallelism)
	for elem := range fm.in {
		sem <- struct{}{}
		go func(element T) {
			defer func() { <-sem }()
			if fm.HasStateErr() {
				return
			}
			result, err := grecovers.RecoverVGFn(func() (interface{}, error) {
				return fm.flatMapFunction(fm.Context(), element)
			})()
			if err != nil {
				fm.SetStateErr(err)
				return
			}
			for _, item := range result.([]R) {
				fm.out <- item
			}
		}(elem.(T))
	}
	for i := 0; i < int(fm.parallelism); i++ {
		sem <- struct{}{}
	}

	close(fm.out)
}

// passThrough retransmits incoming elements as is.
//
// in  -- 1 -- 2 ---- 3 -- 4 ------ 5 --
//
// out -- 1 -- 2 ---- 3 -- 4 ------ 5 --
type passThrough struct {
	FlowState
	in  chan interface{}
	out chan interface{}
}

// Verify passThrough satisfies the Transfer interface.
var _ Transfer = (*passThrough)(nil)

// newPassThrough returns a new passThrough instance.
func newPassThrough() *passThrough {
	passThrough := &passThrough{
		in:  make(chan interface{}),
		out: make(chan interface{}),
	}
	go passThrough.doStream()

	return passThrough
}

// Via streams data through the given flow
func (pt *passThrough) Via(flow Transfer) Transfer {
	flow.setState(pt.State())
	go pt.transmit(flow)
	return flow
}

// To streams data to the given sink
func (pt *passThrough) To(sink Sink) {
	sink.setSinkState(pt.State())
	go pt.transmit(sink)
}

// Out returns an output channel for sending data
func (pt *passThrough) Out() <-chan interface{} {
	return pt.out
}

// In returns an input channel for receiving data
func (pt *passThrough) In() chan<- interface{} {
	return pt.in
}

func (pt *passThrough) transmit(inlet Inlet) {
	for elem := range pt.Out() {
		inlet.In() <- elem
	}
	close(inlet.In())
}

func (pt *passThrough) doStream() {
	for elem := range pt.in {
		pt.out <- elem
	}
	close(pt.out)
}

// ReduceFunction combines the current element with the last reduced value.
type ReduceFunction[T any] func(context.Context, T, T) (T, error)

// Reduce represents a “rolling” reduce on a data stream.
// Combines the current element with the last reduced value and emits the new value.
//
// in  -- 1 -- 2 ---- 3 -- 4 ------ 5 --
//
// [ --------- ReduceFunction --------- ]
//
// out -- 1 -- 2' --- 3' - 4' ----- 5' -
type Reduce[T any] struct {
	FlowState
	reduceFunction ReduceFunction[T]
	in             chan interface{}
	out            chan interface{}
	lastReduced    interface{}
}

// Verify Reduce satisfies the Transfer interface.
var _ Transfer = (*Reduce[any])(nil)

// NewReduce returns a new Reduce instance.
//
// reduceFunction combines the current element with the last reduced value.
func NewReduce[T any](reduceFunction ReduceFunction[T]) *Reduce[T] {
	reduce := &Reduce[T]{
		reduceFunction: reduceFunction,
		in:             make(chan interface{}),
		out:            make(chan interface{}),
	}
	go reduce.doStream()
	return reduce
}

// Via streams data through the given flow
func (r *Reduce[T]) Via(flow Transfer) Transfer {
	flow.setState(r.State())
	go r.transmit(flow)
	return flow
}

// To streams data to the given sink
func (r *Reduce[T]) To(sink Sink) {
	sink.setSinkState(r.State())
	go r.transmit(sink)
}

// Out returns an output channel for sending data
func (r *Reduce[T]) Out() <-chan interface{} {
	return r.out
}

// In returns an input channel for receiving data
func (r *Reduce[T]) In() chan<- interface{} {
	return r.in
}

func (r *Reduce[T]) transmit(inlet Inlet) {
	for element := range r.Out() {
		inlet.In() <- element
	}
	close(inlet.In())
}

func (r *Reduce[T]) doStream() {
	for element := range r.in {
		if r.HasStateErr() {
			continue
		}
		if r.lastReduced == nil {
			r.lastReduced = element
		} else {
			lastReduced, err := grecovers.RecoverVGFn(func() (interface{}, error) {
				return r.reduceFunction(r.Context(), r.lastReduced.(T), element.(T))
			})()
			if err != nil {
				r.SetStateErr(err)
				continue
			}
			r.lastReduced = lastReduced
		}
		r.out <- r.lastReduced
	}
	close(r.out)
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

func parallel(parallelism []uint) uint {
	if len(parallelism) > 0 {
		return parallelism[0]
	}
	return 1
}
