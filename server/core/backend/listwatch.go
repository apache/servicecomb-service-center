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
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"golang.org/x/net/context"
	"sync"
)

type ListWatcher struct {
	Client registry.Registry
	Prefix string

	rev int64
}

func (lw *ListWatcher) List(op ListWatchConfig) (*registry.PluginResponse, error) {
	otCtx, _ := context.WithTimeout(op.Context, op.Timeout)
	resp, err := lw.Client.Do(otCtx, registry.WatchPrefixOpOptions(lw.Prefix)...)
	if err != nil {
		util.Logger().Errorf(err, "list prefix %s failed, current rev: %d", lw.Prefix, lw.Revision())
		return nil, err
	}
	lw.setRevision(resp.Revision)
	return resp, nil
}

func (lw *ListWatcher) Revision() int64 {
	return lw.rev
}

func (lw *ListWatcher) setRevision(rev int64) {
	lw.rev = rev
}

func (lw *ListWatcher) Watch(op ListWatchConfig) *Watcher {
	return newWatcher(lw, op)
}

func (lw *ListWatcher) doWatch(ctx context.Context, f func(*registry.PluginResponse)) error {
	rev := lw.Revision()
	opts := append(
		registry.WatchPrefixOpOptions(lw.Prefix),
		registry.WithRev(rev+1),
		registry.WithWatchCallback(
			func(message string, resp *registry.PluginResponse) error {
				if resp == nil || len(resp.Kvs) == 0 {
					return fmt.Errorf("unknown event %s", resp)
				}

				util.Logger().Infof("caught event %s, watch prefix %s, start rev %d+1,", resp, lw.Prefix, rev)

				lw.setRevision(resp.Revision)

				f(resp)
				return nil
			}))

	err := lw.Client.Watch(ctx, opts...)
	if err != nil { // compact可能会导致watch失败 or message body size lager than 4MB
		util.Logger().Errorf(err, "watch prefix %s failed, start rev: %d+1->%d->0", lw.Prefix, rev, lw.Revision())

		lw.setRevision(0)
		f(nil)
	}
	return err
}

type Watcher struct {
	ListOps ListWatchConfig
	lw      *ListWatcher
	bus     chan *registry.PluginResponse
	stopCh  chan struct{}
	stop    bool
	mux     sync.Mutex
}

func (w *Watcher) EventBus() <-chan *registry.PluginResponse {
	return w.bus
}

func (w *Watcher) process(_ context.Context) {
	stopCh := make(chan struct{})
	ctx, cancel := context.WithTimeout(w.ListOps.Context, w.ListOps.Timeout)
	util.Go(func(_ context.Context) {
		defer close(stopCh)
		w.lw.doWatch(ctx, w.sendEvent)
	})

	select {
	case <-stopCh:
		// timed out or exception
		w.Stop()
	case <-w.stopCh:
		cancel()
	}
}

func (w *Watcher) sendEvent(resp *registry.PluginResponse) {
	defer util.RecoverAndReport()
	w.bus <- resp
}

func (w *Watcher) Stop() {
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

func newWatcher(lw *ListWatcher, listOps ListWatchConfig) *Watcher {
	w := &Watcher{
		ListOps: listOps,
		lw:      lw,
		bus:     make(chan *registry.PluginResponse, EVENT_BUS_MAX_SIZE),
		stopCh:  make(chan struct{}),
	}
	util.Go(w.process)
	return w
}
