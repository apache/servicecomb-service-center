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

package etcd

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/log"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/discovery"
	mgr "github.com/apache/incubator-servicecomb-service-center/server/plugin"
)

const etcd = "etcd"

func init() {
	mgr.RegisterPlugin(mgr.Plugin{mgr.DISCOVERY, "buildin", NewRepository})
}

type KvEntityRepository struct {
}

func (r *KvEntityRepository) NewEntity(t discovery.StoreType, cfg *discovery.Config) discovery.Entity {
	if cfg == nil {
		// do not new instance
		log.Warnf("'%s' config is nil, new default entity", t)
		return DefaultKvEntity()
	}
	e := NewKvEntity(t.String(), cfg)
	e.Run()
	return e
}

func NewRepository() mgr.PluginInstance {
	return &KvEntityRepository{}
}
