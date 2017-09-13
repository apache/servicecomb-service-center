//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package store

import (
	"fmt"
	"github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/util"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
	"sync"
	"time"
)

const EVENT_BUS_MAX_SIZE = 1000

type Event struct {
	Revision int64
	Type     proto.EventType
	Key      string
	Object   interface{}
}

type ListOptions struct {
	Timeout time.Duration
	Context context.Context
}

func (lo *ListOptions) String() string {
	return fmt.Sprintf("{timeout: %s}", lo.Timeout)
}

type ListWatcher struct {
	Client registry.Registry
	Key    string

	rev int64
}

func (lw *ListWatcher) List(op *ListOptions) ([]*mvccpb.KeyValue, error) {
	otCtx, _ := context.WithTimeout(op.Context, op.Timeout)
	resp, err := lw.Client.Do(otCtx, registry.WithWatchPrefix(lw.Key)...)
	if err != nil {
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

func (lw *ListWatcher) Watch(op *ListOptions) *Watcher {
	return newWatcher(lw, op)
}

func (lw *ListWatcher) doWatch(ctx context.Context, f func(evt []*Event)) error {
	opts := append(
		registry.WithWatchPrefix(lw.Key),
		registry.WithRev(lw.Revision()+1),
		registry.WithWatchCallback(
			func(message string, resp *registry.PluginResponse) error {
				if resp == nil || len(resp.Kvs) == 0 {
					return fmt.Errorf("unknown event %s", resp)
				}

				lw.setRevision(resp.Revision)

				evts := make([]*Event, len(resp.Kvs))
				for i, kv := range resp.Kvs {
					evt := &Event{Key: lw.Key, Revision: kv.ModRevision}
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
	if err != nil { // compact可能会导致watch失败
		lw.setRevision(0)
		f([]*Event{errEvent(lw.Key, err)})
	}
	return err
}

type Watcher struct {
	ListOps *ListOptions
	lw      *ListWatcher
	bus     chan []*Event
	stopCh  chan struct{}
	stop    bool
	mux     sync.Mutex
}

func (w *Watcher) EventBus() <-chan []*Event {
	return w.bus
}

func (w *Watcher) process() {
	stopCh := make(chan struct{})
	ctx, cancel := context.WithTimeout(w.ListOps.Context, w.ListOps.Timeout)
	go func() {
		defer close(stopCh)
		w.lw.doWatch(ctx, w.sendEvent)
	}()

	select {
	case <-stopCh:
		// time out
		w.Stop()
	case <-w.stopCh:
		cancel()
	}
}

func (w *Watcher) sendEvent(evt []*Event) {
	defer util.RecoverAndReport()
	w.bus <- evt
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

func errEvent(key string, err error) *Event {
	return &Event{
		Type:   proto.EVT_ERROR,
		Key:    key,
		Object: err,
	}
}

func newWatcher(lw *ListWatcher, listOps *ListOptions) *Watcher {
	w := &Watcher{
		ListOps: listOps,
		lw:      lw,
		bus:     make(chan []*Event, EVENT_BUS_MAX_SIZE),
		stopCh:  make(chan struct{}),
	}
	go w.process()
	return w
}
