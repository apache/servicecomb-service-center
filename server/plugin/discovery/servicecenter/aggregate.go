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
	"context"
	"crypto/tls"
	"strings"
	"sync"

	client2 "github.com/apache/servicecomb-service-center/client"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/model"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	mgr "github.com/apache/servicecomb-service-center/server/plugin"
	"github.com/apache/servicecomb-service-center/server/plugin/discovery"
	"github.com/apache/servicecomb-service-center/server/plugin/registry"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
)

var (
	scClient   *SCClientAggregate
	clientOnce sync.Once
	clientTLS  *tls.Config
)

type SCClientAggregate []*client2.Client

func getClientTLS() (*tls.Config, error) {
	if clientTLS != nil {
		return clientTLS, nil
	}
	var err error
	clientTLS, err = mgr.Plugins().TLS().ClientConfig()
	return clientTLS, err
}

func (c *SCClientAggregate) GetScCache(ctx context.Context) (*model.Cache, map[string]error) {
	var caches *model.Cache
	errs := make(map[string]error)
	for _, client := range *c {
		cache, err := client.GetScCache(ctx)
		if err != nil {
			errs[client.Cfg.Name] = err
			continue
		}

		if caches == nil {
			caches = &model.Cache{}
		}
		c.cacheAppend(client.Cfg.Name, &caches.Microservices, &cache.Microservices)
		c.cacheAppend(client.Cfg.Name, &caches.Indexes, &cache.Indexes)
		c.cacheAppend(client.Cfg.Name, &caches.Aliases, &cache.Aliases)
		c.cacheAppend(client.Cfg.Name, &caches.Tags, &cache.Tags)
		c.cacheAppend(client.Cfg.Name, &caches.Rules, &cache.Rules)
		c.cacheAppend(client.Cfg.Name, &caches.RuleIndexes, &cache.RuleIndexes)
		c.cacheAppend(client.Cfg.Name, &caches.DependencyRules, &cache.DependencyRules)
		c.cacheAppend(client.Cfg.Name, &caches.Summaries, &cache.Summaries)
		c.cacheAppend(client.Cfg.Name, &caches.Instances, &cache.Instances)
	}
	return caches, errs
}

func (c *SCClientAggregate) cacheAppend(name string, setter model.Setter, getter model.Getter) {
	getter.ForEach(func(_ int, v *model.KV) bool {
		if len(v.ClusterName) == 0 || v.ClusterName == registry.DefaultClusterName {
			v.ClusterName = name
		}
		setter.SetValue(v)
		return true
	})
}

func (c *SCClientAggregate) GetSchemasByServiceID(ctx context.Context, domainProject, serviceID string) (*discovery.Response, *scerr.Error) {
	dp := strings.Split(domainProject, "/")
	var response discovery.Response
	for _, client := range *c {
		schemas, err := client.GetSchemasByServiceID(ctx, dp[0], dp[1], serviceID)
		if err != nil && err.InternalError() {
			log.Errorf(err, "get schema by serviceID[%s/%s] failed", domainProject, serviceID)
			continue
		}
		if schemas == nil {
			continue
		}
		response.Count = int64(len(schemas))
		for _, schema := range schemas {
			response.Kvs = append(response.Kvs, &discovery.KeyValue{
				Key:         []byte(core.GenerateServiceSchemaKey(domainProject, serviceID, schema.SchemaId)),
				Value:       util.StringToBytesWithNoCopy(schema.Schema),
				ModRevision: 0,
				ClusterName: client.Cfg.Name,
			})
		}
		return &response, nil
	}
	return &response, nil
}

func (c *SCClientAggregate) GetSchemaBySchemaID(ctx context.Context, domainProject, serviceID, schemaID string) (*discovery.Response, *scerr.Error) {
	dp := strings.Split(domainProject, "/")
	var response discovery.Response
	for _, client := range *c {
		schema, err := client.GetSchemaBySchemaID(ctx, dp[0], dp[1], serviceID, schemaID)
		if err != nil && err.InternalError() {
			log.Errorf(err, "get schema by serviceID[%s/%s] failed", domainProject, serviceID)
			continue
		}
		if schema == nil {
			continue
		}
		response.Count = 1
		response.Kvs = append(response.Kvs, &discovery.KeyValue{
			Key:         []byte(core.GenerateServiceSchemaKey(domainProject, serviceID, schema.SchemaId)),
			Value:       util.StringToBytesWithNoCopy(schema.Schema),
			ModRevision: 0,
			ClusterName: client.Cfg.Name,
		})
		return &response, nil
	}
	return &response, nil
}

func (c *SCClientAggregate) GetInstancesByServiceID(ctx context.Context, domain, project, providerID, consumerID string) (*discovery.Response, *scerr.Error) {
	var response discovery.Response
	for _, client := range *c {
		insts, err := client.GetInstancesByServiceID(ctx, domain, project, providerID, consumerID)
		if err != nil && err.InternalError() {
			log.Errorf(err, "consumer[%s] get provider[%s/%s/%s] instances failed", consumerID, domain, project, providerID)
			continue
		}
		if insts == nil {
			continue
		}
		response.Count = int64(len(insts))
		for _, instance := range insts {
			response.Kvs = append(response.Kvs, &discovery.KeyValue{
				Key:         []byte(core.GenerateInstanceKey(domain+"/"+project, providerID, instance.InstanceId)),
				Value:       instance,
				ModRevision: 0,
				ClusterName: client.Cfg.Name,
			})
		}
	}
	return &response, nil
}

func (c *SCClientAggregate) GetInstanceByInstanceID(ctx context.Context, domain, project, providerID, instanceID, consumerID string) (*discovery.Response, *scerr.Error) {
	var response discovery.Response
	for _, client := range *c {
		instance, err := client.GetInstanceByInstanceID(ctx, domain, project, providerID, instanceID, consumerID)
		if err != nil && err.InternalError() {
			log.Errorf(err, "consumer[%s] get provider[%s/%s/%s] instances failed", consumerID, domain, project, providerID)
			continue
		}
		if instance == nil {
			continue
		}
		response.Count = 1
		response.Kvs = append(response.Kvs, &discovery.KeyValue{
			Key:         []byte(core.GenerateInstanceKey(domain+"/"+project, providerID, instance.InstanceId)),
			Value:       instance,
			ModRevision: 0,
			ClusterName: client.Cfg.Name,
		})
		return &response, nil
	}
	return &response, nil
}

func GetOrCreateSCClient() *SCClientAggregate {
	clientOnce.Do(func() {
		scClient = &SCClientAggregate{}
		clusters := registry.Configuration().Clusters
		for name, endpoints := range clusters {
			if len(name) == 0 || name == registry.Configuration().ClusterName {
				continue
			}
			client, err := client2.NewSCClient(client2.Config{Name: name, Endpoints: endpoints})
			if err != nil {
				log.Errorf(err, "new service center[%s]%v client failed", name, endpoints)
				continue
			}
			client.Timeout = registry.Configuration().RequestTimeOut
			// TLS
			if strings.Contains(endpoints[0], "https") {
				client.TLS, err = getClientTLS()
				if err != nil {
					log.Errorf(err, "get service center[%s]%v tls config failed", name, endpoints)
					continue
				}
			}

			*scClient = append(*scClient, client)
			log.Infof("new service center[%s]%v client", name, endpoints)
		}
	})
	return scClient
}
