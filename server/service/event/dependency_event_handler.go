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
	"github.com/apache/servicecomb-service-center/pkg/backoff"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/queue"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/backend"
	pb "github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/server/mux"
	"github.com/apache/servicecomb-service-center/server/plugin/discovery"
	"github.com/apache/servicecomb-service-center/server/plugin/registry"
	serviceUtil "github.com/apache/servicecomb-service-center/server/service/util"
	"time"
)

const defaultEventHandleInterval = 5 * time.Minute

// DependencyEventHandler add or remove the service dependencies
// when user call find instance api or dependence operation api
type DependencyEventHandler struct {
	signals *queue.UniQueue
}

func (h *DependencyEventHandler) Type() discovery.Type {
	return backend.DEPENDENCY_QUEUE
}

func (h *DependencyEventHandler) OnEvent(evt discovery.KvEvent) {
	action := evt.Type
	if action != pb.EVT_CREATE && action != pb.EVT_UPDATE && action != pb.EVT_INIT {
		return
	}
	h.notify()
}

func (h *DependencyEventHandler) notify() {
	h.signals.Put(struct{}{})
}

func (h *DependencyEventHandler) backoff(f func(), retries int) int {
	if f != nil {
		<-time.After(backoff.GetBackoff().Delay(retries))
		f()
	}
	return retries + 1
}

func (h *DependencyEventHandler) tryWithBackoff(success func() error, backoff func(), retries int) (error, int) {
	defer log.Recover()
	lock, err := mux.Try(mux.DepQueueLock)
	if err != nil {
		log.Errorf(err, "try to lock %s failed", mux.DepQueueLock)
		return err, h.backoff(backoff, retries)
	}

	if lock == nil {
		return nil, 0
	}

	defer lock.Unlock()
	err = success()
	if err != nil {
		log.Errorf(err, "handle dependency event failed")
		return err, h.backoff(backoff, retries)
	}

	return nil, 0
}

func (h *DependencyEventHandler) eventLoop() {
	gopool.Go(func(ctx context.Context) {
		// the events will lose, need to handle dependence records periodically
		period := defaultEventHandleInterval
		if core.ServerInfo.Config.CacheTTL > 0 {
			period = core.ServerInfo.Config.CacheTTL
		}
		timer := time.NewTimer(period)
		retries := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-h.signals.Chan():
				_, retries = h.tryWithBackoff(h.Handle, h.notify, retries)
				util.ResetTimer(timer, period)
			case <-timer.C:
				h.notify()
				timer.Reset(period)
			}
		}
	})
}

type DependencyEventHandlerResource struct {
	dep           *pb.ConsumerDependency
	kv            *discovery.KeyValue
	domainProject string
}

func NewDependencyEventHandlerResource(dep *pb.ConsumerDependency, kv *discovery.KeyValue, domainProject string) *DependencyEventHandlerResource {
	return &DependencyEventHandlerResource{
		dep,
		kv,
		domainProject,
	}
}

func isAddToLeft(centerNode *util.Node, addRes interface{}) bool {
	res := addRes.(*DependencyEventHandlerResource)
	compareRes := centerNode.Res.(*DependencyEventHandlerResource)
	if res.kv.ModRevision > compareRes.kv.ModRevision {
		return false
	}
	return true
}

func (h *DependencyEventHandler) Handle() error {
	key := core.GetServiceDependencyQueueRootKey("")
	resp, err := backend.Store().DependencyQueue().Search(context.Background(),
		registry.WithStrKey(key),
		registry.WithPrefix())
	if err != nil {
		return err
	}

	// maintain dependency rules.
	l := len(resp.Kvs)
	if l == 0 {
		return nil
	}

	dependencyTree := util.NewTree(isAddToLeft)

	cleanUpDomainProjects := make(map[string]struct{})
	defer h.CleanUp(cleanUpDomainProjects)

	for _, kv := range resp.Kvs {
		r := kv.Value.(*pb.ConsumerDependency)

		_, domainProject, uuid := core.GetInfoFromDependencyQueueKV(kv.Key)
		if uuid == core.DEPS_QUEUE_UUID {
			cleanUpDomainProjects[domainProject] = struct{}{}
		}
		res := NewDependencyEventHandlerResource(r, kv, domainProject)

		dependencyTree.AddNode(res)
	}

	return dependencyTree.InOrderTraversal(dependencyTree.GetRoot(), h.dependencyRuleHandle)
}

func (h *DependencyEventHandler) dependencyRuleHandle(res interface{}) error {
	ctx := context.WithValue(context.Background(), serviceUtil.CTX_GLOBAL, "1")
	dependencyEventHandlerRes := res.(*DependencyEventHandlerResource)
	r := dependencyEventHandlerRes.dep
	consumerFlag := util.StringJoin([]string{r.Consumer.Environment, r.Consumer.AppId, r.Consumer.ServiceName, r.Consumer.Version}, "/")

	domainProject := dependencyEventHandlerRes.domainProject
	consumerInfo := pb.DependenciesToKeys([]*pb.MicroServiceKey{r.Consumer}, domainProject)[0]
	providersInfo := pb.DependenciesToKeys(r.Providers, domainProject)

	var dep serviceUtil.Dependency
	var err error
	dep.DomainProject = domainProject
	dep.Consumer = consumerInfo
	dep.ProvidersRule = providersInfo
	if r.Override {
		err = serviceUtil.CreateDependencyRule(ctx, &dep)
	} else {
		err = serviceUtil.AddDependencyRule(ctx, &dep)
	}

	if err != nil {
		log.Errorf(err, "modify dependency rule failed, override: %t, consumer %s", r.Override, consumerFlag)
		return fmt.Errorf("override: %t, consumer is %s, %s", r.Override, consumerFlag, err.Error())
	}

	if err = h.removeKV(ctx, dependencyEventHandlerRes.kv); err != nil {
		log.Errorf(err, "remove dependency rule failed, override: %t, consumer %s", r.Override, consumerFlag)
		return err
	}

	log.Infof("maintain dependency [%v] successfully", r)
	return nil
}

func (h *DependencyEventHandler) removeKV(ctx context.Context, kv *discovery.KeyValue) error {
	dResp, err := backend.Registry().TxnWithCmp(ctx, []registry.PluginOp{registry.OpDel(registry.WithKey(kv.Key))},
		[]registry.CompareOp{registry.OpCmp(registry.CmpVer(kv.Key), registry.CMP_EQUAL, kv.Version)},
		nil)
	if err != nil {
		return fmt.Errorf("can not remove the dependency %s request, %s", util.BytesToStringWithNoCopy(kv.Key), err.Error())
	}
	if !dResp.Succeeded {
		log.Infof("the dependency %s request is changed", util.BytesToStringWithNoCopy(kv.Key))
	}
	return nil
}

func (h *DependencyEventHandler) CleanUp(domainProjects map[string]struct{}) {
	for domainProject := range domainProjects {
		ctx := context.WithValue(context.Background(), serviceUtil.CTX_GLOBAL, "1")
		if err := serviceUtil.CleanUpDependencyRules(ctx, domainProject); err != nil {
			log.Errorf(err, "clean up '%s' dependency rules failed", domainProject)
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
