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
package disco_test

import (
	"context"

	_ "github.com/apache/servicecomb-service-center/test"

	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/syncer/config"
)

const (
	microserviceDomain  = "sync-microservice"
	microserviceProject = "sync-microservice"
	tagDomain           = "sync-tag"
	tagProject          = "sync-tag"
	schemaDomain        = "sync-schema"
	schemaProject       = "sync-schema"
	depDomain           = "sync-dependency"
	depProject          = "sync-dependency"
)

func getContext() context.Context {
	return util.WithNoCache(util.SetDomainProject(context.Background(), "default", "default"))
}

func microServiceGetContext() context.Context {
	ctx := util.WithNoCache(util.SetDomainProject(context.Background(), microserviceDomain,
		microserviceProject))
	return util.WithNoCache(util.SetContext(ctx, util.CtxEnableSync, "1"))
}

func schemaContext() context.Context {
	ctx := util.WithNoCache(util.SetDomainProject(context.Background(), schemaDomain, schemaProject))
	return util.WithNoCache(util.SetContext(ctx, util.CtxEnableSync, "1"))
}

func tagContext() context.Context {
	ctx := util.WithNoCache(util.SetDomainProject(context.Background(), tagDomain, tagProject))
	return util.WithNoCache(util.SetContext(ctx, util.CtxEnableSync, "1"))
}

func depContext() context.Context {
	ctx := util.WithNoCache(util.SetDomainProject(context.Background(), depDomain, depProject))
	return util.WithNoCache(util.SetContext(ctx, util.CtxEnableSync, "1"))
}

func initWhiteList() {
	cfg := config.Config{
		Sync: &config.Sync{
			EnableOnStart: true,
			WhiteList: &config.WhiteList{
				Service: &config.Service{
					Rules: []string{"sync*"},
				},
				Account: &config.Account{
					Rules: []string{"sync*"},
				},
				Role: &config.Role{
					Rules: []string{"sync*"},
				},
			},
		},
	}
	config.SetConfig(cfg)
}
