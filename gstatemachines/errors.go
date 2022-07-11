package gstatemachines

import "errors"

var ErrStateNotExist = errors.New("gstatemachines: state not exist")
var ErrStateInvalid = errors.New("gstatemachines: state invalid")

var ErrTransitionAllNotSatisfied = errors.New("gstatemachines: transition all not satisfied")

var ErrStateEmptyTarget = errors.New("gstatemachines: state target invalid")
var ErrStateEmptySource = errors.New("gstatemachines: state source invalid")

var ErrConditionExpressionResultTypeUnmatch = errors.New("gstatemachines: expression result type must be bool")

const conditionExpressionInvalidErrFmt = "gstatemachines: condition expression invalid, expression is: %s, err: %w"
