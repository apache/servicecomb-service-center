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
package etcd

import (
	"context"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"github.com/coreos/etcd/clientv3"
	"testing"
	"time"
)

// tracing
import (
	"fmt"
	_ "github.com/apache/incubator-servicecomb-service-center/server/plugin/infra/tracing/buildin"
)

func TestEtcdClient_Delete(t *testing.T) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: connectRegistryServerTimeout,
	})
	if err != nil {
		panic(err)
	}
	etcd := &EtcdClient{
		err:    make(chan error, 1),
		ready:  make(chan int),
		Client: client,
	}
	_, err = etcd.Do(context.Background(), registry.DEL, registry.WithStrKey("/test_range/"), registry.WithPrefix())
	if err != nil {
		panic(err)
	}
	_, err = etcd.Do(context.Background(), registry.PUT, registry.WithStrKey("/test_range/b"), registry.WithStrValue("b"))
	if err != nil {
		panic(err)
	}
	_, err = etcd.Do(context.Background(), registry.PUT, registry.WithStrKey("/test_range/a"), registry.WithStrValue("a"))
	if err != nil {
		panic(err)
	}
	_, err = etcd.Do(context.Background(), registry.PUT, registry.WithStrKey("/test_range/c"), registry.WithStrValue("c"))
	if err != nil {
		panic(err)
	}
	_, err = etcd.Do(context.Background(), registry.DEL, registry.WithStrKey("/test_range/b"),
		registry.WithStrEndKey("/test_range/d")) // [b, d) !!!
	if err != nil {
		panic(err)
	}
	resp, err := etcd.Do(context.Background(), registry.GET, registry.WithStrKey("/test_range/"), registry.WithPrefix())
	if err != nil {
		panic(err)
	}
	if len(resp.Kvs) != 1 || string(resp.Kvs[0].Key) != "/test_range/a" {
		t.Fatalf("TestEtcdClient_Delete failed, %#v", resp.Kvs)
	}
}
