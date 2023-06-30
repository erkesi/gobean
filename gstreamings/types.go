package gstreamings

import (
	"context"
	"sync"
)

// Inlet represents a type that exposes one open input.
type Inlet interface {
	Stater
	In() chan<- interface{}
}

// Outlet represents a type that exposes one open output.
type Outlet interface {
	Stater
	Out() <-chan interface{}
}

// Source represents a set of stream processing steps that has one open output.
type Source interface {
	Outlet
	Via(Transfer) Transfer
}

// Transfer represents a set of stream processing steps that has one open input and one open output.
type Transfer interface {
	Inlet
	Outlet
	Via(Transfer) Transfer
	To(Sink)
}

type Stater interface {
	State() *state
	Context() context.Context
	setState(state *state)
	SetStateErr(err error)
}

// Sink represents a set of stream processing steps that has one open input.
type Sink interface {
	Inlet
	setSinkState(state *state)
}

type Optional[T any] interface {
	Get() T
	IsPresent() bool
}

func NewOptional[T any](t T) Optional[T] {
	return &optional[T]{val: t, isPresent: true}
}

func NewEmptyOptional[T any]() Optional[T] {
	return &optional[T]{}
}

type optional[T any] struct {
	isPresent bool
	val       T
}

func (o *optional[T]) Get() T {
	return o.val
}

func (o *optional[T]) IsPresent() bool {
	return o.isPresent
}

type StoreSink[T any] interface {
	Sink
}

type MemorySink[T any] interface {
	Sink
	Result() []T
}

func FlowStateWithContext(ctx context.Context) FlowState {
	return FlowState{_state: NewState(ctx)}
}

type FlowState struct {
	_state *state
}

func (bs *FlowState) State() *state {
	return bs._state
}

func (bs *FlowState) Context() context.Context {
	return bs._state.ctx
}

func (bs *FlowState) Done() {
	bs._state.wg.Done()
}

func (bs *FlowState) setState(s *state) {
	bs._state = s
}

func (bs *FlowState) SetStateErr(err error) {
	bs._state.setErr(err)
}

func (bs *FlowState) HasStateErr() bool {
	return bs._state.hasErr()
}

func (bs *FlowState) setSinkState(transState *state) {
	transState.wg.Add(1)
	bs._state = transState
}

func NewState(ctx context.Context) *state {
	return &state{
		ctx: ctx,
	}
}

type state struct {
	rw  sync.RWMutex
	ctx context.Context
	wg  sync.WaitGroup
	err error
}

func (s *state) Wait() error {
	s.wg.Wait()
	return s.error()
}

func (s *state) setErr(err error) {
	s.rw.Lock()
	defer s.rw.Unlock()
	if s.err != nil {
		return
	}
	s.err = err
}

func (s *state) error() error {
	s.rw.RLock()
	defer s.rw.RUnlock()
	return s.err
}

func (s *state) hasErr() bool {
	return s.error() != nil
}
