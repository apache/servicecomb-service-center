package events

import "context"

type contextEvent struct {
	kind string
	ctx  context.Context
}

func NewContextEvent(kind string, ctx context.Context) ContextEvent {
	return &contextEvent{
		kind: kind,
		ctx:  ctx,
	}
}

func (e *contextEvent) Type() string {
	return e.kind
}

func (e *contextEvent) Context() context.Context {
	return e.ctx
}
