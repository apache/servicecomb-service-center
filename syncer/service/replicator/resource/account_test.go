package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"

	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/go-chassis/cari/rbac"
	rbacmodel "github.com/go-chassis/cari/rbac"
	"github.com/go-chassis/cari/sync"
	"github.com/stretchr/testify/assert"
)

type mockAccount struct {
	accounts map[string]*rbac.Account
}

func (f mockAccount) CreateAccount(_ context.Context, a *rbac.Account) error {
	_, ok := f.accounts[a.Name]
	if ok {
		return rbacmodel.NewError(rbacmodel.ErrAccountConflict, "account exist")
	}

	f.accounts[a.Name] = a
	return nil
}

func (f mockAccount) GetAccount(_ context.Context, name string) (*rbac.Account, error) {
	result, ok := f.accounts[name]
	if !ok {
		msg := fmt.Sprintf("account [%s] not exist", name)
		return nil, rbacmodel.NewError(rbacmodel.ErrAccountNotExist, msg)
	}
	return result, nil
}

func (f mockAccount) UpdateAccount(_ context.Context, name string, account *rbac.Account) error {
	_, ok := f.accounts[name]
	if !ok {
		msg := fmt.Sprintf("account [%s] not exist", name)
		return rbacmodel.NewError(rbacmodel.ErrAccountNotExist, msg)
	}
	f.accounts[name] = account
	return nil
}

func (f mockAccount) DeleteAccount(_ context.Context, name string) error {
	_, ok := f.accounts[name]
	if !ok {
		msg := fmt.Sprintf("account [%s] not exist", name)
		return rbacmodel.NewError(rbacmodel.ErrAccountNotExist, msg)
	}
	delete(f.accounts, name)
	return nil
}

func TestOperateAccount(t *testing.T) {
	t.Run("create update delete case", func(t *testing.T) {
		createTime := strconv.FormatInt(time.Now().Unix(), 10)
		input := &rbac.Account{
			ID:         "xxxx",
			Name:       "test",
			Password:   "xxx",
			Roles:      []string{"admin"},
			Status:     "active",
			CreateTime: createTime,
			UpdateTime: createTime,
		}
		value, _ := json.Marshal(input)
		id, _ := v1sync.NewEventID()
		e := &v1sync.Event{
			Id:        id,
			Action:    sync.CreateAction,
			Subject:   Account,
			Opts:      nil,
			Value:     value,
			Timestamp: v1sync.Timestamp(),
		}
		a := &account{
			event: e,
		}
		a.manager = &mockAccount{
			accounts: make(map[string]*rbac.Account),
		}
		ctx := context.Background()
		result := a.LoadCurrentResource(ctx)
		if assert.Nil(t, result) {
			result = a.NeedOperate(ctx)
			if assert.Nil(t, result) {
				result = a.Operate(ctx)
				if assert.NotNil(t, result) && assert.Equal(t, Success, result.Status) {
					data, err := a.manager.GetAccount(ctx, "test")
					assert.Nil(t, err)
					assert.NotNil(t, data)
				}
			}
		}

		updateTime := strconv.FormatInt(time.Now().Unix(), 10)
		input = &rbac.Account{
			ID:         "xxxx",
			Name:       "test",
			Password:   "xxx",
			Roles:      []string{"developer"},
			Status:     "active",
			CreateTime: createTime,
			UpdateTime: updateTime,
		}
		value, _ = json.Marshal(input)

		id, _ = v1sync.NewEventID()
		e1 := &v1sync.Event{
			Id:        id,
			Action:    sync.UpdateAction,
			Subject:   Account,
			Opts:      nil,
			Value:     value,
			Timestamp: v1sync.Timestamp(),
		}
		a1 := &account{
			event:   e1,
			manager: a.manager,
		}
		result = a1.LoadCurrentResource(ctx)
		if assert.Nil(t, result) {
			result = a1.NeedOperate(ctx)
			if assert.Nil(t, result) {
				result = a1.Operate(ctx)
				if assert.NotNil(t, result) && assert.Equal(t, Success, result.Status) {
					data, err := a1.manager.GetAccount(ctx, "test")
					assert.Nil(t, err)
					assert.NotNil(t, data)
					assert.Equal(t, []string{"developer"}, data.Roles)
				}
			}
		}

		id, _ = v1sync.NewEventID()
		e2 := &v1sync.Event{
			Id:        id,
			Action:    sync.DeleteAction,
			Subject:   Account,
			Opts:      nil,
			Value:     value,
			Timestamp: v1sync.Timestamp(),
		}
		a2 := &account{
			event:   e2,
			manager: a.manager,
		}
		result = a2.LoadCurrentResource(ctx)
		if assert.Nil(t, result) {
			result = a2.NeedOperate(ctx)
			if assert.Nil(t, result) {
				result = a2.Operate(ctx)
				if assert.NotNil(t, result) && assert.Equal(t, Success, result.Status) {
					_, err := a1.manager.GetAccount(ctx, "test")
					assert.NotNil(t, err)
					assert.True(t, errsvc.IsErrEqualCode(err, rbacmodel.ErrAccountNotExist))
				}
			}
		}
	})
}

func TestNewAccount(t *testing.T) {
	id, _ := v1sync.NewEventID()
	e := &v1sync.Event{
		Id:        id,
		Action:    sync.DeleteAction,
		Subject:   Account,
		Opts:      nil,
		Value:     nil,
		Timestamp: v1sync.Timestamp(),
	}
	a := NewAccount(e)
	assert.NotNil(t, a)
}
