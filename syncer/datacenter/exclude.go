package datacenter

import (
	scpb "github.com/apache/servicecomb-service-center/server/core/proto"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
)

func (s *store) exclude(data *pb.SyncData) {
	mapping := s.cache.GetAllMapping()
	services := make([]*pb.SyncService, 0, 10)
	for _, svc := range data.Services {
		ins := s.excludeInstances(svc.Instances, mapping)
		if len(ins) > 0 {
			svc.Instances = ins
			services = append(services, svc)
		}
	}
	data.Services = services
}

func (s *store) excludeInstances(ins []*scpb.MicroServiceInstance, mapping pb.SyncMapping) []*scpb.MicroServiceInstance {
	is := make([]*scpb.MicroServiceInstance, 0, len(ins))
	for _, item := range ins {
		_, ok := mapping[item.InstanceId]
		if ok {
			if item.Status != "UP" {
				delete(mapping, item.InstanceId)
			}
			continue
		}
		if item.Status == "UP" {
			is = append(is, item)
		}
	}
	return is
}
