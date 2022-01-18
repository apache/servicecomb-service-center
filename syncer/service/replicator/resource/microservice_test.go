package resource

import (
	"context"
	"encoding/json"
	"fmt"
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
	testServiceID = "b03f3c3624ea5c4baeca66a39fbb3bc49973bfc4"
)

type mockMetadata struct {
	services  map[string]*pb.MicroService
	instances map[string]*pb.MicroServiceInstance
}

func (f mockMetadata) RegisterService(_ context.Context, request *pb.CreateServiceRequest) (*pb.CreateServiceResponse, error) {
	serviceID := request.Service.ServiceId
	_, ok := f.services[serviceID]
	if ok {
		return nil, pb.NewError(pb.ErrServiceAlreadyExists,
			"ServiceID conflict or found the same service with different id.")
	}
	f.services[request.Service.ServiceId] = request.Service
	return &pb.CreateServiceResponse{
		ServiceId: serviceID,
	}, nil
}

func (f mockMetadata) GetService(_ context.Context, in *pb.GetServiceRequest) (*pb.MicroService, error) {
	serviceID := in.ServiceId
	result, ok := f.services[serviceID]
	if !ok {
		return nil, pb.NewError(pb.ErrServiceNotExists, "Service does not exist.")
	}
	return result, nil
}

func (f mockMetadata) PutServiceProperties(_ context.Context, in *pb.UpdateServicePropsRequest) error {
	serviceID := in.ServiceId
	result, ok := f.services[serviceID]
	if !ok {
		return pb.NewError(pb.ErrServiceNotExists, "Service does not exist.")
	}

	result.Properties = in.Properties
	return nil
}

func (f mockMetadata) UnregisterService(_ context.Context, in *pb.DeleteServiceRequest) error {
	serviceID := in.ServiceId
	_, ok := f.services[serviceID]
	if !ok {
		return pb.NewError(pb.ErrServiceNotExists, "Service does not exist.")
	}

	delete(f.services, serviceID)
	return nil
}

func (f mockMetadata) RegisterInstance(_ context.Context, in *pb.RegisterInstanceRequest) (*pb.RegisterInstanceResponse, error) {
	serviceID := in.Instance.ServiceId
	_, ok := f.services[serviceID]
	if !ok {
		return nil, pb.NewError(pb.ErrServiceNotExists, "Service does not exist.")
	}

	instanceID := in.Instance.InstanceId
	f.instances[instanceID] = in.Instance
	return &pb.RegisterInstanceResponse{
		InstanceId: instanceID,
	}, nil
}

func (f mockMetadata) SendHeartbeat(_ context.Context, in *pb.HeartbeatRequest) error {
	instanceID := in.InstanceId

	_, ok := f.instances[instanceID]
	if !ok {
		return pb.NewError(pb.ErrInstanceNotExists, "instance does not exist.")
	}
	return nil
}

func (f mockMetadata) GetInstance(_ context.Context, in *pb.GetOneInstanceRequest) (*pb.GetOneInstanceResponse, error) {
	serviceID := in.ProviderServiceId
	_, ok := f.services[serviceID]
	if !ok {
		return nil, pb.NewError(pb.ErrServiceNotExists,
			fmt.Sprintf("Consumer[%s] does not exist.", serviceID))
	}
	instanceID := in.ProviderInstanceId
	result, ok := f.instances[instanceID]
	if !ok {
		return nil, pb.NewError(pb.ErrInstanceNotExists, "instance not exist")
	}
	return &pb.GetOneInstanceResponse{
		Instance: result,
	}, nil
}

func (f mockMetadata) PutInstance(_ context.Context, in *pb.RegisterInstanceRequest) error {
	instanceID := in.Instance.InstanceId
	_, ok := f.instances[instanceID]
	if !ok {
		return pb.NewError(pb.ErrInstanceNotExists, "Service instance does not exist.")
	}

	f.instances[instanceID] = in.Instance
	return nil
}

func (f mockMetadata) UnregisterInstance(_ context.Context, in *pb.UnregisterInstanceRequest) error {
	instanceID := in.InstanceId
	_, ok := f.instances[instanceID]
	if !ok {
		return pb.NewError(pb.ErrInstanceNotExists, "Instance's leaseId not exist.")
	}
	delete(f.instances, instanceID)
	return nil
}

func testCreateMicroservice(t *testing.T, manager serviceManager) {
	createTime := strconv.FormatInt(time.Now().Unix(), 10)
	service := &pb.MicroService{
		ServiceId:   testServiceID,
		AppId:       "demo-java-chassis-cse-v2",
		ServiceName: "provider",
		Version:     "0.0.1",
		Description: "",
		Level:       "FRONT",
		Status:      "UP",
		Properties: map[string]string{
			"hello": "world",
		},
		Timestamp:    createTime,
		ModTimestamp: createTime,
	}
	input := &pb.CreateServiceRequest{
		Service: service,
	}
	value, _ := json.Marshal(input)
	id, _ := v1sync.NewEventID()
	e := &v1sync.Event{
		Id:        id,
		Action:    sync.CreateAction,
		Subject:   Microservice,
		Value:     value,
		Timestamp: v1sync.Timestamp(),
	}
	a := &microservice{
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
				data, err := a.manager.GetService(ctx, &pb.GetServiceRequest{
					ServiceId: testServiceID,
				})
				assert.Nil(t, err)
				assert.NotNil(t, data)
			}
		}
	}
	return
}

func TestOperateMicroService(t *testing.T) {
	manager := &mockMetadata{
		services: make(map[string]*pb.MicroService),
	}
	testCreateMicroservice(t, manager)
	ctx := context.TODO()

	updateInput := &pb.UpdateServicePropsRequest{
		ServiceId: testServiceID,
		Properties: map[string]string{
			"demo": "update",
		},
	}
	value, _ := json.Marshal(updateInput)

	id, _ := v1sync.NewEventID()
	e1 := &v1sync.Event{
		Id:        id,
		Action:    sync.UpdateAction,
		Subject:   Microservice,
		Opts:      nil,
		Value:     value,
		Timestamp: v1sync.Timestamp(),
	}
	a1 := &microservice{
		event:   e1,
		manager: manager,
	}
	result := a1.LoadCurrentResource(ctx)
	if assert.Nil(t, result) {
		result = a1.NeedOperate(ctx)
		if assert.Nil(t, result) {
			result = a1.Operate(ctx)
			if assert.NotNil(t, result) && assert.Equal(t, Success, result.Status) {
				data, err := a1.manager.GetService(ctx, &pb.GetServiceRequest{
					ServiceId: testServiceID,
				})
				assert.Nil(t, err)
				assert.NotNil(t, data)
				assert.Equal(t, map[string]string{
					"demo": "update",
				}, data.Properties)
			}
		}
	}

	id, _ = v1sync.NewEventID()
	e2 := &v1sync.Event{
		Id:        id,
		Action:    sync.DeleteAction,
		Subject:   Microservice,
		Opts:      nil,
		Value:     value,
		Timestamp: v1sync.Timestamp(),
	}
	a2 := &microservice{
		event:   e2,
		manager: manager,
	}
	result = a2.LoadCurrentResource(ctx)
	if assert.Nil(t, result) {
		result = a2.NeedOperate(ctx)
		if assert.Nil(t, result) {
			result = a2.Operate(ctx)
			if assert.NotNil(t, result) && assert.Equal(t, Success, result.Status) {
				_, err := a1.manager.GetService(ctx, &pb.GetServiceRequest{
					ServiceId: testServiceID,
				})
				assert.NotNil(t, err)
				assert.True(t, errsvc.IsErrEqualCode(err, pb.ErrServiceNotExists))
			}
		}
	}
}

func TestNewMicroservice(t *testing.T) {
	m := NewMicroservice(nil)
	assert.NotNil(t, m)
}
