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

package servicecenter

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"sync"

	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/go-chassis/etcdadpt"

	"github.com/apache/servicecomb-service-center/client"
	"github.com/apache/servicecomb-service-center/datasource/etcd"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/etcd/state/kvstore"
	"github.com/apache/servicecomb-service-center/pkg/dump"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/plugin/security/tlsconf"
)

var (
	scClient   *SCClientAggregate
	clientOnce sync.Once
	clientTLS  *tls.Config
)

type SCClientAggregate []*client.Client

func getClientTLS() (*tls.Config, error) {
	if clientTLS != nil {
		return clientTLS, nil
	}
	var err error
	clientTLS, err = tlsconf.ClientConfig()
	return clientTLS, err
}

func (c *SCClientAggregate) GetScCache(ctx context.Context) (*dump.Cache, map[string]error) {
	var caches *dump.Cache
	errs := make(map[string]error)
	for _, client := range *c {
		cache, err := client.GetScCache(ctx)
		if err != nil {
			errs[client.Cfg.Name] = err
			continue
		}

		if caches == nil {
			caches = &dump.Cache{}
		}
		c.cacheAppend(client.Cfg.Name, &caches.Microservices, &cache.Microservices)
		c.cacheAppend(client.Cfg.Name, &caches.Indexes, &cache.Indexes)
		c.cacheAppend(client.Cfg.Name, &caches.Aliases, &cache.Aliases)
		c.cacheAppend(client.Cfg.Name, &caches.Tags, &cache.Tags)
		c.cacheAppend(client.Cfg.Name, &caches.DependencyRules, &cache.DependencyRules)
		c.cacheAppend(client.Cfg.Name, &caches.Summaries, &cache.Summaries)
		c.cacheAppend(client.Cfg.Name, &caches.Instances, &cache.Instances)
	}
	return caches, errs
}

func (c *SCClientAggregate) cacheAppend(name string, setter dump.Setter, getter dump.Getter) {
	getter.ForEach(func(_ int, v *dump.KV) bool {
		if len(v.ClusterName) == 0 || v.ClusterName == etcdadpt.DefaultClusterName {
			v.ClusterName = name
		}
		setter.SetValue(v)
		return true
	})
}

func (c *SCClientAggregate) GetSchemasByServiceID(ctx context.Context, domainProject, serviceID string) (*kvstore.Response, *errsvc.Error) {
	dp := strings.Split(domainProject, "/")
	var response kvstore.Response
	for _, client := range *c {
		schemas, err := client.GetSchemasByServiceID(ctx, dp[0], dp[1], serviceID)
		if err != nil && err.InternalError() {
			log.Error(fmt.Sprintf("get schema by serviceID[%s/%s] failed", domainProject, serviceID), err)
			continue
		}
		if schemas == nil {
			continue
		}
		response.Count = int64(len(schemas))
		for _, schema := range schemas {
			response.Kvs = append(response.Kvs, &kvstore.KeyValue{
				Key:         []byte(path.GenerateServiceSchemaKey(domainProject, serviceID, schema.SchemaId)),
				Value:       util.StringToBytesWithNoCopy(schema.Schema),
				ModRevision: 0,
				ClusterName: client.Cfg.Name,
			})
		}
		return &response, nil
	}
	return &response, nil
}

func (c *SCClientAggregate) GetSchemaBySchemaID(ctx context.Context, domainProject, serviceID, schemaID string) (*kvstore.Response, *errsvc.Error) {
	dp := strings.Split(domainProject, "/")
	var response kvstore.Response
	for _, client := range *c {
		schema, err := client.GetSchemaBySchemaID(ctx, dp[0], dp[1], serviceID, schemaID)
		if err != nil && err.InternalError() {
			log.Error(fmt.Sprintf("get schema by serviceID[%s/%s] failed", domainProject, serviceID), err)
			continue
		}
		if schema == nil {
			continue
		}
		response.Count = 1
		response.Kvs = append(response.Kvs, &kvstore.KeyValue{
			Key:         []byte(path.GenerateServiceSchemaKey(domainProject, serviceID, schema.SchemaId)),
			Value:       util.StringToBytesWithNoCopy(schema.Schema),
			ModRevision: 0,
			ClusterName: client.Cfg.Name,
		})
		return &response, nil
	}
	return &response, nil
}

func (c *SCClientAggregate) GetInstancesByServiceID(ctx context.Context, domain, project, providerID, consumerID string) (*kvstore.Response, *errsvc.Error) {
	var response kvstore.Response
	for _, client := range *c {
		insts, err := client.GetInstancesByServiceID(ctx, domain, project, providerID, consumerID)
		if err != nil && err.InternalError() {
			log.Error(fmt.Sprintf("consumer[%s] get provider[%s/%s/%s] instances failed", consumerID, domain, project, providerID), err)
			continue
		}
		if insts == nil {
			continue
		}
		response.Count = int64(len(insts))
		for _, instance := range insts {
			response.Kvs = append(response.Kvs, &kvstore.KeyValue{
				Key:         []byte(path.GenerateInstanceKey(domain+"/"+project, providerID, instance.InstanceId)),
				Value:       instance,
				ModRevision: 0,
				ClusterName: client.Cfg.Name,
			})
		}
	}
	return &response, nil
}

func (c *SCClientAggregate) GetInstanceByInstanceID(ctx context.Context, domain, project, providerID, instanceID, consumerID string) (*kvstore.Response, *errsvc.Error) {
	var response kvstore.Response
	for _, client := range *c {
		instance, err := client.GetInstanceByInstanceID(ctx, domain, project, providerID, instanceID, consumerID)
		if err != nil && err.InternalError() {
			log.Error(fmt.Sprintf("consumer[%s] get provider[%s/%s/%s] instances failed", consumerID, domain, project, providerID), err)
			continue
		}
		if instance == nil {
			continue
		}
		response.Count = 1
		response.Kvs = append(response.Kvs, &kvstore.KeyValue{
			Key:         []byte(path.GenerateInstanceKey(domain+"/"+project, providerID, instance.InstanceId)),
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
		// TODO should not use the etcd config
		cfg := etcd.Configuration()
		clusters, err := etcdadpt.ListCluster(context.Background())
		if err != nil {
			log.Fatal("GetOrCreateSCClient failed", err)
		}
		for name, endpoints := range clusters {
			if len(name) == 0 || name == cfg.ClusterName {
				continue
			}
			client, err := client.NewSCClient(client.Config{Name: name, Endpoints: endpoints})
			if err != nil {
				log.Error(fmt.Sprintf("new service center[%s]%v client failed", name, endpoints), err)
				continue
			}
			client.Timeout = cfg.RequestTimeOut
			// TLS
			if strings.Contains(endpoints[0], "https") {
				client.TLS, err = getClientTLS()
				if err != nil {
					log.Error(fmt.Sprintf("get service center[%s]%v tls config failed", name, endpoints), err)
					continue
				}
			}

			*scClient = append(*scClient, client)
			log.Info(fmt.Sprintf("new service center[%s]%v client", name, endpoints))
		}
	})
	return scClient
}
