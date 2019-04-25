package datacenter

import (
	"github.com/apache/servicecomb-service-center/pkg/log"

	"github.com/apache/servicecomb-service-center/syncer/notify"
	"github.com/apache/servicecomb-service-center/syncer/pkg/events"
	"github.com/apache/servicecomb-service-center/syncer/plugins"
	"github.com/apache/servicecomb-service-center/syncer/plugins/repository"
	"github.com/apache/servicecomb-service-center/syncer/plugins/storage"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
)

type Store interface {
	OnEvent(event events.ContextEvent)
	LocalInfo() *pb.SyncData
	Stop()
}

type store struct {
	repo  repository.Repository
	cache storage.Repository
}

func NewStore(endpoints []string) (Store, error) {
	repo, err := plugins.Plugins().Repository().New(endpoints)
	if err != nil {
		return nil, err
	}

	return &store{
		repo:  repo,
		cache: plugins.Plugins().Storage(),
	}, nil
}

func (s *store) Stop() {
	if s.cache == nil{
		return
	}
	s.cache.Stop()
}

func (s *store) LocalInfo() *pb.SyncData {
	return s.cache.GetSyncData()
}

func (s *store) OnEvent(event events.ContextEvent) {
	switch event.Type() {
	case notify.EventTicker:
		s.getLocalDataInfo(event)
	case notify.EventPullByPeer:
		s.syncPeerDataInfo(event)
	default:
	}
}

func (s *store) getLocalDataInfo(event events.ContextEvent) {
	ctx := event.Context()
	data, err := s.repo.GetAll(ctx)
	if err != nil {
		log.Errorf(err, "Syncer discover instances failed")
		return
	}
	s.exclude(data)
	s.cache.SaveSyncData(data)

	events.Dispatch(events.NewContextEvent(notify.EventDiscovery, nil))
}

func (s *store) syncPeerDataInfo(event events.ContextEvent) {
	ctx := event.Context()
	nodeData, ok := ctx.Value(event.Type()).(*pb.NodeDataInfo)
	if !ok {
		log.Error("save peer info failed", nil)
		return
	}
	s.sync(nodeData)
}
