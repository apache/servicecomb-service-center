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
	"crypto/tls"
	"github.com/apache/servicecomb-service-center/pkg/client/sc"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/admin/model"
	pb "github.com/apache/servicecomb-service-center/server/core/proto"
	scerr "github.com/apache/servicecomb-service-center/server/error"
	mgr "github.com/apache/servicecomb-service-center/server/plugin"
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/registry"
	"strings"
)

type SCClientAggregate []*sc.SCClient

var clientTLS *tls.Config

func getClientTLS() (*tls.Config, error) {
	if clientTLS != nil {
		return clientTLS, nil
	}
	var err error
	clientTLS, err = mgr.Plugins().TLS().ClientConfig()
	return clientTLS, err
}

func (c *SCClientAggregate) GetScCache() (*model.Cache, map[string]error) {
	var caches *model.Cache
	errs := make(map[string]error)
	for _, client := range *c {
		cache, err := client.GetScCache()
		if err != nil {
			errs[client.Cfg.Name] = err
			continue
		}

		if caches == nil {
			caches = &model.Cache{}
		}
		caches.Microservices = append(caches.Microservices, cache.Microservices...)
		caches.Indexes = append(caches.Indexes, cache.Indexes...)
		caches.Aliases = append(caches.Aliases, cache.Aliases...)
		caches.Tags = append(caches.Tags, cache.Tags...)
		caches.Rules = append(caches.Rules, cache.Rules...)
		caches.RuleIndexes = append(caches.RuleIndexes, cache.RuleIndexes...)
		caches.DependencyRules = append(caches.DependencyRules, cache.DependencyRules...)
		caches.Summaries = append(caches.Summaries, cache.Summaries...)
		caches.Instances = append(caches.Instances, cache.Instances...)
	}
	return caches, errs
}

func (c *SCClientAggregate) GetSchemasByServiceId(domainProject, serviceId string) ([]*pb.Schema, *scerr.Error) {
	var schemas []*pb.Schema
	for _, client := range *c {
		ss, err := client.GetSchemasByServiceId(domainProject, serviceId)
		if err != nil && err.InternalError() {
			log.Errorf(err, "get schema by serviceId[%s/%s] failed", domainProject, serviceId)
			continue
		}
		schemas = append(schemas, ss...)
	}

	return schemas, nil
}

func (c *SCClientAggregate) GetSchemaBySchemaId(domainProject, serviceId, schemaId string) (schema *pb.Schema, err *scerr.Error) {
	for _, client := range *c {
		schema, err = client.GetSchemaBySchemaId(domainProject, serviceId, schemaId)
		if err != nil && err.InternalError() {
			log.Errorf(err, "get schema by serviceId[%s/%s] failed", domainProject, serviceId)
			continue
		}
		if schema != nil {
			return
		}
	}

	return
}

func NewSCClientAggregate() *SCClientAggregate {
	c := &SCClientAggregate{}
	clusters := registry.Configuration().Clusters
	for name, endpoints := range clusters {
		if len(name) == 0 || name == registry.Configuration().ClusterName {
			continue
		}
		client, err := sc.NewSCClient(sc.Config{Name: name, Endpoints: endpoints})
		if err != nil {
			log.Errorf(err, "new service center[%s]%v client failed", name, endpoints)
			continue
		}
		client.Timeout = registry.Configuration().RequestTimeOut
		// TLS
		if strings.Index(endpoints[0], "https") >= 0 {
			client.TLS, err = getClientTLS()
			if err != nil {
				log.Errorf(err, "get service center[%s]%v tls config failed", name, endpoints)
				continue
			}
		}

		*c = append(*c, client)
		log.Infof("new service center[%s]%v client", name, endpoints)
	}
	return c
}
