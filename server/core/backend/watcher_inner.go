/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package backend

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"golang.org/x/net/context"
	"sync"
)

type innerWatcher struct {
	Cfg    ListWatchConfig
	lw     ListWatch
	bus    chan *registry.PluginResponse
	stopCh chan struct{}
	stop   bool
	mux    sync.Mutex
}

func (w *innerWatcher) EventBus() <-chan *registry.PluginResponse {
	return w.bus
}

func (w *innerWatcher) process(_ context.Context) {
	stopCh := make(chan struct{})
	ctx, cancel := context.WithTimeout(w.Cfg.Context, w.Cfg.Timeout)
	util.Go(func(_ context.Context) {
		defer close(stopCh)
		w.lw.DoWatch(ctx, w.sendEvent)
	})

	select {
	case <-stopCh:
		// timed out or exception
		w.Stop()
	case <-w.stopCh:
		cancel()
	}
}

func (w *innerWatcher) sendEvent(resp *registry.PluginResponse) {
	defer util.RecoverAndReport()
	w.bus <- resp
}

func (w *innerWatcher) Stop() {
	w.mux.Lock()
	if w.stop {
		w.mux.Unlock()
		return
	}
	w.stop = true
	close(w.stopCh)
	close(w.bus)
	w.mux.Unlock()
}

func newInnerWatcher(lw ListWatch, cfg ListWatchConfig) *innerWatcher {
	w := &innerWatcher{
		Cfg:    cfg,
		lw:     lw,
		bus:    make(chan *registry.PluginResponse, EVENT_BUS_MAX_SIZE),
		stopCh: make(chan struct{}),
	}
	util.Go(w.process)
	return w
}
