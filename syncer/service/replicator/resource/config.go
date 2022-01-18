package resource

import (
	"context"

	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"
)

const (
	Config = "config"
)

func NewConfig(e *v1sync.Event) Resource {
	return &config{
		event: e,
	}
}

type config struct {
	event *v1sync.Event

	defaultFailHandler
}

func (c *config) LoadCurrentResource(context.Context) *Result {
	return NonImplementResult()
}

func (c *config) NeedOperate(context.Context) *Result {
	return NonImplementResult()
}

func (c *config) Operate(context.Context) *Result {
	return NonImplementResult()
}
