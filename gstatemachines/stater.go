package gstatemachines

import (
	"context"
	"fmt"
	"reflect"

	"github.com/maja42/goval"
)

type BizStater interface {
	// Entry 进入状态时执行
	Entry(ctx context.Context, event Event, args ...interface{}) error
	// Exit 退出状态时执行
	Exit(ctx context.Context, event Event, args ...interface{}) error
}

type Transition struct {
	Source    *State
	Condition string
	Target    *State
	Actions   []reflect.Value
}

// Satisfied
// event 为变量池
// condition 为表达式
func (t *Transition) Satisfied(event Event) (bool, error) {
	return testExpression(t.Condition, event)
}

func testExpression(expression string, vars map[string]interface{}) (bool, error) {
	if expression == "" {
		return true, nil
	}
	eval := goval.NewEvaluator()
	result, err := eval.Evaluate(expression, vars, nil)
	if err != nil {
		return false, fmt.Errorf(conditionExpressionInvalidErrFmt, expression, err)
	}
	if v, ok := result.(bool); ok {
		return v, nil
	}
	return false, ErrConditionExpressionResultTypeUnmatch
}

type State struct {
	BizStater
	Id             string
	Desc           string
	Transitions    []*Transition
	isStart, isEnd bool
}

func (s *State) String() string {
	return fmt.Sprintf("[State] Id:%s, Desc:%s, IsStart:%t, IsEnd:%t",
		s.Id, s.Desc, s.isStart, s.isEnd)
}

func (s *State) getTransitions() []*Transition {
	return s.Transitions
}

func (s *State) setTransitions(ts []*Transition) {
	s.Transitions = ts
}

func (s *State) Transform(ctx context.Context, event Event, args ...interface{}) (*State, error) {
	for _, transition := range s.Transitions {
		ok, err := transition.Satisfied(event)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		if len(transition.Actions) > 0 {
			var inputArgs []reflect.Value
			inputArgs = append(inputArgs, reflect.ValueOf(ctx))
			inputArgs = append(inputArgs, reflect.ValueOf(event))
			for _, arg := range args {
				inputArgs = append(inputArgs, reflect.ValueOf(arg))
			}
			for _, action := range transition.Actions {
				outValues := action.Call(inputArgs)
				if !outValues[0].IsZero() {
					err = outValues[0].Interface().(error)
				}
				if err != nil {
					return nil, err
				}
			}
		}
		if transition.Target == nil {
			return nil, nil
		}
		return transition.Target, nil
	}
	return nil, ErrTransitionAllNotSatisfied
}
