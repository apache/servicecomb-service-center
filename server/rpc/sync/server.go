package sync

import (
	"context"
	"fmt"

	v1sync "github.com/apache/servicecomb-service-center/api/sync/v1"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

type Server struct {
	v1sync.UnimplementedEventServiceServer
}

func (s *Server) Sync(ctx context.Context, events *v1sync.EventList) (*v1sync.Results, error) {
	log.Info(fmt.Sprintf("Received: %v", events.Events[0].Action))
	return &v1sync.Results{}, nil
}
