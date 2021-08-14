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
	"testing"

	"github.com/apache/servicecomb-service-center/datasource/etcd/state/parser"
	"github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
)

func TestParseInnerValueTypeFunc(t *testing.T) {
	r, err := parser.BytesParser.Unmarshal(nil)
	assert.NoError(t, err)
	assert.Nil(t, r.([]byte))

	r, err = parser.BytesParser.Unmarshal([]byte("a"))
	assert.NoError(t, err)
	assert.Equal(t, []byte("a"), r.([]byte))

	r, err = parser.StringParser.Unmarshal(nil)
	assert.NoError(t, err)
	assert.Equal(t, "", r.(string))

	r, err = parser.StringParser.Unmarshal([]byte("abc"))
	assert.NoError(t, err)
	assert.Equal(t, "abc", r.(string))

	r, err = parser.MapParser.Unmarshal(nil)
	assert.ErrorIs(t, parser.ErrParseNilPoint, err)

	r, err = parser.MapParser.Unmarshal([]byte(`{"a": "abc"}`))
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"a": "abc"}, r.(map[string]string))

	r, err = parser.MapParser.Unmarshal([]byte(`xxx`))
	assert.Error(t, err)

	var m interface{} = new(discovery.MicroService)
	err = parser.JSONUnmarshal(nil, nil)
	assert.ErrorIs(t, parser.ErrParseNilPoint, err)

	err = parser.JSONUnmarshal([]byte(`{"serviceName": "abc"}`), &m)
	assert.NoError(t, err)
	assert.Equal(t, "abc", m.(*discovery.MicroService).ServiceName)
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
