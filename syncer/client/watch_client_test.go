package client

import (
	"testing"

	"github.com/apache/servicecomb-service-center/pkg/dump"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/stretchr/testify/assert"
)

func TestWatchInstance(t *testing.T) {

	addressWithPro := "http://127.0.0.1:8989"
	addressWithoutPro := "127.0.0.1:8888"
	cli := NewWatchClient(addressWithPro)
	assert.Equal(t, addressWithPro, cli.addr)

	cli = NewWatchClient(addressWithoutPro)
	assert.Equal(t, addressWithoutPro, cli.addr)

	err := cli.WatchInstances(fakeAddToQueue)
	assert.Error(t, err)

	cli.WatchInstanceHeartbeat(fakeAddToQueue)

}

func fakeAddToQueue(event *dump.WatchInstanceChangedEvent) {
	log.Debugf("success add instance event to queue:%s", event)
}
