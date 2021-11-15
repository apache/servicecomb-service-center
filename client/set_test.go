package client_test

import (
	"context"
	"testing"

	v1sync "github.com/apache/servicecomb-service-center/api/sync/v1"
	"github.com/apache/servicecomb-service-center/client"
	"github.com/stretchr/testify/assert"
)

func TestNewSetForConfig(t *testing.T) {
	cs, err := client.NewSetForConfig(client.SetConfig{
		Addr: "127.0.0.1:30105",
	})
	assert.NoError(t, err)
	_, err = cs.EventServiceClient.Sync(context.TODO(), &v1sync.EventList{Events: []*v1sync.Event{
		{Action: "create"},
	}})
	assert.NoError(t, err)
}
