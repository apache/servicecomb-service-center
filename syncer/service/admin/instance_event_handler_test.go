package admin

import (
	"sync"
	"testing"
	"time"

	pb "github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"

	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/datasource/etcd/state/kvstore"
)

type mockEventHandler struct {
	processedEvts []kvstore.Event
}

func (h *mockEventHandler) Type() kvstore.Type {
	return sd.TypeInstance
}

func (h *mockEventHandler) OnEvent(evt kvstore.Event) {
	h.processedEvts = append(h.processedEvts, evt)
}

func TestInstanceEventHandler_OnEvent(t *testing.T) {
	h := NewInstanceEventHandler(&HealthChecker{
		shouldTrustPeerServer: false,
	}, 100*time.Millisecond, kvstore.EventHandler(nil))

	evt := kvstore.Event{
		Type: pb.EVT_CREATE,
		KV: &kvstore.KeyValue{
			Value: &pb.MicroServiceInstance{
				InstanceId: "test_instnace",
				ServiceId:  "test_service",
			},
		},
	}
	h.OnEvent(evt)
	assert.Equal(t, 1, len(h.events))

	// 信任对端SC，不处理
	h.events = make(map[string]kvstore.Event)
	h.healthChecker.shouldTrustPeerServer = true
	h.OnEvent(evt)
	assert.Equal(t, 0, len(h.events))

	h.events = make(map[string]kvstore.Event)
	// 忽略的事件，不处理
	evt.Type = pb.EVT_INIT
	h.healthChecker.shouldTrustPeerServer = false
	h.OnEvent(evt)
	assert.Equal(t, 0, len(h.events))

	// 处理
	evt.Type = pb.EVT_CREATE
	h.OnEvent(evt)
	assert.Equal(t, 1, len(h.events))
}

func Test_checkShouldIgnoreEvent(t *testing.T) {
	instanceDef := &pb.MicroServiceInstance{
		InstanceId: "test_instnace",
		ServiceId:  "test_service",
		Properties: map[string]string{
			"engineID":   "test_engineID_fake",
			"engineName": "test_engineName",
		},
	}

	evt := kvstore.Event{
		Type: pb.EVT_CREATE,
		KV: &kvstore.KeyValue{
			Value: instanceDef,
		},
	}
	instance, ignore := checkShouldIgnoreEvent(evt)
	assert.False(t, ignore)
	assert.Equal(t, instanceDef, instance)

	evt.Type = pb.EVT_INIT
	_, ignore = checkShouldIgnoreEvent(evt)
	assert.True(t, ignore)
	evt.Type = pb.EVT_DELETE
	_, ignore = checkShouldIgnoreEvent(evt)
	assert.True(t, ignore)
	evt.Type = pb.EVT_UPDATE
	_, ignore = checkShouldIgnoreEvent(evt)
	assert.False(t, ignore)

	evt.KV.Value = 1
	_, ignore = checkShouldIgnoreEvent(evt)
	assert.True(t, ignore)
	evt.KV.Value = instanceDef
	_, ignore = checkShouldIgnoreEvent(evt)
	assert.False(t, ignore)

	instanceDef.Properties["engineID"] = "test_engineID"
	_, ignore = checkShouldIgnoreEvent(evt)
	assert.True(t, ignore)
	instanceDef.Properties["engineID"] = "test_engineID_fake"
	_, ignore = checkShouldIgnoreEvent(evt)
	assert.False(t, ignore)
}

func TestInstanceEventHandler_run(t *testing.T) {
	mockHandler := &mockEventHandler{}
	h := NewInstanceEventHandler(&HealthChecker{
		shouldTrustPeerServer: false,
	}, 100*time.Millisecond, mockHandler)

	var wg sync.WaitGroup
	numGoroutines := 50
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.OnEvent(kvstore.Event{
				Type: pb.EVT_CREATE,
				KV: &kvstore.KeyValue{
					Value: &pb.MicroServiceInstance{
						InstanceId: "test_instance",
						ServiceId:  "test_service",
					},
				},
			})
			h.OnEvent(kvstore.Event{
				Type: pb.EVT_CREATE,
				KV: &kvstore.KeyValue{
					Value: &pb.MicroServiceInstance{
						InstanceId: "test_instance1",
						ServiceId:  "test_service1",
					},
				},
			})
		}()
	}
	wg.Wait()
	assert.Equal(t, 2, len(h.events)) // 一个服务多个事件仅保留一个
	assert.Equal(t, 0, len(mockHandler.processedEvts))

	time.Sleep(110 * time.Millisecond)
	assert.Equal(t, 2, len(h.events))
	assert.Equal(t, 0, len(mockHandler.processedEvts)) // 对端不可信，只保存，不处理

	h.healthChecker.shouldTrustPeerServer = true
	time.Sleep(110 * time.Millisecond)
	assert.Equal(t, 0, len(h.events))                                  // 对端可信，处理事件，事件处理完成清空
	assert.Equal(t, 2, len(mockHandler.processedEvts))                 // 处理事件
	assert.ElementsMatch(t, []string{"test_service", "test_service1"}, // 处理服务为设置的服务
		[]string{mockHandler.processedEvts[0].KV.Value.(*pb.MicroServiceInstance).ServiceId, mockHandler.processedEvts[1].KV.Value.(*pb.MicroServiceInstance).ServiceId})

	for _, evt := range mockHandler.processedEvts {
		assert.Equal(t, pb.EVT_UPDATE, evt.Type)
	}
}
