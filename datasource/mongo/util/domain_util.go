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
	"strings"

	pb "github.com/go-chassis/cari/discovery"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/db"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

func GetAllServicesAcrossDomainProject(ctx context.Context) (map[string][]*pb.MicroService, error) {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)

	filter := bson.M{"domain": domain, "project": project}

	findRes, err := client.GetMongoClient().Find(ctx, db.CollectionService, filter)
	if err != nil {
		return nil, err
	}

	services := make(map[string][]*pb.MicroService)

	for findRes.Next(ctx) {
		var mongoService db.Service
		err := findRes.Decode(&mongoService)
		if err != nil {
			return nil, err
		}
		domainProject := mongoService.Domain + "/" + mongoService.Project
		services[domainProject] = append(services[domainProject], mongoService.Service)
	}
	return services, nil
}

func CtxFromDomainProject(pCtx context.Context, domainProject string) (ctx context.Context, err error) {
	splitIndex := strings.Index(domainProject, path.SPLIT)
	if splitIndex == -1 {
		return nil, NewError("invalid domainProject: ", domainProject)
	}
	domain := domainProject[:splitIndex]
	project := domainProject[splitIndex+1:]
	return util.SetDomainProject(pCtx, domain, project), nil
}
