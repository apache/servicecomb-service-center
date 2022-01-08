package sync

import (
	"context"
	"fmt"

	v1sync "github.com/apache/servicecomb-service-center/api/sync/v1"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/config"
)

const (
	HealthStatusConnected = "CONNECTED"
	HealthStatusAbnormal  = "ABNORMAL"
	HealthStatusClose     = "CLOSE"
)

type Server struct {
	v1sync.UnimplementedEventServiceServer
}

func (s *Server) Sync(ctx context.Context, events *v1sync.EventList) (*v1sync.Results, error) {
	log.Info(fmt.Sprintf("Received: %v", events.Events[0].Action))
	return &v1sync.Results{}, nil
}

func (s *Server) Health(ctx context.Context, request *v1sync.HealthRequest) (*v1sync.HealthReply, error) {
	syncerEnabled := config.GetBool("sync.enableOnStart", false)
	if !syncerEnabled {
		return &v1sync.HealthReply{Status: HealthStatusClose}, nil
	}
	return &v1sync.HealthReply{Status: HealthStatusConnected}, nil
}
