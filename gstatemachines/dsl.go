package gstatemachines

import (
	"encoding/xml"
)

type StateMachineDSL struct {
	XMLName     xml.Name             `xml:"stateMachine"`
	Name        string               `xml:"name,attr"`
	Version     string               `xml:"version,attr"`
	States      []StateDSL           `xml:"states>state"`
	Transitions []StateTransitionDSL `xml:"transitions>transition"`
}

type StateDSL struct {
	Id      string `xml:"id,attr"`
	Desc    string `xml:",innerxml"`
	IsStart bool   `xml:"isStart,attr"`
	IsEnd   bool   `xml:"isEnd,attr"`
}

type StateTransitionDSL struct {
	Desc      string `xml:",innerxml"`
	SourceId  string `xml:"sourceId,attr"`
	Condition string `xml:"condition,attr"`
	TargetId  string `xml:"targetId,attr"`
	Actions   string `xml:"actions,attr"`
}

func toStateMachineDSL(dsl string) (StateMachineDSL, error) {
	stateMachineDSL := StateMachineDSL{}
	if err := xml.Unmarshal([]byte(dsl), &stateMachineDSL); err != nil {
		return StateMachineDSL{}, err
	}
	return stateMachineDSL, nil
}
