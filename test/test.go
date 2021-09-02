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
	"io"
	"os"
	"path/filepath"

	"github.com/apache/servicecomb-service-center/pkg/util"
	_ "github.com/apache/servicecomb-service-center/server/init"

	_ "github.com/apache/servicecomb-service-center/server/bootstrap"
	//grpc plugin
	_ "github.com/go-chassis/go-chassis-extension/protocol/grpc/server"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/metrics"
	"github.com/apache/servicecomb-service-center/server/service/disco"
	"github.com/go-chassis/go-archaius"
	"github.com/little-cui/etcdadpt"
)

func init() {
	var kind = "etcd"
	_ = archaius.Set("rbac.releaseLockAfter", "3s")
	if IsETCD() {
		_ = archaius.Set("registry.cache.mode", 0)
		_ = archaius.Set("discovery.kind", "etcd")
		_ = archaius.Set("registry.kind", "etcd")
	} else {
		_ = archaius.Set("registry.heartbeat.kind", "checker")
		kind = "mongo"
	}
	_ = datasource.Init(datasource.Options{
		Config: etcdadpt.Config{
			Kind: kind,
		},
	})
	_ = metrics.Init(metrics.Options{})

	core.ServiceAPI = disco.AssembleResources()
}
func createChassisConfig() {
	b := []byte(`
servicecomb:
  registry:
    disabled: true
  protocols:
    grpc:
      listenAddress: 127.0.0.1:30105
`)
	dir := filepath.Join(util.GetAppRoot(), "conf")
	os.Mkdir(dir, 0700)
	file := filepath.Join(dir, "chassis.yaml")
	f1, _ := os.Create(file)
	_, _ = io.WriteString(f1, string(b))

	b2 := []byte(`
servicecomb:
  service:
    name: service-center
    app: servicecomb
    version: 2.0.0
`)
	file2 := filepath.Join(dir, "chassis.yaml")
	f2, _ := os.Create(file2)
	_, _ = io.WriteString(f2, string(b2))
}
func IsETCD() bool {
	t := archaius.Get("TEST_MODE")
	if t == nil {
		t = "etcd"
	}
	return t == "etcd"
}
