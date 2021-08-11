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
package value

import (
	"github.com/apache/servicecomb-service-center/datasource/etcd/state/parser"
	"testing"

	"github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
)

func TestParseInnerValueTypeFunc(t *testing.T) {
	r, err := parser.BytesParser.Unmarshal(nil)
	if err == nil {
		t.Fatalf("BytesParser.Unmarshal failed")
	}
	r, err = parser.BytesParser.Unmarshal([]byte("a"))
	if err != nil {
		t.Fatalf("BytesParser.Unmarshal failed, %s", err.Error())
	}
	if v, ok := r.([]byte); !ok || v[0] != 'a' {
		t.Fatalf("BytesParser.Unmarshal failed, %s", v)
	}

	r, err = parser.StringParser.Unmarshal(nil)
	if err != nil {
		t.Fatalf("StringParser.Unmarshal failed")
	}
	r, err = parser.StringParser.Unmarshal([]byte("abc"))
	if err != nil {
		t.Fatalf("StringParser.Unmarshal failed, %s", err.Error())
	}
	if v, ok := r.(string); !ok || v != "abc" {
		t.Fatalf("StringParser.Unmarshal failed, %s", v)
	}

	r, err = parser.MapParser.Unmarshal(nil)
	if err == nil {
		t.Fatalf("MapParser.Unmarshal failed")
	}
	r, err = parser.MapParser.Unmarshal([]byte(`{"a": "abc"}`))
	if err != nil {
		t.Fatalf("MapParser.Unmarshal failed, %s", err.Error())
	}
	if v, ok := r.(map[string]string); !ok || v["a"] != "abc" {
		t.Fatalf("MapParser.Unmarshal failed, %s", v)
	}

	r, err = parser.MapParser.Unmarshal([]byte(`xxx`))
	if err == nil {
		t.Fatalf("MapParser.Unmarshal failed")
	}

	var m interface{} = new(discovery.MicroService)
	err = parser.JSONUnmarshal(nil, nil)
	if err == nil {
		t.Fatalf("JSONUnmarshal failed")
	}
	err = parser.JSONUnmarshal([]byte(`{"serviceName": "abc"}`), &m)
	if err != nil {
		t.Fatalf("MapParser.Unmarshal failed, %v", err)
	}
	if m.(*discovery.MicroService).ServiceName != "abc" {
		t.Fatalf("MapParser.Unmarshal failed, %s", m)
	}
}

func TestParseValueTypeFunc(t *testing.T) {
	r, err := ServiceParser.Unmarshal([]byte(`xxx`))
	if err == nil || r != nil {
		t.Fatalf("ServiceParser.Unmarshal failed")
	}

	r, err = ServiceParser.Unmarshal([]byte(`{"serviceName": "abc"}`))
	if err != nil {
		t.Fatalf("ServiceParser.Unmarshal failed, %s", err.Error())
	}
	v, ok := r.(*discovery.MicroService)
	assert.True(t, ok)
	assert.Equal(t, "abc", v.ServiceName)

	r, err = InstanceParser.Unmarshal([]byte(`{"hostName": "abc"}`))
	if err != nil {
		t.Fatalf("InstanceParser.Unmarshal failed, %s", err.Error())
	}
	mi, ok := r.(*discovery.MicroServiceInstance)
	assert.True(t, ok)
	assert.Equal(t, "abc", mi.HostName)

	r, err = RuleParser.Unmarshal([]byte(`{"ruleId": "abc"}`))
	if err != nil {
		t.Fatalf("RuleParser.Unmarshal failed, %s", err.Error())
	}
	if v, ok := r.(*discovery.ServiceRule); !ok || v.RuleId != "abc" {
		t.Fatalf("RuleParser.Unmarshal failed, %s", v)
	}

	r, err = DependencyRuleParser.Unmarshal([]byte(`{"Dependency":[{"tenant":"opsadm/southchina","appId":"csezhq","serviceName":"zhqClient","version":"1.0.0"}]}`))
	if err != nil {
		t.Fatalf("DependencyRuleParser.Unmarshal failed, %s", err.Error())
	}
	md, ok := r.(*discovery.MicroServiceDependency)
	assert.True(t, ok)
	assert.Equal(t, "zhqClient", md.Dependency[0].ServiceName)
}
