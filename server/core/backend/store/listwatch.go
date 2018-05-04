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
package store

import (
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
	"sync"
	"time"
)

type ListOptions struct {
	Timeout time.Duration
	Context context.Context
}

func (lo *ListOptions) String() string {
	return fmt.Sprintf("{timeout: %s}", lo.Timeout)
}

type ListWatcher struct {
	Client registry.Registry
	Prefix string

	rev int64
}

func (lw *ListWatcher) List(op ListOptions) ([]*mvccpb.KeyValue, error) {
	otCtx, _ := context.WithTimeout(op.Context, op.Timeout)
	resp, err := lw.Client.Do(otCtx, registry.WatchPrefixOpOptions(lw.Prefix)...)
	if err != nil {
		util.Logger().Errorf(err, "list prefix %s failed, rev: %d->0", lw.Prefix, lw.Revision())
		lw.setRevision(0)
		return nil, err
	}
	lw.setRevision(resp.Revision)
	if len(resp.Kvs) == 0 {
		return nil, nil
	}
	return resp.Kvs, nil
}

func (lw *ListWatcher) Revision() int64 {
	return lw.rev
}

func (lw *ListWatcher) setRevision(rev int64) {
	lw.rev = rev
}

func (lw *ListWatcher) Watch(op ListOptions) *Watcher {
	return newWatcher(lw, op)
}

func (lw *ListWatcher) doWatch(ctx context.Context, f func(evt []KvEvent)) error {
	rev := lw.Revision()
	opts := append(
		registry.WatchPrefixOpOptions(lw.Prefix),
		registry.WithRev(rev+1),
		registry.WithWatchCallback(
			func(message string, resp *registry.PluginResponse) error {
				if resp == nil || len(resp.Kvs) == 0 {
					return fmt.Errorf("unknown event %s", resp)
				}

				util.Logger().Infof("watch prefix %s, start rev %d+1, event: %s", lw.Prefix, rev, resp)

				lw.setRevision(resp.Revision)

				evts := make([]KvEvent, len(resp.Kvs))
				for i, kv := range resp.Kvs {
					evt := KvEvent{Prefix: lw.Prefix, Revision: kv.ModRevision}
					switch {
					case resp.Action == registry.Put && kv.Version == 1:
						evt.Type, evt.Object = proto.EVT_CREATE, kv
					case resp.Action == registry.Put:
						evt.Type, evt.Object = proto.EVT_UPDATE, kv
					case resp.Action == registry.Delete:
						evt.Type, evt.Object = proto.EVT_DELETE, kv
					default:
						return fmt.Errorf("unknown KeyValue %v", kv)
					}
					evts[i] = evt
				}
				f(evts)
				return nil
			}))

	err := lw.Client.Watch(ctx, opts...)
	if err != nil { // compact可能会导致watch失败 or message body size lager than 4MB
		util.Logger().Errorf(err, "watch prefix %s failed, start rev: %d+1->%d->0", lw.Prefix, rev, lw.Revision())

		lw.setRevision(0)
		f([]KvEvent{errEvent(lw.Prefix, err)})
	}
	return err
}

type Watcher struct {
	ListOps ListOptions
	lw      *ListWatcher
	bus     chan []KvEvent
	stopCh  chan struct{}
	stop    bool
	mux     sync.Mutex
}

func (w *Watcher) EventBus() <-chan []KvEvent {
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

func (w *Watcher) sendEvent(evts []KvEvent) {
	defer util.RecoverAndReport()
	w.bus <- evts
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

func errEvent(key string, err error) KvEvent {
	return KvEvent{
		Type:   proto.EVT_ERROR,
		Prefix: key,
		Object: err,
	}
}

func newWatcher(lw *ListWatcher, listOps ListOptions) *Watcher {
	w := &Watcher{
		ListOps: listOps,
		lw:      lw,
		bus:     make(chan []KvEvent, EVENT_BUS_MAX_SIZE),
		stopCh:  make(chan struct{}),
	}
	util.Go(w.process)
	return w
}
