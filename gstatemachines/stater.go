package gstatemachines

import (
	"context"
	"fmt"
)

type Stater interface {
	BaseStater
	// Transition 流转
	Transition(ctx context.Context, event Event) (Stater, error)
	// Validate 校验
	Validate() error
}

type BaseStater interface {
	// Entry 进入状态时执行
	Entry(ctx context.Context, event Event, args ...interface{}) error
	// Action 当时状态时执行
	Action(ctx context.Context, event Event, args ...interface{}) error
	// Exit 退出状态时执行
	Exit(ctx context.Context, event Event, args ...interface{}) error
}

type Transition struct {
	Condition string
	Target    Stater
}

// Satisfied todo https://github.com/maja42/goval
func (t *Transition) Satisfied(event Event) bool {
	return true
}

type State struct {
	BaseStater
	Id             string
	Desc           string
	Transitions    []*Transition
	IsStart, IsEnd bool
}

func (s *State) String() string {
	return fmt.Sprintf("[State] Id:%s, Desc:%s, IsStart:%t, IsEnd:%t", s.Id, s.Desc, s.IsStart, s.IsEnd)
}

func (s *State) Transition(ctx context.Context, event Event) (Stater, error) {
	for _, transition := range s.Transitions {
		if transition.Satisfied(event) {
			return transition.Target, nil
		}
	}
	return nil, ErrTransitionAllNotSatisfied
}

func (s *State) Validate() error {
	if s.IsEnd {
		return ErrStateInvalid
	}
	return nil
}
