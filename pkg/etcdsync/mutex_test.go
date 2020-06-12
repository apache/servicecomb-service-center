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
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/etcdsync"
	"github.com/astaxie/beego"
	"github.com/stretchr/testify/assert"
	"testing"

	_ "github.com/apache/servicecomb-service-center/server/plugin/pkg/registry/etcd"
	_ "github.com/apache/servicecomb-service-center/server/plugin/pkg/tracing/buildin"
)

func init() {
	beego.AppConfig.Set("registry_plugin", "etcd")
	//init plugin
	etcdsync.IsDebug = true
}

func TestLock(t *testing.T) {
	m1, err := etcdsync.Lock("key1", 5, true)
	assert.NoError(t, err)
	assert.NotNil(t, m1)
	t.Log("m1 locked")

	ch := make(chan bool)
	go func() {
		m2, err := etcdsync.Lock("key1", 1, false)

		assert.Nil(t, m2)
		assert.Error(t, err)
		fmt.Println("m2 try failed")

		m2, err = etcdsync.Lock("key1", 1, true)
		assert.Nil(t, m2)
		assert.Error(t, err)
		fmt.Println("m2 timed out")
		ch <- true
	}()
	<-ch

	m3, err := etcdsync.Lock("key1", 2, true)
	assert.NoError(t, err)
	assert.NotNil(t, m3)

	fmt.Println("m3 locked")
	err = m3.Unlock()
	assert.NoError(t, err)

	err = m1.Unlock()
	assert.NoError(t, err)
	fmt.Println("m1 unlocked")
}
func BenchmarkLock(b *testing.B) {
	var g = 0
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lock, _ := etcdsync.Lock("/test", -1, true)
			//do something
			g += 1
			fmt.Println(g)
			lock.Unlock()
		}
	})
	fmt.Println("Parallel:", b.N)
}
