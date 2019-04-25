package repository

import (
	"context"

	scpb "github.com/apache/servicecomb-service-center/server/core/proto"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
)

type Adaptor interface {
	New(endpoints []string) (Repository, error)
}

type Repository interface {
	GetAll(ctx context.Context) (*pb.SyncData, error)
	CreateService(ctx context.Context, domainProject string, service *scpb.MicroService) (string, error)
	DeleteService(ctx context.Context, domainProject, serviceId string) error
	ServiceExistence(ctx context.Context, domainProject string, service *scpb.MicroService) (string, error)
	RegisterInstance(ctx context.Context, domainProject, serviceId string, instance *scpb.MicroServiceInstance) (string, error)
	UnregisterInstance(ctx context.Context, domainProject, serviceId, instanceId string) error
	DiscoveryInstances(ctx context.Context, domainProject, consumerId, providerAppId, providerServiceName, providerVersionRule string) ([]*scpb.MicroServiceInstance, error)
	Heartbeat(ctx context.Context, domainProject, serviceId, instanceId string) error
}
