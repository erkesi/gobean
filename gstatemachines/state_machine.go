package gstatemachines

import (
	"context"
	"github.com/erkesi/gobean/glogs"
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
	glogs.Log.Debugf("executing, sourceStateId is %s", sourceStateId)

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

	glogs.Log.Debugf("executing, sourceStateId is %s, targetStateId is %s", sourceStateId, nextState.GetId())

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

func ToStateMachineDefinition(dsl string, id2BaseState map[string]BaseStater) (*StateMachineDefinition, error) {
	definition := &StateMachineDefinition{}
	stateMachineDsl, err := toStateMachineDSL(dsl)
	if err != nil {
		return nil, err
	}
	// state映射
	definition.Id2State = make(map[string]Stater)
	for key, baseState := range id2BaseState {
		desc := key
		isStart := false
		isEnd := false
		for _, state := range stateMachineDsl.States {
			if state.Id == key {
				desc = state.Desc
				isStart = state.IsStart
				isEnd = state.IsEnd
				break
			}
		}
		definition.Id2State[key] = &State{
			BaseStater:  baseState,
			Id:          key,
			Desc:        desc,
			Transitions: make([]*Transition, 0, 5),
			IsStart:     isStart,
			IsEnd:       isEnd,
		}
	}

	// 描述信息映射
	definition.Name = stateMachineDsl.Name
	definition.Version = stateMachineDsl.Version

	// transition 映射，绑定到state 上
	for _, t := range stateMachineDsl.Transitions {
		if srcState, ok := definition.Id2State[t.SourceId]; ok {
			if targetState, ok := definition.Id2State[t.TargetId]; ok {
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
