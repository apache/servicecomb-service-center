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
package backend

import (
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"testing"
)

func TestParseValueFunc(t *testing.T) {
	r, err := BytesParser.Unmarshal([]byte("a"))
	if err != nil {
		t.Fatalf("BytesParser.Unmarshal failed, %s", err.Error())
	}
	if v, ok := r.([]byte); !ok || v[0] != 'a' {
		t.Fatalf("BytesParser.Unmarshal failed, %s", v)
	}

	r, err = StringParser.Unmarshal([]byte("abc"))
	if err != nil {
		t.Fatalf("StringParser.Unmarshal failed, %s", err.Error())
	}
	if v, ok := r.(string); !ok || v != "abc" {
		t.Fatalf("StringParser.Unmarshal failed, %s", v)
	}

	r, err = ServiceParser.Unmarshal([]byte(`xxx`))
	if err == nil || r != nil {
		t.Fatalf("ServiceParser.Unmarshal failed")
	}

	r, err = ServiceParser.Unmarshal([]byte(`{"serviceName": "abc"}`))
	if err != nil {
		t.Fatalf("ServiceParser.Unmarshal failed, %s", err.Error())
	}
	if v, ok := r.(*proto.MicroService); !ok || v.ServiceName != "abc" {
		t.Fatalf("ServiceParser.Unmarshal failed, %s", v)
	}

	r, err = InstanceParser.Unmarshal([]byte(`{"hostName": "abc"}`))
	if err != nil {
		t.Fatalf("InstanceParser.Unmarshal failed, %s", err.Error())
	}
	if v, ok := r.(*proto.MicroServiceInstance); !ok || v.HostName != "abc" {
		t.Fatalf("InstanceParser.Unmarshal failed, %s", v)
	}

	r, err = RuleParser.Unmarshal([]byte(`{"ruleId": "abc"}`))
	if err != nil {
		t.Fatalf("RuleParser.Unmarshal failed, %s", err.Error())
	}
	if v, ok := r.(*proto.ServiceRule); !ok || v.RuleId != "abc" {
		t.Fatalf("RuleParser.Unmarshal failed, %s", v)
	}

	r, err = DependencyRuleParser.Unmarshal([]byte(`{"Dependency":[{"tenant":"opsadm/southchina","appId":"csezhq","serviceName":"zhqClient","version":"1.0.0"}]}`))
	if err != nil {
		t.Fatalf("DependencyRuleParser.Unmarshal failed, %s", err.Error())
	}
	if v, ok := r.(*proto.MicroServiceDependency); !ok || v.Dependency[0].ServiceName != "zhqClient" {
		t.Fatalf("DependencyRuleParser.Unmarshal failed, %s", v)
	}

	r, err = MapParser.Unmarshal([]byte(`{"a": "abc"}`))
	if err != nil {
		t.Fatalf("MapParser.Unmarshal failed, %s", err.Error())
	}
	if v, ok := r.(map[string]string); !ok || v["a"] != "abc" {
		t.Fatalf("MapParser.Unmarshal failed, %s", v)
	}
}

func TestGetInfoFromKV(t *testing.T) {
	s, d := GetInfoFromSvcKV(&KeyValue{Key: []byte(core.GenerateServiceKey("a/b", "c"))})
	if d != "a/b" || s != "c" {
		t.Fatalf("TestGetInfoFromKV failed")
	}
	s, d = GetInfoFromSvcKV(&KeyValue{Key: []byte("sdf")})
	if d != "" || s != "" {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	var i string
	s, i, d = GetInfoFromInstKV(&KeyValue{Key: []byte(core.GenerateInstanceKey("a/b", "c", "d"))})
	if d != "a/b" || s != "c" || i != "d" {
		t.Fatalf("TestGetInfoFromKV failed")
	}
	s, i, d = GetInfoFromInstKV(&KeyValue{Key: []byte("sdf")})
	if d != "" || s != "" || i != "" {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	d = GetInfoFromDomainKV(&KeyValue{Key: []byte(core.GenerateDomainKey("a"))})
	if d != "a" {
		t.Fatalf("TestGetInfoFromKV failed")
	}
	d = GetInfoFromDomainKV(&KeyValue{Key: []byte("sdf")})
	if d != "" {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	d = GetInfoFromProjectKV(&KeyValue{Key: []byte(core.GenerateProjectKey("a", "b"))})
	if d != "a/b" {
		t.Fatalf("TestGetInfoFromKV failed")
	}
	d = GetInfoFromProjectKV(&KeyValue{Key: []byte("sdf")})
	if d != "" {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	var r string
	s, r, d = GetInfoFromRuleKV(&KeyValue{Key: []byte(core.GenerateServiceRuleKey("a/b", "c", "d"))})
	if d != "a/b" || s != "c" || r != "d" {
		t.Fatalf("TestGetInfoFromKV failed")
	}
	s, r, d = GetInfoFromRuleKV(&KeyValue{Key: []byte("sdf")})
	if d != "" || s != "" || r != "" {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	s, d = GetInfoFromTagKV(&KeyValue{Key: []byte(core.GenerateServiceTagKey("a/b", "c"))})
	if d != "a/b" || s != "c" {
		t.Fatalf("TestGetInfoFromKV failed")
	}
	s, d = GetInfoFromTagKV(&KeyValue{Key: []byte("sdf")})
	if d != "" || s != "" {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	key := GetInfoFromSvcIndexKV(&KeyValue{Key: []byte(core.GenerateServiceIndexKey(&proto.MicroServiceKey{
		Tenant:      "a/b",
		Project:     "",
		AppId:       "c",
		ServiceName: "d",
		Version:     "e",
		Environment: "f",
		Alias:       "g",
	}))})
	if key.Tenant != "a/b" ||
		key.AppId != "c" ||
		key.ServiceName != "d" ||
		key.Version != "e" ||
		key.Environment != "f" ||
		key.Project != "" ||
		key.Alias != "" {
		t.Fatalf("TestGetInfoFromKV failed")
	}
	key = GetInfoFromSvcIndexKV(&KeyValue{Key: []byte("sdf")})
	if key != nil {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	m := GetInfoFromSchemaSummaryKV(&KeyValue{Key: []byte(core.GenerateServiceSchemaSummaryKey("a/b", "c", "d"))})
	if m != "d" {
		t.Fatalf("TestGetInfoFromKV failed")
	}
	m = GetInfoFromSchemaSummaryKV(&KeyValue{Key: []byte("sdf")})
	if m != "" {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	m = GetInfoFromSchemaKV(&KeyValue{Key: []byte(core.GenerateServiceSchemaKey("a/b", "c", "d"))})
	if m != "d" {
		t.Fatalf("TestGetInfoFromKV failed")
	}
	m = GetInfoFromSchemaKV(&KeyValue{Key: []byte("sdf")})
	if m != "" {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	u := ""
	s, d, u = GetInfoFromDependencyQueueKV(&KeyValue{Key: []byte(core.GenerateConsumerDependencyQueueKey("a/b", "c", "d"))})
	if s != "c" || d != "a/b" || u != "d" {
		t.Fatalf("TestGetInfoFromKV failed")
	}
	s, d, u = GetInfoFromDependencyQueueKV(&KeyValue{Key: []byte("sdf")})
	if s != "" || d != "" || u != "" {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	k := GetInfoFromDependencyRuleKV(&KeyValue{Key: []byte(core.GenerateProviderDependencyRuleKey("a/b", &proto.MicroServiceKey{
		Tenant:      "a/b",
		AppId:       "c",
		ServiceName: "*",
	}))})
	if k == nil || k.AppId != "" || k.ServiceName != "*" {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	k = GetInfoFromDependencyRuleKV(&KeyValue{Key: []byte(core.GenerateProviderDependencyRuleKey("a/b", &proto.MicroServiceKey{
		Tenant:      "a/b",
		AppId:       "c",
		ServiceName: "d",
		Version:     "e",
	}))})
	if k == nil || k.AppId != "c" || k.ServiceName != "d" {
		t.Fatalf("TestGetInfoFromKV failed")
	}

	k = GetInfoFromDependencyRuleKV(&KeyValue{Key: []byte("abc")})
	if k != nil {
		t.Fatalf("TestGetInfoFromKV failed")
	}
}
