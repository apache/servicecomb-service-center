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

	"github.com/go-chassis/cari/discovery"

	"github.com/apache/servicecomb-service-center/datasource/mongo/client/dao"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client/model"
	mutil "github.com/apache/servicecomb-service-center/datasource/mongo/util"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

func Filter(ctx context.Context, rules []*model.Rule, consumerID string) (bool, error) {
	consumer, err := GetServiceByID(ctx, consumerID)
	if consumer == nil {
		return false, err
	}

	if len(rules) == 0 {
		return true, nil
	}
	domain := util.ParseDomainProject(ctx)
	project := util.ParseProject(ctx)
	filter := mutil.NewDomainProjectFilter(domain, project, mutil.ServiceServiceID(consumerID))
	tags, err := dao.GetTags(ctx, filter)
	if err != nil {
		return false, err
	}
	matchErr := MatchRules(rules, consumer.Service, tags)
	if matchErr != nil {
		if matchErr.Code == discovery.ErrPermissionDeny {
			return false, nil
		}
		return false, matchErr
	}
	return true, nil
}

func FilterAll(ctx context.Context, consumerIDs []string, rules []*model.Rule) (allow []string, deny []string, err error) {
	l := len(consumerIDs)
	if l == 0 || len(rules) == 0 {
		return consumerIDs, nil, nil
	}

	allowIdx, denyIdx := 0, l
	consumers := make([]string, l)
	for _, consumerID := range consumerIDs {
		ok, err := Filter(ctx, rules, consumerID)
		if err != nil {
			return nil, nil, err
		}
		if ok {
			consumers[allowIdx] = consumerID
			allowIdx++
		} else {
			denyIdx--
			consumers[denyIdx] = consumerID
		}
	}
	return consumers[:allowIdx], consumers[denyIdx:], nil
}
