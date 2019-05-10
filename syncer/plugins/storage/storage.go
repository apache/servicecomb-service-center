package storage

import pb "github.com/apache/servicecomb-service-center/syncer/proto"

// Repository Storage interface
type Repository interface {
	Stop()
	SaveSyncData(data *pb.SyncData)
	GetSyncData() (data *pb.SyncData)
	SaveSyncMapping(nodeName string, mapping pb.SyncMapping)
	GetSyncMapping(nodeName string) (mapping pb.SyncMapping)
	GetAllMapping() (mapping pb.SyncMapping)
}
