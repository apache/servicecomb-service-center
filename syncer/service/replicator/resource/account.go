package resource

import (
	"context"
	"fmt"

	"github.com/apache/servicecomb-service-center/pkg/log"
	servicerbac "github.com/apache/servicecomb-service-center/server/service/rbac"
	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"
	"github.com/go-chassis/cari/pkg/errsvc"
	rbacmodel "github.com/go-chassis/cari/rbac"
)

const (
	Account = "account"
)

func NewAccount(e *v1sync.Event) Resource {
	a := &account{
		event: e,
	}
	a.manager = a
	return a
}

type accountManager interface {
	CreateAccount(ctx context.Context, a *rbacmodel.Account) error
	GetAccount(ctx context.Context, name string) (*rbacmodel.Account, error)
	UpdateAccount(ctx context.Context, name string, account *rbacmodel.Account) error
	DeleteAccount(ctx context.Context, name string) error
}

type account struct {
	event *v1sync.Event

	input       *rbacmodel.Account
	accountName string

	cur *rbacmodel.Account

	defaultFailHandler

	manager accountManager
}

func (a *account) loadInput() error {
	a.input = new(rbacmodel.Account)
	callback := func() {
		a.accountName = a.input.Name
	}
	param := newInputParam(a.input, callback)

	return newInputLoader(
		a.event,
		param,
		param,
		param,
	).loadInput()
}

func (a *account) LoadCurrentResource(ctx context.Context) *Result {
	err := a.loadInput()
	if err != nil {
		return FailResult(err)
	}

	cur, err := a.manager.GetAccount(ctx, a.accountName)
	if err != nil {
		if errsvc.IsErrEqualCode(err, rbacmodel.ErrAccountNotExist) {
			return nil
		}
		return FailResult(err)
	}
	a.cur = cur
	return nil
}

func (a *account) NeedOperate(ctx context.Context) *Result {
	c := &checker{
		curNotNil: a.cur != nil,
		event:     a.event,
		updateTime: func() string {
			return a.cur.UpdateTime
		},
		resourceID: a.input.Name,
	}
	return c.needOperate(ctx)
}

func (a *account) CreateHandle(ctx context.Context) error {
	if a.cur != nil {
		log.Warn(fmt.Sprintf("create action but account exist, %s",
			a.accountName))
		return a.UpdateHandle(ctx)
	}
	return a.manager.CreateAccount(ctx, a.input)
}

func (a *account) UpdateHandle(ctx context.Context) error {
	if a.cur == nil {
		log.Warn(fmt.Sprintf("update action but account not exist, %s",
			a.accountName))
		return a.CreateHandle(ctx)
	}
	return a.manager.UpdateAccount(ctx, a.input.Name, a.input)
}

func (a *account) DeleteHandle(ctx context.Context) error {
	return a.manager.DeleteAccount(ctx, a.input.Name)
}

func (a *account) Operate(ctx context.Context) *Result {
	return newOperator(a).operate(ctx, a.event.Action)
}

func (a *account) CreateAccount(ctx context.Context, at *rbacmodel.Account) error {
	return servicerbac.CreateAccount(ctx, at)
}

func (a *account) GetAccount(ctx context.Context, name string) (*rbacmodel.Account, error) {
	return servicerbac.GetAccount(ctx, name)
}

func (a *account) UpdateAccount(ctx context.Context, name string, account *rbacmodel.Account) error {
	return servicerbac.UpdateAccount(ctx, name, account)
}

func (a *account) DeleteAccount(ctx context.Context, name string) error {
	return servicerbac.DeleteAccount(ctx, name)
}
