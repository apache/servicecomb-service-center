package client

import (
	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"
	"google.golang.org/grpc"
)

// Set is set of grpc clients
type Set struct {
	EventServiceClient v1sync.EventServiceClient
}

func NewSet(conn *grpc.ClientConn) *Set {
	return &Set{
		EventServiceClient: v1sync.NewEventServiceClient(conn),
	}
}
