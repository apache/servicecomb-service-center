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
package backend

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"golang.org/x/net/context"
)

type Response struct {
	Kvs   []*KeyValue
	Count int64
}

func (pr *Response) MaxModRevision() (max int64) {
	for _, kv := range pr.Kvs {
		if max < kv.ModRevision {
			max = kv.ModRevision
		}
	}
	return
}

type Indexer interface {
	Cacher() Cacher
	Search(ctx context.Context, opts ...registry.PluginOpOption) (*Response, error)
	Run()
	Stop()
	Ready() <-chan struct{}
}

func NewIndexer(name string, cfg *Config) Indexer {
	switch {
	case core.ServerInfo.Config.EnableCache && cfg.InitSize > 0:
		return newCacheIndexer(name, cfg)
	default:
		util.Logger().Infof("core will not cache '%s' and ignore all events of it, cache enabled: %v, init size: %d",
			name, core.ServerInfo.Config.EnableCache, cfg.InitSize)
		return newBaseIndexer(cfg)
	}
}
