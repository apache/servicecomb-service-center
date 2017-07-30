package notification

import pb "github.com/ServiceComb/service-center/server/core/proto"

func NewInstanceWatcher(selfServiceId, instanceRoot string) *ListWatcher {
	return NewWatcher(INSTANCE, selfServiceId, instanceRoot)
}

func NewInstanceListWatcher(selfServiceId, instanceRoot string, listFunc func() (results []*pb.WatchInstanceResponse, rev int64)) *ListWatcher {
	return NewListWatcher(INSTANCE, selfServiceId, instanceRoot, listFunc)
}
