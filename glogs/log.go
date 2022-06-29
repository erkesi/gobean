package glogs

var Log Logger

func Init(logger Logger) {
	Log = logger
}

type Logger interface {
	Debugf(format string, v ...interface{})
}
