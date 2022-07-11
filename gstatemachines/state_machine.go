package gstatemachines

import (
	"context"
	"github.com/pkg/errors"
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
	DebugLog("Executing, sourceStateId is " + sourceStateId)
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

	nextState, err := sm.curState.Transform(ctx, event)
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

func ToStateMachineDefinition(dsl string, id2State map[string]Stater) (*StateMachineDefinition, error) {
	definition := &StateMachineDefinition{}
	stateMachineDsl, err := toStateMachineDSL(dsl)
	if err != nil {
		return nil, errors.Wrap(err, "ToStateMachineDefinition error<|>dsl: "+dsl)
	}
	// 描述结构实例化
	definition.Id2State = id2State

	// transition 绑定到state 上
	for _, t := range stateMachineDsl.Transitions {
		if srcState, ok := id2State[t.SourceId]; ok {
			if targetState, ok := id2State[t.TargetId]; ok {
				transitions := srcState.GetTransitions()
				transitions = append(transitions, &Transition{
					Condition: t.Condition,
					Target:    targetState,
				})
				srcState.SetTransitions(transitions)
			} else {
				return nil, ErrStateEmptyTarget
			}
		} else {
			return nil, ErrStateEmptySource
		}
	}

	return definition, nil
}
