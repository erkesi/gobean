package gstatemachines

import (
	"context"
	"fmt"
	"testing"
)

const dls = `<?xml version="1.0" encoding="utf-8"?>
<stateMachine version="1">
    <states>
        <state id="Start" isStart="true">start</state>
        <state id="Task1">task1</state>
        <state id="Reject" isEnd="true">reject</state>
        <state id="End" isEnd="true">end</state>
    </states>
    <transitions>
        <transition sourceId="Start" targetId="Task1" condition="operation==&quot;toTask1&quot;">Start->Task1</transition>
        <transition sourceId="Task1" targetId="Reject" condition="operation==&quot;Reject&quot;">Task1->Reject</transition>
        <transition sourceId="Task1" targetId="End" condition="operation==&quot;End&quot;">Task1->End</transition>
    </transitions>
</stateMachine>`

func TestStateMachine_Generate(t *testing.T) {

	id2State := make(map[string]Stater)
	id2State["Start"] = &State{
		BaseStater: &Task1State{},
		Id:         "Start",
		Desc:       "Start",
		IsStart:    true,
		IsEnd:      false,
	}
	id2State["Task1"] = &State{
		BaseStater: &Task1State{},
		Id:         "Task1",
		Desc:       "Task1",
		IsStart:    false,
		IsEnd:      false,
	}
	id2State["Reject"] = &State{
		BaseStater: &RejectState{},
		Id:         "Reject",
		Desc:       "Reject",
		IsStart:    false,
		IsEnd:      true,
	}
	id2State["End"] = &State{
		BaseStater: &EndState{},
		Id:         "End",
		Desc:       "End",
		IsStart:    false,
		IsEnd:      true,
	}

	definition, err := ToStateMachineDefinition(dls, id2State)
	if err != nil {
		t.Fatal(err)
	}
	stateMachine := &StateMachine{
		Definition: definition,
	}
	enableDebug()
	err = stateMachine.Execute(context.TODO(), "Task1", map[string]interface{}{"operation": "Reject"})
	if err != nil {
		t.Fatal(err)
	}

	err = stateMachine.Execute(context.TODO(), "Task1", map[string]interface{}{"operation": "End"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestStateMachine_Execute(t *testing.T) {
	/*definition, err := ToStateMachineDefinition(dls, map[stateId]*BaseStater )
	if err != nil {
		t.Fatal(err)
	}*/
	stateMachine := &StateMachine{
		Definition: &StateMachineDefinition{
			Name:    "flow",
			Version: "1",
			Id2State: map[string]Stater{"Task1": &State{
				BaseStater: &Task1State{},
				Id:         "Start",
				Desc:       "Start",
				IsStart:    false,
				IsEnd:      false,
			}},
		},
	}

	err := stateMachine.Execute(context.TODO(), "Task1", map[string]interface{}{"operation": "Reject"})
	if err != nil {
		t.Fatal(err)
	}
}

type Task1State struct {
}

func (s *Task1State) Entry(ctx context.Context, event Event, args ...interface{}) error {
	fmt.Printf("Entry Task1State: %v\n", s)
	return nil
}

func (s *Task1State) Action(ctx context.Context, event Event, args ...interface{}) error {
	fmt.Printf("Action Task1State: %v\n", s)
	return nil
}

func (s *Task1State) Exit(ctx context.Context, event Event, args ...interface{}) error {
	fmt.Printf("Exit Task1State: %v\n", s)
	return nil
}

type RejectState struct {
}

func (s *RejectState) Entry(ctx context.Context, event Event, args ...interface{}) error {
	fmt.Printf("RejectState Entry: %v\n", s)
	return nil
}

func (s *RejectState) Action(ctx context.Context, event Event, args ...interface{}) error {
	fmt.Printf("RejectState Action: %v\n", s)
	return nil
}

func (s *RejectState) Exit(ctx context.Context, event Event, args ...interface{}) error {
	fmt.Printf("RejectState Exit: %v\n", s)
	return nil
}

type EndState struct {
}

func (s *EndState) Entry(ctx context.Context, event Event, args ...interface{}) error {
	fmt.Printf("EndState Entry: %v\n", s)
	return nil
}

func (s *EndState) Action(ctx context.Context, event Event, args ...interface{}) error {
	fmt.Printf("EndState Action: %v\n", s)
	return nil
}

func (s *EndState) Exit(ctx context.Context, event Event, args ...interface{}) error {
	fmt.Printf("EndState Exit: %v\n", s)
	return nil
}
