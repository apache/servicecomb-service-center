package resource

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/apache/servicecomb-service-center/eventbase/datasource"
	"github.com/apache/servicecomb-service-center/eventbase/model"
	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"

	"github.com/go-chassis/cari/sync"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("valid subject", func(t *testing.T) {
		e := &v1sync.Event{
			Subject: Account,
		}
		r, result := New(e)
		if assert.Nil(t, result) {
			assert.NotNil(t, r)
		}
	})
	t.Run("invalid subject", func(t *testing.T) {
		e := &v1sync.Event{
			Subject: "not exist",
		}
		r, result := New(e)
		if assert.Equal(t, Skip, result.Status) {
			assert.Nil(t, r)
		}
	})
}

type forkActionHandler struct {
}

func (f forkActionHandler) CreateHandle(_ context.Context) error {
	return nil
}

func (f forkActionHandler) UpdateHandle(_ context.Context) error {
	return nil
}

func (f forkActionHandler) DeleteHandle(_ context.Context) error {
	return nil
}

func Test_newOperator(t *testing.T) {
	f := new(forkActionHandler)
	o := newOperator(f)
	ctx := context.TODO()
	result := o.operate(ctx, sync.CreateAction)
	assert.Equal(t, SuccessResult(), result)

	result = o.operate(ctx, sync.UpdateAction)
	assert.Equal(t, SuccessResult(), result)

	result = o.operate(ctx, sync.DeleteAction)
	assert.Equal(t, SuccessResult(), result)
}

func TestFailResult(t *testing.T) {
	r := FailResult(errors.New("demo"))
	assert.Equal(t, Fail, r.Status)
	assert.Equal(t, "demo", r.Message)
}

func TestSuccessResult(t *testing.T) {
	r := SuccessResult()
	assert.Equal(t, Success, r.Status)
}

func TestSkipResult(t *testing.T) {
	r := SkipResult()
	assert.Equal(t, Skip, r.Status)
}

func TestNonImplementResult(t *testing.T) {
	r := NonImplementResult()
	assert.Equal(t, NonImplement, r.Status)
}

func TestNewResult(t *testing.T) {
	r := NewResult(Skip, "").WithMessage("hello").WithEventID("xxx")
	assert.Equal(t, Skip, r.Status)
	assert.Equal(t, "hello", r.Message)
	assert.Equal(t, "xxx", r.EventID)
	assert.Equal(t, "eventID: xxx, status 2:skip, message: hello", r.Flag())
}

func TestNewInputLoader(t *testing.T) {
	t.Run("string input", func(t *testing.T) {
		input := new(string)
		var ok bool
		cre := newInputParam(input, func() {
			ok = true
		})
		e := &v1sync.Event{
			Id:        "xxx",
			Action:    sync.CreateAction,
			Subject:   "test",
			Opts:      nil,
			Value:     []byte("hello"),
			Timestamp: 0,
		}
		loader := newInputLoader(e, cre, cre, cre)
		err := loader.loadInput()
		if assert.Nil(t, err) {
			assert.Equal(t, "hello", *input)
			assert.True(t, ok)
		}
	})
	t.Run("object", func(t *testing.T) {
		input := make(map[string]string)
		var ok bool
		cre := newInputParam(&input, func() {
			ok = true
		})
		e := &v1sync.Event{
			Id:        "xxx",
			Action:    sync.UpdateAction,
			Subject:   "test",
			Opts:      nil,
			Value:     []byte(`{"hello": "world"}`),
			Timestamp: 0,
		}
		loader := newInputLoader(e, cre, cre, cre)
		err := loader.loadInput()
		if assert.Nil(t, err) {
			assert.Equal(t, map[string]string{"hello": "world"}, input)
			assert.True(t, ok)
		}
	})

	t.Run("delete case", func(t *testing.T) {
		input := new(string)
		var ok bool
		cre := newInputParam(input, func() {
			ok = true
		})
		e := &v1sync.Event{
			Id:        "xxx",
			Action:    sync.DeleteAction,
			Subject:   "test",
			Opts:      nil,
			Value:     []byte("hello"),
			Timestamp: 0,
		}
		loader := newInputLoader(e, cre, cre, cre)
		err := loader.loadInput()
		if assert.Nil(t, err) {
			assert.Equal(t, "hello", *input)
			assert.True(t, ok)
		}
	})
}

type mockTombstoneLoader struct {
	ts *sync.Tombstone
}

func (f *mockTombstoneLoader) get(_ context.Context, _ *model.GetTombstoneRequest) (*sync.Tombstone, error) {
	if f.ts == nil {
		return nil, datasource.ErrTombstoneNotExists
	}
	return f.ts, nil
}

func TestNeedOperate(t *testing.T) {
	t.Run("curNotNil true", func(t *testing.T) {
		e := &v1sync.Event{
			Id:        "xxx",
			Action:    sync.CreateAction,
			Subject:   Instance,
			Opts:      nil,
			Value:     nil,
			Timestamp: v1sync.Timestamp(),
		}
		c := &checker{
			curNotNil: true,
			event:     e,
			updateTime: func() string {
				return strconv.FormatInt(time.Now().Add(-time.Minute).Unix(), 10)
			},
			resourceID: "",
		}
		ctx := context.TODO()
		r := c.needOperate(ctx)
		assert.Nil(t, r)

		c.updateTime = func() string {
			return strconv.FormatInt(time.Now().Add(time.Minute).Unix(), 10)
		}

		r = c.needOperate(ctx)
		if assert.NotNil(t, r) {
			assert.Equal(t, Skip, r.Status)
		}
	})

	t.Run("curNotNil false", func(t *testing.T) {
		e := &v1sync.Event{
			Id:        "xxx",
			Action:    sync.CreateAction,
			Subject:   Instance,
			Opts:      nil,
			Value:     nil,
			Timestamp: v1sync.Timestamp(),
		}
		c := &checker{
			curNotNil: false,
			event:     e,
			updateTime: func() string {
				return strconv.FormatInt(time.Now().Add(-time.Minute).Unix(), 10)
			},
			resourceID: "",
		}
		ctx := context.TODO()
		r := c.needOperate(ctx)
		assert.Nil(t, r)

		c.event.Action = sync.UpdateAction
		r = c.needOperate(ctx)
		assert.Nil(t, r)

		c.resourceID = "xxx"
		c.tombstoneLoader = &mockTombstoneLoader{
			ts: &sync.Tombstone{
				Timestamp: time.Now().Add(time.Minute).UnixNano(),
			},
		}

		r = c.needOperate(ctx)
		if assert.NotNil(t, r) {
			assert.Equal(t, Skip, r.Status)
		}

		c.tombstoneLoader = &mockTombstoneLoader{
			ts: &sync.Tombstone{
				Timestamp: time.Now().Add(-time.Minute).UnixNano(),
			},
		}

		r = c.needOperate(ctx)
		assert.Nil(t, r)
	})
}
