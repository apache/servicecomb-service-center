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

package dao

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-chassis/cari/discovery"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	mutil "github.com/apache/servicecomb-service-center/datasource/mongo/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/pkg/validate"
)

const (
	Provider = "p"
	Consumer = "c"
)

func GetProviderDeps(ctx context.Context, provider *discovery.MicroService) ([]*discovery.MicroServiceKey, error) {
	depRules, err := getServiceDepRules(ctx, Provider, provider)
	if err != nil {
		log.Error("", err)
		return nil, err
	}
	return serviceVersionFilter(ctx, depRules, provider)

}

func getServiceDepRules(ctx context.Context, ruleType string, provider *discovery.MicroService) ([]*DependencyRule, error) {
	filter := bson.D{
		{ColumnServiceType, ruleType},
		{mutil.ConnectWithDot([]string{ColumnServiceKey, ColumnTenant}), util.ParseDomainProject(ctx)},
		{mutil.ConnectWithDot([]string{ColumnServiceKey, ColumnAppID}), provider.AppId},
		{mutil.ConnectWithDot([]string{ColumnServiceKey, ColumnServiceName}), provider.ServiceName},
	}
	return getDependencyRules(ctx, filter)
}

func serviceVersionFilter(ctx context.Context, depRules []*DependencyRule, service *discovery.MicroService) ([]*discovery.MicroServiceKey, error) {
	var allServices []*discovery.MicroServiceKey
	var latestServiceID []string
	var err error
	serviceKey := discovery.MicroServiceToKey(util.ParseDomainProject(ctx), service)
	for _, depRule := range depRules {
		providerVersionRule := depRule.ServiceKey.Version
		if providerVersionRule == "latest" {
			if latestServiceID == nil {
				latestServiceID, err = GetServiceIDs(ctx, providerVersionRule, serviceKey)
				if err != nil {
					log.Error(fmt.Sprintf("get service[%s/%s/%s/%s]'s serviceID failed",
						service.Environment, service.AppId, service.ServiceName, providerVersionRule), err)
					return nil, err
				}
			}
			if len(latestServiceID) == 0 {
				log.Info(fmt.Sprintf("service[%s/%s/%s/%s] does not exist",
					service.Environment, service.AppId, service.ServiceName, providerVersionRule))
				continue
			}
			if service.ServiceId != latestServiceID[0] {
				continue
			}
		} else {
			if !versionMatchRule(service.Version, providerVersionRule) {
				continue
			}
		}
		if len(depRule.Dep.Dependency) > 0 {
			allServices = append(allServices, depRule.Dep.Dependency...)
		}
	}
	return allServices, nil
}

// not prepare for latest scene, should merge it with find serviceids func.
func versionMatchRule(version, versionRule string) bool {
	if len(versionRule) == 0 {
		return false
	}
	rangeIdx := strings.Index(versionRule, "-")
	versionInt, _ := validate.VersionToInt64(version)
	switch {
	case versionRule[len(versionRule)-1:] == "+":
		start, _ := validate.VersionToInt64(versionRule[:len(versionRule)-1])
		return versionInt >= start
	case rangeIdx > 0:
		start, _ := validate.VersionToInt64(versionRule[:rangeIdx])
		end, _ := validate.VersionToInt64(versionRule[rangeIdx+1:])
		return versionInt >= start && versionInt < end
	default:
		return version == versionRule
	}
}

func getDependencyRules(ctx context.Context, filter interface{}) ([]*DependencyRule, error) {
	findRes, err := client.GetMongoClient().Find(ctx, CollectionDep, filter)
	if err != nil {
		return nil, err
	}
	defer findRes.Close(ctx)
	var depRules []*DependencyRule
	for findRes.Next(ctx) {
		var tmp *DependencyRule
		err := findRes.Decode(&tmp)
		if err != nil {
			return nil, err
		}
		depRules = append(depRules, tmp)
	}
	return depRules, nil
}
