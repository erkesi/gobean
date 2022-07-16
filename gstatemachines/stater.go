package gstatemachines

import (
	"context"
	"fmt"
	"github.com/erkesi/gobean/glogs"
	"github.com/maja42/goval"
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

	// GetId 获取id
	GetId() string
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
	Source    Stater
	Condition string
	Target    Stater
}

// Satisfied
// event 为变量池
// condition 为表达式
func (t *Transition) Satisfied(event Event) (bool, error) {
	r, err := testExpression(t.Condition, event)
	if glogs.Log != nil {
		glogs.Log.Debugf("gstatemachines: check condition: %s; result is %t", t.Condition, r)
	}
	return r, err
}

func testExpression(expression string, vars map[string]interface{}) (bool, error) {
	eval := goval.NewEvaluator()
	result, err := eval.Evaluate(expression, vars, nil)
	if err != nil {
		return false, fmt.Errorf(conditionExpressionInvalidErrFmt, expression, err)
	}
	if v, ok := result.(bool); ok {
		return v, nil
	}
	if glogs.Log != nil {
		glogs.Log.Debugf("gstatemachines: expression result transfer error")
	}
	return false, ErrConditionExpressionResultTypeUnmatch
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

func (s *State) GetId() string {
	return s.Id
}

func (s *State) GetTransitions() []*Transition {
	return s.Transitions
}

func (s *State) SetTransitions(ts []*Transition) {
	s.Transitions = ts
}

func (s *State) Transform(ctx context.Context, event Event) (Stater, error) {
	for _, transition := range s.Transitions {
		ok, err := transition.Satisfied(event)
		if err != nil {
			return nil, err
		}
		if ok {
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
