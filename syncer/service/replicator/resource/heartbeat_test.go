package resource

import (
	"context"
	"encoding/json"
	"testing"

	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"

	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/sync"
	"github.com/stretchr/testify/assert"
)

func TestHeartbeat(t *testing.T) {
	manager := &mockMetadata{
		services:  make(map[string]*pb.MicroService),
		instances: make(map[string]*pb.MicroServiceInstance),
	}

	testCreateMicroservice(t, manager)
	testRegisterInstance(t, manager)

	input := &pb.HeartbeatRequest{
		ServiceId:  testServiceID,
		InstanceId: testInstanceID,
	}
	value, _ := json.Marshal(input)
	id, _ := v1sync.NewEventID()
	e := &v1sync.Event{
		Id:        id,
		Action:    sync.UpdateAction,
		Subject:   Heartbeat,
		Opts:      nil,
		Value:     value,
		Timestamp: v1sync.Timestamp(),
	}
	a := &heartbeat{
		event: e,
	}
	a.manager = manager
	ctx := context.Background()
	result := a.LoadCurrentResource(ctx)
	if assert.Nil(t, result) {
		result = a.NeedOperate(ctx)
		if assert.Nil(t, result) {
			result = a.Operate(ctx)
			if assert.NotNil(t, result) && assert.Equal(t, Success, result.Status) {
				data, err := a.manager.GetInstance(ctx, &pb.GetOneInstanceRequest{
					ProviderServiceId:  testServiceID,
					ProviderInstanceId: testInstanceID,
				})
				assert.Nil(t, err)
				assert.NotNil(t, data)
			}
		}
	}

	syncEvent, err := a.FailHandle(ctx, InstNonExist)
	if assert.Nil(t, err) {
		assert.NotNil(t, syncEvent)
		assert.NotNil(t, syncEvent.Value)
		assert.Equal(t, Instance, syncEvent.Subject)
	}

	_, err = a.FailHandle(ctx, Success)
	assert.Nil(t, err)
}

func TestFailHandlerHeartbeat(t *testing.T) {
	h := new(heartbeat)
	assert.True(t, h.CanDrop())
	_, err := h.FailHandle(context.TODO(), NonImplement)
	assert.Nil(t, err)
}
