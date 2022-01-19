package replicator

import (
	"context"
	"fmt"

	"github.com/apache/servicecomb-service-center/client"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rpc"
	"github.com/apache/servicecomb-service-center/pkg/util"
	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"
	"github.com/apache/servicecomb-service-center/syncer/config"
	"github.com/apache/servicecomb-service-center/syncer/service/replicator/resource"

	"github.com/go-chassis/foundation/gopool"
	"google.golang.org/grpc"
)

var (
	manager = NewManager(make(map[string]struct{}, 1000))
)

var (
	conn *grpc.ClientConn
)

func Work() error {
	err := InitSyncClient()
	if err != nil {
		return err
	}

	gopool.Go(func(ctx context.Context) {
		<-ctx.Done()

		Close()
	})

	resource.InitManager()
	return err
}

func InitSyncClient() error {
	cfg := config.GetConfig()
	peer := cfg.Sync.Peers[0]
	log.Info(fmt.Sprintf("peer is %v", peer))
	var err error
	conn, err = rpc.GetRoundRobinLbConn(&rpc.Config{
		Addrs:       peer.Endpoints,
		Scheme:      "grpc",
		ServiceName: "syncer",
	})
	return err
}

func Close() {
	if conn == nil {
		return
	}

	err := conn.Close()
	if err != nil {
		log.Error("close conn failed", err)
	}
}

func Manager() Replicator {
	return manager
}

// Replicator define replicator manager, receive events from event manager
// and send events to remote syncer
type Replicator interface {
	Replicate(ctx context.Context, el *v1sync.EventList) (*v1sync.Results, error)
	Persist(ctx context.Context, el *v1sync.EventList) []*resource.Result
}

func NewManager(cache map[string]struct{}) Replicator {
	return &replicatorManager{
		cache: cache,
	}
}

type replicatorManager struct {
	cache map[string]struct{}
}

func (r *replicatorManager) Replicate(ctx context.Context, el *v1sync.EventList) (*v1sync.Results, error) {
	return r.replicate(ctx, el)
}

func (r *replicatorManager) replicate(ctx context.Context, el *v1sync.EventList) (*v1sync.Results, error) {
	log.Info(fmt.Sprintf("start replicate events %d", len(el.Events)))

	set := client.NewSet(conn)
	result, err := set.EventServiceClient.Sync(ctx, el)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *replicatorManager) Persist(ctx context.Context, el *v1sync.EventList) []*resource.Result {
	if el == nil || len(el.Events) == 0 {
		return []*resource.Result{}
	}

	results := make([]*resource.Result, 0, len(el.Events))
	for _, event := range el.Events {
		log.Info(fmt.Sprintf("start handle event %s", event.Flag()))

		r, result := resource.New(event)
		if result != nil {
			results = append(results, result.WithEventID(event.Id))
			continue
		}

		ctx = util.SetDomain(ctx, event.Opts[string(util.CtxDomain)])
		ctx = util.SetProject(ctx, event.Opts[string(util.CtxProject)])

		result = r.LoadCurrentResource(ctx)
		if result != nil {
			results = append(results, result.WithEventID(event.Id))
			continue
		}

		result = r.NeedOperate(ctx)
		if result != nil {
			results = append(results, result.WithEventID(event.Id))
			continue
		}

		result = r.Operate(ctx)
		results = append(results, result.WithEventID(event.Id))

		log.Info(fmt.Sprintf("operate resource %s, %s", event.Flag(), result.Flag()))
	}

	for _, result := range results {
		log.Info(fmt.Sprintf("handle event result %s", result.Flag()))
	}

	return results
}
