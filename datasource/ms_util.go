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

package datasource

import (
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/go-chassis/cari/discovery"
)

var GlobalServiceNames = make(map[string]struct{})

type GetInstanceCountByDomainResponse struct {
	Err           error
	CountByDomain int64
}

func SetServiceDefaultValue(service *discovery.MicroService) {
	if len(service.AppId) == 0 {
		service.AppId = discovery.AppID
	}
	if len(service.Version) == 0 {
		service.Version = discovery.VERSION
	}
	if len(service.Level) == 0 {
		service.Level = "BACK"
	}
	if len(service.Status) == 0 {
		service.Status = discovery.MS_UP
	}
}

// SetStaticServices calculate the service/application num under a domainProject
func SetStaticServices(statistics *discovery.Statistics, svcKeys []*discovery.MicroServiceKey, svcIDs []string, withShared bool) map[string]string {
	l := len(svcKeys)
	app := make(map[string]struct{}, l)
	svcWithNonVersion := make(map[string]struct{}, l)
	svcIDToNonVerKey := make(map[string]string, l)
	for index, svc := range svcKeys {
		if !withShared && IsGlobal(svc) {
			continue
		}
		if _, ok := app[svc.AppId]; !ok {
			app[svc.AppId] = struct{}{}
		}
		svc.Version = ""
		svcWithNonVersionKey := generateServiceKey(svc)
		if _, ok := svcWithNonVersion[svcWithNonVersionKey]; !ok {
			svcWithNonVersion[svcWithNonVersionKey] = struct{}{}
		}
		svcIDToNonVerKey[svcIDs[index]] = svcWithNonVersionKey
	}
	statistics.Services.Count = int64(len(svcWithNonVersion))
	statistics.Apps.Count = int64(len(app))
	return svcIDToNonVerKey
}

// SetStaticInstances calculate the instance/onlineService num under a domainProject
func SetStaticInstances(statistics *discovery.Statistics, svcIDToNonVerKey map[string]string, instServiceIDs []string) {
	onlineServices := make(map[string]struct{}, len(instServiceIDs))
	for _, sid := range instServiceIDs {
		key, ok := svcIDToNonVerKey[sid]
		if !ok {
			continue
		}
		statistics.Instances.Count++
		if _, ok := onlineServices[key]; !ok {
			onlineServices[key] = struct{}{}
		}
	}
	statistics.Services.OnlineCount = int64(len(onlineServices))
}

func generateServiceKey(key *discovery.MicroServiceKey) string {
	return util.StringJoin([]string{
		key.Environment,
		key.AppId,
		key.ServiceName,
		key.Version,
	}, "/")
}

func TransServiceToKey(domainProject string, service *discovery.MicroService) *discovery.MicroServiceKey {
	return &discovery.MicroServiceKey{
		Tenant:      domainProject,
		Environment: service.Environment,
		AppId:       service.AppId,
		ServiceName: service.ServiceName,
		Version:     service.Version,
	}
}

func RegisterGlobalService(serviceName string) {
	GlobalServiceNames[serviceName] = struct{}{}
}

func IsGlobal(key *discovery.MicroServiceKey) bool {
	if !IsDefaultDomainProject(key.Tenant) {
		return false
	}
	if key.AppId != RegistryAppID {
		return false
	}
	_, ok := GlobalServiceNames[key.ServiceName]
	if !ok {
		_, ok = GlobalServiceNames[key.Alias]
	}
	return ok
}

func IsDefaultDomainProject(domainProject string) bool {
	return domainProject == RegistryDomainProject
}
