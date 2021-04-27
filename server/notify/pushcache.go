package notify

import (
	pb "github.com/go-chassis/cari/discovery"
	"sort"
	"sync"
)

type InstancesResTemp struct {
	ins           *pb.WatchInstanceResponse
	domainProejct string
	rev           int64
}

type PushCache struct {
	m                     sync.Mutex
	serviceToInstancesRes map[string][]*InstancesResTemp
}

func NewPushCache() *PushCache {
	return &PushCache{
		m:                     sync.Mutex{},
		serviceToInstancesRes: make(map[string][]*InstancesResTemp),
	}
}

func (pc *PushCache) AddInstance(domainProject string, service string, rev int64, instance *pb.WatchInstanceResponse) {
	pc.m.Lock()
	defer pc.m.Unlock()
	instancesData, ok := pc.serviceToInstancesRes[service]
	if !ok {
		pc.serviceToInstancesRes[service] = []*InstancesResTemp{&InstancesResTemp{
			ins:           instance,
			rev:           rev,
			domainProejct: domainProject,
		}}
		return
	}
	pc.serviceToInstancesRes[service] = append(instancesData, &InstancesResTemp{
		ins:           instance,
		rev:           rev,
		domainProejct: domainProject,
	})
}

func (pc *PushCache) GetInstances(service string) ([]*pb.WatchInstanceResponse, int64, string) {
	pc.m.Lock()
	defer pc.m.Unlock()
	res, ok := pc.serviceToInstancesRes[service]
	if !ok {
		return []*pb.WatchInstanceResponse{}, -1, ""
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].rev < res[j].rev
	})
	watchRes := make([]*pb.WatchInstanceResponse, 0, len(res))
	for _, v := range res {
		watchRes = append(watchRes, v.ins)
	}
	delete(pc.serviceToInstancesRes, service)
	return watchRes, res[0].rev, res[0].domainProejct
}
