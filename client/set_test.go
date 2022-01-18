package client_test

import (
	"context"
	"testing"

	"github.com/apache/servicecomb-service-center/client"
	"github.com/apache/servicecomb-service-center/pkg/rpc"
	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"
	"github.com/stretchr/testify/assert"
)

func TestNewSetForConfig(t *testing.T) {
	conn, err := rpc.GetPickFirstLbConn(
		&rpc.Config{
			Addrs:       []string{"127.0.0.1:30105"},
			Scheme:      "test",
			ServiceName: "serviceName",
		})
	assert.NoError(t, err)
	defer conn.Close()
	set := client.NewSet(conn)
	_, err = set.EventServiceClient.Sync(context.TODO(), &v1sync.EventList{Events: []*v1sync.Event{
		{Action: "create"},
	}})
	assert.NoError(t, err)
}
