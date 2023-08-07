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
	Via(Flow) Flow
}

// Flow represents a set of stream processing steps that has one open input and one open output.
type Flow interface {
	Inlet
	Outlet
	Via(Flow) Flow
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

func (fs *FlowState) State() *State {
	return fs.state
}

func (fs *FlowState) Wait() error {
	return fs.state.Wait()
}

func (fs *FlowState) Context() context.Context {
	return fs.state.ctx
}

func (fs *FlowState) Done() {
	fs.state.wg.Done()
}

func (fs *FlowState) setState(s *State) {
	fs.state = s
}

func (fs *FlowState) SetStateErr(err error) {
	fs.state.setErr(err)
}

func (fs *FlowState) HasStateErr() bool {
	return fs.state.hasErr()
}

func (fs *FlowState) setSinkState(transState *State) {
	transState.wg.Add(1)
	fs.state = transState
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
