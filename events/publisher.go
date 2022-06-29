package events

import (
	"context"
)

type PublishOpt func(*Option)

func MustHaveSubscriber() PublishOpt {
	return func(option *Option) {
		option.mustHaveSubscriber = true
	}
}

type Option struct {
	mustHaveSubscriber bool
}

type Publisher interface {
	Publish(ctx context.Context, event interface{}, opts ...PublishOpt) error
}
