package resource

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"

	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/go-chassis/cari/sync"
	"github.com/stretchr/testify/assert"
)

var (
	testInstanceID = "cae5f4b0381045a9adcfb3794cf6246b"
)

func testRegisterInstance(t *testing.T, manager metadataManager) {
	createTime := strconv.FormatInt(time.Now().Unix(), 10)
	input := &pb.RegisterInstanceRequest{
		Instance: &pb.MicroServiceInstance{
			InstanceId: testInstanceID,
			ServiceId:  testServiceID,
			Endpoints:  []string{"127.0.0.1:8080"},
			HostName:   "xx",
			Status:     "UP",
			Properties: map[string]string{
				"hello": "world",
			},
			Timestamp:    createTime,
			ModTimestamp: createTime,
			Version:      "0.0.1",
		},
	}
	value, _ := json.Marshal(input)
	id, _ := v1sync.NewEventID()
	e := &v1sync.Event{
		Id:        id,
		Action:    sync.CreateAction,
		Subject:   Instance,
		Value:     value,
		Timestamp: v1sync.Timestamp(),
	}
	a := &instance{
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
}

func TestOperateInstance(t *testing.T) {
	manager := &mockMetadata{
		services:  make(map[string]*pb.MicroService),
		instances: make(map[string]*pb.MicroServiceInstance),
	}
	ctx := context.TODO()
	testCreateMicroservice(t, manager)
	testRegisterInstance(t, manager)
	data, _ := manager.GetInstance(ctx, &pb.GetOneInstanceRequest{
		ProviderServiceId:  testServiceID,
		ProviderInstanceId: testInstanceID,
	})

	updateInput := *data.Instance
	updateInput.Properties = map[string]string{
		"demo": "update",
	}

	value, _ := json.Marshal(&updateInput)
	id, _ := v1sync.NewEventID()
	e1 := &v1sync.Event{
		Id:        id,
		Action:    sync.UpdateAction,
		Subject:   Instance,
		Opts:      nil,
		Value:     value,
		Timestamp: v1sync.Timestamp(),
	}
	a1 := &instance{
		event:   e1,
		manager: manager,
	}
	result := a1.LoadCurrentResource(ctx)
	if assert.Nil(t, result) {
		result = a1.NeedOperate(ctx)
		if assert.Nil(t, result) {
			result = a1.Operate(ctx)
			if assert.NotNil(t, result) && assert.Equal(t, Success, result.Status) {
				data, err := a1.manager.GetInstance(ctx, &pb.GetOneInstanceRequest{
					ProviderServiceId:  testServiceID,
					ProviderInstanceId: testInstanceID,
				})
				assert.Nil(t, err)
				assert.NotNil(t, data)
				assert.Equal(t, map[string]string{
					"demo": "update",
				}, data.Instance.Properties)
			}
		}
	}

	inputDelete := &pb.UnregisterInstanceRequest{
		ServiceId:  testServiceID,
		InstanceId: testInstanceID,
	}
	value, _ = json.Marshal(inputDelete)
	id, _ = v1sync.NewEventID()
	e2 := &v1sync.Event{
		Id:        id,
		Action:    sync.DeleteAction,
		Subject:   Instance,
		Opts:      nil,
		Value:     value,
		Timestamp: v1sync.Timestamp(),
	}
	a2 := &instance{
		event:   e2,
		manager: manager,
	}
	result = a2.LoadCurrentResource(ctx)
	if assert.Nil(t, result) {
		result = a2.NeedOperate(ctx)
		if assert.Nil(t, result) {
			result = a2.Operate(ctx)
			if assert.NotNil(t, result) && assert.Equal(t, Success, result.Status) {
				_, err := a1.manager.GetInstance(ctx, &pb.GetOneInstanceRequest{
					ProviderServiceId:  testServiceID,
					ProviderInstanceId: testInstanceID,
				})
				assert.NotNil(t, err)
				assert.True(t, errsvc.IsErrEqualCode(err, pb.ErrInstanceNotExists))
			}
		}
	}
}

func TestNewInstance(t *testing.T) {
	i := NewInstance(nil)
	assert.NotNil(t, i)
}
