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
	"github.com/apache/servicecomb-service-center/client"
	etcdclient "github.com/apache/servicecomb-service-center/datasource/etcd/client"
	"github.com/apache/servicecomb-service-center/datasource/etcd/kv"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	scerr "github.com/apache/servicecomb-service-center/pkg/registry"
	"strings"

	"context"
)

// ClusterIndexer implements sd.Indexer.
// ClusterIndexer searches data from cache(firstly) and
// other service-centers(secondly).
type ClusterIndexer struct {
	*sd.CacheIndexer
	Client *SCClientAggregate
	Type   sd.Type
}

func (i *ClusterIndexer) Search(ctx context.Context, opts ...etcdclient.PluginOpOption) (resp *sd.Response, err error) {
	op := etcdclient.OpGet(opts...)

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

func (i *ClusterIndexer) search(ctx context.Context, opts ...etcdclient.PluginOpOption) (r *sd.Response, err error) {
	op := etcdclient.OpGet(opts...)
	key := util.BytesToStringWithNoCopy(op.Key)

	ctx = context.WithValue(ctx, client.QueryGlobal, "0")
	switch i.Type {
	case kv.SCHEMA:
		r, err = i.searchSchemas(ctx, op)
	case kv.INSTANCE:
		r, err = i.searchInstances(ctx, op)
	default:
		return &sd.Response{}, nil
	}
	log.Debugf("search '%s' match special options, request sc server, opts: %s", key, op)
	return
}

func (i *ClusterIndexer) searchSchemas(ctx context.Context, op etcdclient.PluginOp) (*sd.Response, error) {
	var (
		resp  *sd.Response
		scErr *scerr.Error
	)
	domainProject, serviceID, schemaID := core.GetInfoFromSchemaKV(op.Key)
	if op.Prefix && len(schemaID) == 0 {
		resp, scErr = i.Client.GetSchemasByServiceID(ctx, domainProject, serviceID)
	} else {
		resp, scErr = i.Client.GetSchemaBySchemaID(ctx, domainProject, serviceID, schemaID)
	}
	if scErr != nil {
		return nil, scErr
	}
	return resp, nil
}

func (i *ClusterIndexer) searchInstances(ctx context.Context, op etcdclient.PluginOp) (r *sd.Response, err error) {
	var (
		resp  *sd.Response
		scErr *scerr.Error
	)
	serviceID, instanceID, domainProject := core.GetInfoFromInstKV(op.Key)
	dp := strings.Split(domainProject, "/")
	if op.Prefix && len(instanceID) == 0 {
		resp, scErr = i.Client.GetInstancesByServiceID(ctx, dp[0], dp[1], serviceID, "")
	} else {
		resp, scErr = i.Client.GetInstanceByInstanceID(ctx, dp[0], dp[1], serviceID, instanceID, "")
	}
	if scErr != nil {
		return nil, scErr
	}
	return resp, nil
}

// Creditable implements sd.Indexer#Creditable.
// ClusterIndexer's search result's are not creditable as SCClientAggregate
// ignores sc clients' errors.
func (i *ClusterIndexer) Creditable() bool {
	return false
}

func NewClusterIndexer(t sd.Type, cache sd.Cache) *ClusterIndexer {
	return &ClusterIndexer{
		CacheIndexer: sd.NewCacheIndexer(cache),
		Client:       GetOrCreateSCClient(),
		Type:         t,
	}
}
