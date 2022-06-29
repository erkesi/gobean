package gextpts

type ExtensionPointer interface {
	Match(values ...interface{}) bool
}
