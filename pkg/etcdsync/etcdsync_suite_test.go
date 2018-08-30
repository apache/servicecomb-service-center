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
package etcdsync

import (
	"fmt"
	_ "github.com/apache/incubator-servicecomb-service-center/server/plugin/infra/registry/etcd"
	_ "github.com/apache/incubator-servicecomb-service-center/server/plugin/infra/tracing/buildin"
	"github.com/astaxie/beego"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func init() {
	beego.AppConfig.Set("registry_plugin", "etcd")
}

var _ = BeforeSuite(func() {
	//init plugin
	IsDebug = true
})

func TestEtcdsync(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Etcdsync Suite")
}

func BenchmarkLock(b *testing.B) {
	var g = 0
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lock, _ := Lock("/test", true)
			//do something
			g += 1
			fmt.Println(g)
			lock.Unlock()
		}
	})
	fmt.Println("Parallel:", b.N)
}
