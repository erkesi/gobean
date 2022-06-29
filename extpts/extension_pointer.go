package extpts

type ExtensionPointer interface {
	Match(values ...interface{}) bool
}
