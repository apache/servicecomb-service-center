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

package sdcommon

import (
	"context"
	"sync"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/go-chassis/foundation/gopool"
)

type EventBus struct {
	Cfg    ListWatchConfig
	Lw     ListWatch
	Bus    chan *ListWatchResp
	stopCh chan struct{}
	stop   bool
	mux    sync.Mutex
}

func (w *EventBus) ResourceEventBus() <-chan *ListWatchResp {
	return w.Bus
}

func (w *EventBus) process(_ context.Context) {
	stopCh := make(chan struct{})
	ctx, cancel := context.WithTimeout(w.Cfg.Context, w.Cfg.Timeout)
	gopool.Go(func(_ context.Context) {
		defer close(stopCh)
		_ = w.Lw.DoWatch(ctx, w.sendEvent)
	})

	select {
	case <-stopCh:
		// timed out or exception
		w.Stop()
		cancel()
	case <-w.stopCh:
		cancel()
	}

}

func (w *EventBus) sendEvent(resp *ListWatchResp) {
	defer log.Recover()
	w.Bus <- resp
}

func (w *EventBus) Stop() {
	w.mux.Lock()
	if w.stop {
		w.mux.Unlock()
		return
	}
	w.stop = true
	close(w.stopCh)
	close(w.Bus)
	w.mux.Unlock()
}

func NewEventBus(lw ListWatch, cfg ListWatchConfig) *EventBus {
	w := &EventBus{
		Cfg:    cfg,
		Lw:     lw,
		Bus:    make(chan *ListWatchResp, EventBusSize),
		stopCh: make(chan struct{}),
	}
	gopool.Go(w.process)
	return w
}
