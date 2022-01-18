package event

import (
	"context"
	"testing"
	"time"

	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"
	"github.com/apache/servicecomb-service-center/syncer/service/replicator/resource"
	"github.com/stretchr/testify/assert"
)

func TestWork(t *testing.T) {
	r := &forkReplicator{
		results: &v1sync.Results{
			Results: map[string]*v1sync.Result{
				"xxx1": {
					Code:    resource.Success,
					Message: "ok",
				},
				"xxx2": {
					Code:    resource.Success,
					Message: "ok",
				},
			},
		},
		err: nil,
	}
	Work()
	em := m.(*eventManager)
	em.replicator = r

	var f *forkResources
	resource.RegisterResources("fork", func(event *v1sync.Event) resource.Resource {
		f = &forkResources{
			loadCurrentResourceResult: nil,
			needOperateResult:         nil,
			operateResult:             resource.SuccessResult(),
			result:                    make(chan struct{}),
		}
		return f
	})
	Send(&Event{
		Event: &v1sync.Event{
			Id:        "xxx1",
			Action:    "create",
			Subject:   "fork",
			Opts:      nil,
			Timestamp: v1sync.Timestamp(),
		},
	})

	ctx := context.TODO()
	Publish(ctx, "create", "fork", map[string]string{
		"hello": "world",
	})
	result := make(chan *Result, 1)
	Send(&Event{
		Event: &v1sync.Event{
			Id:        "xxx2",
			Action:    "create",
			Subject:   "fork",
			Opts:      nil,
			Timestamp: v1sync.Timestamp(),
		},
		Result: result,
	})

	data := <-result
	if assert.NotNil(t, data.Data) {
		assert.Equal(t, resource.Success, data.Data.Code)
		assert.Equal(t, "ok", data.Data.Message)
	}
	time.Sleep(time.Second)
	if assert.NotNil(t, f) {
		<-f.result
	}
}

type forkReplicator struct {
	results *v1sync.Results
	err     error
}

func (f *forkReplicator) Replicate(_ context.Context, _ *v1sync.EventList) (*v1sync.Results, error) {
	return f.results, f.err
}

func (f *forkReplicator) Persist(_ context.Context, _ *v1sync.EventList) []*resource.Result {
	return nil
}

type forkResources struct {
	loadCurrentResourceResult *resource.Result
	needOperateResult         *resource.Result
	operateResult             *resource.Result
	result                    chan struct{}
}

func (f *forkResources) LoadCurrentResource(_ context.Context) *resource.Result {
	return f.loadCurrentResourceResult
}

func (f *forkResources) NeedOperate(_ context.Context) *resource.Result {
	return f.needOperateResult
}

func (f *forkResources) Operate(_ context.Context) *resource.Result {
	return f.operateResult
}

func (f forkResources) FailHandle(_ context.Context, _ int32) (*v1sync.Event, error) {
	f.result <- struct{}{}
	return nil, nil
}

func TestNewManager(t *testing.T) {
	nm := NewManager(ManagerInternal(defaultInternal), Replicator(new(forkReplicator)))
	assert.NotNil(t, nm)
}
