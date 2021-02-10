/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package sd

import (
	"sync"
	"testing"

	"github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
)

func TestAddHandleFuncAndOnEvent(t *testing.T) {
	var funcs []MongoEventFunc
	mongoEventProxy := MongoEventProxy{
		evtHandleFuncs: funcs,
	}
	mongoEvent := MongoEvent{
		DocumentID: "",
		ResourceID: "",
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

func TestEventProxy(t *testing.T) {
	t.Run("when there is no such a proxy in eventProxies", func(t *testing.T) {
		eventProxies = &sync.Map{}
		proxy := EventProxy("new")
		p, ok := eventProxies.Load("new")

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

		p, ok := eventProxies.Load("a")
		assert.Equal(t, true, ok)
		assert.Equal(t, &mongoEventProxy, p)
		assert.NotNil(t, p, "proxy is not nil")
		assert.NotNil(t, proxy.evtHandleFuncs)
	})
}
