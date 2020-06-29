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

package servicecenter

import (
	"github.com/apache/servicecomb-service-center/pkg/client/sc"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/backend"
	"github.com/apache/servicecomb-service-center/server/plugin/discovery"
	"github.com/apache/servicecomb-service-center/server/plugin/registry"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"

	"context"
)

// ClusterIndexer implements discovery.Indexer.
// ClusterIndexer searches data from cache(firstly) and
// other service-centers(secondly).
type ClusterIndexer struct {
	*discovery.CacheIndexer
	Client *SCClientAggregate
	Type   discovery.Type
}

func (i *ClusterIndexer) Search(ctx context.Context, opts ...registry.PluginOpOption) (resp *discovery.Response, err error) {
	op := registry.OpGet(opts...)

	if op.NoCache() {
		return i.search(ctx, opts...)
	}

	resp, err = i.CacheIndexer.Search(ctx, opts...)
	if err != nil {
		return
	}

	if resp.Count > 0 || op.CacheOnly() {
		return resp, nil
	}

	return i.search(ctx, opts...)
}

func (i *ClusterIndexer) search(ctx context.Context, opts ...registry.PluginOpOption) (r *discovery.Response, err error) {
	op := registry.OpGet(opts...)
	key := util.BytesToStringWithNoCopy(op.Key)

	ctx = context.WithValue(ctx, sc.QueryGlobal, "0")
	switch i.Type {
	case backend.SCHEMA:
		r, err = i.searchSchemas(ctx, op)
	case backend.INSTANCE:
		r, err = i.searchInstances(ctx, op)
	default:
		return &discovery.Response{}, nil
	}
	log.Debugf("search '%s' match special options, request sc server, opts: %s", key, op)
	return
}

func (i *ClusterIndexer) searchSchemas(ctx context.Context, op registry.PluginOp) (*discovery.Response, error) {
	var (
		resp  *discovery.Response
		scErr *scerr.Error
	)
	domainProject, serviceId, schemaId := core.GetInfoFromSchemaKV(op.Key)
	if op.Prefix && len(schemaId) == 0 {
		resp, scErr = i.Client.GetSchemasByServiceId(ctx, domainProject, serviceId)
	} else {
		resp, scErr = i.Client.GetSchemaBySchemaId(ctx, domainProject, serviceId, schemaId)
	}
	if scErr != nil {
		return nil, scErr
	}
	return resp, nil
}

func (i *ClusterIndexer) searchInstances(ctx context.Context, op registry.PluginOp) (r *discovery.Response, err error) {
	var (
		resp  *discovery.Response
		scErr *scerr.Error
	)
	serviceId, instanceId, domainProject := core.GetInfoFromInstKV(op.Key)
	if op.Prefix && len(instanceId) == 0 {
		resp, scErr = i.Client.GetInstancesByServiceId(ctx, domainProject, serviceId, "")
	} else {
		resp, scErr = i.Client.GetInstanceByInstanceId(ctx, domainProject, serviceId, instanceId, "")
	}
	if scErr != nil {
		return nil, scErr
	}
	return resp, nil
}

// Creditable implements discovery.Indexer.Creditable.
// ClusterIndexer's search result's are not creditable as SCClientAggregate
// ignores sc clients' errors.
func (i *ClusterIndexer) Creditable() bool {
	return false
}

func NewClusterIndexer(t discovery.Type, cache discovery.Cache) *ClusterIndexer {
	return &ClusterIndexer{
		CacheIndexer: discovery.NewCacheIndexer(cache),
		Client:       GetOrCreateSCClient(),
		Type:         t,
	}
}
