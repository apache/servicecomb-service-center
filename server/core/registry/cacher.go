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
package registry

import (
	"fmt"
	"github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/util"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"time"
)

type Cacher interface {
	Run()
}

type KvCacher struct {
	Cfg *KvCacherConfig
	lw  ListWatcher

	store map[string]*mvccpb.KeyValue
}

func (c *KvCacher) ListAndWatch() error {
	start := time.Now()
	listOps := &ListOptions{
		Timeout: c.Cfg.Timeout,
	}
	rev := c.lw.Revision()
	if rev == 0 {
		kvs, err := c.lw.List(listOps)
		if err != nil {
			util.LOGGER.Errorf(err, "list key %s failed, list options: %+v", c.Cfg.Key, listOps)
			return err
		}
		rev = c.lw.Revision()
		c.sync(c.filter(rev, kvs))
		syncDuration := time.Now().Sub(start)

		if syncDuration > 5*time.Second {
			util.LOGGER.Warnf(nil, "finish to cache key %s, %d items took %s! list options: %+v, rev: %d",
				c.Cfg.Key, len(kvs), syncDuration, listOps, c.lw.Revision())
		} else {
			util.LOGGER.Infof("finish to cache key %s, %d items took %s, list options: %+v, rev: %d",
				c.Cfg.Key, len(kvs), syncDuration, listOps, c.lw.Revision())
		}
	}

	watcher := c.lw.Watch(listOps)
	util.LOGGER.Debugf("finish to new watcher, key %s, list options: %+v, start rev: %d+1",
		c.Cfg.Key, listOps, rev)
	err := c.handleWatcher(watcher)
	if err != nil {
		util.LOGGER.Errorf(err, "handle watcher failed, watch key %s, list options: %+v", c.Cfg.Key, listOps)
		return err
	}
	return nil
}

func (c *KvCacher) handleWatcher(watcher Watcher) error {
	defer watcher.Stop()
	for evt := range watcher.EventBus() {
		if evt.Type == proto.EVT_ERROR {
			err := evt.Object.(error)
			return err
		}
		c.sync([]*Event{evt})
	}
	return nil
}

func (c *KvCacher) sync(evts []*Event) {
	for _, evt := range evts {
		kv := evt.Object.(*mvccpb.KeyValue)
		key := BytesToStringWithNoCopy(kv.Key)
		prevKv, ok := c.store[key]
		switch evt.Type {
		case proto.EVT_CREATE, proto.EVT_UPDATE:
			util.LOGGER.Debugf("sync %s event and notify watcher, cache key %s, %+v", evt.Type, key, kv)
			c.store[key] = kv
			t := evt.Type
			if !ok && evt.Type != proto.EVT_CREATE {
				util.LOGGER.Warnf(nil, "unexpected %s event! it should be %s key %s",
					evt.Type, proto.EVT_CREATE, key)
				t = proto.EVT_CREATE
			}
			if ok && evt.Type != proto.EVT_UPDATE {
				util.LOGGER.Warnf(nil, "unexpected %s event! it should be %s key %s",
					evt.Type, proto.EVT_UPDATE, key)
				t = proto.EVT_UPDATE
			}
			c.Cfg.OnEvent(&KvEvent{
				Revision: evt.Revision,
				Action:   t,
				KV:       kv,
			})
		case proto.EVT_DELETE:
			if ok {
				util.LOGGER.Debugf("remove key %s and notify watcher, %+v", key, kv)
				delete(c.store, key)
				c.Cfg.OnEvent(&KvEvent{
					Revision: evt.Revision,
					Action:   evt.Type,
					KV:       prevKv,
				})
				continue
			}
			util.LOGGER.Warnf(nil, "unexpected %s event! nonexistent key %s", evt.Type, key)
		}
	}
}

func (c *KvCacher) filter(rev int64, items []interface{}) []*Event {
	oc, nc := len(c.store), len(items)
	tc := oc + nc
	if tc == 0 {
		return nil
	}
	max := oc
	if nc > oc {
		max = nc
	}

	newStore := make(map[string]*mvccpb.KeyValue)
	for _, itf := range items {
		kv := itf.(*mvccpb.KeyValue)
		newStore[BytesToStringWithNoCopy(kv.Key)] = kv
	}
	filterStopCh := make(chan struct{})
	eventsCh := make(chan *Event, max)
	go func() {
		for k, v := range c.store {
			_, ok := newStore[k]
			if !ok {
				eventsCh <- &Event{
					Revision: rev,
					Type:     proto.EVT_DELETE,
					WatchKey: c.Cfg.Key,
					Object:   v,
				}
			}
		}
		close(filterStopCh)
	}()

	go func() {
		for k, v := range newStore {
			ov, ok := c.store[k]
			if !ok {
				eventsCh <- &Event{
					Revision: rev,
					Type:     proto.EVT_CREATE,
					WatchKey: c.Cfg.Key,
					Object:   v,
				}
				continue
			}
			if ov.ModRevision < v.ModRevision {
				eventsCh <- &Event{
					Revision: rev,
					Type:     proto.EVT_UPDATE,
					WatchKey: c.Cfg.Key,
					Object:   v,
				}
				continue
			}
		}
		select {
		case <-filterStopCh:
			close(eventsCh)
		}
	}()

	evts := make([]*Event, 0, max)
	for evt := range eventsCh {
		evts = append(evts, evt)
	}
	return evts
}

func (c *KvCacher) Run() {
	util.Go(func(stopCh <-chan struct{}) {
		util.LOGGER.Warnf(nil, "start to list and watch %s", c.Cfg)
		for {
			start := time.Now()
			c.ListAndWatch()
			watchDuration := time.Now().Sub(start)
			nextPeriod := 0 * time.Second
			if watchDuration > 0 && c.Cfg.Period > watchDuration {
				nextPeriod = c.Cfg.Period - watchDuration
			}
			select {
			case <-stopCh:
				return
			case <-time.After(nextPeriod):
			}
		}
	})
}

type KvEvent struct {
	Revision int64
	Action   proto.EventType
	KV       *mvccpb.KeyValue
}

type KvCacherConfig struct {
	Key     string
	Timeout time.Duration
	Period  time.Duration
	OnEvent func(evt *KvEvent) error
}

func (cfg *KvCacherConfig) String() string {
	return fmt.Sprintf("{key: %s, timeout: %s, period: %s}",
		cfg.Key, cfg.Timeout, cfg.Period)
}

func NewKvCacher(cfg *KvCacherConfig) Cacher {
	cacher := &KvCacher{
		Cfg: cfg,
		lw: &KvListWatcher{
			Client: GetRegisterCenter(),
			Key:    cfg.Key,
		},
		store: make(map[string]*mvccpb.KeyValue),
	}
	return cacher
}
