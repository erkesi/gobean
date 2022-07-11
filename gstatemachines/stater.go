package gstatemachines

import (
	"context"
	"fmt"
	"github.com/maja42/goval"
	"strconv"
)

type Stater interface {
	BaseStater
	// Transform 流转
	Transform(ctx context.Context, event Event) (Stater, error)
	// Validate 校验
	Validate() error

	// GetTransitions 获取转换列表
	GetTransitions() []*Transition

	// SetTransitions 更新转换列表
	SetTransitions([]*Transition)
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
// event 为变量池
// condition 为表达式
func (t *Transition) Satisfied(event Event) bool {
	r := testExpression(t.Condition, event)
	DebugLog("Check condition: " + t.Condition + " ; result is " + strconv.FormatBool(r))
	return r
}

func testExpression(expression string, vars map[string]interface{}) bool {
	eval := goval.NewEvaluator()
	result, err := eval.Evaluate(expression, vars, nil)
	if err != nil {
		fmt.Println("Evaluate error ")
		return false
	}
	if v, ok := result.(bool); ok {

		return v
	}
	fmt.Println("transfer error ")
	return false
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

func (s *State) GetTransitions() []*Transition {
	return s.Transitions
}

func (s *State) SetTransitions(ts []*Transition) {
	s.Transitions = ts
}

func (s *State) Transform(ctx context.Context, event Event) (Stater, error) {
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
