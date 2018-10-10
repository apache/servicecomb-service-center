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

package aggregate

import (
	mgr "github.com/apache/incubator-servicecomb-service-center/server/plugin"
	"github.com/apache/incubator-servicecomb-service-center/server/plugin/pkg/discovery"
	"github.com/apache/incubator-servicecomb-service-center/server/plugin/pkg/registry"
	"golang.org/x/net/context"
)

// Aggregator is a discovery service adaptor implement of one kubernetes cluster
type Aggregator []discovery.Adaptor

func (as *Aggregator) Search(ctx context.Context, opts ...registry.PluginOpOption) (*discovery.Response, error) {
	var response discovery.Response
	for _, a := range *as {
		resp, err := a.Search(ctx, opts...)
		if err != nil {
			continue
		}
		response.Kvs = append(response.Kvs, resp.Kvs...)
		response.Count += resp.Count
	}
	return &response, nil
}

func (as *Aggregator) Cache() discovery.Cache {
	return discovery.NullCache
}

func (as *Aggregator) Run() {
	for _, a := range *as {
		a.Run()
	}
}

func (as *Aggregator) Stop() {
	for _, a := range *as {
		a.Stop()
	}
}

func (as *Aggregator) Ready() <-chan struct{} {
	for _, a := range *as {
		<-a.Ready()
	}
	return closedCh
}

func NewAggregator(t discovery.Type, cfg *discovery.Config) (as *Aggregator) {
	as = &Aggregator{}
	for _, name := range repos {
		repo := mgr.Plugins().Get(mgr.DISCOVERY, name).New().(discovery.AdaptorRepository)
		*as = append(*as, repo.New(t, cfg))
	}
	return as
}
