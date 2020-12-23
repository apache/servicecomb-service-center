package sd

import (
	"sync"
	"testing"

	"github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
)

func TestAddHandleFuncAndOnEvent(t *testing.T) {
	funcs := []MongoEventFunc{}
	mongoEventProxy := MongoEventProxy{
		evtHandleFuncs: funcs,
	}
	mongoEvent := MongoEvent{
		DocumentID: "",
		BusinessID: "",
		Type:       discovery.EVT_CREATE,
		Value:      1,
	}
	mongoEventProxy.evtHandleFuncs = funcs
	assert.Equal(t, 0, len(mongoEventProxy.evtHandleFuncs),
		"size of evtHandleFuncs is zero")
	t.Run("AddHandleFunc one time", func(t *testing.T) {
		mongoEventProxy.AddHandleFunc(mongoEventFuncGet())
		mongoEventProxy.OnEvent(mongoEvent)
		assert.Equal(t, 1, len(mongoEventProxy.evtHandleFuncs))
	})
	t.Run("AddHandleFunc three times", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			mongoEventProxy.AddHandleFunc(mongoEventFuncGet())
			mongoEventProxy.OnEvent(mongoEvent)
		}
		assert.Equal(t, 6, len(mongoEventProxy.evtHandleFuncs))
	})
}

type mockInstanceEventHandler struct {
}

func (h *mockInstanceEventHandler) Type() string {
	return instance
}

func (h *mockInstanceEventHandler) OnEvent(MongoEvent) {

}

func TestAddEventHandler(t *testing.T) {
	AddEventHandler(&mockInstanceEventHandler{})

}

func TestInjectConfig(t *testing.T) {
	funcs := []MongoEventFunc{}
	mongoEventProxy := MongoEventProxy{
		evtHandleFuncs: funcs,
	}
	t.Run("inject config without config OnEvent", func(t *testing.T) {
		options := Options{
			Key: "",
		}
		injectedOptions := mongoEventProxy.InjectOptions(&options)
		assert.NotNil(t, injectedOptions.OnEvent)
	})
	t.Run("inject config with config OnEvent", func(t *testing.T) {
		options := Options{
			Key:     "",
			OnEvent: mongoEventFunc,
		}
		injectedOptions := mongoEventProxy.InjectOptions(&options)
		assert.NotNil(t, injectedOptions.OnEvent)
	})
}

func TestEventProxy(t *testing.T) {
	t.Run("when there is no such a proxy in eventProxies", func(t *testing.T) {
		eventProxies = &sync.Map{}
		proxy := EventProxy("new")
		p, ok := eventProxies.Load("new")
		//assert.Equal(t, 1, , "add an new proxy")
		assert.Equal(t, true, ok)
		assert.NotNil(t, p, "proxy is not nil")
		assert.Nil(t, proxy.evtHandleFuncs)
	})

	t.Run("when there is no such a proxy in eventProxies", func(t *testing.T) {
		eventProxies = &sync.Map{}
		mongoEventFunc := []MongoEventFunc{mongoEventFuncGet()}
		mongoEventProxy := MongoEventProxy{
			evtHandleFuncs: mongoEventFunc,
		}
		eventProxies.Store("a", &mongoEventProxy)
		proxy := EventProxy("a")
		//assert.Equal(t, 1, len(eventProxies), "find the proxy")
		p, ok := eventProxies.Load("a")
		assert.Equal(t, true, ok)
		assert.Equal(t, &mongoEventProxy, p)
		assert.NotNil(t, p, "proxy is not nil")
		assert.NotNil(t, proxy.evtHandleFuncs)
	})
}
