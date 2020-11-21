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
package core

import (
	"github.com/apache/servicecomb-service-center/pkg/registry"
	"testing"
)

func TestGetInfoFromKV(t *testing.T) {
	s, d := GetInfoFromSvcKV([]byte(GenerateServiceKey("a/b", "c")))
	if d != "a/b" || s != "c" {
		t.Fatalf("TestGetInfoFromKV failed")
	}
	s, d = GetInfoFromSvcKV([]byte("sdf"))
	if d != "" || s != "" {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	var i string
	s, i, d = GetInfoFromInstKV([]byte(GenerateInstanceKey("a/b", "c", "d")))
	if d != "a/b" || s != "c" || i != "d" {
		t.Fatalf("TestGetInfoFromKV failed")
	}
	s, i, d = GetInfoFromInstKV([]byte("sdf"))
	if d != "" || s != "" || i != "" {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	d = GetInfoFromDomainKV([]byte(GenerateDomainKey("a")))
	if d != "a" {
		t.Fatalf("TestGetInfoFromKV failed")
	}
	d = GetInfoFromDomainKV([]byte("sdf"))
	if d != "" {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	d = GetInfoFromProjectKV([]byte(GenerateProjectKey("a", "b")))
	if d != "a/b" {
		t.Fatalf("TestGetInfoFromKV failed")
	}
	d = GetInfoFromProjectKV([]byte("sdf"))
	if d != "" {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	var r string
	s, r, d = GetInfoFromRuleKV([]byte(GenerateServiceRuleKey("a/b", "c", "d")))
	if d != "a/b" || s != "c" || r != "d" {
		t.Fatalf("TestGetInfoFromKV failed")
	}
	s, r, d = GetInfoFromRuleKV([]byte("sdf"))
	if d != "" || s != "" || r != "" {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	s, d = GetInfoFromTagKV([]byte(GenerateServiceTagKey("a/b", "c")))
	if d != "a/b" || s != "c" {
		t.Fatalf("TestGetInfoFromKV failed")
	}
	s, d = GetInfoFromTagKV([]byte("sdf"))
	if d != "" || s != "" {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	key := GetInfoFromSvcIndexKV([]byte(GenerateServiceIndexKey(&registry.MicroServiceKey{
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
	key = GetInfoFromSvcIndexKV([]byte("sdf"))
	if key != nil {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	d, s, m := GetInfoFromSchemaSummaryKV([]byte(GenerateServiceSchemaSummaryKey("a/b", "c", "d")))
	if m != "d" || s != "c" || d != "a/b" {
		t.Fatalf("TestGetInfoFromKV failed")
	}
	d, s, m = GetInfoFromSchemaSummaryKV([]byte("sdf"))
	if m != "" || s != "" || d != "" {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	d, s, m = GetInfoFromSchemaKV([]byte(GenerateServiceSchemaKey("a/b", "c", "d")))
	if m != "d" || s != "c" || d != "a/b" {
		t.Fatalf("TestGetInfoFromKV failed")
	}
	d, s, m = GetInfoFromSchemaKV([]byte("sdf"))
	if m != "" || s != "" || d != "" {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	u := ""
	s, d, u = GetInfoFromDependencyQueueKV([]byte(GenerateConsumerDependencyQueueKey("a/b", "c", "d")))
	if s != "c" || d != "a/b" || u != "d" {
		t.Fatalf("TestGetInfoFromKV failed")
	}
	s, d, u = GetInfoFromDependencyQueueKV([]byte("sdf"))
	if s != "" || d != "" || u != "" {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	dt, k := GetInfoFromDependencyRuleKV([]byte(GenerateProviderDependencyRuleKey("a/b", &registry.MicroServiceKey{
		Tenant:      "a/b",
		AppId:       "c",
		ServiceName: "*",
	})))
	if dt != DepsProvider || k == nil || k.AppId != "" || k.ServiceName != "*" {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	dt, k = GetInfoFromDependencyRuleKV([]byte(GenerateProviderDependencyRuleKey("a/b", &registry.MicroServiceKey{
		Tenant:      "a/b",
		AppId:       "c",
		ServiceName: "d",
		Version:     "e",
	})))
	if dt != DepsProvider || k == nil || k.AppId != "c" || k.ServiceName != "d" {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	dt, k = GetInfoFromDependencyRuleKV([]byte("abc"))
	if dt != "" || k != nil {
		t.Fatalf("TestGetInfoFromKV failed")
	}
}
