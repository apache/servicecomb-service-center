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

// Package test prepare service center required module before UT
package test

import (
	"context"
	"time"

	_ "github.com/apache/servicecomb-service-center/server/init"

	_ "github.com/apache/servicecomb-service-center/syncer/bootstrap"

	_ "github.com/apache/servicecomb-service-center/server/bootstrap"

	"github.com/apache/servicecomb-service-center/datasource"
	edatasource "github.com/apache/servicecomb-service-center/eventbase/datasource"
	"github.com/apache/servicecomb-service-center/server/metrics"
	"github.com/apache/servicecomb-service-center/server/service/registry"
	"github.com/go-chassis/cari/db"
	"github.com/go-chassis/cari/db/config"
	"github.com/go-chassis/go-archaius"
)

func init() {
	var kind = "etcd"
	var uri = "http://127.0.0.1:2379"
	_ = archaius.Set("rbac.releaseLockAfter", "3s")
	_ = archaius.Set("registry.instance.properties.engineID", "test_engineID")
	_ = archaius.Set("registry.instance.properties.engineName", "test_engineName")
	if IsETCD() {
		_ = archaius.Set("registry.cache.mode", 0)
		_ = archaius.Set("discovery.kind", "etcd")
		_ = archaius.Set("registry.kind", "etcd")
		_ = archaius.Set("registry.etcd.cluster.name", "sc-0")
		_ = archaius.Set("registry.etcd.cluster.endpoints", "sc-0="+uri+",sc-1=http://127.0.0.2:2379")
	} else if IsLOCAL() {
		_ = archaius.Set("registry.cache.mode", 0)
		_ = archaius.Set("discovery.kind", "etcd")
		_ = archaius.Set("registry.kind", "etcd")
		_ = archaius.Set("registry.etcd.cluster.name", "sc-0")
		_ = archaius.Set("registry.etcd.cluster.endpoints", "sc-0="+uri+",sc-1=http://127.0.0.2:2379")
		_ = archaius.Set("schema.root.path", "/data/schemas")
	} else {
		_ = archaius.Set("registry.heartbeat.kind", "checker")
		kind = "mongo"
		uri = "mongodb://127.0.0.1:27017"
	}

	_ = datasource.Init(datasource.Options{
		Kind: kind,
	})
	_ = metrics.Init(metrics.Options{})

	_ = db.Init(&config.Config{
		Kind:    kind,
		URI:     uri,
		Timeout: 10 * time.Second,
	})

	_ = edatasource.Init(&edatasource.Config{
		Kind:   kind,
		Logger: nil,
	})

	_ = registry.SelfRegister(context.Background())
}

func IsETCD() bool {
	t := archaius.Get("TEST_MODE")
	if t == nil {
		t = "etcd"
	}
	return t == "etcd"
}

func IsLOCAL() bool {
	t := archaius.Get("TEST_MODE")
	return t == "local"
}
