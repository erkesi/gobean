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

type StateConf struct {
	transState *state
}

func (bs *StateConf) State() *state {
	return bs.transState
}

func (bs *StateConf) Context() context.Context {
	return bs.transState.ctx
}

func (bs *StateConf) Done() {
	bs.transState.wg.Done()
}

func (bs *StateConf) setState(s *state) {
	bs.transState = s
}

func (bs *StateConf) SetSinkState(transState *state) {
	transState.wg.Add(1)
	bs.transState = transState
}

func newState(ctx context.Context) *state {
	return &state{
		ctx: ctx,
	}
}

type state struct {
	sync.RWMutex
	ctx context.Context
	wg  sync.WaitGroup
	err error
}

func (s *state) Wait() error {
	s.wg.Wait()
	return s.error()
}

func (s *state) SetErr(err error) {
	s.Lock()
	defer s.Unlock()
	if s.err != nil {
		return
	}
	s.err = err
}

func (s *state) error() error {
	s.RLock()
	defer s.RUnlock()
	return s.err
}

func (s *state) HasErr() bool {
	return s.error() != nil
}
