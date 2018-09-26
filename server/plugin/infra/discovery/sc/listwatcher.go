// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sc

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/client/sc"
	"github.com/apache/incubator-servicecomb-service-center/pkg/gopool"
	"github.com/apache/incubator-servicecomb-service-center/pkg/log"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/admin/model"
	"github.com/apache/incubator-servicecomb-service-center/server/core/backend"
	"github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/discovery"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"golang.org/x/net/context"
	"sync"
	"time"
)

var (
	client     *ServiceCenterAggregate
	clientOnce sync.Once
)

type ServiceCenterInterface interface {
	Register(t discovery.Type, cacher *ServiceCenterCacher)
	Sync(ctx context.Context) error
	Ready() <-chan struct{}
}

type ServiceCenterListWatcher struct {
	Client *sc.SCClient

	cachers map[discovery.Type]*ServiceCenterCacher

	ready chan struct{}
}

func (c *ServiceCenterListWatcher) Initialize() {
	c.ready = make(chan struct{})
	c.cachers = make(map[discovery.Type]*ServiceCenterCacher)
}

func (c *ServiceCenterListWatcher) Sync(ctx context.Context) error {
	cache, err := c.Client.GetScCache()
	if err != nil {
		log.Errorf(err, "sync failed")
		return err
	}

	defer util.SafeCloseChan(c.ready)

	c.syncMicroservice(cache)
	c.syncInstance(cache)
	return nil
}

func (c *ServiceCenterListWatcher) syncMicroservice(cache *model.Cache) {
	aliasCacher, ok := c.cachers[backend.SERVICE_ALIAS]
	if ok {
		c.check(aliasCacher, &cache.Aliases)
	}
	indexCacher, ok := c.cachers[backend.SERVICE_INDEX]
	if ok {
		c.check(indexCacher, &cache.Indexes)
	}
	serviceCacher, ok := c.cachers[backend.SERVICE]
	if ok {
		c.check(serviceCacher, &cache.Microservices)
	}
}

func (c *ServiceCenterListWatcher) syncInstance(cache *model.Cache) {
	instCacher, ok := c.cachers[backend.INSTANCE]
	if ok {
		c.check(instCacher, &cache.Instances)
	}
}

func (c *ServiceCenterListWatcher) check(local *ServiceCenterCacher, remote model.Getter) {
	init := !c.IsReady()

	remote.ForEach(func(_ int, v *model.KV) bool {
		kv := local.Cache().Get(v.Key)
		newKv := &discovery.KeyValue{
			Key:            util.StringToBytesWithNoCopy(v.Key),
			Value:          v.Value,
			Version:        v.Rev,
			CreateRevision: v.Rev,
			ModRevision:    v.Rev,
		}
		switch {
		case kv == nil && init:
			local.Notify(proto.EVT_INIT, v.Key, newKv)
		case kv == nil && !init:
			local.Notify(proto.EVT_CREATE, v.Key, newKv)
		case kv.ModRevision != v.Rev:
			local.Notify(proto.EVT_UPDATE, v.Key, newKv)
		}
		return true
	})

	var deletes []*discovery.KeyValue
	local.Cache().ForEach(func(key string, v *discovery.KeyValue) (next bool) {
		var exist bool
		remote.ForEach(func(_ int, v *model.KV) bool {
			exist = v.Key == key
			return !exist
		})
		if !exist {
			deletes = append(deletes, v)
		}
		return true
	})
	for _, v := range deletes {
		local.Notify(proto.EVT_DELETE, util.BytesToStringWithNoCopy(v.Key), v)
	}
}

func (c *ServiceCenterListWatcher) loop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Debug("service center client is stopped")
			return
		case <-time.After(registry.RegistryConfig().AutoSyncInterval):
			// TODO support watching sc
			c.Sync(ctx)
		}
	}
}

// unsafe
func (c *ServiceCenterListWatcher) Register(t discovery.Type, cacher *ServiceCenterCacher) {
	c.cachers[t] = cacher
}

func (c *ServiceCenterListWatcher) Run() {
	gopool.Go(c.loop)
}

func (c *ServiceCenterListWatcher) Ready() <-chan struct{} {
	return c.ready
}

func (c *ServiceCenterListWatcher) IsReady() bool {
	select {
	case <-c.ready:
		return true
	default:
		return false
	}
}

type ServiceCenterAggregate []*ServiceCenterListWatcher

func (s *ServiceCenterAggregate) Initialize() {
	for _, lw := range *s {
		lw.Initialize()
		lw.Run()
	}
}

func (s *ServiceCenterAggregate) Register(t discovery.Type, cacher *ServiceCenterCacher) {
	for _, lw := range *s {
		lw.Register(t, cacher)
	}
}

func (s *ServiceCenterAggregate) Sync(ctx context.Context) error {
	for _, lw := range *s {
		lw.Sync(ctx)
	}
	return nil
}

func (s *ServiceCenterAggregate) Ready() <-chan struct{} {
	for _, lw := range *s {
		<-lw.Ready()
	}
	return closedCh
}

func NewServiceCenterAggregate() *ServiceCenterAggregate {
	c := &ServiceCenterAggregate{}
	clusters := registry.RegistryConfig().Clusters()
	for name, addr := range clusters {
		if len(name) == 0 || name == registry.RegistryConfig().ClusterName {
			continue
		}
		sc.Addr = addr
		client, err := sc.NewSCClient()
		if err != nil {
			log.Errorf(err, "new service center[%s][%s] client failed", name, addr)
			continue
		}
		client.Timeout = registry.RegistryConfig().RequestTimeOut
		*c = append(*c, &ServiceCenterListWatcher{
			Client: client,
		})
		log.Infof("list and watch service center[%s][%s]", name, addr)
	}
	return c
}

func ServiceCenter() ServiceCenterInterface {
	clientOnce.Do(func() {
		client = NewServiceCenterAggregate()
		client.Initialize()
	})
	return client
}
