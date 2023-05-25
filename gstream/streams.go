package gstream

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
	Done()
	SetSinkState(state *state)
}

type MemorySink[T any] interface {
	Sink
	Result() []T
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

func (bs *FlowState) SetSinkState(transState *state) {
	transState.wg.Add(1)
	bs._state = transState
}

func newState(ctx context.Context) *state {
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
