package gstatemachines

import (
	"context"
)

type StateMachineDefinition struct {
	Name     string
	Version  string
	Id2State map[string]Stater
}

type StateMachine struct {
	Definition *StateMachineDefinition
	curState   Stater
}

func (sm *StateMachine) Execute(ctx context.Context, sourceStateId string, event Event, args ...interface{}) error {
	curState, ok := sm.Definition.Id2State[sourceStateId]
	if !ok {
		return ErrStateNotExist
	}
	sm.curState = curState

	err := sm.curState.Validate()
	if err != nil {
		return err
	}

	err = curState.Action(ctx, event, args...)
	if err != nil {
		return err
	}

	nextState, err := sm.curState.Transition(ctx, event)
	if err != nil {
		return err
	}

	err = sm.curState.Exit(ctx, event, args...)
	if err != nil {
		return err
	}
	sm.curState = nextState
	return sm.curState.Entry(ctx, event, args...)
}

func (sm *StateMachine) CurState() Stater {
	return sm.curState
}

func ToStateMachineDefinition(dsl string) (*StateMachineDefinition, error) {
	definition := &StateMachineDefinition{}
	_, err := toStateMachineDSL(dsl)
	if err != nil {
		return nil, err
	}
	return definition, nil
}
