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
package path_test

import (
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/pkg/registry"
	"testing"
)

func TestGetInfoFromKV(t *testing.T) {
	s, d := path.GetInfoFromSvcKV([]byte(path.GenerateServiceKey("a/b", "c")))
	if d != "a/b" || s != "c" {
		t.Fatalf("TestGetInfoFromKV failed")
	}
	s, d = path.GetInfoFromSvcKV([]byte("sdf"))
	if d != "" || s != "" {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	var i string
	s, i, d = path.GetInfoFromInstKV([]byte(path.GenerateInstanceKey("a/b", "c", "d")))
	if d != "a/b" || s != "c" || i != "d" {
		t.Fatalf("TestGetInfoFromKV failed")
	}
	s, i, d = path.GetInfoFromInstKV([]byte("sdf"))
	if d != "" || s != "" || i != "" {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	d = path.GetInfoFromDomainKV([]byte(path.GenerateDomainKey("a")))
	if d != "a" {
		t.Fatalf("TestGetInfoFromKV failed")
	}
	d = path.GetInfoFromDomainKV([]byte("sdf"))
	if d != "" {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	p := ""
	d, p = path.GetInfoFromProjectKV([]byte(path.GenerateProjectKey("a", "b")))
	if d != "a" || p != "b" {
		t.Fatalf("TestGetInfoFromKV failed")
	}
	d, p = path.GetInfoFromProjectKV([]byte("sdf"))
	if d != "" || p != "" {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	var r string
	s, r, d = path.GetInfoFromRuleKV([]byte(path.GenerateServiceRuleKey("a/b", "c", "d")))
	if d != "a/b" || s != "c" || r != "d" {
		t.Fatalf("TestGetInfoFromKV failed")
	}
	s, r, d = path.GetInfoFromRuleKV([]byte("sdf"))
	if d != "" || s != "" || r != "" {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	s, d = path.GetInfoFromTagKV([]byte(path.GenerateServiceTagKey("a/b", "c")))
	if d != "a/b" || s != "c" {
		t.Fatalf("TestGetInfoFromKV failed")
	}
	s, d = path.GetInfoFromTagKV([]byte("sdf"))
	if d != "" || s != "" {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	key := path.GetInfoFromSvcIndexKV([]byte(path.GenerateServiceIndexKey(&registry.MicroServiceKey{
		Tenant:      "a/b",
		AppId:       "c",
		ServiceName: "d",
		Version:     "e",
		Environment: "f",
		Alias:       "g",
	})))
	if key.Tenant != "a/b" ||
		key.AppId != "c" ||
		key.ServiceName != "d" ||
		key.Version != "e" ||
		key.Environment != "f" ||
		key.Alias != "" {
		t.Fatalf("TestGetInfoFromKV failed")
	}
	key = path.GetInfoFromSvcIndexKV([]byte("sdf"))
	if key != nil {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	d, s, m := path.GetInfoFromSchemaSummaryKV([]byte(path.GenerateServiceSchemaSummaryKey("a/b", "c", "d")))
	if m != "d" || s != "c" || d != "a/b" {
		t.Fatalf("TestGetInfoFromKV failed")
	}
	d, s, m = path.GetInfoFromSchemaSummaryKV([]byte("sdf"))
	if m != "" || s != "" || d != "" {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	d, s, m = path.GetInfoFromSchemaKV([]byte(path.GenerateServiceSchemaKey("a/b", "c", "d")))
	if m != "d" || s != "c" || d != "a/b" {
		t.Fatalf("TestGetInfoFromKV failed")
	}
	d, s, m = path.GetInfoFromSchemaKV([]byte("sdf"))
	if m != "" || s != "" || d != "" {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	u := ""
	s, d, u = path.GetInfoFromDependencyQueueKV([]byte(path.GenerateConsumerDependencyQueueKey("a/b", "c", "d")))
	if s != "c" || d != "a/b" || u != "d" {
		t.Fatalf("TestGetInfoFromKV failed")
	}
	s, d, u = path.GetInfoFromDependencyQueueKV([]byte("sdf"))
	if s != "" || d != "" || u != "" {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	dt, k := path.GetInfoFromDependencyRuleKV([]byte(path.GenerateProviderDependencyRuleKey("a/b", &registry.MicroServiceKey{
		Tenant:      "a/b",
		AppId:       "c",
		ServiceName: "*",
	})))
	if dt != path.DepsProvider || k == nil || k.AppId != "" || k.ServiceName != "*" {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	dt, k = path.GetInfoFromDependencyRuleKV([]byte(path.GenerateProviderDependencyRuleKey("a/b", &registry.MicroServiceKey{
		Tenant:      "a/b",
		AppId:       "c",
		ServiceName: "d",
		Version:     "e",
	})))
	if dt != path.DepsProvider || k == nil || k.AppId != "c" || k.ServiceName != "d" {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	dt, k = path.GetInfoFromDependencyRuleKV([]byte("abc"))
	if dt != "" || k != nil {
		t.Fatalf("TestGetInfoFromKV failed")
	}
}
