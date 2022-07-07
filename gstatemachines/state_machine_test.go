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

func TestStateMachine_Execute(t *testing.T) {
	/*definition, err := ToStateMachineDefinition(dls, map[stateId]*BaseStater )
	if err != nil {
		t.Fatal(err)
	}*/
	stateMachine := &StateMachine{
		Definition: &StateMachineDefinition{
			Name:    "flow",
			Version: "1",
			Id2State: map[string]Stater{"Start": &State{
				BaseStater: &StartState{},
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

type StartState struct {
}

func (s *StartState) Entry(ctx context.Context, event Event, args ...interface{}) error {
	fmt.Printf("Entry: %v", s)
	return nil
}

func (s *StartState) Action(ctx context.Context, event Event, args ...interface{}) error {
	fmt.Printf("Action: %v", s)
	return nil
}

func (s *StartState) Exit(ctx context.Context, event Event, args ...interface{}) error {
	fmt.Printf("Exit: %v", s)
	return nil
}

type Task1State struct {
}

func (s *Task1State) Entry(ctx context.Context, event Event, args ...interface{}) error {
	fmt.Printf("Entry: %v", s)
	return nil
}

func (s *Task1State) Action(ctx context.Context, event Event, args ...interface{}) error {
	fmt.Printf("Action: %v", s)
	return nil
}

func (s *Task1State) Exit(ctx context.Context, event Event, args ...interface{}) error {
	fmt.Printf("Exit: %v", s)
	return nil
}

type RejectState struct {
}

func (s *RejectState) Entry(ctx context.Context, event Event, args ...interface{}) error {
	fmt.Printf("Entry: %v", s)
	return nil
}

func (s *RejectState) Action(ctx context.Context, event Event, args ...interface{}) error {
	fmt.Printf("Action: %v", s)
	return nil
}

func (s *RejectState) Exit(ctx context.Context, event Event, args ...interface{}) error {
	fmt.Printf("Exit: %v", s)
	return nil
}

type EndState struct {
}

func (s *EndState) Entry(ctx context.Context, event Event, args ...interface{}) error {
	fmt.Printf("Entry: %v", s)
	return nil
}

func (s *EndState) Action(ctx context.Context, event Event, args ...interface{}) error {
	fmt.Printf("Action: %v", s)
	return nil
}

func (s *EndState) Exit(ctx context.Context, event Event, args ...interface{}) error {
	fmt.Printf("Exit: %v", s)
	return nil
}
