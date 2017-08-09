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
	"golang.org/x/net/context"
	"sync"
	"time"
)

const EVENT_BUS_MAX_SIZE = 1000

type Event struct {
	Revision int64
	Type     proto.EventType
	WatchKey string
	Object   interface{}
}

type Watcher interface {
	EventBus() <-chan *Event
	Stop()
}

type ListOptions struct {
	Timeout time.Duration
	Context context.Context
}

type ListWatcher interface {
	Revision() int64
	List(op *ListOptions) ([]interface{}, error)
	Watch(op *ListOptions) Watcher
}

type KvListWatcher struct {
	Client registry.Registry
	Key    string

	rev int64
}

func (lw *KvListWatcher) Revision() int64 {
	return lw.rev
}

func (lw *KvListWatcher) List(op *ListOptions) ([]interface{}, error) {
	otCtx, _ := context.WithTimeout(op.Context, op.Timeout)
	resp, err := lw.Client.Do(otCtx, registry.WithWatchPrefix(lw.Key))
	if err != nil {
		return nil, err
	}
	lw.upgradeRevision(resp.Revision)
	if len(resp.Kvs) == 0 {
		return nil, nil
	}
	itfs := make([]interface{}, len(resp.Kvs))
	for idx, kv := range resp.Kvs {
		itfs[idx] = kv
	}
	return itfs, nil
}

func (lw *KvListWatcher) Watch(op *ListOptions) Watcher {
	return newKvWatcher(lw, op)
}

func (lw *KvListWatcher) upgradeRevision(rev int64) {
	lw.rev = rev
}

func (lw *KvListWatcher) doWatch(ctx context.Context, f func(evt *Event)) error {
	ops := registry.WithWatchPrefix(lw.Key)
	ops.WithRev = lw.Revision() + 1
	err := lw.Client.Watch(ctx, ops, func(message string, evt *registry.PluginResponse) error {
		if lw.Revision() < evt.Revision {
			lw.upgradeRevision(evt.Revision)
		}
		sendEvt := &Event{
			Revision: evt.Revision,
			Type:     proto.EVT_ERROR,
			WatchKey: lw.Key,
			Object:   fmt.Errorf("unknown event %+v", evt),
		}
		if evt != nil && len(evt.Kvs) > 0 {
			switch {
			case evt.Action == registry.PUT && evt.Kvs[0].Version == 1:
				sendEvt.Type, sendEvt.Object = proto.EVT_CREATE, evt.Kvs[0]
			case evt.Action == registry.PUT:
				sendEvt.Type, sendEvt.Object = proto.EVT_UPDATE, evt.Kvs[0]
			case evt.Action == registry.DELETE:
				kv := evt.PrevKv
				if kv == nil {
					// TODO 内嵌无法获取
					kv = evt.Kvs[0]
				}
				sendEvt.Type, sendEvt.Object = proto.EVT_DELETE, kv
			}
		}
		f(sendEvt)
		return nil
	})
	if err != nil { // compact可能会导致watch失败
		lw.upgradeRevision(0)
		f(errEvent(lw.Key, err))
	}
	return err
}

type KvWatcher struct {
	ListOps *ListOptions
	lw      *KvListWatcher
	bus     chan *Event
	stopCh  chan struct{}
	stop    bool
	mux     sync.Mutex
}

func (w *KvWatcher) EventBus() <-chan *Event {
	return w.bus
}

func (w *KvWatcher) process() {
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

func (w *KvWatcher) sendEvent(evt *Event) {
	w.bus <- evt
	// LOG ignore event
}

func (w *KvWatcher) Stop() {
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

func errEvent(watchKey string, err error) *Event {
	return &Event{
		Type:     proto.EVT_ERROR,
		WatchKey: watchKey,
		Object:   err,
	}
}

func newKvWatcher(lw *KvListWatcher, listOps *ListOptions) Watcher {
	w := &KvWatcher{
		ListOps: listOps,
		lw:      lw,
		bus:     make(chan *Event, EVENT_BUS_MAX_SIZE),
		stopCh:  make(chan struct{}),
	}
	go w.process()
	return w
}
