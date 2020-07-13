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

//Package test prepare service center required module before UT
package test

import (
	mgr "github.com/apache/servicecomb-service-center/server/plugin"
	"github.com/apache/servicecomb-service-center/server/plugin/discovery/etcd"
	etcd2 "github.com/apache/servicecomb-service-center/server/plugin/registry/etcd"
	plain "github.com/apache/servicecomb-service-center/server/plugin/security/buildin"
	"github.com/apache/servicecomb-service-center/server/plugin/tracing/pzipkin"
	"github.com/astaxie/beego"
)

func init() {
	beego.AppConfig.Set("registry_plugin", "etcd")
	mgr.RegisterPlugin(mgr.Plugin{mgr.REGISTRY, "etcd", etcd2.NewRegistry})
	mgr.RegisterPlugin(mgr.Plugin{mgr.DISCOVERY, "buildin", etcd.NewRepository})
	mgr.RegisterPlugin(mgr.Plugin{mgr.DISCOVERY, "etcd", etcd.NewRepository})
	mgr.RegisterPlugin(mgr.Plugin{mgr.CIPHER, "buildin", plain.New})
	mgr.RegisterPlugin(mgr.Plugin{mgr.TRACING, "buildin", pzipkin.New})
}
