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

package util

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/db"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

func GetTags(ctx context.Context, domain string, project string, serviceID string) (tags map[string]string, err error) {
	filter := bson.M{
		db.ColumnDomain:  domain,
		db.ColumnProject: project,
		StringBuilder([]string{db.ColumnService, db.ColumnServiceID}): serviceID,
	}
	result, err := client.GetMongoClient().FindOne(ctx, db.CollectionService, filter)
	if err != nil {
		return nil, err
	}
	if result.Err() != nil {
		return nil, result.Err()
	}
	var service db.Service
	err = result.Decode(&service)
	if err != nil {
		log.Error("type conversion error", err)
		return nil, err
	}
	return service.Tags, nil
}

func FilterServicesByTags(services []*db.Service, tags []string) []*db.Service {
	if len(tags) == 0 {
		return services
	}
	var newServices []*db.Service
	for _, service := range services {
		index := 0
		for ; index < len(tags); index++ {
			if _, ok := service.Tags[tags[index]]; !ok {
				break
			}
		}
		if index == len(tags) {
			newServices = append(newServices, service)
		}
	}
	return newServices
}
