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

import _ "github.com/apache/incubator-servicecomb-service-center/server/core" // initialize

// rest
import _ "github.com/apache/incubator-servicecomb-service-center/server/rest/controller/v3"
import _ "github.com/apache/incubator-servicecomb-service-center/server/rest/controller/v4"

// registry
import _ "github.com/apache/incubator-servicecomb-service-center/server/plugin/infra/registry/etcd"
import _ "github.com/apache/incubator-servicecomb-service-center/server/plugin/infra/registry/embededetcd"

// cipher
import _ "github.com/apache/incubator-servicecomb-service-center/server/plugin/infra/security/buildin"

// quota
import _ "github.com/apache/incubator-servicecomb-service-center/server/plugin/infra/quota/buildin"
import _ "github.com/apache/incubator-servicecomb-service-center/server/plugin/infra/quota/unlimit"

// auth
import _ "github.com/apache/incubator-servicecomb-service-center/server/plugin/infra/auth/buildin"

// uuid
import _ "github.com/apache/incubator-servicecomb-service-center/server/plugin/infra/uuid/buildin"

// tracing
import _ "github.com/apache/incubator-servicecomb-service-center/server/plugin/infra/tracing/buildin"

// module
import _ "github.com/apache/incubator-servicecomb-service-center/server/govern"

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/handler/auth"
	"github.com/apache/incubator-servicecomb-service-center/server/handler/cache"
	"github.com/apache/incubator-servicecomb-service-center/server/handler/context"
	"github.com/apache/incubator-servicecomb-service-center/server/handler/metric"
	"github.com/apache/incubator-servicecomb-service-center/server/handler/tracing"
	"github.com/apache/incubator-servicecomb-service-center/server/interceptor"
	"github.com/apache/incubator-servicecomb-service-center/server/interceptor/access"
	"github.com/apache/incubator-servicecomb-service-center/server/interceptor/cors"
	"github.com/apache/incubator-servicecomb-service-center/server/interceptor/ratelimiter"
)

func init() {
	util.Logger().Info("BootStrap ServiceComb.io Edition")

	// intercept requests before routing.
	interceptor.RegisterInterceptFunc(access.Intercept)
	interceptor.RegisterInterceptFunc(ratelimiter.Intercept)
	interceptor.RegisterInterceptFunc(cors.Intercept)

	// handle requests after routing.
	metric.RegisterHandlers()
	tracing.RegisterHandlers()
	auth.RegisterHandlers()
	context.RegisterHandlers()
	cache.RegisterHandlers()
}
