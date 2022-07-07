package gextpts

import "context"

type ExtensionPointer interface {
	Match(ctx context.Context, values ...interface{}) bool
}
