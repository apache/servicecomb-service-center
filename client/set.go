package client

import (
	"fmt"

	v1sync "github.com/apache/servicecomb-service-center/api/sync/v1"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"google.golang.org/grpc"
)

// SetConfig is client configs
type SetConfig struct {
	Addr string
}

// Set is set of grpc clients
type Set struct {
	EventServiceClient v1sync.EventServiceClient
}

// NewSetForConfig dial grpc connection and create all grpc clients
func NewSetForConfig(c SetConfig) (*Set, error) {
	conn, err := grpc.Dial(c.Addr, grpc.WithInsecure())
	if err != nil {
		log.Error(fmt.Sprintf("can not connect: %s", err), nil)
		return nil, err
	}
	return &Set{
		EventServiceClient: v1sync.NewEventServiceClient(conn),
	}, nil
}
