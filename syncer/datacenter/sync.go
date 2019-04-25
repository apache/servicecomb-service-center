package datacenter

import (
	"context"
	"github.com/apache/servicecomb-service-center/pkg/log"

	scpb "github.com/apache/servicecomb-service-center/server/core/proto"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
)

func (s *store) sync(nodeData *pb.NodeDataInfo) {
	allInstances, mapping := s.syncServiceInstances(nodeData.DataInfo.Services, s.cache.GetSyncMapping(nodeData.NodeName))
	mapping = s.deleteInstances(allInstances, mapping)
	s.cache.SaveSyncMapping(nodeData.NodeName, mapping)
}

func (s *store) syncServiceInstances(services []*pb.SyncService, mapping pb.SyncMapping) ([]*scpb.MicroServiceInstance, pb.SyncMapping) {
	var err error
	ctx := context.Background()
	allInstances := make([]*scpb.MicroServiceInstance, 0, 10)
	for _, svc := range services {
		serviceID := ""
		for _, inst := range svc.Instances {
			allInstances = append(allInstances, inst)
			syncKey, ok := mapping[inst.InstanceId]
			if ok && syncKey.InstanceID != ""{
				err = s.repo.Heartbeat(ctx, svc.DomainProject, syncKey.ServiceID, syncKey.InstanceID)
				if err != nil {
					log.Errorf(err, "Syncer heartbeat instance failed")
				}
				continue
			}

			if serviceID == "" {
				if serviceID, _ = s.repo.ServiceExistence(ctx, svc.DomainProject, svc.Service); serviceID == "" {
					serviceID, err = s.repo.CreateService(ctx, svc.DomainProject, svc.Service)
					if err != nil {
						log.Errorf(err, "Syncer create service failed")
					}
				}
			}

			instanceID, err := s.repo.RegisterInstance(ctx, svc.DomainProject, serviceID, inst)
			if err != nil {
				log.Errorf(err, "Syncer create service failed")
				continue
			}

			mapping[inst.InstanceId] = &pb.SyncServiceKey{
				DomainProject: svc.DomainProject,
				ServiceID:     serviceID,
				InstanceID:    instanceID,
			}
		}
	}
	return allInstances, mapping
}

func (s *store) deleteInstances(ins []*scpb.MicroServiceInstance, mapping pb.SyncMapping) pb.SyncMapping {
	ctx := context.Background()
	nm := make(pb.SyncMapping)
	for key, val := range mapping {
		skip := false
		for _, instance := range ins {
			skip = instance.InstanceId == key
			if skip {
				break
			}
		}
		if skip {
			nm[key] = val
			continue
		}

		err := s.repo.UnregisterInstance(ctx, val.DomainProject, val.ServiceID, val.InstanceID)
		if err != nil {
			log.Errorf(err,"Syncer delete service failed")
		}
	}
	return nm
}
