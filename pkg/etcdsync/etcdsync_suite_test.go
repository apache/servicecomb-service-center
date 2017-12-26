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
package etcdsync_test

import (
	_ "github.com/apache/incubator-servicecomb-service-center/server/core/registry/embededetcd"
	_ "github.com/apache/incubator-servicecomb-service-center/server/core/registry/etcd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)
import (
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/etcdsync"
	"testing"
)

func init() {
	etcdsync.IsDebug = true
}

func TestEtcdsync(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Etcdsync Suite")
}

func BenchmarkLock(b *testing.B) {
	var g = 0
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lock, _ := etcdsync.Lock("/test")
			defer lock.Unlock()
			//do something
			g += 1
			fmt.Println(g)
		}
	})
	fmt.Println("Parallel:", b.N)
}
