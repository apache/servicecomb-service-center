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

package mongo

import (
	"context"
	"fmt"

	pb "github.com/go-chassis/cari/discovery"

	"github.com/apache/servicecomb-service-center/datasource/mongo/db"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

func GetAllConsumerIds(ctx context.Context, provider *pb.MicroService) (allow []string, deny []string, _ error) {
	if provider == nil || len(provider.ServiceId) == 0 {
		return nil, nil, fmt.Errorf("invalid provider")
	}

	//todo 删除服务，最后实例推送有误差
	domain := util.ParseDomainProject(ctx)
	project := util.ParseProject(ctx)
	providerRules, err := getRulesUtil(ctx, domain, project, provider.ServiceId)
	if err != nil {
		return nil, nil, err
	}

	allow, deny, err = GetConsumerIDsWithFilter(ctx, provider, providerRules)
	if err != nil {
		return nil, nil, err
	}
	return allow, deny, nil
}

func GetConsumerIDsWithFilter(ctx context.Context, provider *pb.MicroService, rules []*db.Rule) (allow []string, deny []string, err error) {
	domainProject := util.ParseDomainProject(ctx)
	dr := NewProviderDependencyRelation(ctx, domainProject, provider)
	consumerIDs, err := dr.GetDependencyConsumerIds()
	if err != nil {
		log.Error(fmt.Sprintf("get service[%s]'s consumerIds failed", provider.ServiceId), err)
		return nil, nil, err
	}
	return FilterAll(ctx, consumerIDs, rules)
}
