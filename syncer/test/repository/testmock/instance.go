package testmock

import (
	"context"

	scpb "github.com/apache/servicecomb-service-center/server/core/proto"
)

var (
	registerInstance func(ctx context.Context, domainProject, serviceId string, instance *scpb.MicroServiceInstance) (string, error)
	unregisterInstance func(ctx context.Context, domainProject, serviceId, instanceId string) error
	discoveryInstances func(ctx context.Context, domainProject, consumerId, providerAppId, providerServiceName, providerVersionRule string) ([]*scpb.MicroServiceInstance, error)
	heartbeat func(ctx context.Context, domainProject, serviceId, instanceId string) error
)

func SetRegisterInstance(handler func(ctx context.Context, domainProject, serviceId string, instance *scpb.MicroServiceInstance) (string, error))  {
	registerInstance = handler
}

func SetUnregisterInstance(handler func(ctx context.Context, domainProject, serviceId, instanceId string) error)  {
	unregisterInstance = handler
}

func SetDiscoveryInstances(handler func(ctx context.Context, domainProject, consumerId, providerAppId, providerServiceName, providerVersionRule string) ([]*scpb.MicroServiceInstance, error))  {
	discoveryInstances = handler
}

func SetHeartbeat(handler func(ctx context.Context, domainProject, serviceId, instanceId string) error)  {
	heartbeat = handler
}

func (c *Client) RegisterInstance(ctx context.Context, domainProject, serviceId string, instance *scpb.MicroServiceInstance) (string, error) {
	if registerInstance != nil{
		return registerInstance(ctx, domainProject, serviceId, instance)
	}
	return "4d41a637471f11e9888cfa163eca30e0", nil
}

func (c *Client) UnregisterInstance(ctx context.Context, domainProject, serviceId, instanceId string) error {
	if unregisterInstance != nil{
		return unregisterInstance(ctx, domainProject, serviceId, instanceId)
	}
	return nil
}

func (c *Client) DiscoveryInstances(ctx context.Context, domainProject, consumerId, providerAppId, providerServiceName, providerVersionRule string) ([]*scpb.MicroServiceInstance, error) {
	if discoveryInstances != nil{
		return discoveryInstances(ctx, domainProject, consumerId, providerAppId, providerServiceName, providerVersionRule)
	}
	return []*scpb.MicroServiceInstance{
		{
			InstanceId: "4d41a637471f11e9888cfa163eca30e0",
			ServiceId:  "5db1b794aa6f8a875d6e68110260b5491ee7e223",
			Endpoints: []string{
				"rest://127.0.0.1:30100/",
			},
			HostName: "testmock",
			Status:   "UP",
			HealthCheck: &scpb.HealthCheck{
				Mode:     "push",
				Interval: 30,
				Times:    3,
			},
			Timestamp:    "1552653537",
			ModTimestamp: "1552653537",
			Version:      "1.1.0",
		},
	}, nil
}

func (c *Client) Heartbeat(ctx context.Context, domainProject, serviceId, instanceId string) error {
	if heartbeat != nil{
		return heartbeat(ctx, domainProject, serviceId, instanceId)
	}
	return nil
}
