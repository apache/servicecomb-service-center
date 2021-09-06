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

package event

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/mux"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/datasource/etcd/state/kvstore"
	serviceUtil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/queue"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/config"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/foundation/backoff"
	"github.com/go-chassis/foundation/gopool"
	"github.com/go-chassis/foundation/stringutil"
	"github.com/go-chassis/foundation/timeutil"
	"github.com/little-cui/etcdadpt"
)

const DepQueueLock mux.ID = "/cse-sr/lock/dep-queue"

// just for unit test
var testMux sync.Mutex

// DependencyEventHandler add or remove the service dependencies
// when user call find instance api or dependence operation api
type DependencyEventHandler struct {
	signals *queue.UniQueue
}

func (h *DependencyEventHandler) Type() kvstore.Type {
	return sd.TypeDependencyQueue
}

func (h *DependencyEventHandler) OnEvent(evt kvstore.Event) {
	action := evt.Type
	if action != pb.EVT_CREATE && action != pb.EVT_UPDATE && action != pb.EVT_INIT {
		return
	}
	h.notify()
}

func (h *DependencyEventHandler) notify() {
	err := h.signals.Put(struct{}{})
	if err != nil {
		log.Error("", err)
	}
}

func (h *DependencyEventHandler) backoff(f func(), retries int) int {
	if f != nil {
		<-time.After(backoff.GetBackoff().Delay(retries))
		f()
	}
	return retries + 1
}

func (h *DependencyEventHandler) tryWithBackoff(success func() error, backoff func(), retries int) (int, error) {
	defer log.Recover()
	lock, err := mux.Try(DepQueueLock)
	if err != nil {
		log.Error(fmt.Sprintf("try to lock %s failed", DepQueueLock), err)
		return h.backoff(backoff, retries), err
	}

	if lock == nil {
		return 0, nil
	}

	defer func() {
		if err := lock.Unlock(); err != nil {
			log.Error("", err)
		}
	}()
	err = success()
	if err != nil {
		log.Error("handle dependency event failed", err)
		return h.backoff(backoff, retries), err
	}

	return 0, nil
}

func (h *DependencyEventHandler) eventLoop() {
	gopool.Go(func(ctx context.Context) {
		// the events will lose, need to handle dependence records periodically
		period := config.GetRegistry().CacheTTL
		timer := time.NewTimer(period)
		retries := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-h.signals.Chan():
				_, err := h.tryWithBackoff(h.Handle, h.notify, retries)
				if err != nil {
					log.Error("", err)
				}
				timeutil.ResetTimer(timer, period)
			case <-timer.C:
				h.notify()
				timer.Reset(period)
			}
		}
	})
}

type DependencyEventHandlerResource struct {
	dep           *pb.ConsumerDependency
	kv            *kvstore.KeyValue
	domainProject string
}

func NewDependencyEventHandlerResource(dep *pb.ConsumerDependency, kv *kvstore.KeyValue, domainProject string) *DependencyEventHandlerResource {
	return &DependencyEventHandlerResource{
		dep,
		kv,
		domainProject,
	}
}

func (h *DependencyEventHandler) Handle() error {
	testMux.Lock()
	defer testMux.Unlock()

	key := path.GetServiceDependencyQueueRootKey("")
	resp, err := sd.DependencyQueue().Search(context.Background(), etcdadpt.WithNoCache(),
		etcdadpt.WithStrKey(key), etcdadpt.WithPrefix())
	if err != nil {
		return err
	}

	// maintain dependency rules.
	l := len(resp.Kvs)
	if l == 0 {
		return nil
	}

	cleanUpDomainProjects := make(map[string]struct{})
	defer h.CleanUp(cleanUpDomainProjects)

	for _, keyValue := range resp.Kvs {
		r, ok := keyValue.Value.(*pb.ConsumerDependency)
		if !ok {
			log.Error("failed to assert consumerDependency", datasource.ErrAssertFail)
			continue
		}

		_, domainProject, uuid := path.GetInfoFromDependencyQueueKV(keyValue.Key)
		if uuid == path.DepsQueueUUID {
			cleanUpDomainProjects[domainProject] = struct{}{}
		}
		res := NewDependencyEventHandlerResource(r, keyValue, domainProject)

		if err := h.dependencyRuleHandle(res); err != nil {
			return err
		}
	}
	return nil
}

func (h *DependencyEventHandler) dependencyRuleHandle(res interface{}) error {
	ctx := util.WithGlobal(context.Background())
	dependencyEventHandlerRes := res.(*DependencyEventHandlerResource)
	r := dependencyEventHandlerRes.dep
	consumerFlag := util.StringJoin([]string{r.Consumer.Environment, r.Consumer.AppId, r.Consumer.ServiceName, r.Consumer.Version}, "/")

	domainProject := dependencyEventHandlerRes.domainProject
	consumerInfo := pb.DependenciesToKeys([]*pb.MicroServiceKey{r.Consumer}, domainProject)[0]
	providersInfo := pb.DependenciesToKeys(r.Providers, domainProject)

	var dep serviceUtil.Dependency
	dep.DomainProject = domainProject
	dep.Consumer = consumerInfo
	dep.ProvidersRule = providersInfo
	err := serviceUtil.AddDependencyRule(ctx, &dep)
	if err != nil {
		log.Error(fmt.Sprintf("modify dependency rule failed, override: %t, consumer %s", r.Override, consumerFlag), err)
		return fmt.Errorf("override: %t, consumer is %s, %s", r.Override, consumerFlag, err.Error())
	}

	if err = h.removeKV(ctx, dependencyEventHandlerRes.kv); err != nil {
		log.Error(fmt.Sprintf("remove dependency rule failed, override: %t, consumer %s", r.Override, consumerFlag), err)
		return err
	}

	log.Info(fmt.Sprintf("maintain dependency [%v] successfully", r))
	return nil
}

func (h *DependencyEventHandler) removeKV(ctx context.Context, kv *kvstore.KeyValue) error {
	dResp, err := etcdadpt.TxnWithCmp(ctx, etcdadpt.Ops(etcdadpt.OpDel(etcdadpt.WithKey(kv.Key))),
		etcdadpt.If(etcdadpt.EqualVer(stringutil.Bytes2str(kv.Key), kv.Version)),
		nil)
	if err != nil {
		return fmt.Errorf("can not remove the dependency %s request, %s", util.BytesToStringWithNoCopy(kv.Key), err.Error())
	}
	if !dResp.Succeeded {
		log.Info(fmt.Sprintf("the dependency %s request is changed", util.BytesToStringWithNoCopy(kv.Key)))
	}
	return nil
}

func (h *DependencyEventHandler) CleanUp(domainProjects map[string]struct{}) {
	for domainProject := range domainProjects {
		ctx := util.WithGlobal(context.Background())
		if err := serviceUtil.CleanUpDependencyRules(ctx, domainProject); err != nil {
			log.Error(fmt.Sprintf("clean up '%s' dependency rules failed", domainProject), err)
		}
	}
}

func NewDependencyEventHandler() *DependencyEventHandler {
	h := &DependencyEventHandler{
		signals: queue.NewUniQueue(),
	}
	h.eventLoop()
	return h
}
