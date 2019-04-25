package servicecenter

import (
	"context"

	scpb "github.com/apache/servicecomb-service-center/server/core/proto"
)

func (c *Client) CreateService(ctx context.Context, domainProject string, service *scpb.MicroService) (string, error) {
	serviceID, err := c.cli.CreateService(ctx, domainProject, service)
	if err != nil {
		return "", err
	}
	return serviceID, nil
}

func (c *Client) DeleteService(ctx context.Context, domainProject, serviceId string) error {
	err := c.cli.DeleteService(ctx, domainProject, serviceId)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) ServiceExistence(ctx context.Context, domainProject string, service *scpb.MicroService) (string, error) {
	serviceID, err := c.cli.ServiceExistence(ctx, domainProject, service.AppId, service.ServiceName, service.Version, service.Environment)
	if err != nil {
		return "", err
	}
	return serviceID, nil
}
