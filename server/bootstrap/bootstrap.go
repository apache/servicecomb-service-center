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
	//cari db
	_ "github.com/go-chassis/cari/db/bootstrap"

	//eventbase
	_ "github.com/apache/servicecomb-service-center/eventbase/bootstrap"

	//etcd
	_ "github.com/apache/servicecomb-service-center/datasource/etcd/bootstrap"

	//mongo
	_ "github.com/apache/servicecomb-service-center/datasource/mongo/bootstrap"

	//local
	_ "github.com/apache/servicecomb-service-center/datasource/local/bootstrap"

	//rest v3 api
	_ "github.com/apache/servicecomb-service-center/server/rest/controller/v3"

	//rest v4 api
	_ "github.com/apache/servicecomb-service-center/server/resource"
	_ "github.com/apache/servicecomb-service-center/server/rest/controller/v4"

	//cipher
	_ "github.com/apache/servicecomb-service-center/server/plugin/security/cipher/buildin"

	//quota
	_ "github.com/apache/servicecomb-service-center/server/plugin/quota/buildin"

	//auth
	_ "github.com/apache/servicecomb-service-center/server/plugin/auth/buildin"

	//uuid
	_ "github.com/apache/servicecomb-service-center/server/plugin/uuid/buildin"
	_ "github.com/apache/servicecomb-service-center/server/plugin/uuid/context"

	//tracing
	_ "github.com/apache/servicecomb-service-center/server/plugin/tracing/pzipkin"

	//tlsconf
	_ "github.com/apache/servicecomb-service-center/server/plugin/security/tlsconf/buildin"

	//module 'admin'
	_ "github.com/apache/servicecomb-service-center/server/rest/admin"

	//governance
	_ "github.com/apache/servicecomb-service-center/server/service/grc/kie"

	//metrics
	_ "github.com/apache/servicecomb-service-center/server/rest/metrics"

	//jobs
	_ "github.com/apache/servicecomb-service-center/server/job/account"
	_ "github.com/apache/servicecomb-service-center/server/job/disco"

	//dlock
	_ "github.com/go-chassis/cari/dlock/bootstrap"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/handler/accesslog"
	"github.com/apache/servicecomb-service-center/server/handler/auth"
	"github.com/apache/servicecomb-service-center/server/handler/context"
	"github.com/apache/servicecomb-service-center/server/handler/exception"
	"github.com/apache/servicecomb-service-center/server/handler/maxbody"
	"github.com/apache/servicecomb-service-center/server/handler/metrics"
	"github.com/apache/servicecomb-service-center/server/handler/route"
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
	exception.RegisterHandlers()
	context.RegisterHandlers()
	accesslog.RegisterHandlers()
	maxbody.RegisterHandlers()
	auth.RegisterHandlers()
	metrics.RegisterHandlers()
	tracing.RegisterHandlers()
	route.RegisterHandlers()
}
