package gstatemachines

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/erkesi/gobean/glogs"
)

type StateMachineDefinition struct {
	Name         string
	Version      string
	StartStateId string
	Id2State     map[string]*State
	Transitions  []*Transition
}

func (d StateMachineDefinition) PlainUML() string {
	plainUML := "@startuml\n\n"
	for _, transition := range d.Transitions {
		target := transition.Source
		if transition.Target != nil {
			target = transition.Target
		}
		condition := transition.Condition
		if len(transition.ActionsRaw) > 0 {
			condition += ",actions(" + transition.ActionsRaw + ")"
		}
		plainUML += fmt.Sprintf("%s --> %s : %s\n", transition.Source.Id,
			target.Id, condition)
	}
	plainUML += "\n@enduml"
	return plainUML
}

type StateMachine struct {
	Definition *StateMachineDefinition
	curState   *State
}

func (sm *StateMachine) Execute(ctx context.Context, sourceStateId string,
	event Event, args ...interface{}) error {
	if glogs.Log != nil {
		glogs.Log.Debugf(ctx, "gstatemachines: executing, sourceStateId is %s", sourceStateId)
	}

	curState, ok := sm.Definition.Id2State[sourceStateId]
	if !ok {
		return ErrStateNotExist
	}
	sm.curState = curState

	for {
		nextState, err := sm.curState.Transform(ctx, event, args)
		if err != nil {
			return err
		}
		if glogs.Log != nil {
			if nextState == nil {
				glogs.Log.Debugf(ctx, "gstatemachines: executing, sourceStateId is %s, targetStateId is %s", sm.curState.Id, sm.curState.Id)
			} else {
				glogs.Log.Debugf(ctx, "gstatemachines: executing, sourceStateId is %s, targetStateId is %s", sm.curState.Id, nextState.Id)
			}
		}
		if nextState == nil {
			return nil
		}
		if glogs.Log != nil {
			glogs.Log.Debugf(ctx, "gstatemachines: executing, exit sourceState(%v)", sm.curState)
		}
		err = sm.curState.Exit(ctx, event, args...)
		if err != nil {
			return err
		}
		sm.curState = nextState
		if glogs.Log != nil {
			glogs.Log.Debugf(ctx, "gstatemachines: executing, entry nextState(%v)", sm.curState)
		}
		err = sm.curState.Entry(ctx, event, args...)
		if errors.Is(err, ErrStateSkip) {
			continue
		}
		return err
	}
}

func (sm *StateMachine) CurState() *State {
	return sm.curState
}

func ToStateMachineDefinition(dsl string, id2BaseState map[string]BizStater) (*StateMachineDefinition, error) {
	definition := &StateMachineDefinition{}
	stateMachineDsl, err := toStateMachineDSL(dsl)
	if err != nil {
		return nil, err
	}
	// state映射
	definition.Id2State = make(map[string]*State)
	for key, baseState := range id2BaseState {
		desc := key
		isStart := false
		isEnd := false
		for _, state := range stateMachineDsl.States {
			if state.IsStart {
				definition.StartStateId = state.Id
			}
			if state.Id == key {
				desc = state.Desc
				isStart = state.IsStart
				isEnd = state.IsEnd
				break
			}
		}
		definition.Id2State[key] = &State{
			BizStater:   baseState,
			Id:          key,
			Desc:        desc,
			Transitions: make([]*Transition, 0, 5),
			isStart:     isStart,
			isEnd:       isEnd,
		}
	}

	// 描述信息映射
	definition.Name = stateMachineDsl.Name
	definition.Version = stateMachineDsl.Version

	// transition 映射，绑定到state 上
	for _, t := range stateMachineDsl.Transitions {
		if sourceState, ok := definition.Id2State[t.SourceId]; ok {
			if _, ok := definition.Id2State[t.TargetId]; len(t.Actions) == 0 && !ok {
				return nil, ErrStateEmptyTarget
			}
			transition := &Transition{
				Source:     sourceState,
				Condition:  t.Condition,
				ActionsRaw: t.Actions,
			}
			if len(t.Actions) > 0 {
				value := reflect.ValueOf(sourceState.BizStater)
				for _, action := range strings.Split(t.Actions, ",") {
					methodValue := value.MethodByName(action)
					if !methodValue.IsValid() {
						return nil, fmt.Errorf(actionInvalidErrFmt, t.SourceId, action)
					}
					transition.Actions = append(transition.Actions, methodValue)
				}
			}
			if targetState, ok := definition.Id2State[t.TargetId]; ok {
				transition.Target = targetState
			}
			transitions := sourceState.getTransitions()
			transitions = append(transitions, transition)
			sourceState.setTransitions(transitions)
		} else {
			return nil, ErrStateEmptySource
		}
	}
	definition.Transitions = recTransitions(definition.StartStateId, definition.Id2State)
	return definition, nil
}

func recTransitions(stateId string, id2State map[string]*State) []*Transition {
	var transitions []*Transition
	var targetStateIds []string
	for _, transition := range id2State[stateId].getTransitions() {
		transitions = append(transitions, transition)
		if transition.Target != nil {
			targetStateIds = append(targetStateIds, transition.Target.Id)
		}
	}
	for _, targetStateId := range targetStateIds {
		transitions = append(transitions, recTransitions(targetStateId, id2State)...)
	}
	return transitions
}
