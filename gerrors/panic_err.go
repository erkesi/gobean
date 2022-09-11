package gerrors

import "fmt"

type PanicError struct {
	msg   string
	stack string
}

func NewPanicError(msg interface{}, stack string) *PanicError {
	return &PanicError{msg: fmt.Sprintf("%v", msg), stack: stack}
}

func (pe *PanicError) Error() string {
	return fmt.Sprintf("panic: %s, %s", pe.msg, pe.stack)
}
