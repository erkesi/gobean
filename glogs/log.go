package glogs

import "context"

var Log Logger

func Init(logger Logger) {
	Log = logger
}

type Logger interface {
	Debugf(ctx context.Context, format string, v ...interface{})
	Errorf(ctx context.Context, format string, v ...interface{})
}
