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

//rest v3 api
import (
	_ "github.com/apache/servicecomb-service-center/server/rest/controller/v3"

	// rest v4 api
	_ "github.com/apache/servicecomb-service-center/server/rest/controller/v4"

	//registry is buildin
	_ "github.com/apache/servicecomb-service-center/server/plugin/registry/buildin"

	//registry etcd
	_ "github.com/apache/servicecomb-service-center/server/plugin/registry/etcd"

	//registry etcd
	_ "github.com/apache/servicecomb-service-center/server/plugin/registry/embededetcd"

	//discovery
	_ "github.com/apache/servicecomb-service-center/server/plugin/discovery/aggregate"

	_ "github.com/apache/servicecomb-service-center/server/plugin/discovery/servicecenter"

	_ "github.com/apache/servicecomb-service-center/server/plugin/discovery/etcd"

	_ "github.com/apache/servicecomb-service-center/server/plugin/discovery/k8s"

	//cipher
	_ "github.com/apache/servicecomb-service-center/server/plugin/security/buildin"

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

	//tls
	_ "github.com/apache/servicecomb-service-center/server/plugin/tls/buildin"

	//module 'govern'
	_ "github.com/apache/servicecomb-service-center/server/rest/govern"

	//module 'broker'
	_ "github.com/apache/servicecomb-service-center/server/broker"

	//module 'admin'
	_ "github.com/apache/servicecomb-service-center/server/rest/admin"

	//metrics
	_ "github.com/apache/servicecomb-service-center/server/metrics/prometheus"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/handler/accesslog"
	"github.com/apache/servicecomb-service-center/server/handler/auth"
	"github.com/apache/servicecomb-service-center/server/handler/cache"
	"github.com/apache/servicecomb-service-center/server/handler/context"
	"github.com/apache/servicecomb-service-center/server/handler/maxbody"
	"github.com/apache/servicecomb-service-center/server/handler/metric"
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
	metric.RegisterHandlers()
	tracing.RegisterHandlers()
	auth.RegisterHandlers()
	context.RegisterHandlers()
	cache.RegisterHandlers()
}
