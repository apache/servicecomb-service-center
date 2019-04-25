package servicecenter

import (
	"strings"

	"github.com/apache/servicecomb-service-center/server/admin/model"
	scpb "github.com/apache/servicecomb-service-center/server/core/proto"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
)

func transform(cache *model.Cache) (data *pb.SyncData) {
	data = &pb.SyncData{Services: make([]*pb.SyncService, 0, 10)}
	for _, svc := range cache.Microservices {
		instances := instancesFromService(svc.Value, cache.Instances)
		data.Services = append(data.Services, &pb.SyncService{
			DomainProject: strings.Join(strings.Split(svc.Key, "/")[4:6], "/"),
			Service:       svc.Value,
			Instances:     instances,
		})
	}
	return
}

func instancesFromService(service *scpb.MicroService, instances []*model.Instance) []*scpb.MicroServiceInstance {
	instList := make([]*scpb.MicroServiceInstance, 0, 10)
	for _, inst := range instances {
		if inst.Value.ServiceId == service.ServiceId {
			instList = append(instList, inst.Value)
		}
	}
	return instList
}
