package rpc

import (
	"context"
	"reflect"
	"testing"

	"github.com/go-chassis/cari/rbac"
	"github.com/stretchr/testify/assert"

	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"
	"github.com/apache/servicecomb-service-center/syncer/config"
	"github.com/apache/servicecomb-service-center/syncer/service/replicator/resource"
)

type testReplicator struct{}

func (testReplicator) Replicate(ctx context.Context, el *v1sync.EventList) (*v1sync.Results, error) {
	return &v1sync.Results{Results: map[string]*v1sync.Result{
		"constant_id": &v1sync.Result{
			Code: resource.Success,
		},
	}}, nil
}

func (testReplicator) Persist(ctx context.Context, el *v1sync.EventList) []*resource.Result {
	return []*resource.Result{
		&resource.Result{
			EventID: "constant_id",
			Status:  resource.Success,
		},
	}
}

func TestServer_Sync(t *testing.T) {
	s := NewServer()

	// rbac enabled, should sync failed and return auth failed message
	config.SetConfig(config.Config{
		Sync: &config.Sync{
			RbacEnabled: true,
		}})
	events := &v1sync.EventList{Events: []*v1sync.Event{
		{
			Id: "evt1",
		},
		{
			Id: "evt2",
		},
	}}

	expectedRst := map[string]*v1sync.Result{
		"evt1": &v1sync.Result{
			Code:    resource.Fail,
			Message: rbac.NewError(rbac.ErrNoAuthHeader, "").Error(),
		},

		"evt2": &v1sync.Result{
			Code:    resource.Fail,
			Message: rbac.NewError(rbac.ErrNoAuthHeader, "").Error(),
		},
	}
	rst, err := s.Sync(context.Background(), events)
	assert.NoError(t, err)
	assert.True(t, reflect.DeepEqual(expectedRst, rst.Results))

	rst, err = s.Sync(context.Background(), nil) // nil input
	assert.NoError(t, err)
	assert.Equal(t, 0, len(rst.Results))

	// rbac disabled, should sync success(with the mock replicator)
	config.SetConfig(config.Config{
		Sync: &config.Sync{
			RbacEnabled: false,
		}})
	expectedRst = map[string]*v1sync.Result{
		"constant_id": &v1sync.Result{
			Code: resource.Success,
		},
	}
	s.replicator = &testReplicator{}
	rst, err = s.Sync(context.Background(), events)
	assert.NoError(t, err)
	assert.True(t, reflect.DeepEqual(expectedRst, rst.Results))
}
