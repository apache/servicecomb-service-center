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
package bootstrap

import (
	//etcd
	_ "github.com/apache/servicecomb-service-center/datasource/etcd/bootstrap"

	//mongo
	_ "github.com/apache/servicecomb-service-center/datasource/mongo/bootstrap"

	//rest v3 api
	_ "github.com/apache/servicecomb-service-center/server/rest/controller/v3"

	//rest v4 api
	_ "github.com/apache/servicecomb-service-center/server/rest/controller/v4"

	_ "github.com/apache/servicecomb-service-center/server/resource"

	//cipher
	_ "github.com/apache/servicecomb-service-center/server/plugin/security/cipher/buildin"

	//quota
	_ "github.com/apache/servicecomb-service-center/server/plugin/quota/buildin"

	_ "github.com/apache/servicecomb-service-center/server/plugin/quota/unlimit"

	//auth
	_ "github.com/apache/servicecomb-service-center/server/plugin/auth/buildin"

	//uuid
	_ "github.com/apache/servicecomb-service-center/server/plugin/uuid/buildin"

	_ "github.com/apache/servicecomb-service-center/server/plugin/uuid/context"

	//tracing
	_ "github.com/apache/servicecomb-service-center/server/plugin/tracing/pzipkin"

	//tlsconf
	_ "github.com/apache/servicecomb-service-center/server/plugin/security/tlsconf/buildin"

	//module 'govern'
	_ "github.com/apache/servicecomb-service-center/server/rest/govern"

	//module 'admin'
	_ "github.com/apache/servicecomb-service-center/server/rest/admin"

	//module 'syncer'
	_ "github.com/apache/servicecomb-service-center/server/rest/syncer"

	//governance
	_ "github.com/apache/servicecomb-service-center/server/service/gov/kie"

	//metrics
	_ "github.com/apache/servicecomb-service-center/server/rest/prometheus"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/broker"
	"github.com/apache/servicecomb-service-center/server/handler/accesslog"
	"github.com/apache/servicecomb-service-center/server/handler/auth"
	"github.com/apache/servicecomb-service-center/server/handler/cache"
	"github.com/apache/servicecomb-service-center/server/handler/context"
	"github.com/apache/servicecomb-service-center/server/handler/maxbody"
	"github.com/apache/servicecomb-service-center/server/handler/metrics"
	"github.com/apache/servicecomb-service-center/server/handler/tracing"
	"github.com/apache/servicecomb-service-center/server/interceptor"
	"github.com/apache/servicecomb-service-center/server/interceptor/access"
	"github.com/apache/servicecomb-service-center/server/interceptor/cors"
)

func init() {
	log.Info("BootStrap ServiceComb.io Edition")

	// intercept requests before routing.
	interceptor.RegisterInterceptFunc(access.Intercept)
	interceptor.RegisterInterceptFunc(cors.Intercept)

	// handle requests after routing.
	accesslog.RegisterHandlers()
	maxbody.RegisterHandlers()
	metrics.RegisterHandlers()
	tracing.RegisterHandlers()
	auth.RegisterHandlers()
	context.RegisterHandlers()
	cache.RegisterHandlers()

	// init broker
	broker.Init()
}
