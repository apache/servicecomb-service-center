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
	"github.com/apache/incubator-servicecomb-service-center/server/core/backend/store"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"github.com/apache/incubator-servicecomb-service-center/server/mux"
	serviceUtil "github.com/apache/incubator-servicecomb-service-center/server/service/util"
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
	if action != pb.EVT_CREATE && action != pb.EVT_INIT {
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

				err = h.handle()
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

func (h *DependencyEventHandler) handle() error {
	key := core.GetServiceDependencyQueueRootKey("")
	resp, err := store.Store().DependencyQueue().Search(context.Background(),
		registry.WithStrKey(key),
		registry.WithPrefix())
	if err != nil {
		return err
	}

	// maintain dependency rules.
	for _, kv := range resp.Kvs {
		var (
			r   *pb.ConsumerDependency
			ctx context.Context = context.Background()
		)

		consumerId, domainProject, data := pb.GetInfoFromDependencyQueueKV(kv)

		err := json.Unmarshal(data, r)
		if err != nil {
			return fmt.Errorf("unmarshal consumer %s dependency failed, %s", consumerId, err.Error())
		}

		serviceUtil.SetDependencyDefaultValue(r)

		consumerFlag := util.StringJoin([]string{r.Consumer.AppId, r.Consumer.ServiceName, r.Consumer.Version}, "/")
		consumerInfo := pb.DependenciesToKeys([]*pb.DependencyKey{r.Consumer}, domainProject)[0]
		providersInfo := pb.DependenciesToKeys(r.Providers, domainProject)

		consumerId, err = serviceUtil.GetServiceId(ctx, consumerInfo)
		if err != nil {
			return fmt.Errorf("get consumer %s id failed, override: %t, %s", consumerFlag, r.Override, err.Error())
		}
		if len(consumerId) == 0 {
			util.Logger().Errorf(nil, "maintain dependency failed, override: %t, consumer %s does not exist",
				r.Override, consumerFlag)
			return nil
		}

		var dep serviceUtil.Dependency
		dep.DomainProject = domainProject
		dep.Consumer = consumerInfo
		dep.ProvidersRule = providersInfo
		dep.ConsumerId = consumerId
		if r.Override {
			err = serviceUtil.CreateDependencyRule(ctx, &dep)
		} else {
			err = serviceUtil.AddDependencyRule(ctx, &dep)
		}

		if err != nil {
			return fmt.Errorf("override: %t, consumer is %s, %s", r.Override, consumerFlag, err.Error())
		}

		util.Logger().Infof("maintain dependency %+v successfully", r)
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
