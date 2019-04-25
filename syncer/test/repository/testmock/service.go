package testmock

import (
	"context"

	scpb "github.com/apache/servicecomb-service-center/server/core/proto"
)

var (
	createServiceHandler func(ctx context.Context, domainProject string, service *scpb.MicroService) (string, error)
	deleteService func(ctx context.Context, domainProject, serviceId string) error
	serviceExistence func(ctx context.Context, domainProject string, service *scpb.MicroService) (string, error)
)

func SetCreateService(handler func(ctx context.Context, domainProject string, service *scpb.MicroService) (string, error))  {
	createServiceHandler = handler
}

func SetDeleteService(handler func(ctx context.Context, domainProject, serviceId string) error)  {
	deleteService = handler
}

func SetServiceExistence(handler func(ctx context.Context, domainProject string, service *scpb.MicroService) (string, error))  {
	serviceExistence = handler
}

func (c *Client) CreateService(ctx context.Context, domainProject string, service *scpb.MicroService) (string, error) {
	if createServiceHandler != nil{
		return createServiceHandler(ctx, domainProject, service)
	}
	return "5db1b794aa6f8a875d6e68110260b5491ee7e223", nil
}

func (c *Client) DeleteService(ctx context.Context, domainProject, serviceId string) error {
	if createServiceHandler != nil{
		return deleteService(ctx, domainProject, serviceId)
	}
	return nil
}

func (c *Client) ServiceExistence(ctx context.Context, domainProject string, service *scpb.MicroService) (string, error) {
	if serviceExistence != nil{
		return serviceExistence(ctx, domainProject, service)
	}
	return "5db1b794aa6f8a875d6e68110260b5491ee7e223", nil
}
