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

// rest
import _ "github.com/apache/incubator-servicecomb-service-center/server/rest/controller/v3"
import _ "github.com/apache/incubator-servicecomb-service-center/server/rest/controller/v4"

// registry
import _ "github.com/apache/incubator-servicecomb-service-center/server/plugin/pkg/registry/buildin"
import _ "github.com/apache/incubator-servicecomb-service-center/server/plugin/pkg/registry/etcd"
import _ "github.com/apache/incubator-servicecomb-service-center/server/plugin/pkg/registry/embededetcd"

// discovery
import _ "github.com/apache/incubator-servicecomb-service-center/server/plugin/pkg/discovery/aggregate"
import _ "github.com/apache/incubator-servicecomb-service-center/server/plugin/pkg/discovery/sc"
import _ "github.com/apache/incubator-servicecomb-service-center/server/plugin/pkg/discovery/etcd"
import _ "github.com/apache/incubator-servicecomb-service-center/server/plugin/pkg/discovery/k8s"

// cipher
import _ "github.com/apache/incubator-servicecomb-service-center/server/plugin/pkg/security/buildin"

// quota
import _ "github.com/apache/incubator-servicecomb-service-center/server/plugin/pkg/quota/buildin"
import _ "github.com/apache/incubator-servicecomb-service-center/server/plugin/pkg/quota/unlimit"

// auth
import _ "github.com/apache/incubator-servicecomb-service-center/server/plugin/pkg/auth/buildin"

// uuid
import _ "github.com/apache/incubator-servicecomb-service-center/server/plugin/pkg/uuid/buildin"

// tracing
import _ "github.com/apache/incubator-servicecomb-service-center/server/plugin/pkg/tracing/buildin"

// tls
import _ "github.com/apache/incubator-servicecomb-service-center/server/plugin/pkg/tls/buildin"

// module 'govern'
import _ "github.com/apache/incubator-servicecomb-service-center/server/govern"

// module 'broker'
import _ "github.com/apache/incubator-servicecomb-service-center/server/broker"

// module 'admin'
import _ "github.com/apache/incubator-servicecomb-service-center/server/admin"

// metrics
import _ "github.com/apache/incubator-servicecomb-service-center/server/metric"

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/log"
	"github.com/apache/incubator-servicecomb-service-center/server/handler/auth"
	"github.com/apache/incubator-servicecomb-service-center/server/handler/cache"
	"github.com/apache/incubator-servicecomb-service-center/server/handler/context"
	"github.com/apache/incubator-servicecomb-service-center/server/handler/maxbody"
	"github.com/apache/incubator-servicecomb-service-center/server/handler/metric"
	"github.com/apache/incubator-servicecomb-service-center/server/handler/tracing"
	"github.com/apache/incubator-servicecomb-service-center/server/interceptor"
	"github.com/apache/incubator-servicecomb-service-center/server/interceptor/access"
	"github.com/apache/incubator-servicecomb-service-center/server/interceptor/cors"
)

func init() {
	log.Info("BootStrap ServiceComb.io Edition")

	// intercept requests before routing.
	interceptor.RegisterInterceptFunc(access.Intercept)
	interceptor.RegisterInterceptFunc(cors.Intercept)

	// handle requests after routing.
	maxbody.RegisterHandlers()
	metric.RegisterHandlers()
	tracing.RegisterHandlers()
	auth.RegisterHandlers()
	context.RegisterHandlers()
	cache.RegisterHandlers()
}
