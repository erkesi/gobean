package gdataflow

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
	To(Sink)
}

type Stater interface {
	TransState() *transState
	Context() context.Context
	SetTransState(state *transState)
}

// Sink represents a set of stream processing steps that has one open input.
type Sink interface {
	Inlet
	Done()
	SetSinkTransState(state *transState)
}

type TransStateConf struct {
	transState *transState
}

func (bs *TransStateConf) TransState() *transState {
	return bs.transState
}

func (bs *TransStateConf) Context() context.Context {
	return bs.transState.ctx
}

func (bs *TransStateConf) Done() {
	bs.transState.wg.Done()
}

func (bs *TransStateConf) SetTransState(transState *transState) {
	bs.transState = transState
}

func (bs *TransStateConf) SetSinkTransState(transState *transState) {
	transState.wg.Add(1)
	bs.transState = transState
}

func newTransState(ctx context.Context) *transState {
	return &transState{
		ctx: ctx,
	}
}

type transState struct {
	sync.RWMutex
	ctx context.Context
	wg  sync.WaitGroup
	err error
}

func (s *transState) Wait() {
	s.wg.Wait()
}

func (s *transState) SetErr(err error) {
	s.Lock()
	defer s.Unlock()
	if s.err != nil {
		return
	}
	s.err = err
}

func (s *transState) Err() error {
	s.RLock()
	defer s.RUnlock()
	return s.err
}

func (s *transState) IsErr() bool {
	return s.Err() != nil
}
