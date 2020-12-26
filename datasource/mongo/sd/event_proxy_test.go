package sd

import (
	"github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAddHandleFuncAndOnEvent(t *testing.T) {
	funcs := []MongoEventFunc{}
	kvEventProxy := KvEventProxy{
		evtHandleFuncs: funcs,
	}
	mongoEvent := MongoEvent {
		DocumentId: "",
		BusinessId: "",
		Type:     discovery.EVT_CREATE,
		Value: 1,
	}
	kvEventProxy.evtHandleFuncs = funcs
	assert.Equal(t, 0, len(kvEventProxy.evtHandleFuncs),
		"size of evtHandleFuncs is zero")
	t.Run("AddHandleFunc one time", func(t *testing.T) {
		kvEventProxy.AddHandleFunc(mongoEventFuncGet())
		kvEventProxy.OnEvent(mongoEvent)
		assert.Equal(t, 1, len(kvEventProxy.evtHandleFuncs))
	})
	t.Run("AddHandleFunc three times", func(t *testing.T) {
		for i := 0; i <5; i++ {
			kvEventProxy.AddHandleFunc(mongoEventFuncGet())
			kvEventProxy.OnEvent(mongoEvent)
		}
		assert.Equal(t, 5, len(kvEventProxy.evtHandleFuncs))
	})
}

func TestInjectConfig(t *testing.T) {
	funcs := []MongoEventFunc{}
	kvEventProxy := KvEventProxy{
		evtHandleFuncs: funcs,
	}
	t.Run("inject config without config OnEvent", func(t *testing.T) {
		cfg := Config{
			Key: "",
		}
		cfg1 := kvEventProxy.InjectConfig(&cfg)
		assert.NotNil(t, cfg1.OnEvent)
	})
	t.Run("inject config with config OnEvent", func(t *testing.T) {
		cfg := Config{
			Key: "",
			OnEvent: mongoEventFunc,
		}
		cfg1 := kvEventProxy.InjectConfig(&cfg)
		assert.NotNil(t, cfg1.OnEvent)
	})
}

func TestEventProxy(t *testing.T) {
	t.Run("when there is no such a proxy in eventProxies", func(t *testing.T) {
		eventProxies = make(map[string]*KvEventProxy)
		proxy := EventProxy("new")
		assert.Equal(t, 1, len(eventProxies), "add an new proxy")
		assert.NotNil(t,  eventProxies["new"], "proxy is not nil")
		assert.Nil(t, proxy.evtHandleFuncs)
	})

	t.Run("when there is no such a proxy in eventProxies", func(t *testing.T) {
		eventProxies = make(map[string]*KvEventProxy)
		mongoEventFunc := []MongoEventFunc{mongoEventFuncGet()}
		kvEventProxy := KvEventProxy{
			evtHandleFuncs: mongoEventFunc,
		}
		eventProxies["a"] = &kvEventProxy
		proxy := EventProxy("a")
		assert.Equal(t, 1, len(eventProxies), "find the proxy")
		assert.Equal(t, &kvEventProxy, eventProxies["a"])
		assert.NotNil(t,  eventProxies["a"], "proxy is not nil")
		assert.NotNil(t, proxy.evtHandleFuncs)
	})
}