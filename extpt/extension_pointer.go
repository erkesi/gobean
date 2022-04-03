package extpt

type ExtensionPointer interface {
	Match(values ...interface{}) bool
}
