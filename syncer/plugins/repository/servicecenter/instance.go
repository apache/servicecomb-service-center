package servicecenter

import (
	"context"

	scpb "github.com/apache/servicecomb-service-center/server/core/proto"
)

func (c *Client) RegisterInstance(ctx context.Context, domainProject, serviceId string, instance *scpb.MicroServiceInstance) (string, error) {
	instanceID, err := c.cli.RegisterInstance(ctx, domainProject, serviceId, instance)
	if err != nil {
		return "", err
	}
	return instanceID, nil
}

func (c *Client) UnregisterInstance(ctx context.Context, domainProject, serviceId, instanceId string) error {
	err := c.cli.UnregisterInstance(ctx, domainProject, serviceId, instanceId)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) DiscoveryInstances(ctx context.Context, domainProject, consumerId, providerAppId, providerServiceName, providerVersionRule string) ([]*scpb.MicroServiceInstance, error) {
	instances, err := c.cli.DiscoveryInstances(ctx, domainProject, consumerId, providerAppId, providerServiceName, providerVersionRule)
	if err != nil {
		return nil, err
	}
	return instances, nil
}

func (c *Client) Heartbeat(ctx context.Context, domainProject, serviceId, instanceId string) error {
	err := c.cli.Heartbeat(ctx, domainProject, serviceId, instanceId)
	if err != nil {
		return err
	}
	return nil
}
