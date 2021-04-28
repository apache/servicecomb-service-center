package mongo

import (
	"context"
	"github.com/go-chassis/cari/discovery"
)

type InstanceRegisterEvent struct {
	Ctx        context.Context
	Request    *discovery.RegisterInstanceRequest
	isCustomID bool
	failedTime int
}
