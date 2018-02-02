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
	"encoding/json"
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/apache/incubator-servicecomb-service-center/server/core/backend"
	"github.com/apache/incubator-servicecomb-service-center/server/core/backend/store"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"github.com/apache/incubator-servicecomb-service-center/server/mux"
	serviceUtil "github.com/apache/incubator-servicecomb-service-center/server/service/util"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
	"time"
)

type DependencyEventHandler struct {
	signals *util.UniQueue
}

func (h *DependencyEventHandler) Type() store.StoreType {
	return store.DEPENDENCY_QUEUE
}

func (h *DependencyEventHandler) OnEvent(evt *store.KvEvent) {
	action := evt.Action
	if action != pb.EVT_CREATE && action != pb.EVT_UPDATE && action != pb.EVT_INIT {
		return
	}

	h.signals.Put(context.Background(), struct{}{})
}

func (h *DependencyEventHandler) loop() {
	util.Go(func(stopCh <-chan struct{}) {
		waitDelayIndex := 0
		waitDelay := []int{1, 1, 5, 10, 20, 30, 60}
		retry := func() {
			if waitDelayIndex >= len(waitDelay) {
				waitDelayIndex = 0
			}
			<-time.After(time.Duration(waitDelay[waitDelayIndex]) * time.Second)
			waitDelayIndex++

			h.signals.Put(context.Background(), struct{}{})
		}
		for {
			select {
			case <-stopCh:
				return
			case <-h.signals.Chan():
				lock, err := mux.Try(mux.DEP_QUEUE_LOCK)
				if err != nil {
					util.Logger().Errorf(err, "try to lock %s failed", mux.DEP_QUEUE_LOCK)
					retry()
					continue
				}

				if lock == nil {
					continue
				}

				err = h.Handle()
				lock.Unlock()
				if err != nil {
					util.Logger().Errorf(err, "handle dependency event failed")
					retry()
					continue
				}
			}
		}
	})
}

type DependencyEventHandlerResource struct {
	dep           *pb.ConsumerDependency
	kv            *mvccpb.KeyValue
	domainProject string
}

func NewDependencyEventHandlerResource(dep *pb.ConsumerDependency, kv *mvccpb.KeyValue, domainProject string) *DependencyEventHandlerResource {
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
	resp, err := store.Store().DependencyQueue().Search(context.Background(),
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

	ctx := context.Background()

	dependencyTree := util.NewTree(isAddToLeft)

	for _, kv := range resp.Kvs {
		r := &pb.ConsumerDependency{}
		consumerId, domainProject, data := pb.GetInfoFromDependencyQueueKV(kv)

		err := json.Unmarshal(data, r)
		if err != nil {
			util.Logger().Errorf(err, "maintain dependency failed, unmarshal failed, consumer %s dependency: %s",
				consumerId, util.BytesToStringWithNoCopy(data))

			if err = h.removeKV(ctx, kv); err != nil {
				return err
			}
			continue
		}

		res := NewDependencyEventHandlerResource(r, kv, domainProject)

		dependencyTree.AddNode(res)
	}

	return dependencyTree.InOrderTraversal(dependencyTree.GetRoot(), h.dependencyRuleHandle)
}

func (h *DependencyEventHandler) dependencyRuleHandle(res interface{}) error {
	ctx := context.Background()
	dependencyEventHandlerRes := res.(*DependencyEventHandlerResource)
	r := dependencyEventHandlerRes.dep
	consumerFlag := util.StringJoin([]string{r.Consumer.AppId, r.Consumer.ServiceName, r.Consumer.Version}, "/")


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
		util.Logger().Errorf(err, "modify dependency rule failed, override: %t, consumer %s", r.Override, consumerFlag)
		return fmt.Errorf("override: %t, consumer is %s, %s", r.Override, consumerFlag, err.Error())
	}

	if err = h.removeKV(ctx, dependencyEventHandlerRes.kv); err != nil {
		util.Logger().Errorf(err, "remove dependency rule failed, override: %t, consumer %s", r.Override, consumerFlag)
		return err
	}

	util.Logger().Infof("maintain dependency %v successfully, override: %t", r, r.Override)
	return nil
}

func (h *DependencyEventHandler) removeKV(ctx context.Context, kv *mvccpb.KeyValue) error {
	dResp, err := backend.Registry().TxnWithCmp(ctx, []registry.PluginOp{registry.OpDel(registry.WithKey(kv.Key))},
		[]registry.CompareOp{registry.OpCmp(registry.CmpVer(kv.Key), registry.CMP_EQUAL, kv.Version)},
		nil)
	if err != nil {
		return fmt.Errorf("can not remove the dependency %s request, %s", util.BytesToStringWithNoCopy(kv.Key), err.Error())
	}
	if !dResp.Succeeded {
		util.Logger().Infof("the dependency %s request is changed", util.BytesToStringWithNoCopy(kv.Key))
	}
	return nil
}

func NewDependencyEventHandler() *DependencyEventHandler {
	h := &DependencyEventHandler{
		signals: util.NewUniQueue(),
	}
	h.loop()
	return h
}
