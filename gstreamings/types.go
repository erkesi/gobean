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

// Streaming represents a set of stream processing steps that has one open output.
type Streaming interface {
	Outlet
	Via(Transfer) Transfer
}

// Transfer represents a set of stream processing steps that has one open input and one open output.
type Transfer interface {
	Inlet
	Outlet
	Via(Transfer) Transfer
	To(Sink) Stater
}

type Stater interface {
	State() *State
	Wait() error
	Context() context.Context
	setState(state *State)
	SetStateErr(err error)
}

// Sink represents a set of stream processing steps that has one open input.
type Sink interface {
	Inlet
	setSinkState(state *State)
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
	return FlowState{state: newState(ctx)}
}

type FlowState struct {
	state *State
}

func (bs *FlowState) State() *State {
	return bs.state
}

func (bs *FlowState) Wait() error {
	return bs.state.Wait()
}

func (bs *FlowState) Context() context.Context {
	return bs.state.ctx
}

func (bs *FlowState) Done() {
	bs.state.wg.Done()
}

func (bs *FlowState) setState(s *State) {
	bs.state = s
}

func (bs *FlowState) SetStateErr(err error) {
	bs.state.setErr(err)
}

func (bs *FlowState) HasStateErr() bool {
	return bs.state.hasErr()
}

func (bs *FlowState) setSinkState(transState *State) {
	transState.wg.Add(1)
	bs.state = transState
}

func newState(ctx context.Context) *State {
	return &State{
		ctx: ctx,
	}
}

type State struct {
	rw  sync.RWMutex
	ctx context.Context
	wg  sync.WaitGroup
	err error
}

func (s *State) Wait() error {
	s.wg.Wait()
	return s.error()
}

func (s *State) setErr(err error) {
	s.rw.Lock()
	defer s.rw.Unlock()
	if s.err != nil {
		return
	}
	s.err = err
}

func (s *State) error() error {
	s.rw.RLock()
	defer s.rw.RUnlock()
	return s.err
}

func (s *State) hasErr() bool {
	return s.error() != nil
}
