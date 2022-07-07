package gstatemachines

import (
	"encoding/xml"
)

type StateMachineDSL struct {
	XMLName     xml.Name            `xml:"stateMachine"`
	Version     string              `xml:"version,attr"`
	States      []StateDSL          `xml:"states>state"`
	Transitions []TaskTransitionDSL `xml:"transitions>transition"`
}

type StateDSL struct {
	Id      string `xml:"id,attr"`
	Desc    string `xml:",innerxml"`
	IsStart bool   `xml:"isStart,attr"`
	IsEnd   bool   `xml:"isEnd,attr"`
}

type TaskTransitionDSL struct {
	Desc      string `xml:",innerxml"`
	SourceId  string `xml:"sourceId,attr"`
	Condition string `xml:"condition,attr"`
	TargetId  string `xml:"targetId,attr"`
}

func ToStateMachineDSL(dsl string) (StateMachineDSL, error) {
	stateMachineDSL := StateMachineDSL{}
	if err := xml.Unmarshal([]byte(dsl), &stateMachineDSL); err != nil {
		return StateMachineDSL{}, err
	}
	return stateMachineDSL, nil
}
