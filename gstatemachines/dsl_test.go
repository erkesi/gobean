package gstatemachines

import (
	"testing"
)

func TestToStateMachineDSL(t *testing.T) {
	dls := `<?xml version="1.0" encoding="utf-8"?>
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

	stateMachineDSL, err := ToStateMachineDSL(dls)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(stateMachineDSL)
}
