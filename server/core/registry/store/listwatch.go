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

var globalRevision int64 = 0
var revisionMux sync.RWMutex

func SetRevision(rev int64) {
	if globalRevision >= rev {
		return
	}
	revisionMux.Lock()
	if globalRevision < rev {
		globalRevision = rev
	}
	revisionMux.Unlock()
}

func Revision() (rev int64) {
	revisionMux.RLock()
	rev = globalRevision
	revisionMux.RUnlock()
	return
}

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

type ListWatcher struct {
	Client registry.Registry
	Key    string

	modRev int64
}

func (lw *ListWatcher) ModRevision() int64 {
	return lw.modRev
}

func (lw *ListWatcher) List(op *ListOptions) ([]*mvccpb.KeyValue, error) {
	otCtx, _ := context.WithTimeout(op.Context, op.Timeout)
	resp, err := lw.Client.Do(otCtx, registry.WithWatchPrefix(lw.Key))
	if err != nil {
		return nil, err
	}
	lw.setModRevision(resp.Revision)
	if len(resp.Kvs) == 0 {
		return nil, nil
	}
	return resp.Kvs, nil
}

func (lw *ListWatcher) Watch(op *ListOptions) *Watcher {
	return newWatcher(lw, op)
}

func (lw *ListWatcher) setModRevision(rev int64) {
	lw.modRev = rev
	SetRevision(rev)
}

func (lw *ListWatcher) doWatch(ctx context.Context, f func(evt *Event)) error {
	ops := registry.WithWatchPrefix(lw.Key)
	ops.WithRev = lw.ModRevision() + 1
	err := lw.Client.Watch(ctx, ops, func(message string, evt *registry.PluginResponse) error {
		if lw.ModRevision() < evt.Revision {
			lw.setModRevision(evt.Revision)
		}

		sendEvt := errEvent(lw.Key, fmt.Errorf("unknown event %+v", evt))
		sendEvt.Revision = evt.Revision

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
		lw.setModRevision(0)
		f(errEvent(lw.Key, err))
	}
	return err
}

type Watcher struct {
	ListOps *ListOptions
	lw      *ListWatcher
	bus     chan *Event
	stopCh  chan struct{}
	stop    bool
	mux     sync.Mutex
}

func (w *Watcher) EventBus() <-chan *Event {
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

func (w *Watcher) sendEvent(evt *Event) {
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
		bus:     make(chan *Event, EVENT_BUS_MAX_SIZE),
		stopCh:  make(chan struct{}),
	}
	go w.process()
	return w
}
