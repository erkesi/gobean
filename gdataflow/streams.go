package gdataflow

import "sync"

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
	State() *State
	SetState(state *State)
}

type Done interface {
	Done()
}

// Sink represents a set of stream processing steps that has one open input.
type Sink interface {
	Inlet
	Done
	SetSinkState(state *State)
}

type FlowState struct {
	state *State
}

func (bs *FlowState) State() *State {
	return bs.state
}

func (bs *FlowState) Done() {
	bs.state.wg.Done()
}

func (bs *FlowState) SetState(state *State) {
	bs.state = state
}

func (bs *FlowState) SetSinkState(state *State) {
	state.wg.Add(1)
	bs.state = state
}

type State struct {
	sync.RWMutex
	wg  sync.WaitGroup
	err error
}

func (s *State) Wait() {
	s.wg.Wait()
}

func (s *State) SetErr(err error) {
	s.Lock()
	defer s.Unlock()
	if s.err != nil {
		return
	}
	s.err = err
}

func (s *State) Err() error {
	s.RLock()
	defer s.RUnlock()
	return s.err
}

func (s *State) IsErr() bool {
	return s.Err() != nil
}
