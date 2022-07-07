package gstatemachines

import "errors"

var ErrStateNotExist = errors.New("state not exist")
var ErrStateInvalid = errors.New("state invalid")

var ErrTransitionAllNotSatisfied = errors.New("transition all not satisfied")
