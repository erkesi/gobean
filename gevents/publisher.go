package gevents

import (
	"context"
)

type PublishOption func(*pubOptions)

func WithMustHaveSubscriber() PublishOption {
	return func(option *pubOptions) {
		option.mustHaveSubscriber = true
	}
}

type pubOptions struct {
	mustHaveSubscriber bool
}

type Publisher interface {
	Publish(ctx context.Context, event interface{}, opts ...PublishOption) error
}
