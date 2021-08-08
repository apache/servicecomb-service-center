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
	"fmt"

	pb "github.com/go-chassis/cari/discovery"

	"github.com/apache/servicecomb-service-center/datasource/etcd/cache"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/datasource/etcd/state/kvstore"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

// DependencyRuleEventHandler reset the find instances cache
// when provider dependency rule is changed
type DependencyRuleEventHandler struct {
}

func (h *DependencyRuleEventHandler) Type() kvstore.Type {
	return sd.TypeDependencyRule
}

func (h *DependencyRuleEventHandler) OnEvent(evt kvstore.Event) {
	action := evt.Type
	if action != pb.EVT_UPDATE && action != pb.EVT_DELETE {
		return
	}
	t, providerKey := path.GetInfoFromDependencyRuleKV(evt.KV.Key)
	if t != path.DepsProvider {
		return
	}
	log.Debug(fmt.Sprintf("caught [%s] provider rule[%s/%s/%s/%s] event",
		action, providerKey.Environment, providerKey.AppId, providerKey.ServiceName, providerKey.Version))
	cache.DependencyRule.Remove(providerKey)
}

func NewDependencyRuleEventHandler() *DependencyRuleEventHandler {
	return &DependencyRuleEventHandler{}
}
