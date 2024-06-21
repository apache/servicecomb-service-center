package resource

import (
	"context"
	"sync"

	pb "github.com/go-chassis/cari/discovery"
	ev "github.com/go-chassis/cari/env"
	"github.com/go-chassis/cari/pkg/errsvc"

	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"
)

var envManager environmentManager

const (
	Environment = "environment"
)

func NewEnvironment(e *v1sync.Event) Resource {
	m := &environment{
		event: e,
	}
	m.manager = environmentManage()
	return m
}

var envOnce sync.Once

func environmentManage() environmentManager {
	envOnce.Do(InitEnvironmentManager)
	return envManager
}

func InitEnvironmentManager() {
	envManager = new(metadataManage)
}

type environment struct {
	event *v1sync.Event

	createInput *ev.CreateEnvironmentRequest
	updateInput *ev.UpdateEnvironmentRequest
	deleteInput *ev.DeleteEnvironmentRequest

	envID string

	cur *ev.Environment

	manager environmentManager

	defaultFailHandler
}

type environmentManager interface {
	RegisterEnvironment(ctx context.Context, request *ev.CreateEnvironmentRequest) (*ev.CreateEnvironmentResponse, error)
	GetEnvironment(ctx context.Context, in *ev.GetEnvironmentRequest) (*ev.Environment, error)
	UpdateEnvironment(ctx context.Context, request *ev.UpdateEnvironmentRequest) error
	UnregisterEnvironment(ctx context.Context, request *ev.DeleteEnvironmentRequest) error
}

func (e *environment) loadInput() error {
	e.createInput = new(ev.CreateEnvironmentRequest)
	cre := newInputParam(e.createInput, func() {
		e.envID = e.createInput.Environment.ID
	})

	e.updateInput = new(ev.UpdateEnvironmentRequest)
	upd := newInputParam(e.updateInput, func() {
		e.envID = e.updateInput.Environment.ID
	})

	e.deleteInput = new(ev.DeleteEnvironmentRequest)
	del := newInputParam(e.deleteInput, func() {
		e.envID = e.deleteInput.EnvironmentId
	})

	return newInputLoader(
		e.event,
		cre,
		upd,
		del,
	).loadInput()
}

func (e *environment) LoadCurrentResource(ctx context.Context) *Result {
	err := e.loadInput()
	if err != nil {
		return FailResult(err)
	}

	cur, err := e.manager.GetEnvironment(ctx, &ev.GetEnvironmentRequest{
		EnvironmentId: e.envID,
	})
	if err != nil {
		if errsvc.IsErrEqualCode(err, pb.ErrEnvironmentNotExists) {
			return nil
		}

		return FailResult(err)
	}
	e.cur = cur
	return nil
}

func (e *environment) NeedOperate(ctx context.Context) *Result {
	c := &checker{
		curNotNil: e.cur != nil,
		event:     e.event,
		updateTime: func() (int64, error) {
			return formatUpdateTimeSecond(e.cur.ModTimestamp)
		},
		resourceID: e.envID,
	}
	c.tombstoneLoader = c
	return c.needOperate(ctx)
}

func (e *environment) CreateHandle(ctx context.Context) error {
	_, err := e.manager.RegisterEnvironment(ctx, e.createInput)
	return err
}

func (e *environment) UpdateHandle(ctx context.Context) error {
	return e.manager.UpdateEnvironment(ctx, e.updateInput)
}

func (e *environment) DeleteHandle(ctx context.Context) error {
	return e.manager.UnregisterEnvironment(ctx, e.deleteInput)
}

func (e *environment) Operate(ctx context.Context) *Result {
	return newOperator(e).operate(ctx, e.event.Action)
}
