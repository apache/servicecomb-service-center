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
	"github.com/apache/servicecomb-service-center/datasource/etcd/client"
	"github.com/apache/servicecomb-service-center/datasource/etcd/kv"
	"github.com/apache/servicecomb-service-center/datasource/etcd/mux"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	serviceUtil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/pkg/backoff"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/queue"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/proto"
	"time"
)

const defaultEventHandleInterval = 5 * time.Minute

// DependencyEventHandler add or remove the service dependencies
// when user call find instance api or dependence operation api
type DependencyEventHandler struct {
	signals *queue.UniQueue
}

func (h *DependencyEventHandler) Type() sd.Type {
	return kv.DependencyQueue
}

func (h *DependencyEventHandler) OnEvent(evt sd.KvEvent) {
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
	lock, err := mux.Try(mux.DepQueueLock)
	if err != nil {
		log.Errorf(err, "try to lock %s failed", mux.DepQueueLock)
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
		log.Errorf(err, "handle dependency event failed")
		return h.backoff(backoff, retries), err
	}

	return 0, nil
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
				_, err := h.tryWithBackoff(h.Handle, h.notify, retries)
				if err != nil {
					log.Error("", err)
				}
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
	kv            *sd.KeyValue
	domainProject string
}

func NewDependencyEventHandlerResource(dep *pb.ConsumerDependency, kv *sd.KeyValue, domainProject string) *DependencyEventHandlerResource {
	return &DependencyEventHandlerResource{
		dep,
		kv,
		domainProject,
	}
}

func isAddToLeft(centerNode *util.Node, addRes interface{}) bool {
	res := addRes.(*DependencyEventHandlerResource)
	compareRes := centerNode.Res.(*DependencyEventHandlerResource)
	return res.kv.ModRevision <= compareRes.kv.ModRevision
}

func (h *DependencyEventHandler) Handle() error {
	key := core.GetServiceDependencyQueueRootKey("")
	resp, err := kv.Store().DependencyQueue().Search(context.Background(),
		client.WithStrKey(key),
		client.WithPrefix())
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
		if uuid == core.DepsQueueUUID {
			cleanUpDomainProjects[domainProject] = struct{}{}
		}
		res := NewDependencyEventHandlerResource(r, kv, domainProject)

		dependencyTree.AddNode(res)
	}

	return dependencyTree.InOrderTraversal(dependencyTree.GetRoot(), h.dependencyRuleHandle)
}

func (h *DependencyEventHandler) dependencyRuleHandle(res interface{}) error {
	ctx := context.WithValue(context.Background(), util.CtxGlobal, "1")
	dependencyEventHandlerRes := res.(*DependencyEventHandlerResource)
	r := dependencyEventHandlerRes.dep
	consumerFlag := util.StringJoin([]string{r.Consumer.Environment, r.Consumer.AppId, r.Consumer.ServiceName, r.Consumer.Version}, "/")

	domainProject := dependencyEventHandlerRes.domainProject
	consumerInfo := proto.DependenciesToKeys([]*pb.MicroServiceKey{r.Consumer}, domainProject)[0]
	providersInfo := proto.DependenciesToKeys(r.Providers, domainProject)

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

func (h *DependencyEventHandler) removeKV(ctx context.Context, kv *sd.KeyValue) error {
	dResp, err := client.Instance().TxnWithCmp(ctx, []client.PluginOp{client.OpDel(client.WithKey(kv.Key))},
		[]client.CompareOp{client.OpCmp(client.CmpVer(kv.Key), client.CmpEqual, kv.Version)},
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
		ctx := context.WithValue(context.Background(), util.CtxGlobal, "1")
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
