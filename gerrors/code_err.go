package gerrors

import (
	"errors"
	"fmt"
)

type CodeError struct {
	code int
	msg  string
	err  error
}

func NewCodeErrorOfMessage(code int, msg string) *CodeError {
	return NewCodeErrorWarp(code, msg, nil)
}

func NewCodeErrorOfMessagef(code int, format string, args ...interface{}) *CodeError {
	msg := fmt.Sprintf(format, args...)
	return NewCodeErrorWarp(code, msg, nil)
}

func NewCodeError(code int, err error) *CodeError {
	return NewCodeErrorWarp(code, err.Error(), err)
}

func NewCodeErrorWarp(code int, msg string, err error) *CodeError {
	return &CodeError{code: code, msg: msg, err: err}
}

func (ce *CodeError) Code() int {
	return ce.code
}

func (ce *CodeError) Msg() string {
	return ce.msg
}

func (ce *CodeError) Err() error {
	if ce.err == nil {
		return errors.New(ce.msg)
	}
	return ce.err
}

func (ce *CodeError) Error() string {
	return fmt.Sprintf("gerrors: code:%d, msg:%s, err:%v", ce.code, ce.msg, ce.err)
}
